package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	Scheme = runtime.NewScheme()
	Codecs = serializer.NewCodecFactory(Scheme)
)

func init() {
	SchemeBuilder.Register(&ClusterAgent{}, &ClusterAgentList{})
	SchemeBuilder.Register(&HostEndpoint{}, &HostEndpointList{})
}

// Adds the list of known types to Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(GroupVersion,
		&ClusterAgent{},
		&ClusterAgentList{},
		&HostEndpoint{},
		&HostEndpointList{},
	)
	metav1.AddToGroupVersion(scheme, GroupVersion)
	return nil
}
