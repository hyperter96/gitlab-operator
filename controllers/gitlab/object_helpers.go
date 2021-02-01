package gitlab

import (
	"fmt"

	gitlabutils "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

var (
	localUser int64 = 1000
)

const (
	// GitLabShellComponentName is the common name of GitLab Shell.
	GitLabShellComponentName = "gitlab-shell"

	// TaskRunnerComponentName is the common name of GitLab Task Runner.
	TaskRunnerComponentName = "task-runner"

	// GitLabExporterComponentName is the common name of GitLab Exporter.
	GitLabExporterComponentName = "gitlab-exporter"

	// WebserviceComponentName is the common name of Webservice.
	WebserviceComponentName = "webservice"
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

// ExporterService returns the Service for the GitLab Exporter component.
func ExporterService(adapter CustomResourceAdapter) *corev1.Service {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}
	result := template.Query().ServiceByComponent(GitLabExporterComponentName)

	return patchGitLabExporterService(adapter, result)
}

// ExporterDeployment returns the Deployment for the GitLab Exporter component.
func ExporterDeployment(adapter CustomResourceAdapter) *appsv1.Deployment {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().DeploymentByComponent(GitLabExporterComponentName)

	return patchGitLabExporterDeployment(adapter, result)
}

// ExporterConfigMaps returns the ConfigMaps for the GitLab Exporter component.
func ExporterConfigMaps(adapter CustomResourceAdapter) []*corev1.ConfigMap {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	exporterCfgMap := template.Query().ConfigMapByName(
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), GitLabExporterComponentName))

	result := []*corev1.ConfigMap{exporterCfgMap}

	return patchGitLabExporterConfigMaps(adapter, result)
}

// WebserviceDeployment returns the Deployment for the Webservice component.
func WebserviceDeployment(adapter CustomResourceAdapter) *appsv1.Deployment {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().DeploymentByComponent(WebserviceComponentName)

	return patchWebserviceDeployment(adapter, result)
}

// WebserviceConfigMaps returns the ConfigMaps for the Webservice component.
func WebserviceConfigMaps(adapter CustomResourceAdapter) []*corev1.ConfigMap {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().ConfigMapsByLabels(map[string]string{
		"app": WebserviceComponentName,
	})

	return patchWebserviceConfigMaps(adapter, result)
}

// WebserviceService returns the Service for the Webservice component.
func WebserviceService(adapter CustomResourceAdapter) *corev1.Service {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().ServiceByComponent(WebserviceComponentName)

	return patchWebserviceService(adapter, result)
}

func patchGitLabShellDeployment(adapter CustomResourceAdapter, deployment *appsv1.Deployment) *appsv1.Deployment {
	updateCommonDeployments(GitLabShellComponentName, deployment)

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

func patchGitLabExporterService(adapter CustomResourceAdapter, service *corev1.Service) *corev1.Service {
	updateCommonLabels(adapter.ReleaseName(), GitLabExporterComponentName, &service.ObjectMeta.Labels)
	updateCommonLabels(adapter.ReleaseName(), GitLabExporterComponentName, &service.Spec.Selector)

	return service
}

func patchGitLabExporterDeployment(adapter CustomResourceAdapter, deployment *appsv1.Deployment) *appsv1.Deployment {
	updateCommonDeployments(GitLabExporterComponentName, deployment)

	return deployment
}

func patchGitLabExporterConfigMaps(adapter CustomResourceAdapter, configMaps []*corev1.ConfigMap) []*corev1.ConfigMap {
	for _, c := range configMaps {
		updateCommonLabels(adapter.ReleaseName(), GitLabExporterComponentName, &c.ObjectMeta.Labels)
	}

	return configMaps
}

// TaskRunnerDeployment returns the Deployment of the Task Runner component.
func TaskRunnerDeployment(adapter CustomResourceAdapter) *appsv1.Deployment {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().DeploymentByComponent(TaskRunnerComponentName)

	return patchTaskRunnerDeployment(adapter, result)
}

// TaskRunnerConfigMap returns the ConfigMaps of the Task Runner component.
func TaskRunnerConfigMap(adapter CustomResourceAdapter) *corev1.ConfigMap {
	var result *corev1.ConfigMap
	template, err := GetTemplate(adapter)

	if err != nil {
		return result
		/* WARNING: This should return an error instead. */
	}

	result = template.Query().ConfigMapByName(
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), TaskRunnerComponentName))

	return patchTaskRunnerConfigMap(adapter, result)
}

