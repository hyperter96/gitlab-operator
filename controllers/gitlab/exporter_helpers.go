package gitlab

import (
	"fmt"

	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/helpers"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// ExporterService returns the Service for the GitLab Exporter component.
func ExporterService(adapter helpers.CustomResourceAdapter) *corev1.Service {
	template, err := helpers.GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}
	result := template.Query().ServiceByComponent(GitLabExporterComponentName)

	return result
}

// ExporterDeployment returns the Deployment for the GitLab Exporter component.
func ExporterDeployment(adapter helpers.CustomResourceAdapter) *appsv1.Deployment {
	template, err := helpers.GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().DeploymentByComponent(GitLabExporterComponentName)

	return result
}

// ExporterConfigMaps returns the ConfigMaps for the GitLab Exporter component.
func ExporterConfigMaps(adapter helpers.CustomResourceAdapter) []*corev1.ConfigMap {
	template, err := helpers.GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	exporterCfgMap := template.Query().ConfigMapByName(
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), GitLabExporterComponentName))

	result := []*corev1.ConfigMap{exporterCfgMap}

	return result
}
