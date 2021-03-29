package gitlab

import (
	"fmt"

	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/helpers"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// ShellDeployment returns the Deployment of GitLab Shell component.
func ShellDeployment(adapter helpers.CustomResourceAdapter) *appsv1.Deployment {
	template, err := helpers.GetTemplate(adapter)
	if err != nil {
		return nil
		/* WARNING: This should return an error instead. */
	}

	return template.Query().DeploymentByComponent(GitLabShellComponentName)
}

// ShellConfigMaps returns the ConfigMaps of GitLab Shell component.
func ShellConfigMaps(adapter helpers.CustomResourceAdapter) []*corev1.ConfigMap {
	template, err := helpers.GetTemplate(adapter)
	if err != nil {
		return []*corev1.ConfigMap{}
		/* WARNING: This should return an error instead. */
	}

	shellCfgMap := template.Query().ConfigMapByName(
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), GitLabShellComponentName))
	sshdCfgMap := template.Query().ConfigMapByName(
		fmt.Sprintf("%s-%s-sshd", adapter.ReleaseName(), GitLabShellComponentName))

	result := []*corev1.ConfigMap{
		shellCfgMap,
		sshdCfgMap,
	}

	return result
}

// ShellService returns the Service of GitLab Shell component.
func ShellService(adapter helpers.CustomResourceAdapter) *corev1.Service {
	template, err := helpers.GetTemplate(adapter)
	if err != nil {
		return nil
		/* WARNING: This should return an error instead. */
	}

	result := template.Query().ServiceByComponent(GitLabShellComponentName)

	return result
}