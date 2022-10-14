package gitlab

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab"
)

// RegistryService returns the Service of the Registry component.
func RegistryService(template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(ServiceKind, RegistryComponentName)
}

// RegistryDeployment returns the Deployment of the Registry component.
func RegistryDeployment(template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(DeploymentKind, RegistryComponentName)
}

// RegistryConfigMap returns the ConfigMap of the Registry component.
func RegistryConfigMap(adapter gitlab.Adapter, template helm.Template) client.Object {
	return template.Query().ObjectByKindAndName(ConfigMapKind,
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), RegistryComponentName))
}

// RegistryIngress returns the Ingress of the Registry component.
func RegistryIngress(template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(IngressKind, RegistryComponentName)
}
