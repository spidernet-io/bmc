package hoststatus

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spidernet-io/bmc/pkg/agent/hoststatus/data"
	bmcv1beta1 "github.com/spidernet-io/bmc/pkg/k8s/apis/bmc.spidernet.io/v1beta1"
	"github.com/spidernet-io/bmc/pkg/log"
)

// getSecretData 从 Secret 中获取用户名和密码
func (c *hostStatusController) getSecretData(secretName, secretNamespace string) (string, string, error) {
	// 检查是否与 AgentObjSpec.Endpoint 中的配置相同
	if secretName == c.config.AgentObjSpec.Endpoint.SecretName &&
		secretNamespace == c.config.AgentObjSpec.Endpoint.SecretNamespace {
		// 如果相同，直接返回配置中的认证信息
		return c.config.Username, c.config.Password, nil
	}

	// 如果不同，从 Secret 中获取认证信息
	secret, err := c.kubeClient.CoreV1().Secrets(secretNamespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		return "", "", fmt.Errorf("failed to get secret %s/%s: %v", secretNamespace, secretName, err)
	}

	username := string(secret.Data["username"])
	password := string(secret.Data["password"])
	return username, password, nil
}

// processHostStatus 处理 HostStatus 对象
func (c *hostStatusController) processHostStatus(hostStatus *bmcv1beta1.HostStatus) error {
	username, password, err := c.getSecretData(
		hostStatus.Status.Basic.SecretName,
		hostStatus.Status.Basic.SecretNamespace,
	)
	if err != nil {
		return fmt.Errorf("failed to get secret data: %v", err)
	}

	data.HostCacheDatabase.Add(hostStatus.Name, data.HostConnectCon{
		Info:     &hostStatus.Status.Basic,
		Username: username,
		Password: password,
	})

	return nil
}

// handleHostStatusAdd handles the addition of a HostStatus resource
func (c *hostStatusController) handleHostStatusAdd(obj interface{}) {
	hostStatus := obj.(*bmcv1beta1.HostStatus)
	log.Logger.Debugf("HostStatus added: %s (Type: %s, IP: %s)",
		hostStatus.Name, hostStatus.Status.Basic.Type, hostStatus.Status.Basic.IpAddr)

	// 添加到工作队列进行处理
	c.workqueue.Add(hostStatus)
}

// handleHostStatusUpdate handles updates to a HostStatus resource
func (c *hostStatusController) handleHostStatusUpdate(old, new interface{}) {
	hostStatus := new.(*bmcv1beta1.HostStatus)
	log.Logger.Debugf("HostStatus updated: %s (Type: %s, IP: %s, Health: %v)",
		hostStatus.Name, hostStatus.Status.Basic.Type, hostStatus.Status.Basic.IpAddr, hostStatus.Status.HealthReady)

	// 添加到工作队列进行处理
	c.workqueue.Add(hostStatus)
}

// handleHostStatusDelete handles the deletion of a HostStatus resource
func (c *hostStatusController) handleHostStatusDelete(obj interface{}) {
	hostStatus := obj.(*bmcv1beta1.HostStatus)
	log.Logger.Debugf("HostStatus deleted: %s", hostStatus.Name)

	data.HostCacheDatabase.Delete(hostStatus.Name)
}
