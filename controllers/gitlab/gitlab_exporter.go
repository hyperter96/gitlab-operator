package gitlab

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab"
)

// ExporterService returns the Service for the GitLab Exporter component.
func ExporterService(template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(ServiceKind, GitLabExporterComponentName)
}

// ExporterDeployment returns the Deployment for the GitLab Exporter component.
func ExporterDeployment(template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(DeploymentKind, GitLabExporterComponentName)
}

// ExporterConfigMaps returns the ConfigMaps for the GitLab Exporter component.
func ExporterConfigMaps(adapter gitlab.Adapter, template helm.Template) []client.Object {
	exporterCfgMap := template.Query().ObjectByKindAndName(ConfigMapKind,
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), GitLabExporterComponentName))

	return []client.Object{exporterCfgMap}
}
