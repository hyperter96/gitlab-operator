package gitlab

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	gitlabExporterEnabled  = "gitlab.gitlab-exporter.enabled"
	exporterEnabledDefault = true
)

// ExporterEnabled returns `true` if enabled and `false` if not.
func ExporterEnabled(adapter CustomResourceAdapter) bool {
	return adapter.Values().GetBool(gitlabExporterEnabled, exporterEnabledDefault)
}

// ExporterService returns the Service for the GitLab Exporter component.
func ExporterService(adapter CustomResourceAdapter) client.Object {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().ObjectByKindAndComponent(ServiceKind, GitLabExporterComponentName)

	return result
}

// ExporterDeployment returns the Deployment for the GitLab Exporter component.
func ExporterDeployment(adapter CustomResourceAdapter) client.Object {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().ObjectByKindAndComponent(DeploymentKind, GitLabExporterComponentName)

	return result
}

// ExporterConfigMaps returns the ConfigMaps for the GitLab Exporter component.
func ExporterConfigMaps(adapter CustomResourceAdapter) []client.Object {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	exporterCfgMap := template.Query().ObjectByKindAndName(ConfigMapKind,
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), GitLabExporterComponentName))

	result := []client.Object{exporterCfgMap}

	return result
}
