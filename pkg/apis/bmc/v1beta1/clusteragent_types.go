package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="boolean",JSONPath=".status.ready"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"

// ClusterAgentSpec defines the desired state of ClusterAgent
type ClusterAgentSpec struct {
	// AgentYaml contains the agent configuration
	// +kubebuilder:validation:Required
	AgentYaml AgentConfig `json:"agentYaml"`

	// Endpoint contains the endpoint configuration
	// +optional
	Endpoint *EndpointConfig `json:"endpoint,omitempty"`

	// Feature contains the feature configuration
	// +optional
	Feature *FeatureConfig `json:"feature,omitempty"`
}

// AgentConfig defines the configuration for the agent
type AgentConfig struct {
	// UnderlayInterface specifies the network interface configuration for underlay network
	// +kubebuilder:validation:Required
	UnderlayInterface string `json:"underlayInterface"`

	// Image is the agent container image
	// +optional
	Image string `json:"image,omitempty"`

	// Replicas is the number of agents to run
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=0
	Replicas *int32 `json:"replicas"`

	// NodeAffinity defines scheduling constraints for the agent pods
	// +optional
	NodeAffinity *corev1.NodeAffinity `json:"nodeAffinity,omitempty"`

	// NodeName is a request to schedule this pod onto a specific node
	// +optional
	NodeName string `json:"nodeName,omitempty"`
}

// EndpointConfig defines the endpoint configuration for the agent
type EndpointConfig struct {
	// Port is the endpoint port
	// +kubebuilder:default=443
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int32 `json:"port,omitempty"`

	// SecretName is the name of the secret containing the TLS certificates
	// +optional
	SecretName string `json:"secretName,omitempty"`

	// SecretNamespace is the namespace of the secret containing the TLS certificates
	// +optional
	SecretNamespace string `json:"secretNamespace,omitempty"`

	// HTTPS enables HTTPS for the endpoint
	// +kubebuilder:default=true
	HTTPS bool `json:"https,omitempty"`
}

// FeatureConfig defines the feature configuration for the agent
type FeatureConfig struct {
	// EnableDhcpServer enables the DHCP server
	// +kubebuilder:default=true
	EnableDhcpServer bool `json:"enableDhcpServer,omitempty"`

	// EnableDhcpDiscovery enables DHCP discovery
	// +kubebuilder:default=true
	EnableDhcpDiscovery bool `json:"enableDhcpDiscovery,omitempty"`

	// DhcpServerInterface specifies the interface for DHCP server
	// +kubebuilder:default="net1"
	DhcpServerInterface string `json:"dhcpServerInterface,omitempty"`

	// RedfishMetrics enables redfish metrics collection
	// +kubebuilder:default=false
	RedfishMetrics bool `json:"redfishMetrics,omitempty"`

	// EnableGuiProxy enables GUI proxy
	// +kubebuilder:default=true
	EnableGuiProxy bool `json:"enableGuiProxy,omitempty"`
}

// ClusterAgentStatus defines the observed state of ClusterAgent
type ClusterAgentStatus struct {
	// Whether the agent is ready
	// +optional
	Ready bool `json:"ready,omitempty"`
}

// ClusterAgent represents a cluster-wide agent deployment
type ClusterAgent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterAgentSpec   `json:"spec,omitempty"`
	Status ClusterAgentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterAgentList contains a list of ClusterAgent
type ClusterAgentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterAgent `json:"items"`
}
