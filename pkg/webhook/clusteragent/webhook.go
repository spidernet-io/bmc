package clusteragent

import (
	"context"
	"fmt"
	"os"
	"regexp"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	bmcv1beta1 "github.com/spidernet-io/bmc/pkg/apis/bmc/v1beta1"
	"github.com/spidernet-io/bmc/pkg/log"
	"go.uber.org/zap"
)

// ClusterAgentWebhook validates ClusterAgent resources
type ClusterAgentWebhook struct {
	Client client.Client
}

func (w *ClusterAgentWebhook) SetupWebhookWithManager(mgr ctrl.Manager) error {
	w.Client = mgr.GetClient()
	return ctrl.NewWebhookManagedBy(mgr).
		For(&bmcv1beta1.ClusterAgent{}).
		WithValidator(w).
		WithDefaulter(w).
		Complete()
}

// Default implements webhook.Defaulter
func (w *ClusterAgentWebhook) Default(ctx context.Context, obj runtime.Object) error {
	clusterAgent, ok := obj.(*bmcv1beta1.ClusterAgent)
	if !ok {
		return fmt.Errorf("object is not a ClusterAgent")
	}

	logger := log.Logger.With(
		zap.String("webhook", "clusteragent"),
		zap.String("name", clusterAgent.Name),
		zap.String("ns", clusterAgent.Namespace),
		zap.String("action", "default"),
	)
	logger.Info("Setting default values for ClusterAgent")

	// Set default image from AGENT_IMAGE environment variable if not specified
	if clusterAgent.Spec.AgentYaml.Image == "" {
		agentImage := os.Getenv("AGENT_IMAGE")
		if agentImage == "" {
			logger.Error("AGENT_IMAGE environment variable not set")
			return fmt.Errorf("AGENT_IMAGE environment variable not set")
		}
		logger.Infof("Setting default image: %s", agentImage)
		clusterAgent.Spec.AgentYaml.Image = agentImage
	}

	// Set default replicas to 1 if not specified
	if clusterAgent.Spec.AgentYaml.Replicas == nil {
		defaultReplicas := int32(1)
		logger.Infof("Setting default replicas: %d", defaultReplicas)
		clusterAgent.Spec.AgentYaml.Replicas = &defaultReplicas
	}

	// Initialize endpoint if not specified
	if clusterAgent.Spec.Endpoint == nil {
		logger.Info("Initializing endpoint with default values")
		clusterAgent.Spec.Endpoint = &bmcv1beta1.EndpointConfig{
			Port:    443,
			HTTPS:   true,
		}
		logger.Infof("Set default endpoint values - Port: %d, HTTPS: %v", 
			clusterAgent.Spec.Endpoint.Port, 
			clusterAgent.Spec.Endpoint.HTTPS)
	} else {
		// Set default values for endpoint fields if not specified
		if clusterAgent.Spec.Endpoint.Port == 0 {
			logger.Info("Setting default endpoint port: 443")
			clusterAgent.Spec.Endpoint.Port = 443
		}
		if !clusterAgent.Spec.Endpoint.HTTPS {
			logger.Info("Setting default endpoint HTTPS: true")
			clusterAgent.Spec.Endpoint.HTTPS = true
		}
	}

	// Initialize feature if not specified
	if clusterAgent.Spec.Feature == nil {
		logger.Info("Initializing feature with default values")
		clusterAgent.Spec.Feature = &bmcv1beta1.FeatureConfig{
			EnableDhcpServer: true,
			DhcpServerConfig: &bmcv1beta1.DhcpServerConfig{
				EnableDhcpDiscovery: true,
				DhcpServerInterface: "net1",
			},
			RedfishMetrics: false,
			EnableGuiProxy: true,
		}
		logger.Infof("Set default feature values - EnableDhcpServer: %v, DhcpServerConfig: %+v, RedfishMetrics: %v, EnableGuiProxy: %v",
			clusterAgent.Spec.Feature.EnableDhcpServer,
			clusterAgent.Spec.Feature.DhcpServerConfig,
			clusterAgent.Spec.Feature.RedfishMetrics,
			clusterAgent.Spec.Feature.EnableGuiProxy)
	} else {
		// Set default values for feature fields if not specified
		if !clusterAgent.Spec.Feature.EnableDhcpServer {
			logger.Info("Setting default feature EnableDhcpServer: true")
			clusterAgent.Spec.Feature.EnableDhcpServer = true
		}

		// Initialize DhcpServerConfig if not specified
		if clusterAgent.Spec.Feature.EnableDhcpServer && clusterAgent.Spec.Feature.DhcpServerConfig == nil {
			logger.Info("Initializing DhcpServerConfig with default values")
			clusterAgent.Spec.Feature.DhcpServerConfig = &bmcv1beta1.DhcpServerConfig{
				EnableDhcpDiscovery: true,
				DhcpServerInterface: "net1",
			}
		}

		if !clusterAgent.Spec.Feature.EnableGuiProxy {
			logger.Info("Setting default feature EnableGuiProxy: true")
			clusterAgent.Spec.Feature.EnableGuiProxy = true
		}
	}

	logger.Info("Finished setting default values for ClusterAgent")
	return nil
}

// ValidateCreate implements webhook.Validator
func (w *ClusterAgentWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	clusterAgent, ok := obj.(*bmcv1beta1.ClusterAgent)
	if !ok {
		return nil, fmt.Errorf("object is not a ClusterAgent")
	}

	logger := log.Logger.With(
		zap.String("webhook", "clusteragent"),
		zap.String("name", clusterAgent.Name),
		zap.String("ns", clusterAgent.Namespace),
		zap.String("action", "validate-create"),
	)
	logger.Info("Validating ClusterAgent creation")

	if err := w.validateClusterAgent(ctx, clusterAgent); err != nil {
		logger.Errorf("Validation failed: %v", err)
		return nil, err
	}

	logger.Info("ClusterAgent validation successful")
	return nil, nil
}

