package gitlab

import (
	"fmt"

	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/helpers"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
)

// RegistryService returns the Service of the Registry component.
func RegistryService(adapter helpers.CustomResourceAdapter) *corev1.Service {
	template, err := helpers.GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}
	result := template.Query().ServiceByComponent(RegistryComponentName)

	return result
}

// RegistryDeployment returns the Deployment of the Registry component.
func RegistryDeployment(adapter helpers.CustomResourceAdapter) *appsv1.Deployment {
	template, err := helpers.GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().DeploymentByComponent(RegistryComponentName)

	return result
}

// RegistryConfigMap returns the ConfigMap of the Registry component.
func RegistryConfigMap(adapter helpers.CustomResourceAdapter) *corev1.ConfigMap {
	template, err := helpers.GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().ConfigMapByName(
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), RegistryComponentName))

	return result
}

// RegistryIngress returns the Ingress of the Registry component.
func RegistryIngress(adapter helpers.CustomResourceAdapter) *extensionsv1beta1.Ingress {
	template, err := helpers.GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}
	result := template.Query().IngressByComponent(RegistryComponentName)

	return result
}
