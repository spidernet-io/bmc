package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ClusterAgent represents a cluster-wide agent deployment
type ClusterAgent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterAgentSpec   `json:"spec,omitempty"`
	Status ClusterAgentStatus `json:"status,omitempty"`
}

// ClusterAgentSpec defines the desired state of ClusterAgent
type ClusterAgentSpec struct {
	// Interface specifies the network interface configuration
	Interface   string `json:"interface"`
	ClusterName string `json:"clusterName"`
	// Image is the agent container image
	Image string `json:"image"`
	// Replicas is the number of agents to run
	Replicas int32 `json:"replicas"`
}

// ClusterAgentStatus defines the observed state of ClusterAgent
type ClusterAgentStatus struct {
	// Ready indicates whether all pods in the ClusterAgent deployment are running
	Ready bool `json:"ready"`
	// AllocatedIPCount is the number of allocated IPs
	AllocatedIPCount int `json:"allocatedIPCount"`
	// AllocatedIPs is a map of allocated IPs
	AllocatedIPs map[string]string `json:"allocatedIPs"`
	// TotalIPCount is the total number of IPs
	TotalIPCount int `json:"totalIPCount"`
	// AvailableReplicas is the number of agents currently running
	AvailableReplicas int32 `json:"availableReplicas,omitempty"`
	// LastUpdated is the last time the status was updated
	LastUpdated *metav1.Time `json:"lastUpdated,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterAgentList contains a list of ClusterAgent
type ClusterAgentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterAgent `json:"items"`
}
