/*
Copyright 2024 The Spidernet Authors.

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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HostStatusSpec defines the desired state of HostStatus
type HostStatusSpec struct {
	// ClusterAgent is the name of the ClusterAgent that this HostStatus belongs to
	// +kubebuilder:validation:Required
	ClusterAgent string `json:"clusterAgent"`

	// Basic contains basic information about the host
	// +kubebuilder:validation:Required
	Basic HostBasicInfo `json:"basic"`

	// Info contains additional information about the host
	// +optional
	Info *HostInfo `json:"info,omitempty"`
}

// HostBasicInfo contains basic information about the host
type HostBasicInfo struct {
	// IP is the IP address of the host
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$
	IP string `json:"ip"`

	// MAC is the MAC address of the host
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$
	MAC string `json:"mac"`
}

// HostInfo contains additional information about the host
type HostInfo struct {
	// Hostname is the hostname of the host
	// +optional
	Hostname string `json:"hostname,omitempty"`

	// Vendor is the vendor of the host
	// +optional
	Vendor string `json:"vendor,omitempty"`

	// Model is the model of the host
	// +optional
	Model string `json:"model,omitempty"`

	// SerialNumber is the serial number of the host
	// +optional
	SerialNumber string `json:"serialNumber,omitempty"`

	// BMCVersion is the version of the BMC
	// +optional
	BMCVersion string `json:"bmcVersion,omitempty"`

	// BIOSVersion is the version of the BIOS
	// +optional
	BIOSVersion string `json:"biosVersion,omitempty"`

	// CPUInfo contains information about the CPU
	// +optional
	CPUInfo *CPUInfo `json:"cpuInfo,omitempty"`

	// MemoryInfo contains information about the memory
	// +optional
	MemoryInfo *MemoryInfo `json:"memoryInfo,omitempty"`
}

// CPUInfo contains information about the CPU
type CPUInfo struct {
	// Count is the number of CPUs
	// +optional
	Count int32 `json:"count,omitempty"`

	// Model is the model of the CPU
	// +optional
	Model string `json:"model,omitempty"`

	// Speed is the speed of the CPU in MHz
	// +optional
	Speed int32 `json:"speed,omitempty"`
}

// MemoryInfo contains information about the memory
type MemoryInfo struct {
	// TotalGB is the total memory in GB
	// +optional
	TotalGB int32 `json:"totalGB,omitempty"`

	// Type is the type of memory
	// +optional
	Type string `json:"type,omitempty"`

	// Speed is the speed of memory in MHz
	// +optional
	Speed int32 `json:"speed,omitempty"`
}

// HostStatusStatus defines the observed state of HostStatus
type HostStatusStatus struct {
	// HealthReady indicates if the host is healthy and ready
	// +optional
	HealthReady bool `json:"healthReady,omitempty"`

	// LastUpdateTime is the last time this status was updated
	// +optional
	LastUpdateTime *metav1.Time `json:"lastUpdateTime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="READY",type="boolean",JSONPath=".status.healthReady"
//+kubebuilder:printcolumn:name="IP",type="string",JSONPath=".spec.basic.ip"
//+kubebuilder:printcolumn:name="MAC",type="string",JSONPath=".spec.basic.mac"
//+kubebuilder:printcolumn:name="CLUSTER_AGENT",type="string",JSONPath=".spec.clusterAgent"
//+kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"

// HostStatus is the Schema for the hoststatuses API
type HostStatus struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HostStatusSpec   `json:"spec,omitempty"`
	Status HostStatusStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// HostStatusList contains a list of HostStatus
type HostStatusList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HostStatus `json:"items"`
}
