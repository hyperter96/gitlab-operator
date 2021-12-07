package gitlab

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

const (
	globalPagesEnabled  = "global.pages.enabled"
	pagesEnabledDefault = false
)

// PagesEnabled returns `true` if enabled and `false` if not.
func PagesEnabled(adapter CustomResourceAdapter) bool {
	enabled, _ := GetBoolValue(adapter.Values(), globalPagesEnabled, pagesEnabledDefault)
	return enabled
}

// PagesConfigMap returns the ConfigMap for the GitLab Pages component.
func PagesConfigMap(adapter CustomResourceAdapter) *corev1.ConfigMap {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	cfgMapName := fmt.Sprintf("%s-%s", adapter.ReleaseName(), PagesComponentName)

	return template.Query().ConfigMapByName(cfgMapName)
}

// PagesService returns the Service for the GitLab Pages component.
func PagesService(adapter CustomResourceAdapter) *corev1.Service {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	return template.Query().ServiceByComponent(PagesComponentName)
}

// PagesDeployment returns the Deployment for the GitLab Pages component.
func PagesDeployment(adapter CustomResourceAdapter) *appsv1.Deployment {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	return template.Query().DeploymentByComponent(PagesComponentName)
}

// PagesIngress returns the Ingress for the GitLab Pages component.
func PagesIngress(adapter CustomResourceAdapter) *networkingv1.Ingress {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	return template.Query().IngressByComponent(PagesComponentName)
}
