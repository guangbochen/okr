package runtime

import (
	"fmt"

	"github.com/rancher/system-agent/pkg/applyinator"

	"github.com/oneblock-ai/okr/pkg/k3s/config"
	"github.com/oneblock-ai/okr/pkg/k3s/images"
)

func ToInstruction(cfg *config.RuntimeConfig, imageOverride string, systemDefaultRegistry string, k8sVersion string) (*applyinator.OneTimeInstruction, error) {
	runtime := config.GetRuntime(k8sVersion)

	var env []string
	if len(cfg.Server) != 0 {
		env = addEnv(env, "K3S_URL", cfg.Server)
	}

	env = addEnv(env, "K3S_TOKEN", cfg.Token)
	env = addEnv(env, "RESTART_STAMP", images.GetInstallerImage(imageOverride, systemDefaultRegistry, k8sVersion))
	return &applyinator.OneTimeInstruction{
		CommonInstruction: applyinator.CommonInstruction{
			Name:  string(runtime),
			Env:   env,
			Image: images.GetInstallerImage(imageOverride, systemDefaultRegistry, k8sVersion),
		},
		SaveOutput: true,
	}, nil
}

func addEnv(env []string, key, value string) []string {
	return append(env, fmt.Sprintf("%s=%s", key, value))
}
