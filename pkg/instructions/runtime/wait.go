package runtime

import (
	"fmt"
	"os"

	"github.com/rancher/system-agent/pkg/applyinator"

	"github.com/oneblock-ai/okr/pkg/kubectl"
	"github.com/oneblock-ai/okr/pkg/self"
)

func ToWaitKubernetesInstruction(imageOverride, systemDefaultRegistry, k8sVersion string) (*applyinator.OneTimeInstruction, error) {
	cmd, err := self.Self()
	if err != nil {
		return nil, fmt.Errorf("resolving location of %s: %w", os.Args[0], err)
	}
	return &applyinator.OneTimeInstruction{
		CommonInstruction: applyinator.CommonInstruction{
			Name: "wait-kubernetes-provisioned",
			Args: []string{"retry", kubectl.Command(k8sVersion), "-n", "fleet-local", "wait",
				"--for=condition=Provisioned=true", "clusters.provisioning.cattle.io", "local"},
			Env:     kubectl.Env(k8sVersion),
			Command: cmd,
		},
		SaveOutput: true,
	}, nil
}
