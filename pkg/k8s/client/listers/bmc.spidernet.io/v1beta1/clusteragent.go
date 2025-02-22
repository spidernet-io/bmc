// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

// Code generated by lister-gen. DO NOT EDIT.

package v1beta1

import (
	bmcspidernetiov1beta1 "github.com/spidernet-io/bmc/pkg/k8s/apis/bmc.spidernet.io/v1beta1"
	labels "k8s.io/apimachinery/pkg/labels"
	listers "k8s.io/client-go/listers"
	cache "k8s.io/client-go/tools/cache"
)

// ClusterAgentLister helps list ClusterAgents.
// All objects returned here must be treated as read-only.
type ClusterAgentLister interface {
	// List lists all ClusterAgents in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*bmcspidernetiov1beta1.ClusterAgent, err error)
	// Get retrieves the ClusterAgent from the index for a given name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*bmcspidernetiov1beta1.ClusterAgent, error)
	ClusterAgentListerExpansion
}

// clusterAgentLister implements the ClusterAgentLister interface.
type clusterAgentLister struct {
	listers.ResourceIndexer[*bmcspidernetiov1beta1.ClusterAgent]
}

// NewClusterAgentLister returns a new ClusterAgentLister.
func NewClusterAgentLister(indexer cache.Indexer) ClusterAgentLister {
	return &clusterAgentLister{listers.New[*bmcspidernetiov1beta1.ClusterAgent](indexer, bmcspidernetiov1beta1.Resource("clusteragent"))}
}
