package gitlab

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// TaskRunnerDeployment returns the Deployment of the Task Runner component.
func TaskRunnerDeployment(adapter CustomResourceAdapter) *appsv1.Deployment {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().DeploymentByComponent(TaskRunnerComponentName)

	return result
}

// TaskRunnerConfigMap returns the ConfigMaps of the Task Runner component.
func TaskRunnerConfigMap(adapter CustomResourceAdapter) *corev1.ConfigMap {
	var result *corev1.ConfigMap

	template, err := GetTemplate(adapter)

	if err != nil {
		return nil // WARNING: This should return an error instead.
	}

	result = template.Query().ConfigMapByName(
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), TaskRunnerComponentName))

	return result
}
