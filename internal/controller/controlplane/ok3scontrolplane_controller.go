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

package controlplane

import (
	"context"
	"fmt"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/annotations"
	"sigs.k8s.io/cluster-api/util/predicates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	controlplanev1 "github.com/oneblock-ai/okr/api/controlplane/v1"
	"github.com/oneblock-ai/okr/pkg/scope"
	"github.com/oneblock-ai/okr/pkg/services"
)

// Reconciler reconciles a Ok3sControlPlane object
type Reconciler struct {
	client.Client
	Scheme           *runtime.Scheme
	WatchFilterValue string
}

// SetupWithManager sets up the bootstrap with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager, ctx context.Context) error {
	logger := log.FromContext(ctx)

	controlPlane := &controlplanev1.Ok3sControlPlane{}
	c, err := ctrl.NewControllerManagedBy(mgr).
		For(controlPlane).
		WithEventFilter(predicates.ResourceNotPausedAndHasFilterLabel(logger, r.WatchFilterValue)).
		Build(r)

	if err != nil {
		return fmt.Errorf("failed setting up the control-plane controller manager: %w", err)
	}

	if err = c.Watch(
		source.Kind(nil, &clusterv1.Cluster{}),
		handler.EnqueueRequestsFromMapFunc(util.ClusterToInfrastructureMapFunc(ctx, controlPlane.GroupVersionKind(), mgr.GetClient(), &controlplanev1.Ok3sControlPlane{})),
		predicates.ClusterUnpausedAndInfrastructureReady(logger),
	); err != nil {
		return fmt.Errorf("failed adding a watch for ready clusters: %w", err)
	}

	return nil
}

//+kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=ok3scontrolplanes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=ok3scontrolplanes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=ok3scontrolplanes/finalizers,verbs=update
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;clusters/status,verbs=get;list;watch

// Reconcile the Ok3sControlPlane object against the actual cluster state, and then
// perform operations to make the current cluster state closer to the desired state.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// TODO(user): your logic here
	cp := &controlplanev1.Ok3sControlPlane{}
	err := r.Get(ctx, req.NamespacedName, cp)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// Fetch the Cluster.
	cluster, err := util.GetOwnerCluster(ctx, r.Client, cp.ObjectMeta)
	if err != nil {
		logger.Error(err, "Failed to retrieve owner Cluster from the API Server")

		return ctrl.Result{}, err
	}

	if cluster == nil {
		logger.Info("Cluster Controller has not yet set OwnerRef")

		return ctrl.Result{Requeue: true}, nil
	}

	logger = logger.WithValues("cluster", cluster.Name)

	if annotations.IsPaused(cluster, cp) {
		logger.Info("Reconciliation is paused for this object")

		return ctrl.Result{}, nil
	}

	cpScope, err := scope.NewControlPlaneScope(scope.ControlPlaneScopeParams{
		Client:         r.Client,
		Cluster:        cluster,
		ControlPlane:   cp,
		ControllerName: strings.ToLower(cp.Kind),
		Logger:         &logger,
	})
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("failed to create scope: %w", err)
	}

	defer func() {
		//TODO: update conditions if needed
		if err = cpScope.Close(); err != nil {
		}
	}()

	if !cp.ObjectMeta.DeletionTimestamp.IsZero() {
		// Handle deletion reconciliation loop.
		return r.reconcileDelete(ctx, cpScope)
	}

	// Handle normal reconciliation loop.
	return r.reconcileNormal(ctx, cpScope)

	//return ctrl.Result{}, nil
}

func (r *Reconciler) reconcileNormal(ctx context.Context, cpScope *scope.ControlPlaneScope) (res ctrl.Result, reterr error) {
	cpScope.Logger.Info("Reconciling Ok3sControlPlane")

	if controllerutil.AddFinalizer(cpScope.ControlPlane, controlplanev1.Ok3sControlPlaneFinalizer) {
		if err := cpScope.PatchObject(); err != nil {
			return ctrl.Result{}, err
		}
	}

	reconcilers := []services.ReconcilerWithResult{
		//cluster.NewService(cpScope),
	}

	for _, r := range reconcilers {
		res, err := r.Reconcile(ctx)
		if err != nil {
			cpScope.Logger.Error(err, "Reconcile error")
			//record.Warnf(clusterScope.GCPCluster, "GCPClusterReconcile", "Reconcile error - %v", err)
			return ctrl.Result{}, err
		}
		if res.Requeue || res.RequeueAfter > 0 {
			return res, nil
		}
	}

	return reconcile.Result{}, nil
}

func (r *Reconciler) reconcileDelete(ctx context.Context, cpScope *scope.ControlPlaneScope) (ctrl.Result, error) {
	cpScope.Logger.Info("Reconciling Ok3sControlPlane delete")

	reconcilers := []services.ReconcilerWithResult{
		//cluster.NewService(cpScope),
	}

	for _, r := range reconcilers {
		res, err := r.Delete(ctx)
		if err != nil {
			cpScope.Logger.Error(err, "Reconcile error")
			//record.Warnf(clusterScope.GCPCluster, "GCPClusterReconcile", "Reconcile error - %v", err)
			return ctrl.Result{}, err
		}
		if res.Requeue || res.RequeueAfter > 0 {
			return res, nil
		}
	}

	controllerutil.RemoveFinalizer(cpScope.ControlPlane, controlplanev1.Ok3sControlPlaneFinalizer)

	return reconcile.Result{}, nil
}
