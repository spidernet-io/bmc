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
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	bmcv1beta1 "github.com/spidernet-io/bmc/pkg/apis/bmc/v1beta1"
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
	logger := log.FromContext(ctx)
	clusterAgent, ok := obj.(*bmcv1beta1.ClusterAgent)
	if !ok {
		return fmt.Errorf("object is not a ClusterAgent")
	}

	// Set default image from AGENT_IMAGE environment variable if not specified
	if clusterAgent.Spec.Image == "" {
		agentImage := os.Getenv("AGENT_IMAGE")
		if agentImage == "" {
			return fmt.Errorf("AGENT_IMAGE environment variable not set")
		}
		logger.Info("Setting default image", "image", agentImage)
		clusterAgent.Spec.Image = agentImage
	}

	// Set default replicas to 1 if not specified
	if clusterAgent.Spec.Replicas == nil {
		defaultReplicas := int32(1)
		logger.Info("Setting default replicas", "replicas", defaultReplicas)
		clusterAgent.Spec.Replicas = &defaultReplicas
	}

	return nil
}

// +kubebuilder:webhook:path=/validate-bmc-spidernet-io-v1beta1-clusteragent,mutating=true,failurePolicy=fail,sideEffects=None,groups=bmc.spidernet.io,resources=clusteragents,verbs=create;update,versions=v1beta1,name=vclusteragent.kb.io,admissionReviewVersions=v1

// ValidateCreate implements webhook.Validator
func (w *ClusterAgentWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	logger := log.FromContext(ctx)
	clusterAgent, ok := obj.(*bmcv1beta1.ClusterAgent)
	if !ok {
		return nil, fmt.Errorf("object is not a ClusterAgent")
	}

	logger.Info("Validating ClusterAgent creation",
		"name", clusterAgent.Name,
		"clusterName", clusterAgent.Spec.ClusterName)

	if err := w.validateClusterAgent(ctx, clusterAgent); err != nil {
		return nil, err
	}

	return nil, nil
}

// ValidateUpdate implements webhook.Validator
func (w *ClusterAgentWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	logger := log.FromContext(ctx)
	clusterAgent, ok := newObj.(*bmcv1beta1.ClusterAgent)
	if !ok {
		return nil, fmt.Errorf("object is not a ClusterAgent")
	}

	logger.Info("Validating ClusterAgent update",
		"name", clusterAgent.Name,
		"clusterName", clusterAgent.Spec.ClusterName)

	if err := w.validateClusterAgent(ctx, clusterAgent); err != nil {
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
	if clusterAgent.Spec.ClusterName == "" {
		return fmt.Errorf("clusterName is required")
	}
	if clusterAgent.Spec.UnderlayInterface == "" {
		return fmt.Errorf("underlayInterface is required")
	}

	// Validate clusterName format
	if err := validateClusterName(clusterAgent.Spec.ClusterName); err != nil {
		return err
	}

	// Validate replicas is non-negative
	if clusterAgent.Spec.Replicas != nil && *clusterAgent.Spec.Replicas < 0 {
		return fmt.Errorf("replicas must be greater than or equal to 0")
	}

	// Check for clusterName uniqueness
	var clusterAgentList bmcv1beta1.ClusterAgentList
	if err := w.Client.List(ctx, &clusterAgentList); err != nil {
		return fmt.Errorf("failed to list ClusterAgents: %v", err)
	}

	for _, ca := range clusterAgentList.Items {
		if ca.Name != clusterAgent.Name && ca.Spec.ClusterName == clusterAgent.Spec.ClusterName {
			return fmt.Errorf("clusterName %s is already used by ClusterAgent %s",
				clusterAgent.Spec.ClusterName, ca.Name)
		}
	}

	return nil
}

func validateClusterName(name string) error {
	// Check if name is lowercase
	if !isLowerCase(name) {
		return fmt.Errorf("clusterName must be lowercase: %s", name)
	}

	// Check if name is valid for Kubernetes resources
	if errs := validation.IsDNS1123Label(name); len(errs) > 0 {
		return fmt.Errorf("invalid clusterName %q: %s", name, errs[0])
	}

	return nil
}

func isLowerCase(s string) bool {
	return regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`).MatchString(s)
}
