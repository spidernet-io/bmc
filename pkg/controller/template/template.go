package template

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

const (
	TemplateDir = "/etc/bmc/templates"
)

// TemplateData contains the data needed to render the templates
type TemplateData struct {
	Name              string
	Namespace         string
	ClusterName       string
	Image            string
	Interface        string
	Replicas         int32
	ServiceAccountName string
	RoleName          string
}

// RenderTemplate reads a template file and renders it with the given data
func RenderTemplate(templateName string, data *TemplateData) (*unstructured.Unstructured, error) {
	templatePath := filepath.Join(TemplateDir, templateName)
	templateContent, err := ioutil.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file %s: %v", templatePath, err)
	}

	tmpl, err := template.New(templateName).Parse(string(templateContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %v", templateName, err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return nil, fmt.Errorf("failed to execute template %s: %v", templateName, err)
	}

	// Convert the rendered template to an unstructured object
	var obj unstructured.Unstructured
	err = yaml.Unmarshal(buf.Bytes(), &obj)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal template %s: %v", templateName, err)
	}

	return &obj, nil
}

// RenderAllAgentResources renders all the resources needed for an agent
func RenderAllAgentResources(data *TemplateData) (map[string]*unstructured.Unstructured, error) {
	templates := []string{
		"agent-serviceaccount.yaml",
		"agent-role.yaml",
		"agent-rolebinding.yaml",
		"agent-deployment.yaml",
	}

	resources := make(map[string]*unstructured.Unstructured)
	for _, tmpl := range templates {
		obj, err := RenderTemplate(tmpl, data)
		if err != nil {
			return nil, err
		}
		resources[strings.TrimSuffix(tmpl, ".yaml")] = obj
	}

	return resources, nil
}
