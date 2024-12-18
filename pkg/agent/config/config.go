package config

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"

	"github.com/spidernet-io/bmc/pkg/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

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

// EndpointConfig represents the endpoint configuration
type EndpointConfig struct {
	Port            int32
	HTTPS           bool
	SecretName      string
	SecretNamespace string
	TLSPath         string // Directory to store TLS certificates
}

// FeatureConfig represents the feature configuration
type FeatureConfig struct {
	EnableDhcpServer    bool
	EnableDhcpDiscovery bool
	DhcpServerInterface string
	RedfishMetrics      bool
	EnableGuiProxy      bool
}

// AgentConfig represents the agent configuration
type AgentConfig struct {
	ClusterAgentName string
	Endpoint         *EndpointConfig
	Feature          *FeatureConfig
}

// ValidateEndpointConfig validates the endpoint configuration
func (c *AgentConfig) ValidateEndpointConfig(clientset *kubernetes.Clientset) error {
	if c.Endpoint == nil {
		return fmt.Errorf("endpoint configuration is required")
	}

	if c.Endpoint.HTTPS {
		if c.Endpoint.SecretName == "" || c.Endpoint.SecretNamespace == "" {
			return fmt.Errorf("when HTTPS is enabled, both secretName and secretNamespace must be specified")
		}

		if c.Endpoint.TLSPath == "" {
			return fmt.Errorf("TLSPath must be specified when HTTPS is enabled")
		}

		// Create TLS directory if it doesn't exist
		if err := os.MkdirAll(c.Endpoint.TLSPath, 0700); err != nil {
			return fmt.Errorf("failed to create TLS directory: %v", err)
		}

		// Get the secret
		secret, err := clientset.CoreV1().Secrets(c.Endpoint.SecretNamespace).Get(context.TODO(), c.Endpoint.SecretName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get TLS secret: %v", err)
		}

		// Store TLS files
		if err := c.storeTLSFiles(secret); err != nil {
			return fmt.Errorf("failed to store TLS files: %v", err)
		}
	}

	// Check if port is valid
	if c.Endpoint.Port <= 0 || c.Endpoint.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", c.Endpoint.Port)
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

		filePath := filepath.Join(c.Endpoint.TLSPath, fileName)
		if err := ioutil.WriteFile(filePath, data, 0600); err != nil {
			return fmt.Errorf("failed to write %s: %v", fileName, err)
		}
		log.Logger.Infof("Stored %s to %s", fileName, filePath)
	}

	return nil
}

// ValidateFeatureConfig validates the feature configuration
func (c *AgentConfig) ValidateFeatureConfig() error {
	if c.Feature == nil {
		return fmt.Errorf("feature configuration is required")
	}

	if c.Feature.EnableDhcpServer {
		if c.Feature.DhcpServerInterface == "" {
			return fmt.Errorf("dhcp server interface must be specified when dhcp server is enabled")
		}

		// Check if interface exists
		_, err := net.InterfaceByName(c.Feature.DhcpServerInterface)
		if err != nil {
			return fmt.Errorf("dhcp server interface %s not found: %v", c.Feature.DhcpServerInterface, err)
		}
	}

	return nil
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

	// Create agent config from ClusterAgent spec
	agentConfig := &AgentConfig{
		ClusterAgentName: agentName,
		Endpoint: &EndpointConfig{
			TLSPath:         DefaultTLSPath,
			Port:            clusterAgent.Spec.Endpoint.Port,
			HTTPS:           clusterAgent.Spec.Endpoint.HTTPS,
			SecretName:      clusterAgent.Spec.Endpoint.SecretName,
			SecretNamespace: clusterAgent.Spec.Endpoint.SecretNamespace,
		},
		Feature: &FeatureConfig{
			EnableDhcpServer:    clusterAgent.Spec.Feature.EnableDhcpServer,
			EnableDhcpDiscovery: clusterAgent.Spec.Feature.EnableDhcpDiscovery,
			DhcpServerInterface: clusterAgent.Spec.Feature.DhcpServerInterface,
			RedfishMetrics:      clusterAgent.Spec.Feature.RedfishMetrics,
			EnableGuiProxy:      clusterAgent.Spec.Feature.EnableGuiProxy,
		},
	}

	return agentConfig, nil
}
