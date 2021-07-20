package gitlab

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
)

// WebserviceDeployments returns the Deployments for the Webservice component.
func WebserviceDeployments(adapter CustomResourceAdapter) []*appsv1.Deployment {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().DeploymentsByLabels(map[string]string{
		"app": WebserviceComponentName,
	})

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

// WebserviceServices returns the Services for the Webservice component.
func WebserviceServices(adapter CustomResourceAdapter) []*corev1.Service {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().ServicesByLabels(map[string]string{
		"app": WebserviceComponentName,
	})

	return result
}

// WebserviceIngresses returns the Ingresses for the Webservice component.
func WebserviceIngresses(adapter CustomResourceAdapter) []*extensionsv1beta1.Ingress {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().IngressesByLabels(map[string]string{
		"app": WebserviceComponentName,
	})

	return result
}
