package gitlab

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	gitlabSidekiqEnabled  = "gitlab.sidekiq.enabled"
	sidekiqEnabledDefault = true
)

// SidekiqEnabled returns `true` if Sidekiq is enabled, and `false` if not.
func SidekiqEnabled(adapter CustomResourceAdapter) bool {
	return adapter.Values().GetBool(gitlabSidekiqEnabled, sidekiqEnabledDefault)
}

// SidekiqDeployments returns the Deployments of the Sidekiq component.
func SidekiqDeployments(adapter CustomResourceAdapter) []*appsv1.Deployment {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().DeploymentsByLabels(map[string]string{
		"app": SidekiqComponentName,
	})

	return result
}

// SidekiqConfigMaps returns the ConfigMaps of the Sidekiq component.
func SidekiqConfigMaps(adapter CustomResourceAdapter) []*corev1.ConfigMap {
	template, err := GetTemplate(adapter)
	if err != nil {
		return []*corev1.ConfigMap{} // WARNING: this should return an error instead.
	}

	result := template.Query().ConfigMapsByLabels(map[string]string{
		"app": SidekiqComponentName,
	})

	for _, cm := range result {
		setInstallationType(cm)
	}

	return result
}
