package gitlab

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
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
func SidekiqDeployments(adapter CustomResourceAdapter) []client.Object {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().ObjectsByKindAndLabels(DeploymentKind, map[string]string{
		"app": SidekiqComponentName,
	})

	return result
}

// SidekiqConfigMaps returns the ConfigMaps of the Sidekiq component.
func SidekiqConfigMaps(adapter CustomResourceAdapter) []client.Object {
	template, err := GetTemplate(adapter)
	if err != nil {
		return []client.Object{} // WARNING: this should return an error instead.
	}

	result := template.Query().ObjectsByKindAndLabels(ConfigMapKind, map[string]string{
		"app": SidekiqComponentName,
	})

	for _, cm := range result {
		setInstallationType(cm)
	}

	return result
}
