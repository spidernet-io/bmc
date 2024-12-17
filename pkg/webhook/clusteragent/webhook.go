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

	// Set default image from AGENT_IMAGE environment variable if not specified
	if clusterAgent.Spec.AgentYaml.Image == "" {
		agentImage := os.Getenv("AGENT_IMAGE")
		if agentImage == "" {
			return fmt.Errorf("AGENT_IMAGE environment variable not set")
		}
		log.Logger.Infof("Setting default image: %s", agentImage)
		clusterAgent.Spec.AgentYaml.Image = agentImage
	}

	// Set default replicas to 1 if not specified
	if clusterAgent.Spec.AgentYaml.Replicas == nil {
		defaultReplicas := int32(1)
		log.Logger.Infof("Setting default replicas: %d", defaultReplicas)
		clusterAgent.Spec.AgentYaml.Replicas = &defaultReplicas
	}

	return nil
}

// ValidateCreate implements webhook.Validator
func (w *ClusterAgentWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	clusterAgent, ok := obj.(*bmcv1beta1.ClusterAgent)
	if !ok {
		return nil, fmt.Errorf("object is not a ClusterAgent")
	}

	log.Logger.Infof("Validating ClusterAgent creation: %s", clusterAgent.Name)

	if err := w.validateClusterAgent(ctx, clusterAgent); err != nil {
		log.Logger.Errorf("ClusterAgent validation failed: %v", err)
		return nil, err
	}

	return nil, nil
}

// ValidateUpdate implements webhook.Validator
func (w *ClusterAgentWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	clusterAgent, ok := newObj.(*bmcv1beta1.ClusterAgent)
	if !ok {
		return nil, fmt.Errorf("object is not a ClusterAgent")
	}

	log.Logger.Infof("Validating ClusterAgent update: %s", clusterAgent.Name)

	if err := w.validateClusterAgent(ctx, clusterAgent); err != nil {
		log.Logger.Errorf("ClusterAgent validation failed: %v", err)
		return nil, err
	}

	return nil, nil
}

// ValidateDelete implements webhook.Validator
func (w *ClusterAgentWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (w *ClusterAgentWebhook) validateClusterAgent(ctx context.Context, clusterAgent *bmcv1beta1.ClusterAgent) error {
	// Validate required fields
	if clusterAgent.Name == "" {
		return fmt.Errorf("name is required")
	}
	if clusterAgent.Spec.AgentYaml.UnderlayInterface == "" {
		return fmt.Errorf("underlayInterface is required")
	}

	// Validate name format
	if err := validateClusterName(clusterAgent.Name); err != nil {
		return err
	}

	// Validate replicas is non-negative
	if clusterAgent.Spec.AgentYaml.Replicas != nil && *clusterAgent.Spec.AgentYaml.Replicas < 0 {
		return fmt.Errorf("replicas must be greater than or equal to 0")
	}

	// Check for name uniqueness is not needed as k8s already ensures name uniqueness within a namespace

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
