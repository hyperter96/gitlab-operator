package gitlab

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
)

const (
	gitlabShellEnabled        = "gitlab.gitlab-shell.enabled"
	gitlabShellEnabledDefault = true
)

// ShellEnabled returns `true` if enabled, and `false` if not.
func ShellEnabled(adapter CustomResourceAdapter) bool {
	return adapter.Values().GetBool(gitlabShellEnabled, gitlabShellEnabledDefault)
}

// ShellDeployment returns the Deployment of GitLab Shell component.
func ShellDeployment(template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(DeploymentKind, GitLabShellComponentName)
}

// ShellConfigMaps returns the ConfigMaps of GitLab Shell component.
func ShellConfigMaps(adapter CustomResourceAdapter, template helm.Template) []client.Object {
	shellCfgMap := template.Query().ObjectByKindAndName(ConfigMapKind,
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), GitLabShellComponentName))
	sshdCfgMap := template.Query().ObjectByKindAndName(ConfigMapKind,
		fmt.Sprintf("%s-%s-sshd", adapter.ReleaseName(), GitLabShellComponentName))
	tcpCfgMap := template.Query().ObjectByKindAndName(ConfigMapKind,
		fmt.Sprintf("%s-nginx-ingress-tcp", adapter.ReleaseName()))

	return []client.Object{
		shellCfgMap,
		sshdCfgMap,
		tcpCfgMap,
	}
}

// ShellService returns the Service of GitLab Shell component.
func ShellService(template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(ServiceKind, GitLabShellComponentName)
}