// ValidateUpdate implements webhook.Validator
func (w *ClusterAgentWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	clusterAgent, ok := newObj.(*bmcv1beta1.ClusterAgent)
	if !ok {
		return nil, fmt.Errorf("object is not a ClusterAgent")
	}

	logger := log.Logger.With(
		zap.String("webhook", "clusteragent"),
		zap.String("name", clusterAgent.Name),
		zap.String("ns", clusterAgent.Namespace),
		zap.String("action", "validate-update"),
	)
	logger.Info("Validating ClusterAgent update")

	if err := w.validateClusterAgent(ctx, clusterAgent); err != nil {
		logger.Errorf("Validation failed: %v", err)
		return nil, err
	}

	logger.Info("ClusterAgent update validation successful")
	return nil, nil
}

// ValidateDelete implements webhook.Validator
func (w *ClusterAgentWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (w *ClusterAgentWebhook) validateClusterAgent(ctx context.Context, clusterAgent *bmcv1beta1.ClusterAgent) error {
	logger := log.Logger.With(
		zap.String("webhook", "clusteragent"),
		zap.String("name", clusterAgent.Name),
		zap.String("ns", clusterAgent.Namespace),
		zap.String("action", "validate"),
	)

	// Validate required fields
	if clusterAgent.Name == "" {
		logger.Error("name is required")
		return fmt.Errorf("name is required")
	}
	if clusterAgent.Spec.AgentYaml.UnderlayInterface == "" {
		logger.Error("underlayInterface is required")
		return fmt.Errorf("underlayInterface is required")
	}

	// Validate name format
	if err := validateClusterName(clusterAgent.Name); err != nil {
		logger.Error(err.Error())
		return err
	}

	// Validate DHCP server configuration
	if clusterAgent.Spec.Feature != nil && clusterAgent.Spec.Feature.EnableDhcpServer {
		if clusterAgent.Spec.Feature.DhcpServerConfig == nil {
			logger.Error("dhcpServerConfig is required when enableDhcpServer is true")
			return fmt.Errorf("dhcpServerConfig is required when enableDhcpServer is true")
		}

		config := clusterAgent.Spec.Feature.DhcpServerConfig

		// Validate required fields
		if config.DhcpServerInterface == "" {
			logger.Error("dhcpServerInterface is required in dhcpServerConfig")
			return fmt.Errorf("dhcpServerInterface is required in dhcpServerConfig")
		}

		if config.Subnet == "" {
			logger.Error("subnet is required in dhcpServerConfig")
			return fmt.Errorf("subnet is required in dhcpServerConfig")
		}

		if config.IpRange == "" {
			logger.Error("ipRange is required in dhcpServerConfig")
			return fmt.Errorf("ipRange is required in dhcpServerConfig")
		}

		if config.Gateway == "" {
			logger.Error("gateway is required in dhcpServerConfig")
			return fmt.Errorf("gateway is required in dhcpServerConfig")
		}

		// Validate IP formats
		subnetRegex := regexp.MustCompile(`^([0-9]{1,3}\.){3}[0-9]{1,3}/([0-9]|[1-2][0-9]|3[0-2])$`)
		if !subnetRegex.MatchString(config.Subnet) {
			logger.Error("invalid subnet format")
			return fmt.Errorf("invalid subnet format: %s", config.Subnet)
		}

		ipRangeRegex := regexp.MustCompile(`^([0-9]{1,3}\.){3}[0-9]{1,3}-([0-9]{1,3}\.){3}[0-9]{1,3}$`)
		if !ipRangeRegex.MatchString(config.IpRange) {
			logger.Error("invalid ipRange format")
			return fmt.Errorf("invalid ipRange format: %s", config.IpRange)
		}

		ipRegex := regexp.MustCompile(`^([0-9]{1,3}\.){3}[0-9]{1,3}$`)
		if !ipRegex.MatchString(config.Gateway) {
			logger.Error("invalid gateway format")
			return fmt.Errorf("invalid gateway format: %s", config.Gateway)
		}

		if config.SelfIp != "" && !ipRegex.MatchString(config.SelfIp) {
			logger.Error("invalid selfIp format")
			return fmt.Errorf("invalid selfIp format: %s", config.SelfIp)
		}
	}

	// Validate replicas
	if clusterAgent.Spec.AgentYaml.Replicas != nil && *clusterAgent.Spec.AgentYaml.Replicas < 0 {
		logger.Error("replicas must be greater than or equal to 0")
		return fmt.Errorf("replicas must be greater than or equal to 0")
	}

	logger.Info("ClusterAgent validation successful")
	return nil
}

func validateClusterName(name string) error {
	// Check if name is lowercase
	if !isLowerCase(name) {
		return fmt.Errorf("name must be lowercase: %s", name)
	}

	// Check if name is valid for Kubernetes resources
	if errs := validation.IsDNS1123Label(name); len(errs) > 0 {
		return fmt.Errorf("invalid name %q: %s", name, errs[0])
	}

	return nil
}

func isLowerCase(s string) bool {
	return regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`).MatchString(s)
}

// +kubebuilder:webhook:path=/validate-bmc-spidernet-io-v1beta1-clusteragent,mutating=true,failurePolicy=fail,sideEffects=None,groups=bmc.spidernet.io,resources=clusteragents,verbs=create;update,versions=v1beta1,name=vclusteragent.kb.io,admissionReviewVersions=v1
