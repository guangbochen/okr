package config

import "strings"

var (
	RuntimeK3S     Runtime = "k3s"
	RuntimeUnknown Runtime = "unknown"
)

type Runtime string

type RuntimeConfig struct {
	Server          string                 `json:"server,omitempty"`
	Role            string                 `json:"role,omitempty"`
	SANS            []string               `json:"tlsSans,omitempty"`
	NodeName        string                 `json:"nodeName,omitempty"`
	Address         string                 `json:"address,omitempty"`
	InternalAddress string                 `json:"internalAddress,omitempty"`
	Taints          []string               `json:"taints,omitempty"`
	Labels          []string               `json:"labels,omitempty"`
	Token           string                 `json:"token,omitempty"`
	ConfigValues    map[string]interface{} `json:"extraConfig,omitempty"`
}

func GetRuntime(kubernetesVersion string) Runtime {
	if strings.Contains(kubernetesVersion, "k3s") {
		return RuntimeK3S
	}
	return RuntimeUnknown
}
