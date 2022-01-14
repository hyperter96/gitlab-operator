package gitlab

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

const (
	gitlabWebserviceEnabled  = "gitlab.webservice.enabled"
	webserviceEnabledDefault = true
)

// WebserviceEnabled returns `true` if Webservice is enabled, and `false` if not.
func WebserviceEnabled(adapter CustomResourceAdapter) bool {
	return adapter.Values().GetBool(gitlabWebserviceEnabled, webserviceEnabledDefault)
}

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

	for _, cm := range result {
		setInstallationType(cm)
	}

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
func WebserviceIngresses(adapter CustomResourceAdapter) []*networkingv1.Ingress {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().IngressesByLabels(map[string]string{
		"app": WebserviceComponentName,
	})

	return result
}
