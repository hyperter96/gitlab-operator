package gitlab

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
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
func WebserviceDeployments(adapter CustomResourceAdapter) []client.Object {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().ObjectsByKindAndLabels(DeploymentKind, map[string]string{
		"app": WebserviceComponentName,
	})

	return result
}

// WebserviceConfigMaps returns the ConfigMaps for the Webservice component.
func WebserviceConfigMaps(adapter CustomResourceAdapter) []client.Object {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().ObjectsByKindAndLabels(ConfigMapKind, map[string]string{
		"app": WebserviceComponentName,
	})

	for _, cm := range result {
		setInstallationType(cm)
	}

	return result
}

// WebserviceServices returns the Services for the Webservice component.
func WebserviceServices(adapter CustomResourceAdapter) []client.Object {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().ObjectsByKindAndLabels(ServiceKind, map[string]string{
		"app": WebserviceComponentName,
	})

	return result
}

// WebserviceIngresses returns the Ingresses for the Webservice component.
func WebserviceIngresses(adapter CustomResourceAdapter) []client.Object {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().ObjectsByKindAndLabels(IngressKind, map[string]string{
		"app": WebserviceComponentName,
	})

	return result
}