func patchTaskRunnerDeployment(adapter CustomResourceAdapter, deployment *appsv1.Deployment) *appsv1.Deployment {
	updateCommonDeployments(TaskRunnerComponentName, deployment)

	return deployment
}

func patchTaskRunnerConfigMap(adapter CustomResourceAdapter, configMap *corev1.ConfigMap) *corev1.ConfigMap {
	updateCommonLabels(adapter.ReleaseName(), TaskRunnerComponentName, &configMap.ObjectMeta.Labels)

	return configMap
}

func updateCommonDeployments(componentName string, deployment *appsv1.Deployment) {
	updateCommonLabels(deployment.ObjectMeta.Labels["release"], componentName, &deployment.ObjectMeta.Labels)
	updateCommonLabels(deployment.ObjectMeta.Labels["release"], componentName, &deployment.Spec.Selector.MatchLabels)
	updateCommonLabels(deployment.ObjectMeta.Labels["release"], componentName, &deployment.Spec.Template.ObjectMeta.Labels)

	if deployment.Spec.Template.Spec.SecurityContext == nil {
		deployment.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{}
	}

	var replicas int32 = 1
	var volCfgMapDefaultMode int32 = 420

	deployment.Spec.Replicas = &replicas
	deployment.Spec.Template.Spec.SecurityContext.FSGroup = &localUser
	deployment.Spec.Template.Spec.SecurityContext.RunAsUser = &localUser
	deployment.Spec.Template.Spec.ServiceAccountName = AppServiceAccount
	for _, v := range deployment.Spec.Template.Spec.Volumes {
		if v.VolumeSource.ConfigMap != nil {
			v.VolumeSource.ConfigMap.DefaultMode = &volCfgMapDefaultMode
		}
	}
}

func patchWebserviceService(adapter CustomResourceAdapter, service *corev1.Service) *corev1.Service {
	updateCommonLabels(adapter.ReleaseName(), WebserviceComponentName, &service.ObjectMeta.Labels)
	updateCommonLabels(adapter.ReleaseName(), WebserviceComponentName, &service.Spec.Selector)

	return service
}

func patchWebserviceDeployment(adapter CustomResourceAdapter, deployment *appsv1.Deployment) *appsv1.Deployment {
	updateCommonLabels(adapter.ReleaseName(), WebserviceComponentName, &deployment.ObjectMeta.Labels)
	updateCommonLabels(adapter.ReleaseName(), WebserviceComponentName, &deployment.Spec.Selector.MatchLabels)
	updateCommonLabels(adapter.ReleaseName(), WebserviceComponentName, &deployment.Spec.Template.ObjectMeta.Labels)

	if deployment.Spec.Template.Spec.SecurityContext == nil {
		deployment.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{}
	}

	var replicas int32 = 1
	var volCfgMapDefaultMode int32 = 420

	deployment.Spec.Replicas = &replicas
	deployment.Spec.Template.Spec.SecurityContext.FSGroup = &localUser
	deployment.Spec.Template.Spec.SecurityContext.RunAsUser = &localUser
	deployment.Spec.Template.Spec.ServiceAccountName = AppServiceAccount
	for _, v := range deployment.Spec.Template.Spec.Volumes {
		if v.VolumeSource.ConfigMap != nil {
			v.VolumeSource.ConfigMap.DefaultMode = &volCfgMapDefaultMode
		}
	}

	return deployment
}

func patchWebserviceConfigMaps(adapter CustomResourceAdapter, configMaps []*corev1.ConfigMap) []*corev1.ConfigMap {
	for _, c := range configMaps {
		updateCommonLabels(adapter.ReleaseName(), GitLabExporterComponentName, &c.ObjectMeta.Labels)
	}

	return configMaps
}

func updateCommonLabels(releaseName, componentName string, labels *map[string]string) {
	for k, v := range gitlabutils.Label(releaseName, componentName, gitlabutils.GitlabType) {
		(*labels)[k] = v
	}
}
