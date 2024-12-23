package hoststatus

import (
	"context"
	"time"

	bmcv1beta1 "github.com/spidernet-io/bmc/pkg/k8s/apis/bmc.spidernet.io/v1beta1"
	"github.com/spidernet-io/bmc/pkg/log"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *hostStatusController) handleHostEndpointDelete(obj interface{}) {
	hostEndpoint := obj.(*bmcv1beta1.HostEndpoint)
	log.Logger.Debugf("HostEndpoint %s deleted, associated HostStatus will be garbage collected", hostEndpoint.Name)
}

func (c *hostStatusController) handleHostEndpointAdd(obj interface{}) {
	hostEndpoint := obj.(*bmcv1beta1.HostEndpoint)
	log.Logger.Debugf("Processing HostEndpoint: %s", hostEndpoint.Name)

	log.Logger.Debugf("HostEndpoint %s details - ClusterAgent: %s, Current ClusterAgent: %s",
		hostEndpoint.Name, hostEndpoint.Spec.ClusterAgent, c.config.ClusterAgentName)

	if hostEndpoint.Spec.ClusterAgent != c.config.ClusterAgentName {
		log.Logger.Debugf("Skipping HostEndpoint %s: belongs to different cluster agent (%s)",
			hostEndpoint.Name, hostEndpoint.Spec.ClusterAgent)
		return
	}

	// Add to workqueue for retry handling
	log.Logger.Debugf("Adding HostEndpoint %s to workqueue", hostEndpoint.Name)
	c.workqueue.Add(hostEndpoint)
}

//--------------------------------------------

func (c *hostStatusController) processHostEndpoint(hostEndpoint *bmcv1beta1.HostEndpoint) error {
	name := formatHostStatusName(c.config.ClusterAgentName, hostEndpoint.Spec.IPAddr)
	log.Logger.Debugf("Processing HostEndpoint %s (IP: %s)", hostEndpoint.Name, hostEndpoint.Spec.IPAddr)

	// Try to get existing HostStatus
	existing, err := c.client.HostStatuses().Get(context.Background(), name, metav1.GetOptions{})
	if err == nil {
		// HostStatus exists, check if spec changed
		if specEqual(existing.Status.Basic, hostEndpoint.Spec) {
			log.Logger.Debugf("HostStatus %s exists with same spec, no update needed", name)
			return nil
		}

		// Spec changed, update the object
		log.Logger.Infof("Updating HostStatus %s due to spec change", name)

		// Create a copy of the existing object to avoid modifying the cache
		updated := existing.DeepCopy()
		updated.Status.LastUpdateTime = time.Now().UTC().Format(time.RFC3339)
		updated.Status.Basic = bmcv1beta1.BasicInfo{
			Type:            bmcv1beta1.HostTypeEndpoint,
			IpAddr:          hostEndpoint.Spec.IPAddr,
			SecretName:      hostEndpoint.Spec.SecretName,
			SecretNamespace: hostEndpoint.Spec.SecretNamespace,
			Https:           *hostEndpoint.Spec.HTTPS,
			Port:            *hostEndpoint.Spec.Port,
		}

		if _, err := c.client.HostStatuses().UpdateStatus(context.Background(), updated, metav1.UpdateOptions{}); err != nil {
			if errors.IsConflict(err) {
				log.Logger.Debugf("Conflict updating HostStatus %s, will retry", name)
				return err
			}
			log.Logger.Errorf("Failed to update HostStatus %s: %v", name, err)
			return err
		}
		log.Logger.Infof("Successfully updated HostStatus %s", name)
		log.Logger.Debugf("Updated HostStatus details - IP: %s, Secret: %s/%s, Port: %d",
			updated.Status.Basic.IpAddr,
			updated.Status.Basic.SecretNamespace,
			updated.Status.Basic.SecretName,
			updated.Status.Basic.Port)
		return nil
	}

	if !errors.IsNotFound(err) {
		log.Logger.Errorf("Failed to get HostStatus %s: %v", name, err)
		return err
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
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         bmcv1beta1.APIVersion,
					Kind:               bmcv1beta1.KindHostEndpoint,
					Name:               hostEndpoint.Name,
					UID:                hostEndpoint.UID,
					Controller:         &[]bool{true}[0],
					BlockOwnerDeletion: &[]bool{true}[0],
				},
			},
		},
	}
	log.Logger.Debugf("Creating new HostStatus %s", name)

	// must create it and then update the status !!!!
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
		HealthReady:    false,
		ClusterAgent:   hostEndpoint.Spec.ClusterAgent,
		LastUpdateTime: time.Now().UTC().Format(time.RFC3339),
		Basic: bmcv1beta1.BasicInfo{
			Type:            bmcv1beta1.HostTypeEndpoint,
			IpAddr:          hostEndpoint.Spec.IPAddr,
			SecretName:      hostEndpoint.Spec.SecretName,
			SecretNamespace: hostEndpoint.Spec.SecretNamespace,
			Https:           *hostEndpoint.Spec.HTTPS,
			Port:            *hostEndpoint.Spec.Port,
		},
		Info: bmcv1beta1.Info{},
	}

	if _, err := c.client.HostStatuses().UpdateStatus(context.Background(), created, metav1.UpdateOptions{}); err != nil {
		log.Logger.Errorf("Failed to update status of HostStatus %s: %v", name, err)
		return err
	}

	log.Logger.Infof("Successfully created HostStatus %s", name)
	log.Logger.Debugf("HostStatus details - IP: %s, Secret: %s/%s, Port: %d",
		created.Status.Basic.IpAddr,
		created.Status.Basic.SecretNamespace,
		created.Status.Basic.SecretName,
		created.Status.Basic.Port)
	return nil
}

// specEqual checks if the HostStatus basic info matches the HostEndpoint spec
func specEqual(basic bmcv1beta1.BasicInfo, spec bmcv1beta1.HostEndpointSpec) bool {
	return basic.IpAddr == spec.IPAddr &&
		basic.SecretName == spec.SecretName &&
		basic.SecretNamespace == spec.SecretNamespace &&
		basic.Https == *spec.HTTPS &&
		basic.Port == *spec.Port
}
