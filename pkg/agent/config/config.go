package config

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/spidernet-io/bmc/pkg/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	bmcv1beta1 "github.com/spidernet-io/bmc/pkg/apis/bmc/v1beta1"
)

// TLS file names
const (
	TLSCertFile = "tls.crt"
	TLSKeyFile  = "tls.key"
	CAFile      = "ca.crt"
)

// Default paths
const (
	DefaultTLSPath = "/var/run/bmc/tls" // Default path for storing TLS certificates
)

// AgentConfig represents the agent configuration
type AgentConfig struct {
	ClusterAgentName string
	agentObjSpec     bmcv1beta1.ClusterAgentSpec
	TLSPath          string
}

// ValidateEndpointConfig validates the endpoint configuration
func (c *AgentConfig) ValidateEndpointConfig(clientset *kubernetes.Clientset) error {
	if c.agentObjSpec.Endpoint == nil {
		return fmt.Errorf("endpoint configuration is required")
	}

	if c.agentObjSpec.Endpoint.HTTPS {
		if c.agentObjSpec.Endpoint.SecretName == "" || c.agentObjSpec.Endpoint.SecretNamespace == "" {
			return fmt.Errorf("when HTTPS is enabled, both secretName and secretNamespace must be specified")
		}

		// Create TLS directory if it doesn't exist
		if err := os.MkdirAll(DefaultTLSPath, 0700); err != nil {
			return fmt.Errorf("failed to create TLS directory: %v", err)
		}

		// Get the secret
		secret, err := clientset.CoreV1().Secrets(c.agentObjSpec.Endpoint.SecretNamespace).Get(
			context.TODO(),
			c.agentObjSpec.Endpoint.SecretName,
			metav1.GetOptions{},
		)
		if err != nil {
			return fmt.Errorf("failed to get TLS secret: %v", err)
		}

		// Store TLS files
		if err := c.storeTLSFiles(secret); err != nil {
			return fmt.Errorf("failed to store TLS files: %v", err)
		}
		c.TLSPath = DefaultTLSPath
	} else {
		c.TLSPath = ""
	}

	// Check if port is valid
	if c.agentObjSpec.Endpoint.Port <= 0 || c.agentObjSpec.Endpoint.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", c.agentObjSpec.Endpoint.Port)
	}

	return nil
}

// storeTLSFiles stores TLS files from the secret to the local directory
func (c *AgentConfig) storeTLSFiles(secret *corev1.Secret) error {
	// Map of secret keys to file names
	files := map[string]string{
		"tls.crt": TLSCertFile,
		"tls.key": TLSKeyFile,
		"ca.crt":  CAFile,
	}

	for secretKey, fileName := range files {
		data, ok := secret.Data[secretKey]
		if !ok {
			if secretKey == "ca.crt" {
				// CA is optional
				continue
			}
			return fmt.Errorf("required key %s not found in secret", secretKey)
		}

		filePath := filepath.Join(DefaultTLSPath, fileName)
		if err := ioutil.WriteFile(filePath, data, 0600); err != nil {
			return fmt.Errorf("failed to write %s: %v", fileName, err)
		}
		log.Logger.Infof("Stored %s to %s", fileName, filePath)
	}

	return nil
}

// ValidateFeatureConfig validates the feature configuration
func (c *AgentConfig) ValidateFeatureConfig() error {
	if c.agentObjSpec.Feature == nil {
		return fmt.Errorf("feature configuration is required")
	}

	if c.agentObjSpec.Feature.EnableDhcpServer {
		if c.agentObjSpec.Feature.DhcpServerConfig == nil {
			return fmt.Errorf("dhcp server config must be specified when dhcp server is enabled")
		}

		config := c.agentObjSpec.Feature.DhcpServerConfig

		if config.DhcpServerInterface == "" {
			return fmt.Errorf("dhcp server interface must be specified when dhcp server is enabled")
		}

		// Check if interface exists
		_, err := net.InterfaceByName(config.DhcpServerInterface)
		if err != nil {
			return fmt.Errorf("dhcp server interface %s not found: %v", config.DhcpServerInterface, err)
		}
	}

	return nil
}

