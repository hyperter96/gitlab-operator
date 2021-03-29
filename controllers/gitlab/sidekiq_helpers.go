package gitlab

import (
	"fmt"

	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/helpers"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// SidekiqDeployment returns the Deployment of the Sidekiq component.
func SidekiqDeployment(adapter helpers.CustomResourceAdapter) *appsv1.Deployment {
	template, err := helpers.GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().DeploymentByComponent(SidekiqComponentName)

	return result
}

// SidekiqConfigMaps returns the ConfigMaps of the Sidekiq component.
func SidekiqConfigMaps(adapter helpers.CustomResourceAdapter) []*corev1.ConfigMap {
	template, err := helpers.GetTemplate(adapter)
	if err != nil {
		return []*corev1.ConfigMap{} // WARNING: this should return an error instead.
	}

	queueCfgMap := template.Query().ConfigMapByName(
		fmt.Sprintf("%s-%s-%s", adapter.ReleaseName(), SidekiqComponentName, "all-in-1"))
	mainCfgMap := template.Query().ConfigMapByName(
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), SidekiqComponentName))

	result := []*corev1.ConfigMap{
		queueCfgMap,
		mainCfgMap,
	}

	return result
}
