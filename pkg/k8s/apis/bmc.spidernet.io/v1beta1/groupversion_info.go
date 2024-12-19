// Package v1beta1 contains API Schema definitions for the bmc v1beta1 API group
// +kubebuilder:object:generate=true
// +groupName=bmc.spidernet.io
package v1beta1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: "bmc.spidernet.io", Version: "v1beta1"}

	// Resource takes an unqualified resource and returns a Group qualified GroupResource
	Resource = func(resource string) schema.GroupResource {
		return SchemeGroupVersion.WithResource(resource).GroupResource()
	}

	// GroupResource takes an unqualified resource and returns a Group qualified GroupResource
	GroupResource = func(resource string) schema.GroupResource {
		return SchemeGroupVersion.WithResource(resource).GroupResource()
	}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

func init() {
	SchemeBuilder.Register(&ClusterAgent{}, &ClusterAgentList{})
	SchemeBuilder.Register(&HostEndpoint{}, &HostEndpointList{})
}
