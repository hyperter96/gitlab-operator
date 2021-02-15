package gitlab

import (
	"fmt"
	"strings"

	gitlabutils "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/utils"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	// ManagerServiceAccount is the name of the ServiceAccount that GitLab controller uses.
	ManagerServiceAccount = "gitlab-manager"

	// GitLabShellComponentName is the common name of GitLab Shell.
	GitLabShellComponentName = "gitlab-shell"

	// TaskRunnerComponentName is the common name of GitLab Task Runner.
	TaskRunnerComponentName = "task-runner"

	// MigrationsComponentName is the common name of Migrations.
	MigrationsComponentName = "migrations"

	// GitLabExporterComponentName is the common name of GitLab Exporter.
	GitLabExporterComponentName = "gitlab-exporter"

	// WebserviceComponentName is the common name of Webservice.
	WebserviceComponentName = "webservice"

	// SharedSecretsComponentName is the common name of Shared Secrets.
	SharedSecretsComponentName = "shared-secrets"

	// GitalyComponentName is the common name of Gitaly.
	GitalyComponentName = "gitaly"

	// SidekiqComponentName is the common name of Sidekiq.
	SidekiqComponentName = "sidekiq"

	// LocalUserID is the SecurityContext user.
	LocalUserID = "1000"
)

var (
	localUser  int64 = 1000
	gitalyUser int64 = 1000
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

// SharedSecretsConfigMap returns the ConfigMaps of Shared Secret component.
func SharedSecretsConfigMap(adapter CustomResourceAdapter) (*corev1.ConfigMap, error) {
	template, err := GetTemplate(adapter)

	if err != nil {
		return nil, err
	}

	cfgMap := template.Query().ConfigMapByComponent(SharedSecretsComponentName)

	return patchSharedSecretsConfigMap(adapter, cfgMap), nil
}

// SharedSecretsJob returns the Job for Shared Secret component.
func SharedSecretsJob(adapter CustomResourceAdapter) (*batchv1.Job, error) {
	template, err := GetTemplate(adapter)

	if err != nil {
		return nil, err
	}

	jobs := template.Query().JobsByLabels(map[string]string{
		"app": SharedSecretsComponentName,
	})

	return patchSharedSecretsJobs(adapter, jobs), nil
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

// GitalyStatefulSet returns the StatefulSet of Gitaly component.
func GitalyStatefulSet(adapter CustomResourceAdapter) *appsv1.StatefulSet {

	template, err := GetTemplate(adapter)

	if err != nil {
		return nil
		/* WARNING: This should return an error instead. */
	}

	result := template.Query().StatefulSetByComponent(GitalyComponentName)

	return patchGitalyStatefulSet(adapter, result)
}

// GitalyConfigMap returns the ConfigMap of Gitaly component.
func GitalyConfigMap(adapter CustomResourceAdapter) *corev1.ConfigMap {
	template, err := GetTemplate(adapter)

	if err != nil {
		return nil
		/* WARNING: This should return an error instead. */
	}

	result := template.Query().ConfigMapByComponent(GitalyComponentName)

	return patchGitalyConfigMaps(adapter, result)
}

// GitalyService returns the Service of GitLab Shell component.
func GitalyService(adapter CustomResourceAdapter) *corev1.Service {
	template, err := GetTemplate(adapter)

	if err != nil {
		return nil
		/* WARNING: This should return an error instead. */
	}

	result := template.Query().ServiceByComponent(GitalyComponentName)

	return patchGitalyService(adapter, result)
}

func patchWebserviceService(adapter CustomResourceAdapter, service *corev1.Service) *corev1.Service {
	updateCommonLabels(adapter.ReleaseName(), WebserviceComponentName, &service.ObjectMeta.Labels)
	updateCommonLabels(adapter.ReleaseName(), WebserviceComponentName, &service.Spec.Selector)

	return service
}

func patchWebserviceDeployment(adapter CustomResourceAdapter, deployment *appsv1.Deployment) *appsv1.Deployment {
	updateCommonDeployments(WebserviceComponentName, deployment)

	return deployment
}

func patchWebserviceConfigMaps(adapter CustomResourceAdapter, configMaps []*corev1.ConfigMap) []*corev1.ConfigMap {
	for _, c := range configMaps {
		updateCommonLabels(adapter.ReleaseName(), GitLabExporterComponentName, &c.ObjectMeta.Labels)
	}

	return configMaps
}

func patchTaskRunnerDeployment(adapter CustomResourceAdapter, deployment *appsv1.Deployment) *appsv1.Deployment {
	updateCommonDeployments(TaskRunnerComponentName, deployment)

	return deployment
}

func patchTaskRunnerConfigMap(adapter CustomResourceAdapter, configMap *corev1.ConfigMap) *corev1.ConfigMap {
	updateCommonLabels(adapter.ReleaseName(), TaskRunnerComponentName, &configMap.ObjectMeta.Labels)

	return configMap
}

func patchSharedSecretsConfigMap(adapter CustomResourceAdapter, configMap *corev1.ConfigMap) *corev1.ConfigMap {
	updateCommonLabels(adapter.ReleaseName(), SharedSecretsComponentName, &configMap.ObjectMeta.Labels)

	return configMap
}

func patchSharedSecretsJobs(adapter CustomResourceAdapter, jobs []*batchv1.Job) *batchv1.Job {
	for _, j := range jobs {
		if !strings.HasSuffix(j.ObjectMeta.Name, "-selfsign") {
			updateCommonLabels(adapter.ReleaseName(), SharedSecretsComponentName, &j.ObjectMeta.Labels)
			updateCommonLabels(adapter.ReleaseName(), SharedSecretsComponentName, &j.Spec.Template.Labels)
			return j
		}
	}

	return nil
}

func patchGitalyStatefulSet(adapter CustomResourceAdapter, statefulSet *appsv1.StatefulSet) *appsv1.StatefulSet {
	updateCommonLabels(adapter.ReleaseName(), GitalyComponentName, &statefulSet.ObjectMeta.Labels)
	updateCommonLabels(adapter.ReleaseName(), GitalyComponentName, &statefulSet.Spec.Selector.MatchLabels)
	updateCommonLabels(adapter.ReleaseName(), GitalyComponentName, &statefulSet.Spec.Template.ObjectMeta.Labels)
	updateCommonLabels(adapter.ReleaseName(), GitalyComponentName, &statefulSet.Spec.VolumeClaimTemplates[0].Labels)
	updateCommonLabels(adapter.ReleaseName(), GitalyComponentName,
		&statefulSet.Spec.Template.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].PodAffinityTerm.LabelSelector.MatchLabels)

	if statefulSet.Spec.Template.Spec.SecurityContext == nil {
		statefulSet.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{}
	}

	var volCfgMapDefaultMode int32 = 420

	statefulSet.Spec.Template.Spec.SecurityContext.FSGroup = &gitalyUser
	statefulSet.Spec.Template.Spec.SecurityContext.RunAsUser = &gitalyUser
	statefulSet.Spec.Template.Spec.ServiceAccountName = AppServiceAccount
	for _, v := range statefulSet.Spec.Template.Spec.Volumes {
		if v.VolumeSource.ConfigMap != nil {
			v.VolumeSource.ConfigMap.DefaultMode = &volCfgMapDefaultMode
		}
	}

	return statefulSet
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

func patchGitalyConfigMaps(adapter CustomResourceAdapter, cfgMap *corev1.ConfigMap) *corev1.ConfigMap {
	updateCommonLabels(adapter.ReleaseName(), GitalyComponentName, &cfgMap.ObjectMeta.Labels)

	return cfgMap
}

func patchGitalyService(adapter CustomResourceAdapter, service *corev1.Service) *corev1.Service {
	updateCommonLabels(adapter.ReleaseName(), GitalyComponentName, &service.ObjectMeta.Labels)
	updateCommonLabels(adapter.ReleaseName(), GitalyComponentName, &service.Spec.Selector)

	return service
}

// SidekiqDeployment returns the Deployment of the Sidekiq component.
func SidekiqDeployment(adapter CustomResourceAdapter) *appsv1.Deployment {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().DeploymentByComponent(SidekiqComponentName)

	return patchSidekiqDeployment(result)
}

func patchSidekiqDeployment(deployment *appsv1.Deployment) *appsv1.Deployment {
	updateCommonDeployments(SidekiqComponentName, deployment)

	return deployment
}

// SidekiqConfigMaps returns the ConfigMaps of the Sidekiq component.
func SidekiqConfigMaps(adapter CustomResourceAdapter) []*corev1.ConfigMap {
	template, err := GetTemplate(adapter)
	if err != nil {
		return []*corev1.ConfigMap{} // WARNING: this should return an error instead.
	}

	queueCfgMap := template.Query().ConfigMapByName(
		fmt.Sprintf("%s-%s-%s", adapter.ReleaseName(), SidekiqComponentName, "all-in-1"))
	mainCfgMap := template.Query().ConfigMapByName(
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), SidekiqComponentName))

	result := []*corev1.ConfigMap{
		queueCfgMap,
		mainCfgMap,
	}

	return patchSidekiqConfigMaps(adapter, result)
}

func patchSidekiqConfigMaps(adapter CustomResourceAdapter, configMaps []*corev1.ConfigMap) []*corev1.ConfigMap {
	for _, c := range configMaps {
		updateCommonLabels(adapter.ReleaseName(), SidekiqComponentName, &c.ObjectMeta.Labels)
	}

	return configMaps
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

func updateCommonLabels(releaseName, componentName string, labels *map[string]string) {
	for k, v := range gitlabutils.Label(releaseName, componentName, gitlabutils.GitlabType) {
		(*labels)[k] = v
	}
}
