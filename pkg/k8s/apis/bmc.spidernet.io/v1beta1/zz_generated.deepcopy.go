//go:build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1beta1

import (
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AgentConfig) DeepCopyInto(out *AgentConfig) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	if in.NodeAffinity != nil {
		in, out := &in.NodeAffinity, &out.NodeAffinity
		*out = new(v1.NodeAffinity)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AgentConfig.
func (in *AgentConfig) DeepCopy() *AgentConfig {
	if in == nil {
		return nil
	}
	out := new(AgentConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BasicInfo) DeepCopyInto(out *BasicInfo) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BasicInfo.
func (in *BasicInfo) DeepCopy() *BasicInfo {
	if in == nil {
		return nil
	}
	out := new(BasicInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterAgent) DeepCopyInto(out *ClusterAgent) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterAgent.
func (in *ClusterAgent) DeepCopy() *ClusterAgent {
	if in == nil {
		return nil
	}
	out := new(ClusterAgent)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ClusterAgent) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
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

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterAgentList.
func (in *ClusterAgentList) DeepCopy() *ClusterAgentList {
	if in == nil {
		return nil
	}
	out := new(ClusterAgentList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ClusterAgentList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterAgentSpec) DeepCopyInto(out *ClusterAgentSpec) {
	*out = *in
	in.AgentYaml.DeepCopyInto(&out.AgentYaml)
	if in.Endpoint != nil {
		in, out := &in.Endpoint, &out.Endpoint
		*out = new(EndpointConfig)
		**out = **in
	}
	if in.Feature != nil {
		in, out := &in.Feature, &out.Feature
		*out = new(FeatureConfig)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterAgentSpec.
func (in *ClusterAgentSpec) DeepCopy() *ClusterAgentSpec {
	if in == nil {
		return nil
	}
	out := new(ClusterAgentSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterAgentStatus) DeepCopyInto(out *ClusterAgentStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterAgentStatus.
func (in *ClusterAgentStatus) DeepCopy() *ClusterAgentStatus {
	if in == nil {
		return nil
	}
	out := new(ClusterAgentStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DhcpServerConfig) DeepCopyInto(out *DhcpServerConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DhcpServerConfig.
func (in *DhcpServerConfig) DeepCopy() *DhcpServerConfig {
	if in == nil {
		return nil
	}
	out := new(DhcpServerConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EndpointConfig) DeepCopyInto(out *EndpointConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EndpointConfig.
func (in *EndpointConfig) DeepCopy() *EndpointConfig {
	if in == nil {
		return nil
	}
	out := new(EndpointConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FeatureConfig) DeepCopyInto(out *FeatureConfig) {
	*out = *in
	if in.DhcpServerConfig != nil {
		in, out := &in.DhcpServerConfig, &out.DhcpServerConfig
		*out = new(DhcpServerConfig)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FeatureConfig.
func (in *FeatureConfig) DeepCopy() *FeatureConfig {
	if in == nil {
		return nil
	}
	out := new(FeatureConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HostStatus) DeepCopyInto(out *HostStatus) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HostStatus.
func (in *HostStatus) DeepCopy() *HostStatus {
	if in == nil {
		return nil
	}
	out := new(HostStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *HostStatus) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HostStatusList) DeepCopyInto(out *HostStatusList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]HostStatus, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HostStatusList.
func (in *HostStatusList) DeepCopy() *HostStatusList {
	if in == nil {
		return nil
	}
	out := new(HostStatusList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *HostStatusList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HostStatusStatus) DeepCopyInto(out *HostStatusStatus) {
	*out = *in
	out.Basic = in.Basic
	out.Info = in.Info
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HostStatusStatus.
func (in *HostStatusStatus) DeepCopy() *HostStatusStatus {
	if in == nil {
		return nil
	}
	out := new(HostStatusStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Info) DeepCopyInto(out *Info) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Info.
func (in *Info) DeepCopy() *Info {
	if in == nil {
		return nil
	}
	out := new(Info)
	in.DeepCopyInto(out)
	return out
}
