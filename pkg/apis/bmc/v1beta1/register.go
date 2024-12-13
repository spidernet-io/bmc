package v1beta1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var Scheme = runtime.NewScheme()
var Codecs = serializer.NewCodecFactory(Scheme)
var ParameterCodec = runtime.NewParameterCodec(Scheme)

func init() {
	SchemeBuilder.Register(&ClusterAgent{}, &ClusterAgentList{})
}

// RegisterTypes adds all types of this clientset into the given scheme.
func RegisterTypes(scheme *runtime.Scheme) error {
	SchemeGroupVersion := schema.GroupVersion{Group: "bmc.spidernet.io", Version: "v1beta1"}
	scheme.AddKnownTypes(SchemeGroupVersion,
		&ClusterAgent{},
		&ClusterAgentList{},
	)
	return nil
}
