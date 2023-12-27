/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bsutil "sigs.k8s.io/cluster-api/bootstrap/util"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/annotations"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/cluster-api/util/kubeconfig"
	"sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/cluster-api/util/secret"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	bootstrapv1 "github.com/oneblock-ai/okr/api/bootstrap/v1"
	"github.com/oneblock-ai/okr/pkg/token"
	olog "github.com/oneblock-ai/okr/pkg/utils/log"
)

// InitLocker is a lock that is used around k3s init.
type InitLocker interface {
	Lock(ctx context.Context, cluster *clusterv1.Cluster, machine *clusterv1.Machine) bool
	Unlock(ctx context.Context, cluster *clusterv1.Cluster) bool
}

// Ok3sConfigReconciler reconciles a Ok3sConfig object
type Ok3sConfigReconciler struct {
	client.Client
	Log         logr.Logger
	Scheme      *runtime.Scheme
	K3sInitLock InitLocker
}

// Scope is a scoped struct used during reconciliation.
type Scope struct {
	logr.Logger
	Config      *bootstrapv1.Ok3sConfig
	ConfigOwner *bsutil.ConfigOwner
	Cluster     *clusterv1.Cluster
}

// SetupWithManager sets up the bootstrap with the Manager.
func (r *Ok3sConfigReconciler) SetupWithManager(mgr ctrl.Manager, ctx context.Context) error {
	//if r.K3sInitLock == nil {
	//	r.K3sInitLock = locking.NewControlPlaneInitMutex(ctrl.Log.WithName("init-locker"), mgr.GetClient())
	//}

	return ctrl.NewControllerManagedBy(mgr).
		For(&bootstrapv1.Ok3sConfig{}).
		Complete(r)
}

//+kubebuilder:rbac:groups=bootstrap.cluster.x-k8s.io,resources=ok3sconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=bootstrap.cluster.x-k8s.io,resources=ok3sconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=bootstrap.cluster.x-k8s.io,resources=ok3sconfigs/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;clusters/status,verbs=get;list;watch
// +kubebuilder:rbac:groups=exp.cluster.x-k8s.io,resources=machinepools;machinepools/status,verbs=get;list;watch

