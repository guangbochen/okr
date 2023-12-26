/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Ok3sConfig is the Schema for the ok3sconfigs API
type Ok3sConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   Ok3sConfigSpec   `json:"spec,omitempty"`
	Status Ok3sConfigStatus `json:"status,omitempty"`
}

// Ok3sConfigSpec defines the desired state of K3sConfig.
type Ok3sConfigSpec struct {
	// Files specifies extra files to be passed to user_data upon creation.
	// +optional
	//Files []File `json:"files,omitempty"`

	// Version specifies the k3s version
	// +optional
	Version string `json:"version,omitempty"`

	// ServerConfig specifies configuration for the agent nodes
	// +optional
	ServerConfig KThreesServerConfig `json:"serverConfig,omitempty"`
	// PreK3sCommands specifies extra commands to run before k3s setup runs

	// AgentConfig specifies configuration for the agent nodes
	// +optional
	AgentConfig KThreesAgentConfig `json:"agentConfig,omitempty"`

	// +optional
	PreK3sCommands []string `json:"preK3sCommands,omitempty"`

	// PostK3sCommands specifies extra commands to run after k3s setup runs
	// +optional
	PostK3sCommands []string `json:"postK3sCommands,omitempty"`
}

type KThreesServerConfig struct {
	// KubeAPIServerArgs is a customized flag for kube-apiserver process
	// +optional
	KubeAPIServerArgs []string `json:"kubeAPIServerArg,omitempty"`

	// KubeControllerManagerArgs is a customized flag for kube-bootstrap-manager process
	// +optional
	KubeControllerManagerArgs []string `json:"kubeControllerManagerArgs,omitempty"`

	// KubeSchedulerArgs is a customized flag for kube-scheduler process
	// +optional
	KubeSchedulerArgs []string `json:"kubeSchedulerArgs,omitempty"`

	// TLSSan Add additional hostname or IP as a Subject Alternative Name in the TLS cert
	// +optional
	TLSSan []string `json:"tlsSan,omitempty"`

	// BindAddress k3s bind address (default: 0.0.0.0)
	// +optional
	BindAddress string `json:"bindAddress,omitempty"`

	// HTTPSListenPort HTTPS listen port (default: 6443)
	// +optional
	HTTPSListenPort string `json:"httpsListenPort,omitempty"`

	// AdvertiseAddress IP address that apiserver uses to advertise to members of the cluster (default: node-external-ip/node-ip)
	// +optional
	AdvertiseAddress string `json:"advertiseAddress,omitempty"`

	// AdvertisePort Port that apiserver uses to advertise to members of the cluster (default: listen-port) (default: 0)
	// +optional
	AdvertisePort string `json:"advertisePort,omitempty"`

	// ClusterCidr  Network CIDR to use for pod IPs (default: "10.42.0.0/16")
	// +optional
	ClusterCidr string `json:"clusterCidr,omitempty"`

	// ServiceCidr Network CIDR to use for services IPs (default: "10.43.0.0/16")
	// +optional
	ServiceCidr string `json:"serviceCidr,omitempty"`

	// ClusterDNS  Cluster IP for coredns service. Should be in your service-cidr range (default: 10.43.0.10)
	// +optional
	ClusterDNS string `json:"clusterDNS,omitempty"`

	// ClusterDomain Cluster Domain (default: "cluster.local")
	// +optional
	ClusterDomain string `json:"clusterDomain,omitempty"`

	// DisableComponents  specifies extra commands to run before k3s setup runs
	// +optional
	DisableComponents []string `json:"disableComponents,omitempty"`

	// DisableExternalCloudProvider suppresses the 'cloud-provider=external' kubelet argument. (default: false)
	// +optional
	DisableExternalCloudProvider bool `json:"disableExternalCloudProvider,omitempty"`
}

type KThreesAgentConfig struct {
	// NodeLabels  Registering and starting kubelet with set of labels
	// +optional
	NodeLabels []string `json:"nodeLabels,omitempty"`

	// NodeTaints Registering kubelet with set of taints
	// +optional
	NodeTaints []string `json:"nodeTaints,omitempty"`

	// TODO: take in a object or secret and write to file. this is not useful
	// PrivateRegistry  registry configuration file (default: "/etc/rancher/k3s/registries.yaml")
	// +optional
	PrivateRegistry string `json:"privateRegistry,omitempty"`

	// KubeletArgs Customized flag for kubelet process
	// +optional
	KubeletArgs []string `json:"kubeletArgs,omitempty"`

	// KubeProxyArgs Customized flag for kube-proxy process
	// +optional
	KubeProxyArgs []string `json:"kubeProxyArgs,omitempty"`

	// NodeName Name of the Node
	// +optional
	NodeName string `json:"nodeName,omitempty"`
}

// KThreesConfigStatus defines the observed state of KThreesConfig.
type KThreesConfigStatus struct {
	// Ready indicates the BootstrapData field is ready to be consumed
	Ready bool `json:"ready,omitempty"`

	BootstrapData []byte `json:"bootstrapData,omitempty"`

	// DataSecretName is the name of the secret that stores the bootstrap data scripts.
	// +optional
	DataSecretName *string `json:"dataSecretName,omitempty"`

	// FailureReason will be set on non-retryable errors
	// +optional
	FailureReason string `json:"failureReason,omitempty"`

	// FailureMessage will be set on non-retryable errors
	// +optional
	FailureMessage string `json:"failureMessage,omitempty"`

	// ObservedGeneration is the latest generation observed by the bootstrap.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions defines current service state of the KThreesConfig.
	// +optional
	//Conditions clusterv1.Conditions `json:"conditions,omitempty"`
}

// Ok3sConfigStatus defines the observed state of Ok3sConfig
type Ok3sConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true

// Ok3sConfigList contains a list of Ok3sConfig
type Ok3sConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Ok3sConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Ok3sConfig{}, &Ok3sConfigList{})
}
