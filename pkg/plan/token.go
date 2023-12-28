package plan

import (
	"os"

	"github.com/rancher/wrangler/v2/pkg/data/convert"
	"github.com/rancher/wrangler/v2/pkg/randomtoken"
	"github.com/rancher/wrangler/v2/pkg/yaml"

	"github.com/oneblock-ai/okr/pkg/config"
	"github.com/oneblock-ai/okr/pkg/instructions/runtime"
	"github.com/oneblock-ai/okr/pkg/versions"
)

func assignTokenIfUnset(cfg *config.Config) error {
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

func existingToken(cfg *config.Config) (string, error) {
	k8sVersion, err := versions.K8sVersion(cfg.KubernetesVersion)
	if err != nil {
		return "", err
	}

	cfgFile := runtime.GetKubeRuntimeConfigLocation(config.GetRuntime(k8sVersion))
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
