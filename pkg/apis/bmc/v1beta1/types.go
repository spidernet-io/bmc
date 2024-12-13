package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterAgent is the Schema for the clusteragents API
type ClusterAgent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterAgentSpec   `json:"spec,omitempty"`
	Status ClusterAgentStatus `json:"status,omitempty"`
}

// DeepCopyInto copies all properties of this object into another object of the same type
func (in *ClusterAgent) DeepCopyInto(out *ClusterAgent) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy implements runtime.Object interface
func (in *ClusterAgent) DeepCopy() *ClusterAgent {
	if in == nil {
		return nil
	}
	out := new(ClusterAgent)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject implements runtime.Object interface
func (in *ClusterAgent) DeepCopyObject() runtime.Object {
	if in == nil {
		return nil
	}
	return in.DeepCopy()
}

// ClusterAgentSpec defines the desired state of ClusterAgent
type ClusterAgentSpec struct {
	// Interface specifies the network interface configuration
	Interface   string `json:"interface"`
	ClusterName string `json:"clusterName"`
}

// DeepCopyInto copies all properties of this object into another object of the same type
func (in *ClusterAgentSpec) DeepCopyInto(out *ClusterAgentSpec) {
	*out = *in
}

// DeepCopy implements runtime.Object interface
func (in *ClusterAgentSpec) DeepCopy() *ClusterAgentSpec {
	if in == nil {
		return nil
	}
	out := new(ClusterAgentSpec)
	in.DeepCopyInto(out)
	return out
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
}

// DeepCopyInto copies all properties of this object into another object of the same type
func (in *ClusterAgentStatus) DeepCopyInto(out *ClusterAgentStatus) {
	*out = *in
	if in.AllocatedIPs != nil {
		in, out := &in.AllocatedIPs, &out.AllocatedIPs
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy implements runtime.Object interface
func (in *ClusterAgentStatus) DeepCopy() *ClusterAgentStatus {
	if in == nil {
		return nil
	}
	out := new(ClusterAgentStatus)
	in.DeepCopyInto(out)
	return out
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterAgentList contains a list of ClusterAgent
type ClusterAgentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterAgent `json:"items"`
}

// DeepCopyInto copies all properties of this object into another object of the same type
func (in *ClusterAgentList) DeepCopyInto(out *ClusterAgentList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ClusterAgent, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy implements runtime.Object interface
func (in *ClusterAgentList) DeepCopy() *ClusterAgentList {
	if in == nil {
		return nil
	}
	out := new(ClusterAgentList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject implements runtime.Object interface
func (in *ClusterAgentList) DeepCopyObject() runtime.Object {
	if in == nil {
		return nil
	}
	return in.DeepCopy()
}

// BMCServer represents a BMC server configuration
type BMCServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BMCServerSpec   `json:"spec,omitempty"`
	Status BMCServerStatus `json:"status,omitempty"`
}

// DeepCopyInto copies all properties of this object into another object of the same type
func (in *BMCServer) DeepCopyInto(out *BMCServer) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy implements runtime.Object interface
func (in *BMCServer) DeepCopy() *BMCServer {
	if in == nil {
		return nil
	}
	out := new(BMCServer)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject implements runtime.Object interface
func (in *BMCServer) DeepCopyObject() runtime.Object {
	if in == nil {
		return nil
	}
	return in.DeepCopy()
}

// BMCServerSpec defines the desired state of BMCServer
type BMCServerSpec struct {
	// Host is the BMC server hostname or IP address
	Host string `json:"host"`
	// Username for BMC authentication
	Username string `json:"username"`
	// Password for BMC authentication
	Password string `json:"password"`
}

// DeepCopyInto copies all properties of this object into another object of the same type
func (in *BMCServerSpec) DeepCopyInto(out *BMCServerSpec) {
	*out = *in
}

// DeepCopy implements runtime.Object interface
func (in *BMCServerSpec) DeepCopy() *BMCServerSpec {
	if in == nil {
		return nil
	}
	out := new(BMCServerSpec)
	in.DeepCopyInto(out)
	return out
}

// BMCServerStatus defines the observed state of BMCServer
type BMCServerStatus struct {
	// Connected indicates whether the BMC server is accessible
	Connected bool `json:"connected"`
	// LastConnected is the timestamp of the last successful connection
	LastConnected *metav1.Time `json:"lastConnected,omitempty"`
}

// DeepCopyInto copies all properties of this object into another object of the same type
func (in *BMCServerStatus) DeepCopyInto(out *BMCServerStatus) {
	*out = *in
	if in.LastConnected != nil {
		in, out := &in.LastConnected, &out.LastConnected
		*out = (*in).DeepCopy()
	}
}

// DeepCopy implements runtime.Object interface
func (in *BMCServerStatus) DeepCopy() *BMCServerStatus {
	if in == nil {
		return nil
	}
	out := new(BMCServerStatus)
	in.DeepCopyInto(out)
	return out
}

// BMCServerList contains a list of BMCServer
type BMCServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BMCServer `json:"items"`
}

// DeepCopyInto copies all properties of this object into another object of the same type
func (in *BMCServerList) DeepCopyInto(out *BMCServerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]BMCServer, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy implements runtime.Object interface
func (in *BMCServerList) DeepCopy() *BMCServerList {
	if in == nil {
		return nil
	}
	out := new(BMCServerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject implements runtime.Object interface
func (in *BMCServerList) DeepCopyObject() runtime.Object {
	if in == nil {
		return nil
	}
	return in.DeepCopy()
}

// AddToScheme adds all types of this clientset into the given scheme.
func AddToScheme(scheme *runtime.Scheme) error {
	return SchemeBuilder.AddToScheme(scheme)
}

// SchemeBuilder is used to add go types to the GroupVersionKind scheme
var SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

// addKnownTypes adds the set of types defined in this package to the supplied scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&ClusterAgent{},
		&ClusterAgentList{},
		&BMCServer{},
		&BMCServerList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{Group: "bmc.spidernet.io", Version: "v1beta1"}
