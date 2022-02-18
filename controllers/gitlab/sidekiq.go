package gitlab

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
)

const (
	gitlabSidekiqEnabled = "gitlab.sidekiq.enabled"
)

// SidekiqEnabled returns `true` if Sidekiq is enabled, and `false` if not.
func SidekiqEnabled(adapter CustomResourceAdapter) bool {
	return adapter.Values().GetBool(gitlabSidekiqEnabled)
}

// SidekiqDeployments returns the Deployments of the Sidekiq component.
func SidekiqDeployments(template helm.Template) []client.Object {
	return template.Query().ObjectsByKindAndLabels(DeploymentKind, map[string]string{
		"app": SidekiqComponentName,
	})
}

// SidekiqConfigMaps returns the ConfigMaps of the Sidekiq component.
func SidekiqConfigMaps(template helm.Template) []client.Object {
	result := template.Query().ObjectsByKindAndLabels(ConfigMapKind, map[string]string{
		"app": SidekiqComponentName,
	})

	for _, cm := range result {
		setInstallationType(cm)
	}

	return result
}
