package hoststatus

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spidernet-io/bmc/pkg/agent/hoststatus/data"
	bmcv1beta1 "github.com/spidernet-io/bmc/pkg/k8s/apis/bmc.spidernet.io/v1beta1"
	"github.com/spidernet-io/bmc/pkg/log"
)

// getSecretData 从 Secret 中获取用户名和密码
func (c *hostStatusController) getSecretData(secretName, secretNamespace string) (string, string, error) {
	log.Logger.Debugf("Attempting to get secret data for %s/%s", secretNamespace, secretName)

	// 检查是否与 AgentObjSpec.Endpoint 中的配置相同
	if secretName == c.config.AgentObjSpec.Endpoint.SecretName &&
		secretNamespace == c.config.AgentObjSpec.Endpoint.SecretNamespace {
		// 如果相同，直接返回配置中的认证信息
		log.Logger.Debugf("Using credentials from agent config for %s/%s", secretNamespace, secretName)
		return c.config.Username, c.config.Password, nil
	}

	log.Logger.Debugf("Fetching secret from Kubernetes API for %s/%s", secretNamespace, secretName)
	// 如果不同，从 Secret 中获取认证信息
	secret, err := c.kubeClient.CoreV1().Secrets(secretNamespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		log.Logger.Errorf("Failed to get secret %s/%s: %v", secretNamespace, secretName, err)
		return "", "", err
	}

	username := string(secret.Data["username"])
	password := string(secret.Data["password"])
	log.Logger.Debugf("Successfully retrieved secret data for %s/%s", secretNamespace, secretName)
	return username, password, nil
}

// processHostStatus 处理 HostStatus 对象
func (c *hostStatusController) processHostStatus(hostStatus *bmcv1beta1.HostStatus) error {

	if len(hostStatus.Status.Basic.IpAddr) == 0 {
		// the hostStatus is created firstly and then be updated with its status
		log.Logger.Debugf("ingore hostStatus %s just created", hostStatus.Name)
		return nil
	}

	log.Logger.Debugf("Processing HostStatus: %s (Type: %s, IP: %s, Health: %v)",
		hostStatus.Name,
		hostStatus.Status.Basic.Type,
		hostStatus.Status.Basic.IpAddr,
		hostStatus.Status.Healthy)

	username, password, err := c.getSecretData(
		hostStatus.Status.Basic.SecretName,
		hostStatus.Status.Basic.SecretNamespace,
	)
	if err != nil {
		log.Logger.Errorf("Failed to get secret data for HostStatus %s: %v", hostStatus.Name, err)
		return err
	}

	log.Logger.Debugf("Adding/Updating HostStatus %s in cache with username: %s",
		hostStatus.Name, username)

	data.HostCacheDatabase.Add(hostStatus.Name, data.HostConnectCon{
		Info:     &hostStatus.Status.Basic,
		Username: username,
		Password: password,
	})
	// update the crd
	c.UpdateHostStatusWrapper(hostStatus.Name)

	log.Logger.Debugf("Successfully processed HostStatus %s", hostStatus.Name)
	return nil
}

// handleHostStatusAdd handles the addition of a HostStatus resource
func (c *hostStatusController) handleHostStatusAdd(obj interface{}) {
	hostStatus := obj.(*bmcv1beta1.HostStatus)
	log.Logger.Debugf("HostStatus added: %s (Type: %s, IP: %s)",
		hostStatus.Name, hostStatus.Status.Basic.Type, hostStatus.Status.Basic.IpAddr)

	// 添加到工作队列进行处理
	log.Logger.Debugf("Adding HostStatus %s to workqueue for processing", hostStatus.Name)
	c.workqueue.Add(hostStatus)
}

// handleHostStatusUpdate handles updates to a HostStatus resource
func (c *hostStatusController) handleHostStatusUpdate(old, new interface{}) {
	hostStatus := new.(*bmcv1beta1.HostStatus)
	log.Logger.Debugf("HostStatus updated: %s (Type: %s, IP: %s, Health: %v)",
		hostStatus.Name, hostStatus.Status.Basic.Type, hostStatus.Status.Basic.IpAddr, hostStatus.Status.Healthy)

	// 添加到工作队列进行处理
	c.workqueue.Add(hostStatus)
}

// handleHostStatusDelete handles the deletion of a HostStatus resource
func (c *hostStatusController) handleHostStatusDelete(obj interface{}) {
	hostStatus := obj.(*bmcv1beta1.HostStatus)
	log.Logger.Debugf("HostStatus deleted: %s", hostStatus.Name)

	data.HostCacheDatabase.Delete(hostStatus.Name)
}
