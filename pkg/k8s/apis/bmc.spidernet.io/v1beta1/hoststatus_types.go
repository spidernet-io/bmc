package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="CLUSTERAGENT",type="string",JSONPath=".status.clusterAgent"
// +kubebuilder:printcolumn:name="HEALTHREADY",type="boolean",JSONPath=".status.healthReady"
// +kubebuilder:printcolumn:name="IPADDR",type="string",JSONPath=".status.basic.ipAddr"
// +kubebuilder:printcolumn:name="TYPE",type="string",JSONPath=".status.basic.type"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"

type HostStatus struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status HostStatusStatus `json:"status,omitempty"`
}

type HostStatusStatus struct {
	HealthReady    bool              `json:"healthReady"`
	ClusterAgent   string            `json:"clusterAgent"`
	LastUpdateTime string            `json:"lastUpdateTime"`
	Basic          BasicInfo         `json:"basic"`
	Info           map[string]string `json:"info"`
}

type BasicInfo struct {
	Type            string `json:"type"`
	IpAddr          string `json:"ipAddr"`
	SecretName      string `json:"secretName,omitempty"`
	SecretNamespace string `json:"secretNamespace,omitempty"`
	Https           bool   `json:"https"`
	Port            int32  `json:"port"`
	Mac             string `json:"mac,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type HostStatusList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []HostStatus `json:"items"`
}
