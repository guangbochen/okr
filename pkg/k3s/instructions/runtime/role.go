package runtime

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/rancher/system-agent/pkg/applyinator"
	"github.com/rancher/wrangler/v2/pkg/data/convert"
	"sigs.k8s.io/yaml"

	"github.com/oneblock-ai/okr/pkg/k3s/config"
)

var (
	normalizeNames = map[string]string{
		"tlsSans":         "tls-san",
		"nodeName":        "node-name",
		"internalAddress": "internal-address",
		"taints":          "node-taint",
		"labels":          "node-label",
	}
)

func ToFile(config *config.RuntimeConfig, runtime config.Runtime, clusterInit bool) (*applyinator.File, error) {
	data, err := ToConfig(config, clusterInit)
	if err != nil {
		return nil, err
	}
	return &applyinator.File{
		Content: base64.StdEncoding.EncodeToString(data),
		Path:    GetKubeRuntimeConfigLocation(runtime),
	}, nil
}

func ToConfig(config *config.RuntimeConfig, clusterInit bool) ([]byte, error) {
	configObjects := []interface{}{
		config.ConfigValues,
	}

	result := map[string]interface{}{}
	for _, data := range configObjects {
		data, err := convert.EncodeToMap(data)
		if err != nil {
			return nil, err
		}
		delete(data, "extraConfig")
		delete(data, "role")
		for oldKey, newKey := range normalizeNames {
			value, ok := data[oldKey]
			if !ok {
				continue
			}
			delete(data, oldKey)
			data[newKey] = value
		}
		for k, v := range data {
			newKey := strings.ReplaceAll(convert.ToYAMLKey(k), "_", "-")
			result[newKey] = v
		}

		if clusterInit {
			result["cluster-init"] = "true"
		}
	}

	return yaml.Marshal(result)
}

func GetKubeRuntimeConfigLocation(runtime config.Runtime) string {
	return fmt.Sprintf("/etc/rancher/%s/config.yaml.d/40-okr.yaml", runtime)
}
