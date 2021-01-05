package gitlab

import (
	"fmt"

	gitlabutils "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	// GitLabShellComponentName is the common name of GitLab Shell.
	GitLabShellComponentName = "gitlab-shell"
)

// ShellDeployment returns the Deployment of GitLab Shell component.
func ShellDeployment(adapter CustomResourceAdapter) *appsv1.Deployment {

	template, err := GetTemplate(adapter)

	if err != nil {
		return nil
		/* WARNING: This should return an error instead. */
	}

	result := template.Query().DeploymentByComponent(GitLabShellComponentName)

	return patchGitLabShellDeployment(adapter, result)
}

// ShellConfigMaps returns the ConfigMaps of GitLab Shell component.
func ShellConfigMaps(adapter CustomResourceAdapter) []*corev1.ConfigMap {
	template, err := GetTemplate(adapter)

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

	return patchGitLabShellConfigMaps(adapter, result)
}

// ShellService returns the Service of GitLab Shell component.
func ShellService(adapter CustomResourceAdapter) *corev1.Service {
	template, err := GetTemplate(adapter)

	if err != nil {
		return nil
		/* WARNING: This should return an error instead. */
	}

	result := template.Query().ServiceByComponent(GitLabShellComponentName)

	return patchGitLabShellService(adapter, result)
}

func patchGitLabShellDeployment(adapter CustomResourceAdapter, deployment *appsv1.Deployment) *appsv1.Deployment {
	updateCommonLabels(adapter.ReleaseName(), GitLabShellComponentName, &deployment.ObjectMeta.Labels)
	updateCommonLabels(adapter.ReleaseName(), GitLabShellComponentName, &deployment.Spec.Selector.MatchLabels)
	updateCommonLabels(adapter.ReleaseName(), GitLabShellComponentName, &deployment.Spec.Template.ObjectMeta.Labels)

	if deployment.Spec.Template.Spec.SecurityContext == nil {
		deployment.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{}
	}

	var userID int64 = 1000
	var replicas int32 = 1
	var volCfgMapDefaultMode int32 = 420

	deployment.Spec.Replicas = &replicas
	deployment.Spec.Template.Spec.SecurityContext.FSGroup = &userID
	deployment.Spec.Template.Spec.SecurityContext.RunAsUser = &userID
	deployment.Spec.Template.Spec.ServiceAccountName = AppServiceAccount
	for _, v := range deployment.Spec.Template.Spec.Volumes {
		if v.VolumeSource.ConfigMap != nil {
			v.VolumeSource.ConfigMap.DefaultMode = &volCfgMapDefaultMode
		}
	}

	return deployment
}

func patchGitLabShellConfigMaps(adapter CustomResourceAdapter, configMaps []*corev1.ConfigMap) []*corev1.ConfigMap {
	for _, c := range configMaps {
		updateCommonLabels(adapter.ReleaseName(), GitLabShellComponentName, &c.ObjectMeta.Labels)
	}

	return configMaps
}

func patchGitLabShellService(adapter CustomResourceAdapter, service *corev1.Service) *corev1.Service {
	updateCommonLabels(adapter.ReleaseName(), GitLabShellComponentName, &service.ObjectMeta.Labels)
	updateCommonLabels(adapter.ReleaseName(), GitLabShellComponentName, &service.Spec.Selector)

	return service
}

func updateCommonLabels(releaseName, componentName string, labels *map[string]string) {
	for k, v := range gitlabutils.Label(releaseName, componentName, gitlabutils.GitlabType) {
		(*labels)[k] = v
	}
}
