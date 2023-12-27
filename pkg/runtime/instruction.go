package runtime

import (
	"github.com/rancher/system-agent/pkg/applyinator"

	"github.com/oneblock-ai/okr/pkg/config"
	"github.com/oneblock-ai/okr/pkg/images"
)

func ToInstruction(imageOverride string, systemDefaultRegistry string, k8sVersion string) (*applyinator.OneTimeInstruction, error) {
	runtime := config.GetRuntime(k8sVersion)
	return &applyinator.OneTimeInstruction{
		CommonInstruction: applyinator.CommonInstruction{
			Name: string(runtime),
			Env: []string{
				"RESTART_STAMP=" + images.GetInstallerImage(imageOverride, systemDefaultRegistry, k8sVersion),
			},
			Image: images.GetInstallerImage(imageOverride, systemDefaultRegistry, k8sVersion),
		},
		SaveOutput: true,
	}, nil
}

//func ToUpgradeInstruction(k8sVersion string) (*applyinator.Instruction, error) {
//	cmd, err := self.Self()
//	if err != nil {
//		return nil, fmt.Errorf("resolving location of %s: %w", os.Args[0], err)
//	}
//	patch, err := json.Marshal(map[string]interface{}{
//		"spec": map[string]interface{}{
//			"kubernetesVersion": k8sVersion,
//		},
//	})
//	if err != nil {
//		return nil, err
//	}
//	return &applyinator.Instruction{
//		Name:       "patch-kubernetes-version",
//		SaveOutput: true,
//		Args:       []string{"retry", kubectl.Command(k8sVersion), "--type=merge", "-n", "fleet-local", "patch", "clusters.provisioning.cattle.io", "local", "-p", string(patch)},
//		Env:        kubectl.Env(k8sVersion),
//		Command:    cmd,
//	}, nil
//}
