package hoststatus

import (
	"context"
	"time"

	dhcptypes "github.com/spidernet-io/bmc/pkg/dhcpserver/types"
	bmcv1beta1 "github.com/spidernet-io/bmc/pkg/k8s/apis/bmc.spidernet.io/v1beta1"
	"github.com/spidernet-io/bmc/pkg/log"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// retryDelay is the delay before retrying a failed operation
	retryDelay = time.Second
)

func (c *hostStatusController) GetDHCPEventChan() (chan<- dhcptypes.ClientInfo, chan<- dhcptypes.ClientInfo) {
	return c.addChan, c.deleteChan
}

func (c *hostStatusController) processDHCPEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-c.addChan:
			if err := c.handleDHCPAdd(event); err != nil {
				if shouldRetry(err) {
					log.Logger.Debugf("Retrying DHCP add event for IP %s after %v due to: %v",
						event.IP, retryDelay, err)
					go func(e dhcptypes.ClientInfo) {
						time.Sleep(retryDelay)
						c.addChan <- e
					}(event)
				}
			}
		case event := <-c.deleteChan:
			if err := c.handleDHCPDelete(event); err != nil {
				if shouldRetry(err) {
					log.Logger.Debugf("Retrying DHCP delete event for IP %s after %v due to: %v",
						event.IP, retryDelay, err)
					go func(e dhcptypes.ClientInfo) {
						time.Sleep(retryDelay)
						c.deleteChan <- e
					}(event)
				}
			}
		}
	}
}

func (c *hostStatusController) handleDHCPAdd(client dhcptypes.ClientInfo) error {

	name := formatHostStatusName(c.config.ClusterAgentName, client.IP)
	log.Logger.Debugf("Processing DHCP add event - IP: %s, MAC: %s, Active: %v, Lease: %s -> %s",
		client.IP, client.MAC, client.Active, client.StartTime, client.EndTime)

	if c.config.AgentObjSpec.Feature.DhcpServerConfig != nil && c.config.AgentObjSpec.Feature.DhcpServerConfig.EnableDhcpDiscovery == false {
		log.Logger.Warnf("DhcpDiscovery is disabled, so ignore DHCP add event - IP: %s, MAC: %s, Active: %v, Lease: %s -> %s",
			client.IP, client.MAC, client.Active, client.StartTime, client.EndTime)
		return nil
	}

	// Try to get existing HostStatus
	existing := &bmcv1beta1.HostStatus{}
	existing, err := c.client.HostStatuses().Get(context.Background(), name, metav1.GetOptions{})

	if err == nil {
		// HostStatus exists, check if MAC changed
		if existing.Status.Basic.Mac == client.MAC {
			log.Logger.Debugf("HostStatus %s exists with same MAC %s, no update needed", name, client.MAC)
			return nil
		}
		// MAC changed, update the object
		log.Logger.Infof("Updating HostStatus %s: MAC changed from %s to %s",
			name, existing.Status.Basic.Mac, client.MAC)

		// Create a copy of the existing object to avoid modifying the cache
		updated := existing.DeepCopy()
		updated.Status.LastUpdateTime = time.Now().UTC().Format(time.RFC3339)
		updated.Status.Basic.Mac = client.MAC

		if _, err := c.client.HostStatuses().UpdateStatus(context.Background(), updated, metav1.UpdateOptions{}); err != nil {
			if errors.IsConflict(err) {
				log.Logger.Debugf("Conflict updating HostStatus %s, will retry", name)
				return err
			}
			log.Logger.Errorf("Failed to update HostStatus %s: %v", name, err)
			return err
		}
		log.Logger.Infof("Successfully updated HostStatus %s", name)
		log.Logger.Debugf("Updated DHCP client details - IP: %s, MAC: %s, Lease: %s -> %s",
			client.IP, client.MAC, client.StartTime, client.EndTime)
		return nil
	}

	// HostStatus doesn't exist, create new one
	// IMPORTANT: When creating a new HostStatus, we must follow a two-step process:
	// 1. First create the resource with only metadata (no status). This is because
	//    the Kubernetes API server does not allow setting status during creation.
	// 2. Then update the status separately using UpdateStatus. If we try to set
	//    status during creation, the status will be silently ignored, leading to
	//    a HostStatus without any status information until the next reconciliation.
	hostStatus := &bmcv1beta1.HostStatus{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	log.Logger.Debugf("Creating new HostStatus %s", name)

	created, err := c.client.HostStatuses().Create(context.Background(), hostStatus, metav1.CreateOptions{})
	if err != nil {
		log.Logger.Errorf("Failed to create HostStatus %s: %v", name, err)
		return err
	}

	// Now update the status
	// This is the second step of the two-step process. After creating the resource,
	// we update its status. This ensures that the status is properly set and visible
	// immediately, without requiring a controller restart or reconciliation.
	created.Status = bmcv1beta1.HostStatusStatus{
		Healthy:        false,
		ClusterAgent:   c.config.ClusterAgentName,
		LastUpdateTime: time.Now().UTC().Format(time.RFC3339),
		Basic: bmcv1beta1.BasicInfo{
			Type:   bmcv1beta1.HostTypeDHCP,
			IpAddr: client.IP,
			Mac:    client.MAC,
			Port:   c.config.AgentObjSpec.Endpoint.Port,
			Https:  c.config.AgentObjSpec.Endpoint.HTTPS,
		},
		Info: map[string]string{},
	}

	if _, err := c.client.HostStatuses().UpdateStatus(context.Background(), created, metav1.UpdateOptions{}); err != nil {
		log.Logger.Errorf("Failed to update status of HostStatus %s: %v", name, err)
		return err
	}

	log.Logger.Infof("Successfully created HostStatus %s", name)
	log.Logger.Debugf("DHCP client details - IP: %s, MAC: %s, Lease: %s -> %s",
		client.IP, client.MAC, client.StartTime, client.EndTime)
	return nil
}

func (c *hostStatusController) handleDHCPDelete(client dhcptypes.ClientInfo) error {
	name := formatHostStatusName(c.config.ClusterAgentName, client.IP)
	log.Logger.Debugf("Processing DHCP delete event - IP: %s, MAC: %s", client.IP, client.MAC)

	if c.config.AgentObjSpec.Feature.DhcpServerConfig != nil && c.config.AgentObjSpec.Feature.DhcpServerConfig.EnableDhcpDiscovery == false {
		log.Logger.Warnf("DhcpDiscovery is disabled, so ignore DHCP delete event - IP: %s, MAC: %s, Active: %v, Lease: %s -> %s",
			client.IP, client.MAC, client.Active, client.StartTime, client.EndTime)
		return nil
	}

	if err := c.client.HostStatuses().Delete(context.Background(), name, metav1.DeleteOptions{}); err != nil {
		if errors.IsNotFound(err) {
			log.Logger.Debugf("HostStatus %s not found, already deleted", name)
			return nil
		}
		log.Logger.Errorf("Failed to delete HostStatus %s: %v", name, err)
		return err
	}
	log.Logger.Infof("Successfully deleted HostStatus %s", name)
	return nil
}
