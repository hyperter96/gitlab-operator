package gitlab

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
)

// WebserviceDeployment returns the Deployment for the Webservice component.
func WebserviceDeployment(adapter CustomResourceAdapter) *appsv1.Deployment {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().DeploymentByComponent(WebserviceComponentName)

	return result
}

// WebserviceConfigMaps returns the ConfigMaps for the Webservice component.
func WebserviceConfigMaps(adapter CustomResourceAdapter) []*corev1.ConfigMap {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().ConfigMapsByLabels(map[string]string{
		"app": WebserviceComponentName,
	})

	return result
}

// WebserviceService returns the Service for the Webservice component.
func WebserviceService(adapter CustomResourceAdapter) *corev1.Service {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().ServiceByComponent(WebserviceComponentName)

	return result
}

// WebserviceIngress returns the Ingress for the Webservice component.
func WebserviceIngress(adapter CustomResourceAdapter) *extensionsv1beta1.Ingress {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().IngressByComponent(WebserviceComponentName)

	return result
}
