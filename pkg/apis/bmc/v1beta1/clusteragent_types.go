package v1beta1

import (
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
	// +optional
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=0
	Replicas *int32 `json:"replicas,omitempty"`
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
