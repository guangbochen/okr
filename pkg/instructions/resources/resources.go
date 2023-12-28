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

	"github.com/oneblock-ai/okr/pkg/config"
	"github.com/oneblock-ai/okr/pkg/instructions/images"
	"github.com/oneblock-ai/okr/pkg/kubectl"
	"github.com/oneblock-ai/okr/pkg/self"
	"github.com/oneblock-ai/okr/pkg/utils"
)

//const (
//	localRKEStateSecretType = "rke.cattle.io/cluster-state"
//)

//func writeCattleID(id string) error {
//	if err := os.MkdirAll("/etc/rancher", 0755); err != nil {
//		return fmt.Errorf("mkdir /etc/rancher: %w", err)
//	}
//	if err := os.MkdirAll("/etc/rancher/agent", 0700); err != nil {
//		return fmt.Errorf("mkdir /etc/rancher/agent: %w", err)
//	}
//	return os.WriteFile("/etc/rancher/agent/cattle-id", []byte(id), 0400)
//}
//
//func getCattleID() (string, error) {
//	data, err := os.ReadFile("/etc/rancher/agent/cattle-id")
//	if os.IsNotExist(err) {
//	} else if err != nil {
//		return "", err
//	}
//	id := strings.TrimSpace(string(data))
//	if id == "" {
//		id, err = randomtoken.Generate()
//		if err != nil {
//			return "", err
//		}
//		return id, writeCattleID(id)
//	}
//	return id, nil
//}

func ToBootstrapFile(config *config.Config, path string) (*applyinator.File, error) {
	nodeName := config.NodeName
	if nodeName == "" {
		hostname, err := os.Hostname()
		if err != nil {
			return nil, fmt.Errorf("looking up hostname: %w", err)
		}
		nodeName = strings.Split(hostname, ".")[0]
	}

	//k8sVersion, err := versions.K8sVersion(config.KubernetesVersion)
	//if err != nil {
	//	return nil, err
	//}

	//token := config.Token
	//if token == "" {
	//	token, err = randomtoken.Generate()
	//	if err != nil {
	//		return nil, err
	//	}
	//}

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
		//}, utils.GenericMap{
		//	Data: map[string]interface{}{
		//		"kind":       "Namespace",
		//		"apiVersion": "v1",
		//		"metadata": map[string]interface{}{
		//			"name": "fleet-local",
		//		},
		//	},
		//}, utils.GenericMap{
		//	Data: map[string]interface{}{
		//		"kind":       "Cluster",
		//		"apiVersion": "provisioning.cattle.io/v1",
		//		"metadata": map[string]interface{}{
		//			"name":      "local",
		//			"namespace": "fleet-local",
		//		},
		//		"spec": map[string]interface{}{
		//			"kubernetesVersion": k8sVersion,
		//			"rkeConfig": map[string]interface{}{
		//				"controlPlaneConfig": config.ConfigValues,
		//			},
		//		},
		//	},
		//}, utils.GenericMap{
		//	Data: map[string]interface{}{
		//		"kind":       "Secret",
		//		"apiVersion": "v1",
		//		"metadata": map[string]interface{}{
		//			"name":      "local-rke-state",
		//			"namespace": "fleet-local",
		//		},
		//		"type": localRKEStateSecretType,
		//		"data": map[string]interface{}{
		//			"serverToken": []byte(token),
		//			"agentToken":  []byte(token),
		//		},
		//	},
		//}, utils.GenericMap{
		//	Data: map[string]interface{}{
		//		"kind":       "ClusterRegistrationToken",
		//		"apiVersion": "management.cattle.io/v3",
		//		"metadata": map[string]interface{}{
		//			"name":      "default-token",
		//			"namespace": "local",
		//		},
		//		"spec": map[string]interface{}{
		//			"clusterName": "local",
		//		},
		//		"status": map[string]interface{}{
		//			"token": token,
		//		},
		//	},
		//}, utils.GenericMap{
		//	Data: map[string]interface{}{
		//		"apiVersion": "catalog.cattle.io/v1",
		//		"kind":       "ClusterRepo",
		//		"metadata": map[string]interface{}{
		//			"name": "rancher-stable",
		//		},
		//		"spec": map[string]interface{}{
		//			"url": "https://releases.rancher.com/server-charts/stable",
		//		},
		//	},
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

func ToInstruction(imageOverride, systemDefaultRegistry, k8sVersion, dataDir string) (*applyinator.PeriodicInstruction, error) {
	bootstrap := GetBootstrapManifests(dataDir)
	cmd, err := self.Self()
	if err != nil {
		return nil, fmt.Errorf("resolving location of %s: %w", os.Args[0], err)
	}
	return &applyinator.PeriodicInstruction{
		CommonInstruction: applyinator.CommonInstruction{
			Name:    "bootstrap",
			Image:   images.GetInstallerImage(imageOverride, systemDefaultRegistry, k8sVersion),
			Args:    []string{kubectl.Command(k8sVersion), "apply", "--validate='ignore'", "-f", bootstrap},
			Command: cmd,
			Env:     kubectl.Env(k8sVersion),
		},
		SaveStderrOutput: true,
		PeriodSeconds:    15,
	}, nil
}