// GetDetailString returns a detailed string representation of the AgentConfig
func (c *AgentConfig) GetDetailString() string {
	var details strings.Builder

	details.WriteString(fmt.Sprintf("ClusterAgentName: %s\n", c.ClusterAgentName))
	details.WriteString(fmt.Sprintf("TLSPath: %s\n", c.TLSPath))
	details.WriteString("AgentSpec:\n")

	// AgentYaml details
	details.WriteString("  AgentYaml:\n")
	details.WriteString(fmt.Sprintf("    UnderlayInterface: %s\n", c.agentObjSpec.AgentYaml.UnderlayInterface))
	details.WriteString(fmt.Sprintf("    Image: %s\n", c.agentObjSpec.AgentYaml.Image))
	if c.agentObjSpec.AgentYaml.Replicas != nil {
		details.WriteString(fmt.Sprintf("    Replicas: %d\n", *c.agentObjSpec.AgentYaml.Replicas))
	}
	if c.agentObjSpec.AgentYaml.NodeName != "" {
		details.WriteString(fmt.Sprintf("    NodeName: %s\n", c.agentObjSpec.AgentYaml.NodeName))
	}

	// Endpoint details
	if c.agentObjSpec.Endpoint != nil {
		details.WriteString("  Endpoint:\n")
		details.WriteString(fmt.Sprintf("    Port: %d\n", c.agentObjSpec.Endpoint.Port))
		details.WriteString(fmt.Sprintf("    HTTPS: %v\n", c.agentObjSpec.Endpoint.HTTPS))
		if c.agentObjSpec.Endpoint.SecretName != "" {
			details.WriteString(fmt.Sprintf("    SecretName: %s\n", c.agentObjSpec.Endpoint.SecretName))
		}
		if c.agentObjSpec.Endpoint.SecretNamespace != "" {
			details.WriteString(fmt.Sprintf("    SecretNamespace: %s\n", c.agentObjSpec.Endpoint.SecretNamespace))
		}
	}

	// Feature details
	if c.agentObjSpec.Feature != nil {
		details.WriteString("  Feature:\n")
		details.WriteString(fmt.Sprintf("    EnableDhcpServer: %v\n", c.agentObjSpec.Feature.EnableDhcpServer))

		// DHCP Server Config details
		if c.agentObjSpec.Feature.DhcpServerConfig != nil {
			details.WriteString("    DhcpServerConfig:\n")
			config := c.agentObjSpec.Feature.DhcpServerConfig
			details.WriteString(fmt.Sprintf("      EnableDhcpDiscovery: %v\n", config.EnableDhcpDiscovery))
			details.WriteString(fmt.Sprintf("      DhcpServerInterface: %s\n", config.DhcpServerInterface))
			details.WriteString(fmt.Sprintf("      Subnet: %s\n", config.Subnet))
			details.WriteString(fmt.Sprintf("      IpRange: %s\n", config.IpRange))
			details.WriteString(fmt.Sprintf("      Gateway: %s\n", config.Gateway))
			if config.SelfIp != "" {
				details.WriteString(fmt.Sprintf("      SelfIp: %s\n", config.SelfIp))
			}
		}

		details.WriteString(fmt.Sprintf("    RedfishMetrics: %v\n", c.agentObjSpec.Feature.RedfishMetrics))
		details.WriteString(fmt.Sprintf("    EnableGuiProxy: %v\n", c.agentObjSpec.Feature.EnableGuiProxy))
	}

	return details.String()
}

// LoadAgentConfig loads the agent configuration from environment and ClusterAgent instance
func LoadAgentConfig(k8sClient *kubernetes.Clientset) (*AgentConfig, error) {
	// Get agent name from environment
	agentName := os.Getenv("CLUSTERAGENT_NAME")
	if agentName == "" {
		return nil, fmt.Errorf("CLUSTERAGENT_NAME environment variable not set")
	}

	// Create bmc client config
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get in-cluster config: %v", err)
	}

	// Add bmc scheme and set GroupVersion
	scheme := runtime.NewScheme()
	bmcv1beta1.AddToScheme(scheme)
	restConfig.GroupVersion = &bmcv1beta1.GroupVersion
	restConfig.APIPath = "/apis"
	restConfig.NegotiatedSerializer = serializer.NewCodecFactory(scheme)

	// Create REST client for ClusterAgent
	restClient, err := rest.RESTClientFor(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create REST client: %v", err)
	}

	// Get ClusterAgent instance
	clusterAgent := &bmcv1beta1.ClusterAgent{}
	err = restClient.Get().
		Resource("clusteragents").
		Name(agentName).
		Do(context.TODO()).
		Into(clusterAgent)
	if err != nil {
		return nil, fmt.Errorf("failed to get ClusterAgent instance: %v", err)
	}

	// Create agent config
	agentConfig := &AgentConfig{
		ClusterAgentName: agentName,
		agentObjSpec:     clusterAgent.Spec,
	}

	// Validate endpoint configuration
	if err := agentConfig.ValidateEndpointConfig(k8sClient); err != nil {
		return nil, fmt.Errorf("invalid endpoint configuration: %v", err)
	}

	// Validate feature configuration
	if err := agentConfig.ValidateFeatureConfig(); err != nil {
		return nil, fmt.Errorf("invalid feature configuration: %v", err)
	}

	log.Logger.Debugf("Agent configuration loaded successfully: %+v", agentConfig)
	return agentConfig, nil
}
