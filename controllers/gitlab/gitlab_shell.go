package gitlab

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
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
func ShellDeployment(adapter CustomResourceAdapter) client.Object {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: This should return an error instead.
	}

	return template.Query().ObjectByKindAndComponent(DeploymentKind, GitLabShellComponentName)
}

// ShellConfigMaps returns the ConfigMaps of GitLab Shell component.
func ShellConfigMaps(adapter CustomResourceAdapter) []client.Object {
	template, err := GetTemplate(adapter)
	if err != nil {
		return []client.Object{} // WARNING: This should return an error instead.
	}

	shellCfgMap := template.Query().ObjectByKindAndName(ConfigMapKind,
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), GitLabShellComponentName))
	sshdCfgMap := template.Query().ObjectByKindAndName(ConfigMapKind,
		fmt.Sprintf("%s-%s-sshd", adapter.ReleaseName(), GitLabShellComponentName))
	tcpCfgMap := template.Query().ObjectByKindAndName(ConfigMapKind,
		fmt.Sprintf("%s-nginx-ingress-tcp", adapter.ReleaseName()))

	result := []client.Object{
		shellCfgMap,
		sshdCfgMap,
		tcpCfgMap,
	}

	return result
}

// ShellService returns the Service of GitLab Shell component.
func ShellService(adapter CustomResourceAdapter) client.Object {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: This should return an error instead.
	}

	result := template.Query().ObjectByKindAndComponent(ServiceKind, GitLabShellComponentName)

	return result
}
