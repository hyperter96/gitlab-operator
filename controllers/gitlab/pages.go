package gitlab

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
)

const (
	globalPagesEnabled = "global.pages.enabled"
)

// PagesEnabled returns `true` if enabled and `false` if not.
func PagesEnabled(adapter CustomResourceAdapter) bool {
	return adapter.Values().GetBool(globalPagesEnabled)
}

// PagesConfigMap returns the ConfigMap for the GitLab Pages component.
func PagesConfigMap(adapter CustomResourceAdapter, template helm.Template) client.Object {
	cfgMapName := fmt.Sprintf("%s-%s", adapter.ReleaseName(), PagesComponentName)

	return template.Query().ObjectByKindAndName(ConfigMapKind, cfgMapName)
}

// PagesService returns the Service for the GitLab Pages component.
func PagesService(template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(ServiceKind, PagesComponentName)
}

// PagesDeployment returns the Deployment for the GitLab Pages component.
func PagesDeployment(template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(DeploymentKind, PagesComponentName)
}

// PagesIngress returns the Ingress for the GitLab Pages component.
func PagesIngress(template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(IngressKind, PagesComponentName)
}
