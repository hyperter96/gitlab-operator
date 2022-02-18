package gitlab

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
)

const (
	gitlabShellEnabled   = "gitlab.gitlab-shell.enabled"
	gitlabShellSshDaemon = "gitlab.gitlab-shell.sshDaemon"
)

// ShellEnabled returns `true` if enabled, and `false` if not.
func ShellEnabled(adapter CustomResourceAdapter) bool {
	return adapter.Values().GetBool(gitlabShellEnabled)
}

// ShellDeployment returns the Deployment of GitLab Shell component.
func ShellDeployment(template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(DeploymentKind, GitLabShellComponentName)
}

// ShellSshDaemon returns the SSH daemon of GitLab Shell component.
func ShellSshDaemon(adapter CustomResourceAdapter) string {
	return adapter.Values().GetString(gitlabShellSshDaemon)
}

// ShellConfigMaps returns the ConfigMaps of GitLab Shell component.
func ShellConfigMaps(adapter CustomResourceAdapter, template helm.Template) []client.Object {
	shellCfgMap := template.Query().ObjectByKindAndName(ConfigMapKind,
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), GitLabShellComponentName))
	tcpCfgMap := template.Query().ObjectByKindAndName(ConfigMapKind,
		fmt.Sprintf("%s-nginx-ingress-tcp", adapter.ReleaseName()))

	result := []client.Object{
		shellCfgMap,
		tcpCfgMap,
	}

	if ShellSshDaemon(adapter) == "openssh" {
		sshdCfgMap := template.Query().ObjectByKindAndName(ConfigMapKind,
			fmt.Sprintf("%s-%s-sshd", adapter.ReleaseName(), GitLabShellComponentName))
		result = append(result, sshdCfgMap)
	}

	return result
}

// ShellService returns the Service of GitLab Shell component.
func ShellService(template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(ServiceKind, GitLabShellComponentName)
}
