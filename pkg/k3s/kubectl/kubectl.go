package kubectl

import (
	"fmt"
	"os"

	"github.com/oneblock-ai/okr/pkg/k3s/config"
)

const kubectlDefault = "/usr/local/bin/kubectl"

var kubeconfigs = []string{
	"/etc/rancher/k3s/k3s.yaml",
}

func Env(k8sVersion string) []string {
	runtime := config.GetRuntime(k8sVersion)
	if runtime == config.RuntimeUnknown {
		return []string{}
	}
	return []string{
		fmt.Sprintf("KUBECONFIG=/etc/rancher/%s/%s.yaml", runtime, runtime),
	}
}

func Command(k8sVersion string) string {
	runtime := config.GetRuntime(k8sVersion)
	if runtime == config.RuntimeK3S {
		return kubectlDefault
	}
	return "kubectl"
}

func GetKubeconfig(kubeconfig string) (string, error) {
	if kubeconfig != "" {
		return kubeconfig, nil
	}

	for _, kubeconfig := range kubeconfigs {
		if _, err := os.Stat(kubeconfig); err == nil {
			return kubeconfig, nil
		}
	}
	return "", fmt.Errorf("failed to find kubeconfig file at %v", kubeconfigs)
}
