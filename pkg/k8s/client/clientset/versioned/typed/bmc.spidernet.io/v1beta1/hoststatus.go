// Copyright 2024 Authors of elf-io
// SPDX-License-Identifier: Apache-2.0

// Code generated by client-gen. DO NOT EDIT.

package v1beta1

import (
	context "context"

	bmcspidernetiov1beta1 "github.com/spidernet-io/bmc/pkg/k8s/apis/bmc.spidernet.io/v1beta1"
	scheme "github.com/spidernet-io/bmc/pkg/k8s/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	gentype "k8s.io/client-go/gentype"
)

// HostStatusesGetter has a method to return a HostStatusInterface.
// A group's client should implement this interface.
type HostStatusesGetter interface {
	HostStatuses() HostStatusInterface
}

// HostStatusInterface has methods to work with HostStatus resources.
type HostStatusInterface interface {
	Create(ctx context.Context, hostStatus *bmcspidernetiov1beta1.HostStatus, opts v1.CreateOptions) (*bmcspidernetiov1beta1.HostStatus, error)
	Update(ctx context.Context, hostStatus *bmcspidernetiov1beta1.HostStatus, opts v1.UpdateOptions) (*bmcspidernetiov1beta1.HostStatus, error)
	// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
	UpdateStatus(ctx context.Context, hostStatus *bmcspidernetiov1beta1.HostStatus, opts v1.UpdateOptions) (*bmcspidernetiov1beta1.HostStatus, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*bmcspidernetiov1beta1.HostStatus, error)
	List(ctx context.Context, opts v1.ListOptions) (*bmcspidernetiov1beta1.HostStatusList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *bmcspidernetiov1beta1.HostStatus, err error)
	HostStatusExpansion
}

// hostStatuses implements HostStatusInterface
type hostStatuses struct {
	*gentype.ClientWithList[*bmcspidernetiov1beta1.HostStatus, *bmcspidernetiov1beta1.HostStatusList]
}

// newHostStatuses returns a HostStatuses
func newHostStatuses(c *BmcV1beta1Client) *hostStatuses {
	return &hostStatuses{
		gentype.NewClientWithList[*bmcspidernetiov1beta1.HostStatus, *bmcspidernetiov1beta1.HostStatusList](
			"hoststatuses",
			c.RESTClient(),
			scheme.ParameterCodec,
			"",
			func() *bmcspidernetiov1beta1.HostStatus { return &bmcspidernetiov1beta1.HostStatus{} },
			func() *bmcspidernetiov1beta1.HostStatusList { return &bmcspidernetiov1beta1.HostStatusList{} },
		),
	}
}