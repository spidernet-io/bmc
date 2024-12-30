package hoststatus

import (
	"context"
	"fmt"
	"strings"

	bmcv1beta1 "github.com/spidernet-io/bmc/pkg/k8s/apis/bmc.spidernet.io/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

func formatHostStatusName(agentName, ip string) string {
	return fmt.Sprintf("%s-%s", agentName, strings.ReplaceAll(ip, ".", "-"))
}

// 比较两个Status的内容是否相同，忽略指针等问题
func compareHostStatus(a, b bmcv1beta1.HostStatusStatus) bool {
	if a.Healthy != b.Healthy {
		return false
	}
	if a.ClusterAgent != b.ClusterAgent {
		return false
	}
	if a.LastUpdateTime != b.LastUpdateTime {
		return false
	}

	// 比较Basic字段
	if a.Basic.Type != b.Basic.Type ||
		a.Basic.IpAddr != b.Basic.IpAddr ||
		a.Basic.SecretName != b.Basic.SecretName ||
		a.Basic.SecretNamespace != b.Basic.SecretNamespace ||
		a.Basic.Https != b.Basic.Https ||
		a.Basic.Port != b.Basic.Port ||
		a.Basic.Mac != b.Basic.Mac ||
		a.Basic.ActiveDhcpClient != b.Basic.ActiveDhcpClient {
		return false
	}

	// 比较Info map中的内容
	if len(a.Info) != len(b.Info) {
		return false
	}
	for k, v1 := range a.Info {
		if v2, ok := b.Info[k]; !ok || v1 != v2 {
			return false
		}
	}
	return true
}
