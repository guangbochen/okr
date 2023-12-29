package plan

import (
	"os"

	"github.com/rancher/wrangler/v2/pkg/data/convert"
	"github.com/rancher/wrangler/v2/pkg/randomtoken"
	"github.com/rancher/wrangler/v2/pkg/yaml"

	config2 "github.com/oneblock-ai/okr/pkg/k3s/config"
	"github.com/oneblock-ai/okr/pkg/k3s/instructions/runtime"
	"github.com/oneblock-ai/okr/pkg/k3s/versions"
)

func assignTokenIfUnset(cfg *config2.Config) error {
	if cfg.Token != "" {
		return nil
	}

	token, err := existingToken(cfg)
	if err != nil {
		return err
	}

	if token == "" {
		if token, err = randomtoken.Generate(); err != nil {
			return err
		}
	}

	cfg.Token = token
	return nil
}

func existingToken(cfg *config2.Config) (string, error) {
	k8sVersion, err := versions.K8sVersion(cfg.KubernetesVersion)
	if err != nil {
		return "", err
	}

	cfgFile := runtime.GetKubeRuntimeConfigLocation(config2.GetRuntime(k8sVersion))
	data, err := os.ReadFile(cfgFile)
	if os.IsNotExist(err) {
		return "", nil
	} else if err != nil {
		return "", err
	}

	configMap := map[string]interface{}{}
	if err := yaml.Unmarshal(data, &configMap); err != nil {
		return "", err
	}

	return convert.ToString(configMap["token"]), nil
}
