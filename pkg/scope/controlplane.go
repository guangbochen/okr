package scope

import (
	"context"

	"github.com/go-logr/logr"
	controlplanev1 "github.com/oneblock-ai/okr/api/controlplane/v1"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/patch"
)

type ControlPlaneScopeParams struct {
	Client         client.Client
	Logger         *logr.Logger
	Cluster        *clusterv1.Cluster
	ControlPlane   *controlplanev1.Ok3sControlPlane
	ControllerName string
}

func NewControlPlaneScope(params ControlPlaneScopeParams) (*ControlPlaneScope, error) {
	if params.Cluster == nil {
		return nil, errors.New("failed to generate new scope from nil Cluster")
	}
	if params.ControlPlane == nil {
		return nil, errors.New("failed to generate new scope from nil Ok3sControlPlane")
	}
	if params.Logger == nil {
		return nil, errors.New("failed to generate new scope from nil logger")
	}

	cpScope := &ControlPlaneScope{
		Logger:       params.Logger,
		Client:       params.Client,
		Cluster:      params.Cluster,
		ControlPlane: params.ControlPlane,
		patchHelper:  nil,
	}

	helper, err := patch.NewHelper(params.ControlPlane, params.Client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init patch helper")
	}

	cpScope.patchHelper = helper
	return cpScope, nil
}

type ControlPlaneScope struct {
	Client client.Client

	Cluster      *clusterv1.Cluster
	ControlPlane *controlplanev1.Ok3sControlPlane

	Logger      *logr.Logger
	patchHelper *patch.Helper
}

//func (s *ControlPlaneScope) Runtime() string {
//	runtime := s.KwokCluster.Spec.Runtime
//	if runtime == "" {
//		runtime = "docker"
//	}
//
//	return runtime
//}

func (s *ControlPlaneScope) Name() string {
	return s.Cluster.Name
}

func (s *ControlPlaneScope) PatchObject() error {
	return s.patchHelper.Patch(
		context.TODO(),
		s.ControlPlane,
		// patch.WithOwnedConditions{Conditions: []clusterv1.ConditionType{
		// 	infrav1.XYZCondition,
		// }}
	)
}

// Close closes the current scope persisting the control plane configuration and status.
func (s *ControlPlaneScope) Close() error {
	return s.PatchObject()
}
