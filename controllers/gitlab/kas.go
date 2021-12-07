package gitlab

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

// KasEnabled returns `true` if KAS is enabled, and `false` if not. By default it returns `false`.
func KasEnabled(adapter CustomResourceAdapter) bool {
	enabled, _ := GetBoolValue(adapter.Values(), "global.kas.enabled", false)

	return enabled
}

func KasConfigMap(adapter CustomResourceAdapter) *corev1.ConfigMap {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	return template.Query().ConfigMapByComponent(KasComponentName)
}

func KasDeployment(adapter CustomResourceAdapter) *appsv1.Deployment {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	return template.Query().DeploymentByComponent(KasComponentName)
}

func KasIngress(adapter CustomResourceAdapter) *networkingv1.Ingress {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	return template.Query().IngressByComponent(KasComponentName)
}

func KasService(adapter CustomResourceAdapter) *corev1.Service {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	return template.Query().ServiceByComponent(KasComponentName)
}
