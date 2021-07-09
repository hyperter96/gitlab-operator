package gitlab

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// NGINXConfigMaps returns the ConfigMaps of the NGINX component.
func NGINXConfigMaps(adapter CustomResourceAdapter) []*corev1.ConfigMap {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil
	}

	result := template.Query().ConfigMapsByLabels(map[string]string{
		"app": NGINXComponentName,
	})

	for _, cm := range result {
		cm.SetNamespace(adapter.Namespace())
	}

	return result
}

// NGINXServices returns the Services of the NGINX Component.
func NGINXServices(adapter CustomResourceAdapter) []*corev1.Service {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil
	}

	result := template.Query().ServicesByLabels(map[string]string{
		"app": NGINXComponentName,
	})

	for _, svc := range result {
		svc.SetNamespace(adapter.Namespace())
	}

	return result
}

// NGINXDeployments returns the Deployments of the NGINX Component.
func NGINXDeployments(adapter CustomResourceAdapter) []*appsv1.Deployment {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil
	}

	result := template.Query().DeploymentsByLabels(map[string]string{
		"app": NGINXComponentName,
	})

	for _, dep := range result {
		dep.SetNamespace(adapter.Namespace())
	}

	return result
}
