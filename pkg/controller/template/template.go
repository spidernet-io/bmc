package template

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"
)

var logger = log.Log.WithName("template")

// TemplateData contains the data needed to render the templates
type TemplateData struct {
	// Name is the name of the resource
	Name string
	// Namespace is the namespace where the resource will be created
	Namespace string
	// ClusterName is the name of the cluster this agent belongs to
	ClusterName string
	// Replicas is the number of agent replicas to run
	Replicas int32
	// ServiceAccountName is the name of the ServiceAccount to use
	ServiceAccountName string
	// RoleName is the name of the Role to use
	RoleName string
	// Image is the container image to use for the agent
	Image string
	// UnderlayInterface is the network interface to use
	UnderlayInterface string
}

// RenderTemplate reads a template file and renders it with the given data
func RenderTemplate(templateName string, data *TemplateData) (*unstructured.Unstructured, error) {
	logger.Info("Starting template rendering",
		"templateName", templateName,
		"name", data.Name,
		"namespace", data.Namespace)

	templatePath := filepath.Join("/etc/bmc/templates", templateName)
	templateContent, err := ioutil.ReadFile(templatePath)
	if err != nil {
		logger.Error(err, "Failed to read template file",
			"templatePath", templatePath)
		return nil, fmt.Errorf("failed to read template file %s: %v", templatePath, err)
	}

	logger.Info("Template file read successfully",
		"templatePath", templatePath,
		"contentLength", len(templateContent))

	tmpl, err := template.New(templateName).Parse(string(templateContent))
	if err != nil {
		logger.Error(err, "Failed to parse template",
			"templateName", templateName)
		return nil, fmt.Errorf("failed to parse template %s: %v", templateName, err)
	}

	logger.Info("Template and data",
		"template", tmpl,
		"data", data)

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		logger.Error(err, "Failed to execute template",
			"templateName", templateName)
		return nil, fmt.Errorf("failed to execute template %s: %v", templateName, err)
	}

	logger.Info("Template executed successfully",
		"templateName", templateName,
		"outputLength", buf.Len())

	var obj map[string]interface{}
	err = yaml.Unmarshal(buf.Bytes(), &obj)
	if err != nil {
		logger.Error(err, "Failed to unmarshal template output",
			"templateName", templateName)
		return nil, fmt.Errorf("failed to unmarshal template output %s: %v", templateName, err)
	}
	logger.V(1).Info("Rendered template obj",
		"obj", fmt.Sprintf("%+v", obj))

	result := &unstructured.Unstructured{Object: obj}

	logger.V(1).Info("Rendered template result",
		"result", fmt.Sprintf("%+v", *result))

	logger.Info("Template rendered successfully",
		"templateName", templateName,
		"kind", result.GetKind(),
		"name", result.GetName(),
		"namespace", result.GetNamespace())

	return result, nil
}

// RenderAllAgentResources renders all the resources needed for an agent
func RenderAllAgentResources(data *TemplateData) (map[string]*unstructured.Unstructured, error) {
	logger.Info("Starting to render all agent resources",
		"name", data.Name,
		"namespace", data.Namespace,
		"clusterName", data.ClusterName)

	resources := make(map[string]*unstructured.Unstructured)
	templates := []string{
		"agent-deployment.yaml",
		"agent-serviceaccount.yaml",
		"agent-role.yaml",
		"agent-rolebinding.yaml",
	}

	for _, tmpl := range templates {
		logger.Info("Rendering template",
			"templateName", tmpl)

		obj, err := RenderTemplate(tmpl, data)
		if err != nil {
			logger.Error(err, "Failed to render template",
				"templateName", tmpl)
			return nil, err
		}

		kind := strings.ToLower(obj.GetKind())
		resources[kind] = obj

		logger.Info("Template rendered and added to resources",
			"templateName", tmpl,
			"kind", kind,
			"name", obj.GetName(),
			"namespace", obj.GetNamespace())
	}

	logger.Info("All agent resources rendered successfully",
		"resourceCount", len(resources))

	return resources, nil
}
