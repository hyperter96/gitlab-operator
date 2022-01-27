package gitlab

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	globalPagesEnabled  = "global.pages.enabled"
	pagesEnabledDefault = false
)

// PagesEnabled returns `true` if enabled and `false` if not.
func PagesEnabled(adapter CustomResourceAdapter) bool {
	return adapter.Values().GetBool(globalPagesEnabled, pagesEnabledDefault)
}

// PagesConfigMap returns the ConfigMap for the GitLab Pages component.
func PagesConfigMap(adapter CustomResourceAdapter) client.Object {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	cfgMapName := fmt.Sprintf("%s-%s", adapter.ReleaseName(), PagesComponentName)

	return template.Query().ObjectByKindAndName(ConfigMapKind, cfgMapName)
}

// PagesService returns the Service for the GitLab Pages component.
func PagesService(adapter CustomResourceAdapter) client.Object {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	return template.Query().ObjectByKindAndComponent(ServiceKind, PagesComponentName)
}

// PagesDeployment returns the Deployment for the GitLab Pages component.
func PagesDeployment(adapter CustomResourceAdapter) client.Object {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	return template.Query().ObjectByKindAndComponent(DeploymentKind, PagesComponentName)
}

// PagesIngress returns the Ingress for the GitLab Pages component.
func PagesIngress(adapter CustomResourceAdapter) client.Object {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	return template.Query().ObjectByKindAndComponent(IngressKind, PagesComponentName)
}
