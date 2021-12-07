package gitlab

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

const (
	gitlabToolboxEnabled         = "gitlab.%s.enabled"
	toolboxEnabledDefault        = true
	gitlabToolboxCronJobEnabled  = "gitlab.%s.backups.cron.enabled"
	toolboxCronJobEnabledDefault = false
)

// ToolboxEnabled returns `true` if Toolbox is enabled, and `false` if not.
func ToolboxEnabled(adapter CustomResourceAdapter) bool {
	key := fmt.Sprintf(gitlabToolboxEnabled, ToolboxComponentName(adapter.ChartVersion()))
	enabled, _ := GetBoolValue(adapter.Values(), key, toolboxEnabledDefault)

	return enabled
}

// ToolboxCronJobEnabled returns `true` if Toolbox CronJob is enabled, and `false` if not.
func ToolboxCronJobEnabled(adapter CustomResourceAdapter) bool {
	key := fmt.Sprintf(gitlabToolboxCronJobEnabled, ToolboxComponentName(adapter.ChartVersion()))
	enabled, _ := GetBoolValue(adapter.Values(), key, toolboxCronJobEnabledDefault)

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

// ToolboxCronJob returns the CronJob of the Toolbox component.
func ToolboxCronJob(adapter CustomResourceAdapter) *batchv1beta1.CronJob {
	var result *batchv1beta1.CronJob

	template, err := GetTemplate(adapter)

	if err != nil {
		return nil // WARNING: This should return an error instead.
	}

	result = template.Query().CronJobByName(
		fmt.Sprintf("%s-%s-backup", adapter.ReleaseName(), ToolboxComponentName(adapter.ChartVersion())))

	return result
}
