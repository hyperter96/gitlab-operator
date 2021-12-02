package gitlab

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	gitlabToolboxEnabled  = "gitlab.%s.enabled"
	toolboxEnabledDefault = true
)

// ToolboxEnabled returns `true` if Toolbox is enabled, and `false` if not.
func ToolboxEnabled(adapter CustomResourceAdapter) bool {
	key := fmt.Sprintf(gitlabToolboxEnabled, ToolboxComponentName(adapter.ChartVersion()))
	enabled, _ := GetBoolValue(adapter.Values(), key, toolboxEnabledDefault)

	return enabled
}

// ToolboxDeployment returns the Deployment of the Toolbox component.
func ToolboxDeployment(adapter CustomResourceAdapter) *appsv1.Deployment {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().DeploymentByComponent(ToolboxComponentName(adapter.ChartVersion()))

	return result
}

// ToolboxConfigMap returns the ConfigMaps of the Toolbox component.
func ToolboxConfigMap(adapter CustomResourceAdapter) *corev1.ConfigMap {
	var result *corev1.ConfigMap

	template, err := GetTemplate(adapter)

	if err != nil {
		return nil // WARNING: This should return an error instead.
	}

	result = template.Query().ConfigMapByName(
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), ToolboxComponentName(adapter.ChartVersion())))

	return result
}
