package gitlab

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// MailroomService returns the Service for the Mailroom component.
func MailroomService(adapter CustomResourceAdapter) *corev1.Service {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}
	result := template.Query().ServiceByComponent(MailroomComponentName)

	return result
}

// MailroomDeployment returns the Deployment for the Mailroom component.
func MailroomDeployment(adapter CustomResourceAdapter) *appsv1.Deployment {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().DeploymentByComponent(MailroomComponentName)

	return result
}

// MailroomConfigMaps returns the ConfigMaps for the Mailroom component.
func MailroomConfigMaps(adapter CustomResourceAdapter) []*corev1.ConfigMap {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	exporterCfgMap := template.Query().ConfigMapByName(
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), MailroomComponentName))

	result := []*corev1.ConfigMap{exporterCfgMap}

	return result
}
