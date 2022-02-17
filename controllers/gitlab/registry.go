package gitlab

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
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
func RegistryService(template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(ServiceKind, RegistryComponentName)
}

// RegistryDeployment returns the Deployment of the Registry component.
func RegistryDeployment(template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(DeploymentKind, RegistryComponentName)
}

// RegistryConfigMap returns the ConfigMap of the Registry component.
func RegistryConfigMap(adapter CustomResourceAdapter, template helm.Template) client.Object {
	return template.Query().ObjectByKindAndName(ConfigMapKind,
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), RegistryComponentName))
}

// RegistryIngress returns the Ingress of the Registry component.
func RegistryIngress(template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(IngressKind, RegistryComponentName)
}
