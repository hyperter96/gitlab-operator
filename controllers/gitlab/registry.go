package gitlab

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
)

// RegistryService returns the Service of the Registry component.
func RegistryService(adapter CustomResourceAdapter) *corev1.Service {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}
	result := template.Query().ServiceByComponent(RegistryComponentName)

	return result
}

// RegistryDeployment returns the Deployment of the Registry component.
func RegistryDeployment(adapter CustomResourceAdapter) *appsv1.Deployment {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().DeploymentByComponent(RegistryComponentName)

	return result
}

// RegistryConfigMap returns the ConfigMap of the Registry component.
func RegistryConfigMap(adapter CustomResourceAdapter) *corev1.ConfigMap {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().ConfigMapByName(
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), RegistryComponentName))

	return result
}

// RegistryIngress returns the Ingress of the Registry component.
func RegistryIngress(adapter CustomResourceAdapter) *extensionsv1beta1.Ingress {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}
	result := template.Query().IngressByComponent(RegistryComponentName)

	return result
}
