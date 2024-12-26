package hoststatus

import (
	"context"
	"fmt"
	"strings"

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
