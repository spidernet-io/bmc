// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

// Code generated by informer-gen. DO NOT EDIT.

package v1beta1

import (
	context "context"
	time "time"

	apisbmcspidernetiov1beta1 "github.com/spidernet-io/bmc/pkg/k8s/apis/bmc.spidernet.io/v1beta1"
	versioned "github.com/spidernet-io/bmc/pkg/k8s/client/clientset/versioned"
	internalinterfaces "github.com/spidernet-io/bmc/pkg/k8s/client/informers/externalversions/internalinterfaces"
	bmcspidernetiov1beta1 "github.com/spidernet-io/bmc/pkg/k8s/client/listers/bmc.spidernet.io/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// HostOperationInformer provides access to a shared informer and lister for
// HostOperations.
type HostOperationInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() bmcspidernetiov1beta1.HostOperationLister
}

type hostOperationInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewHostOperationInformer constructs a new informer for HostOperation type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewHostOperationInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredHostOperationInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredHostOperationInformer constructs a new informer for HostOperation type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredHostOperationInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.BmcV1beta1().HostOperations().List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.BmcV1beta1().HostOperations().Watch(context.TODO(), options)
			},
		},
		&apisbmcspidernetiov1beta1.HostOperation{},
		resyncPeriod,
		indexers,
	)
}

func (f *hostOperationInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredHostOperationInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *hostOperationInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&apisbmcspidernetiov1beta1.HostOperation{}, f.defaultInformer)
}

func (f *hostOperationInformer) Lister() bmcspidernetiov1beta1.HostOperationLister {
	return bmcspidernetiov1beta1.NewHostOperationLister(f.Informer().GetIndexer())
}