package v1beta1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto copies all properties of this object into another object of the same type
func (in *ClusterAgent) DeepCopyInto(out *ClusterAgent) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	if in.Status.AllocatedIPs != nil {
		in, out := &in.Status.AllocatedIPs, &out.Status.AllocatedIPs
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Status.LastUpdated != nil {
		in, out := &in.Status.LastUpdated, &out.Status.LastUpdated
		*out = (*in).DeepCopy()
	}
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
