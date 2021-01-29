package gitlab

import (
	"fmt"

	gitlabutils "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/utils"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	// GitLabShellComponentName is the common name of GitLab Shell.
	GitLabShellComponentName = "gitlab-shell"

	// TaskRunnerComponentName is the common name of GitLab Task Runner.
	TaskRunnerComponentName = "task-runner"

	// MigrationsComponentName is the common name of Migrations.
	MigrationsComponentName = "migrations"

	// GitLabExporterComponentName is the common name of GitLab Exporter.
	GitLabExporterComponentName = "gitlab-exporter"
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

// MigrationsConfigMap returns the ConfigMaps of Migrations component.
func MigrationsConfigMap(adapter CustomResourceAdapter) *corev1.ConfigMap {
	var result *corev1.ConfigMap
	template, err := GetTemplate(adapter)

	if err != nil {
		return result
		/* WARNING: This should return an error instead. */
	}

	result = template.Query().ConfigMapByName(
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), MigrationsComponentName))

	return patchMigrationsConfigMap(adapter, result)
}

// MigrationsJob returns the Job for Migrations component.
func MigrationsJob(adapter CustomResourceAdapter) (*batchv1.Job, error) {
	template, err := GetTemplate(adapter)

	if err != nil {
		return nil, err
	}

	job := template.Query().JobByComponent(MigrationsComponentName)

	return patchMigrationsJob(adapter, job), nil
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
}

func patchMigrationsConfigMap(adapter CustomResourceAdapter, configMap *corev1.ConfigMap) *corev1.ConfigMap {
	updateCommonLabels(adapter.ReleaseName(), MigrationsComponentName, &configMap.ObjectMeta.Labels)

	return configMap
}

func patchMigrationsJob(adapter CustomResourceAdapter, job *batchv1.Job) *batchv1.Job {
	updateCommonLabels(adapter.ReleaseName(), MigrationsComponentName, &job.ObjectMeta.Labels)

	job.Spec.Template.Spec.ServiceAccountName = AppServiceAccount

	return job
}

func updateCommonLabels(releaseName, componentName string, labels *map[string]string) {
	for k, v := range gitlabutils.Label(releaseName, componentName, gitlabutils.GitlabType) {
		(*labels)[k] = v
	}
}
