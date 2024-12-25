// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

// Code generated by informer-gen. DO NOT EDIT.

package v1beta1

import (
	internalinterfaces "github.com/spidernet-io/bmc/pkg/k8s/client/informers/externalversions/internalinterfaces"
)

// Interface provides access to all the informers in this group version.
type Interface interface {
	// ClusterAgents returns a ClusterAgentInformer.
	ClusterAgents() ClusterAgentInformer
	// HostEndpoints returns a HostEndpointInformer.
	HostEndpoints() HostEndpointInformer
	// HostOperations returns a HostOperationInformer.
	HostOperations() HostOperationInformer
	// HostStatuses returns a HostStatusInformer.
	HostStatuses() HostStatusInformer
}

type version struct {
	factory          internalinterfaces.SharedInformerFactory
	namespace        string
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// New returns a new Interface.
func New(f internalinterfaces.SharedInformerFactory, namespace string, tweakListOptions internalinterfaces.TweakListOptionsFunc) Interface {
	return &version{factory: f, namespace: namespace, tweakListOptions: tweakListOptions}
}

// ClusterAgents returns a ClusterAgentInformer.
func (v *version) ClusterAgents() ClusterAgentInformer {
	return &clusterAgentInformer{factory: v.factory, tweakListOptions: v.tweakListOptions}
}

// HostEndpoints returns a HostEndpointInformer.
func (v *version) HostEndpoints() HostEndpointInformer {
	return &hostEndpointInformer{factory: v.factory, tweakListOptions: v.tweakListOptions}
}

// HostOperations returns a HostOperationInformer.
func (v *version) HostOperations() HostOperationInformer {
	return &hostOperationInformer{factory: v.factory, tweakListOptions: v.tweakListOptions}
}

// HostStatuses returns a HostStatusInformer.
func (v *version) HostStatuses() HostStatusInformer {
	return &hostStatusInformer{factory: v.factory, tweakListOptions: v.tweakListOptions}
}