func (r *Ok3sConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, retErr error) {
	log := r.Log.WithValues("kthreesconfig", req.NamespacedName)

	// Lookup the ok3s config
	config := &bootstrapv1.Ok3sConfig{}
	if err := r.Client.Get(ctx, req.NamespacedName, config); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		log.Error(err, "Failed to get owner")
		return ctrl.Result{}, err
	}

	// AddOwners adds the owners of Ok3sConfig as k/v pairs to the logger.
	// Specifically, it will add Ok3sControlPlane, MachineSet and MachineDeployment.
	ctx, log, err := olog.AddOwners(ctx, r.Client, config)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Look up the owner of this ok3s config if there is one
	configOwner, err := bsutil.GetConfigOwner(ctx, r.Client, config)
	if apierrors.IsNotFound(err) {
		// Could not find the owner yet, this is not an error and will reconcile when the owner gets set.
		return ctrl.Result{}, nil
	}
	if err != nil {
		log.Error(err, "Failed to get owner")
		return ctrl.Result{}, err
	}
	if configOwner == nil {
		return ctrl.Result{}, nil
	}

	log = log.WithValues("kind", configOwner.GetKind(), "version", configOwner.GetResourceVersion(), "name", configOwner.GetName())

	// Lookup the cluster the config owner is associated with
	cluster, err := util.GetClusterByName(ctx, r.Client, configOwner.GetNamespace(), configOwner.ClusterName())
	if err != nil {
		if errors.Is(err, util.ErrNoCluster) {
			log.Info(fmt.Sprintf("%s does not belong to a cluster yet, waiting until it's part of a cluster", configOwner.GetKind()))
			return ctrl.Result{}, nil
		}

		if apierrors.IsNotFound(err) {
			log.Info("Cluster does not exist yet, waiting until it is created")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Could not get cluster with metadata")
		return ctrl.Result{}, err
	}

	if annotations.IsPaused(cluster, config) {
		log.Info("Reconciliation is paused for this object")
		return ctrl.Result{}, nil
	}

	scope := &Scope{
		Logger:      log,
		Config:      config,
		ConfigOwner: configOwner,
		Cluster:     cluster,
	}

	// Initialize the patch helper.
	patchHelper, err := patch.NewHelper(config, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Attempt to Patch the K3sConfig object and status after each reconciliation if no error occurs.
	defer func() {
		// always update the readyCondition; the summary is represented using the "1 of x completed" notation.
		conditions.SetSummary(config,
			conditions.WithConditions(
				bootstrapv1.DataSecretAvailableCondition,
				bootstrapv1.CertificatesAvailableCondition,
			),
		)
		// Patch ObservedGeneration only if the reconciliation completed successfully
		var patchOpts []patch.Option
		if retErr == nil {
			patchOpts = append(patchOpts, patch.WithStatusObservedGeneration{})
		}
		if err := patchHelper.Patch(ctx, config, patchOpts...); err != nil {
			log.Error(retErr, "Failed to patch config")
			if retErr == nil {
				retErr = err
			}
		}
	}()

	switch {
	// Wait for the infrastructure to be ready.
	case !cluster.Status.InfrastructureReady:
		log.Info("Cluster infrastructure is not ready, waiting")
		conditions.MarkFalse(config, bootstrapv1.DataSecretAvailableCondition, bootstrapv1.WaitingForClusterInfrastructureReason, clusterv1.ConditionSeverityInfo, "")
		return ctrl.Result{}, nil
	// Reconcile status for machines that already have a secret reference, but our status isn't up-to-date.
	// This case solves the pivoting scenario (or a backup restore) which doesn't preserve the status subresource on objects.
	case configOwner.DataSecretName() != nil && (!config.Status.Ready || config.Status.DataSecretName == nil):
		config.Status.Ready = true
		config.Status.DataSecretName = configOwner.DataSecretName()
		conditions.MarkTrue(config, bootstrapv1.DataSecretAvailableCondition)
		return ctrl.Result{}, nil
	// Status is ready means a config has been generated.
	case config.Status.Ready:
		return ctrl.Result{}, nil
	}

	// Note: can't use IsFalse here because we need to handle the absence of the condition as well as false.
	if !conditions.IsTrue(cluster, clusterv1.ControlPlaneInitializedCondition) {
		return r.handleClusterNotInitialized(ctx, scope)
	}

	// Every other case it's a join scenario
	// Nb. in this case ClusterConfiguration and InitConfiguration should not be defined by users, but in case of misconfigurations, CABPK3s simply ignore them

	// Unlock any locks that might have been set during init process
	r.K3sInitLock.Unlock(ctx, cluster)

	// it's a control plane join
	if configOwner.IsControlPlaneMachine() {
		return ctrl.Result{}, r.joinControlplane(ctx, scope)
	}

	// It's a worker join
	return ctrl.Result{}, r.joinWorker(ctx, scope)
}

func (r *Ok3sConfigReconciler) joinControlplane(ctx context.Context, scope *Scope) error {
	machine := &clusterv1.Machine{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(scope.ConfigOwner.Object, machine); err != nil {
		return fmt.Errorf("cannot convert %s to Machine: %w", scope.ConfigOwner.GetKind(), err)
	}

	// injects into config.Version values from top level object
	r.reconcileTopLevelObjectSettings(scope.Cluster, machine, scope.Config)

	serverURL := fmt.Sprintf("https://%s", scope.Cluster.Spec.ControlPlaneEndpoint.String())

	token, err := token.Lookup(ctx, r.Client, client.ObjectKeyFromObject(scope.Cluster))
	if err != nil {
		conditions.MarkFalse(scope.Config, bootstrapv1.DataSecretAvailableCondition, bootstrapv1.DataSecretGenerationFailedReason, clusterv1.ConditionSeverityWarning, err.Error())
		return err
	}

	configStruct := k3s.GenerateJoinControlPlaneConfig(serverURL, *token,
		scope.Cluster.Spec.ControlPlaneEndpoint.Host,
		scope.Config.Spec.ServerConfig,
		scope.Config.Spec.AgentConfig)
	b, err := kubeyaml.Marshal(configStruct)
	if err != nil {
		return err
	}

	//workerConfigFile := bootstrapv1.File{
	//	Path:        k3s.DefaultK3sConfigLocation,
	//	Content:     string(b),
	//	Owner:       "root:root",
	//	Permissions: "0640",
	//}

	files, err := r.resolveFiles(ctx, scope.Config)
	if err != nil {
		conditions.MarkFalse(scope.Config, bootstrapv1.DataSecretAvailableCondition, bootstrapv1.DataSecretGenerationFailedReason, clusterv1.ConditionSeverityWarning, err.Error())
		return err
	}

	cpInput := &cloudinit.ControlPlaneInput{
		BaseUserData: cloudinit.BaseUserData{
			PreK3sCommands:  scope.Config.Spec.PreK3sCommands,
			PostK3sCommands: scope.Config.Spec.PostK3sCommands,
			AdditionalFiles: files,
			ConfigFile:      workerConfigFile,
			K3sVersion:      scope.Config.Spec.Version,
		},
	}

	cloudInitData, err := cloudinit.NewJoinControlPlane(cpInput)
	if err != nil {
		return err
	}

	if err := r.storeBootstrapData(ctx, scope, cloudInitData); err != nil {
		scope.Error(err, "Failed to store bootstrap data")
		return err
	}
	return nil
}

func (r *KThreesConfigReconciler) joinWorker(ctx context.Context, scope *Scope) error {
	machine := &clusterv1.Machine{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(scope.ConfigOwner.Object, machine); err != nil {
		return fmt.Errorf("cannot convert %s to Machine: %w", scope.ConfigOwner.GetKind(), err)
	}

	// injects into config.Version values from top level object
	r.reconcileTopLevelObjectSettings(scope.Cluster, machine, scope.Config)

	serverURL := fmt.Sprintf("https://%s", scope.Cluster.Spec.ControlPlaneEndpoint.String())

	tokn, err := token.Lookup(ctx, r.Client, client.ObjectKeyFromObject(scope.Cluster))
	if err != nil {
		conditions.MarkFalse(scope.Config, bootstrapv1.DataSecretAvailableCondition, bootstrapv1.DataSecretGenerationFailedReason, clusterv1.ConditionSeverityWarning, err.Error())
		return err
	}

	configStruct := k3s.GenerateWorkerConfig(serverURL, *tokn, scope.Config.Spec.ServerConfig, scope.Config.Spec.AgentConfig)

	b, err := kubeyaml.Marshal(configStruct)
	if err != nil {
		return err
	}

	workerConfigFile := bootstrapv1.File{
		Path:        k3s.DefaultK3sConfigLocation,
		Content:     string(b),
		Owner:       "root:root",
		Permissions: "0640",
	}

	files, err := r.resolveFiles(ctx, scope.Config)
	if err != nil {
		conditions.MarkFalse(scope.Config, bootstrapv1.DataSecretAvailableCondition, bootstrapv1.DataSecretGenerationFailedReason, clusterv1.ConditionSeverityWarning, err.Error())
		return err
	}

	winput := &cloudinit.WorkerInput{
		BaseUserData: cloudinit.BaseUserData{
			PreK3sCommands:  scope.Config.Spec.PreK3sCommands,
			PostK3sCommands: scope.Config.Spec.PostK3sCommands,
			AdditionalFiles: files,
			ConfigFile:      workerConfigFile,
			K3sVersion:      scope.Config.Spec.Version,
		},
	}

	cloudInitData, err := cloudinit.NewWorker(winput)
	if err != nil {
		return err
	}

	if err := r.storeBootstrapData(ctx, scope, cloudInitData); err != nil {
		scope.Error(err, "Failed to store bootstrap data")
		return err
	}

	return nil
}

func (r *Ok3sConfigReconciler) handleClusterNotInitialized(ctx context.Context, scope *Scope) (_ ctrl.Result, reterr error) {
	// initialize the DataSecretAvailableCondition if missing.
	// this is required in order to avoid the condition's LastTransitionTime to flicker in case of errors surfacing
	// using the DataSecretGeneratedFailedReason
	if conditions.GetReason(scope.Config, bootstrapv1.DataSecretAvailableCondition) != bootstrapv1.DataSecretGenerationFailedReason {
		conditions.MarkFalse(scope.Config, bootstrapv1.DataSecretAvailableCondition, clusterv1.WaitingForControlPlaneAvailableReason, clusterv1.ConditionSeverityInfo, "")
	}

	// if it's NOT a control plane machine, requeue
	if !scope.ConfigOwner.IsControlPlaneMachine() {
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	machine := &clusterv1.Machine{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(scope.ConfigOwner.Object, machine); err != nil {
		return ctrl.Result{}, fmt.Errorf("cannot convert %s to Machine: %w", scope.ConfigOwner.GetKind(), err)
	}

	// acquire the init lock so that only the first machine configured
	// as control plane get processed here
	// if not the first, requeue

	if !r.KThreesInitLock.Lock(ctx, scope.Cluster, machine) {
		scope.Info("A control plane is already being initialized, requeing until control plane is ready")
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	defer func() {
		if reterr != nil {
			if !r.KThreesInitLock.Unlock(ctx, scope.Cluster) {
				reterr = kerrors.NewAggregate([]error{reterr, ErrFailedUnlock})
			}
		}
	}()

	scope.Info("Creating BootstrapData for the init control plane")

	// injects into config.ClusterConfiguration values from top level object
	r.reconcileTopLevelObjectSettings(scope.Cluster, machine, scope.Config)

	certificates := secret.NewCertificatesForInitialControlPlane(&scope.Config.Spec)
	err := certificates.LookupOrGenerate(
		ctx,
		r.Client,
		util.ObjectKey(scope.Cluster),
		*metav1.NewControllerRef(scope.Config, bootstrapv1.GroupVersion.WithKind("KThreesConfig")),
	)
	if err != nil {
		conditions.MarkFalse(scope.Config, bootstrapv1.CertificatesAvailableCondition, bootstrapv1.CertificatesGenerationFailedReason, clusterv1.ConditionSeverityWarning, err.Error())
		return ctrl.Result{}, err
	}
	conditions.MarkTrue(scope.Config, bootstrapv1.CertificatesAvailableCondition)

	token, err := token.Lookup(ctx, r.Client, client.ObjectKeyFromObject(scope.Cluster))
	if err != nil {
		return ctrl.Result{}, err
	}

	// TODO support k3s great feature of external backends.
	// For now just use the etcd option
	configStruct := k3s.GenerateInitControlPlaneConfig(
		scope.Cluster.Spec.ControlPlaneEndpoint.Host,
		*token,
		scope.Config.Spec.ServerConfig,
		scope.Config.Spec.AgentConfig)

	b, err := kubeyaml.Marshal(configStruct)
	if err != nil {
		return ctrl.Result{}, err
	}

	initConfigFile := bootstrapv1.File{
		Path:        k3s.DefaultK3sConfigLocation,
		Content:     string(b),
		Owner:       "root:root",
		Permissions: "0640",
	}

	files, err := r.resolveFiles(ctx, scope.Config)
	if err != nil {
		conditions.MarkFalse(scope.Config, bootstrapv1.DataSecretAvailableCondition, bootstrapv1.DataSecretGenerationFailedReason, clusterv1.ConditionSeverityWarning, err.Error())
		return ctrl.Result{}, err
	}

	cpinput := &cloudinit.ControlPlaneInput{
		BaseUserData: cloudinit.BaseUserData{
			PreK3sCommands:  scope.Config.Spec.PreK3sCommands,
			PostK3sCommands: scope.Config.Spec.PostK3sCommands,
			AdditionalFiles: files,
			ConfigFile:      initConfigFile,
			K3sVersion:      scope.Config.Spec.Version,
		},
		Certificates: certificates,
	}

	cloudInitData, err := cloudinit.NewInitControlPlane(cpinput)
	if err != nil {
		return ctrl.Result{}, err
	}

	if err := r.storeBootstrapData(ctx, scope, cloudInitData); err != nil {
		scope.Error(err, "Failed to store bootstrap data")
		return ctrl.Result{}, err
	}

	// TODO: move to controlplane provider
	return r.reconcileKubeconfig(ctx, scope)
}

// storeBootstrapData creates a new secret with the data passed in as input,
// sets the reference in the configuration status and ready to true.
func (r *Ok3sConfigReconciler) storeBootstrapData(ctx context.Context, scope *Scope, data []byte) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      scope.Config.Name,
			Namespace: scope.Config.Namespace,
			Labels: map[string]string{
				clusterv1.ClusterNameLabel: scope.Cluster.Name,
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: bootstrapv1.GroupVersion.String(),
					Kind:       "KThreesConfig",
					Name:       scope.Config.Name,
					UID:        scope.Config.UID,
					Controller: pointer.Bool(true),
				},
			},
		},
		Data: map[string][]byte{
			"value": data,
		},
		Type: clusterv1.ClusterSecretType,
	}

	// as secret creation and scope.Config status patch are not atomic operations
	// it is possible that secret creation happens but the config.Status patches are not applied
	if err := r.Client.Create(ctx, secret); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create bootstrap data secret for KThreesConfig %s/%s: %w", scope.Config.Namespace, scope.Config.Name, err)
		}
		r.Log.Info("bootstrap data secret for KThreesConfig already exists, updating", "secret", secret.Name, "KThreesConfig", scope.Config.Name)
		if err := r.Client.Update(ctx, secret); err != nil {
			return fmt.Errorf("failed to update bootstrap data secret for KThreesConfig %s/%s: %w", scope.Config.Namespace, scope.Config.Name, err)
		}
	}

	scope.Config.Status.DataSecretName = pointer.String(secret.Name)
	scope.Config.Status.Ready = true
	conditions.MarkTrue(scope.Config, bootstrapv1.DataSecretAvailableCondition)
	return nil
}

