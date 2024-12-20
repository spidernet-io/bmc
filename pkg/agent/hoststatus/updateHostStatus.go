package hoststatus

import (
	bmcv1beta1 "github.com/spidernet-io/bmc/pkg/k8s/apis/bmc.spidernet.io/v1beta1"
	"github.com/spidernet-io/bmc/pkg/log"
)

// handleHostStatusAdd handles the addition of a HostStatus resource
func (c *hostStatusController) handleHostStatusAdd(obj interface{}) {
	hostStatus := obj.(*bmcv1beta1.HostStatus)
	log.Logger.Debugf("HostStatus added: %s (Type: %s, IP: %s)",
		hostStatus.Name, hostStatus.Status.Basic.Type, hostStatus.Status.Basic.IpAddr)
}

// handleHostStatusUpdate handles updates to a HostStatus resource
func (c *hostStatusController) handleHostStatusUpdate(old, new interface{}) {
	hostStatus := new.(*bmcv1beta1.HostStatus)
	log.Logger.Debugf("HostStatus updated: %s (Type: %s, IP: %s, Health: %v)",
		hostStatus.Name, hostStatus.Status.Basic.Type, hostStatus.Status.Basic.IpAddr, hostStatus.Status.HealthReady)
}

// handleHostStatusDelete handles the deletion of a HostStatus resource
func (c *hostStatusController) handleHostStatusDelete(obj interface{}) {
	hostStatus := obj.(*bmcv1beta1.HostStatus)
	log.Logger.Debugf("HostStatus deleted: %s", hostStatus.Name)
}
