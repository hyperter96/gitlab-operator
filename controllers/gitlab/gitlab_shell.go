package gitlab

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	gitlabShellEnabled        = "gitlab.gitlab-shell.enabled"
	gitlabShellEnabledDefault = true
)

// ShellEnabled returns `true` if enabled, and `false` if not.
func ShellEnabled(adapter CustomResourceAdapter) bool {
	enabled, _ := GetBoolValue(adapter.Values(), gitlabShellEnabled, gitlabShellEnabledDefault)

	return enabled
}

// ShellDeployment returns the Deployment of GitLab Shell component.
func ShellDeployment(adapter CustomResourceAdapter) *appsv1.Deployment {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: This should return an error instead.
	}

	return template.Query().DeploymentByComponent(GitLabShellComponentName)
}

// ShellConfigMaps returns the ConfigMaps of GitLab Shell component.
func ShellConfigMaps(adapter CustomResourceAdapter) []*corev1.ConfigMap {
	template, err := GetTemplate(adapter)
	if err != nil {
		return []*corev1.ConfigMap{} // WARNING: This should return an error instead.
	}

	shellCfgMap := template.Query().ConfigMapByName(
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), GitLabShellComponentName))
	sshdCfgMap := template.Query().ConfigMapByName(
		fmt.Sprintf("%s-%s-sshd", adapter.ReleaseName(), GitLabShellComponentName))
	tcpCfgMap := template.Query().ConfigMapByName(
		fmt.Sprintf("%s-nginx-ingress-tcp", adapter.ReleaseName()))

	result := []*corev1.ConfigMap{
		shellCfgMap,
		sshdCfgMap,
		tcpCfgMap,
	}

	return result
}

// ShellService returns the Service of GitLab Shell component.
func ShellService(adapter CustomResourceAdapter) *corev1.Service {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: This should return an error instead.
	}

	result := template.Query().ServiceByComponent(GitLabShellComponentName)

	return result
}
