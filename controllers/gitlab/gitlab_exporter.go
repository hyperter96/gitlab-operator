package gitlab

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	gitlabExporterEnabled  = "gitlab.gitlab-exporter.enabled"
	exporterEnabledDefault = true
)

// ExporterEnabled returns `true` if enabled and `false` if not.
func ExporterEnabled(adapter CustomResourceAdapter) bool {
	enabled, _ := GetBoolValue(adapter.Values(), gitlabExporterEnabled, exporterEnabledDefault)

	return enabled
}

// ExporterService returns the Service for the GitLab Exporter component.
func ExporterService(adapter CustomResourceAdapter) *corev1.Service {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().ServiceByComponent(GitLabExporterComponentName)

	return result
}

// ExporterDeployment returns the Deployment for the GitLab Exporter component.
func ExporterDeployment(adapter CustomResourceAdapter) *appsv1.Deployment {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().DeploymentByComponent(GitLabExporterComponentName)

	return result
}

// ExporterConfigMaps returns the ConfigMaps for the GitLab Exporter component.
func ExporterConfigMaps(adapter CustomResourceAdapter) []*corev1.ConfigMap {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	exporterCfgMap := template.Query().ConfigMapByName(
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), GitLabExporterComponentName))

	result := []*corev1.ConfigMap{exporterCfgMap}

	return result
}
