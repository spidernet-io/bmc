package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// ClusterAgentSpec defines the desired state of ClusterAgent
type ClusterAgentSpec struct {
	// UnderlayInterface specifies the network interface configuration for underlay network
	// +kubebuilder:validation:Required
	UnderlayInterface string `json:"underlayInterface"`

	// ClusterName specifies the name of the cluster
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=^[a-z0-9][a-z0-9-]*[a-z0-9]$
	ClusterName string `json:"clusterName"`

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
