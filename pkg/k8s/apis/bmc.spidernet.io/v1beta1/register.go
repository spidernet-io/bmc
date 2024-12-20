// Package v1beta1 contains API Schema definitions for the bmc v1beta1 API group
// +kubebuilder:object:generate=true
// +groupName=bmc.spidernet.io

package v1beta1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// API Group and Version constants
const (
	// GroupName is the group name used in this package
	GroupName = "bmc.spidernet.io"
	// Version is the API version
	Version = "v1beta1"
	// APIVersion is the full API version string
	APIVersion = GroupName + "/" + Version
)

// Resource Kinds
const (
	// KindHostEndpoint is the kind name for HostEndpoint resource
	KindHostEndpoint = "HostEndpoint"
	// KindHostStatus is the kind name for HostStatus resource
	KindHostStatus = "HostStatus"
	// KindClusterAgent is the kind name for ClusterAgent resource
	KindClusterAgent = "ClusterAgent"
)

var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: Version}

var (
	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}

	AddToScheme = SchemeBuilder.AddToScheme
)

var (

	// Resource takes an unqualified resource and returns a Group qualified GroupResource
	Resource = func(resource string) schema.GroupResource {
		return SchemeGroupVersion.WithResource(resource).GroupResource()
	}

	// GroupResource takes an unqualified resource and returns a Group qualified GroupResource
	GroupResource = func(resource string) schema.GroupResource {
		return SchemeGroupVersion.WithResource(resource).GroupResource()
	}
)

func init() {
	SchemeBuilder.Register(&ClusterAgent{}, &ClusterAgentList{})
	SchemeBuilder.Register(&HostEndpoint{}, &HostEndpointList{})
	SchemeBuilder.Register(&HostStatus{}, &HostStatusList{})
}