func (r *Ok3sConfigReconciler) reconcileKubeconfig(ctx context.Context, scope *Scope) (ctrl.Result, error) {
	logger := r.Log.WithValues("cluster", scope.Cluster.Name, "namespace", scope.Cluster.Namespace)

	_, err := secret.Get(ctx, r.Client, util.ObjectKey(scope.Cluster), secret.Kubeconfig)
	switch {
	case apierrors.IsNotFound(err):
		if err := kubeconfig.CreateSecret(ctx, r.Client, scope.Cluster); err != nil {
			if errors.Is(err, kubeconfig.ErrDependentCertificateNotFound) {
				logger.Info("could not find secret for cluster, requeuing", "secret", secret.ClusterCA)
				return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
			}
			return ctrl.Result{}, err
		}
	case err != nil:
		return ctrl.Result{}, fmt.Errorf("failed to retrieve Kubeconfig Secret for Cluster %q in namespace %q: %w", scope.Cluster.Name, scope.Cluster.Namespace, err)
	}

	return ctrl.Result{}, nil
}

func (r *Ok3sConfigReconciler) reconcileTopLevelObjectSettings(_ *clusterv1.Cluster, machine *clusterv1.Machine, config *bootstrapv1.Ok3sConfig) {
	log := r.Log.WithValues("kthreesconfig", fmt.Sprintf("%s/%s", config.Namespace, config.Name))

	// If there are no Version settings defined in Config, use Version from machine, if defined
	if config.Spec.Version == "" && machine.Spec.Version != nil {
		config.Spec.Version = *machine.Spec.Version
		log.Info("Altering Config", "Version", config.Spec.Version)
	}
}
