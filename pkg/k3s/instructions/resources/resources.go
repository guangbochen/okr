package resources

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/rancher/system-agent/pkg/applyinator"
	"github.com/rancher/wrangler/v2/pkg/yaml"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/oneblock-ai/okr/pkg/k3s/config"
	"github.com/oneblock-ai/okr/pkg/k3s/images"
	"github.com/oneblock-ai/okr/pkg/k3s/kubectl"
	"github.com/oneblock-ai/okr/pkg/k3s/self"
	"github.com/oneblock-ai/okr/pkg/utils"
)

func ToBootstrapFile(config *config.Config, path string) (*applyinator.File, error) {
	nodeName := config.NodeName
	if nodeName == "" {
		hostname, err := os.Hostname()
		if err != nil {
			return nil, fmt.Errorf("looking up hostname: %w", err)
		}
		nodeName = strings.Split(hostname, ".")[0]
	}

	resources := config.Resources
	return ToFile(append(resources, utils.GenericMap{
		Data: map[string]interface{}{
			"kind":       "Node",
			"apiVersion": "v1",
			"metadata": map[string]interface{}{
				"name": nodeName,
				"labels": map[string]interface{}{
					"node-role.kubernetes.io/etcd": "true",
				},
			},
		},
	}, utils.GenericMap{
		Data: map[string]interface{}{
			"kind":       "Namespace",
			"apiVersion": "v1",
			"metadata": map[string]interface{}{
				"name": "kuberay-system",
			},
		},
	}, utils.GenericMap{
		Data: map[string]interface{}{
			"kind":       "HelmChart",
			"apiVersion": "helm.cattle.io/v1",
			"metadata": map[string]interface{}{
				"name":      "kuberay-operator",
				"namespace": "kube-system",
			},
			"spec": map[string]interface{}{
				"repo":            "https://ray-project.github.io/kuberay-helm",
				"chart":           "kuberay-operator",
				"targetNamespace": "kuberay-system",
				"version":         "1.0.0",
			},
		},
	}), path)
}
func ToFile(resources []utils.GenericMap, path string) (*applyinator.File, error) {
	if len(resources) == 0 {
		return nil, nil
	}

	var objs []runtime.Object
	for _, resource := range resources {
		objs = append(objs, &unstructured.Unstructured{
			Object: resource.Data,
		})
	}

	data, err := yaml.ToBytes(objs)
	if err != nil {
		return nil, err
	}

	return &applyinator.File{
		Content: base64.StdEncoding.EncodeToString(data),
		Path:    path,
	}, nil
}

func GetBootstrapManifests(dataDir string) string {
	return fmt.Sprintf("%s/bootstrapmanifests/okr.yaml", dataDir)
}

func ToInstruction(imageOverride, systemDefaultRegistry, k8sVersion, dataDir string) (*applyinator.OneTimeInstruction, error) {
	bootstrap := GetBootstrapManifests(dataDir)
	cmd, err := self.Self()
	if err != nil {
		return nil, fmt.Errorf("resolving location of %s: %w", os.Args[0], err)
	}
	return &applyinator.OneTimeInstruction{
		CommonInstruction: applyinator.CommonInstruction{
			Name:    "bootstrap",
			Image:   images.GetInstallerImage(imageOverride, systemDefaultRegistry, k8sVersion),
			Args:    []string{"retry", kubectl.Command(k8sVersion), "apply", "--validate=false", "-f", bootstrap},
			Command: cmd,
			Env:     kubectl.Env(k8sVersion),
		},
		SaveOutput: true,
	}, nil
}
