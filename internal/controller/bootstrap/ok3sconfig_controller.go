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
	"log/slog"

	bootstrapv1 "github.com/oneblock-ai/okr/api/bootstrap/v1"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/annotations"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bsutil "sigs.k8s.io/cluster-api/bootstrap/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// InitLocker is a lock that is used around k3s init.
type InitLocker interface {
	Lock(ctx context.Context, cluster *clusterv1.Cluster, machine *clusterv1.Machine) bool
	Unlock(ctx context.Context, cluster *clusterv1.Cluster) bool
}

// Ok3sConfigReconciler reconciles a Ok3sConfig object
type Ok3sConfigReconciler struct {
	client.Client
	log         slog.Logger
	Scheme      *runtime.Scheme
	K3sInitLock InitLocker
}

// SetupWithManager sets up the bootstrap with the Manager.
func (r *Ok3sConfigReconciler) SetupWithManager(mgr ctrl.Manager, ctx context.Context) error {
	//if r.K3sInitLock == nil {
	//	r.K3sInitLock = locking.NewControlPlaneInitMutex(ctrl.Log.WithName("init-locker"), mgr.GetClient())
	//}

	_ = r.log.Enabled(ctx, 0)

	return ctrl.NewControllerManagedBy(mgr).
		For(&bootstrapv1.Ok3sConfig{}).
		Complete(r)
}

//+kubebuilder:rbac:groups=bootstrap.cluster.x-k8s.io,resources=ok3sconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=bootstrap.cluster.x-k8s.io,resources=ok3sconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=bootstrap.cluster.x-k8s.io,resources=ok3sconfigs/finalizers,verbs=update

func (r *Ok3sConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	// Lookup the ok3s config
	config := &bootstrapv1.Ok3sConfig{}
	if err := r.Client.Get(ctx, req.NamespacedName, config); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		r.log.Error("Failed to get config")
		return ctrl.Result{}, err
	}

	// Look up the owner of this ok3s config if there is one
	configOwner, err := bsutil.GetConfigOwner(ctx, r.Client, config)
	if apierrors.IsNotFound(err) {
		// Could not find the owner yet, this is not an error and will rereconcile when the owner gets set.
		return ctrl.Result{}, nil
	}
	if err != nil {
		r.log.Error("Failed to get owner", err.Error())
		return ctrl.Result{}, err
	}
	if configOwner == nil {
		return ctrl.Result{}, nil
	}

	// Lookup the cluster the config owner is associated with
	cluster, err := util.GetClusterByName(ctx, r.Client, configOwner.GetNamespace(), configOwner.ClusterName())
	if err != nil {
		if errors.Is(err, util.ErrNoCluster) {
			r.log.Info(fmt.Sprintf("%s does not belong to a cluster yet, waiting until it's part of a cluster", configOwner.GetKind()))
			return ctrl.Result{}, nil
		}

		if apierrors.IsNotFound(err) {
			r.log.Info("Cluster does not exist yet, waiting until it is created")
			return ctrl.Result{}, nil
		}
		r.log.Error("Could not get cluster with metadata", err.Error())
		return ctrl.Result{}, err
	}

	if annotations.IsPaused(cluster, config) {
		r.log.Info("Reconciliation is paused for this object")
		return ctrl.Result{}, nil
	}

	//scope := &Scope{
	//	Logger:      log,
	//	Config:      config,
	//	ConfigOwner: configOwner,
	//	Cluster:     cluster,
	//}

	return ctrl.Result{}, nil
}
