package gitlab

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	registryEnabled        = "registry.enabled"
	registryEnabledDefault = true
)

// RegistryEnabled returns `true` if the registry is enabled, and `false` if not.
func RegistryEnabled(adapter CustomResourceAdapter) bool {
	return adapter.Values().GetBool(registryEnabled, registryEnabledDefault)
}

// RegistryService returns the Service of the Registry component.
func RegistryService(adapter CustomResourceAdapter) client.Object {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().ObjectByKindAndComponent(ServiceKind, RegistryComponentName)

	return result
}

// RegistryDeployment returns the Deployment of the Registry component.
func RegistryDeployment(adapter CustomResourceAdapter) client.Object {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().ObjectByKindAndComponent(DeploymentKind, RegistryComponentName)

	return result
}

// RegistryConfigMap returns the ConfigMap of the Registry component.
func RegistryConfigMap(adapter CustomResourceAdapter) client.Object {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().ObjectByKindAndName(ConfigMapKind,
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), RegistryComponentName))

	return result
}

// RegistryIngress returns the Ingress of the Registry component.
func RegistryIngress(adapter CustomResourceAdapter) client.Object {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().ObjectByKindAndComponent(IngressKind, RegistryComponentName)

	return result
}
