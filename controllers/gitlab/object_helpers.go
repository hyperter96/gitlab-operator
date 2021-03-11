package gitlab

import (
	"fmt"
	"strings"

	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/helpers"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/settings"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/utils"
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

	// RegistryComponentName is the common name of the Registry.
	RegistryComponentName = "registry"

	// WebserviceComponentName is the common name of Webservice.
	WebserviceComponentName = "webservice"

	// SharedSecretsComponentName is the common name of Shared Secrets.
	SharedSecretsComponentName = "shared-secrets"

	// SelfSignedCertsComponentName is the common name of Self Signed Certs.
	SelfSignedCertsComponentName = "shared-secrets-selfsign"

	// GitalyComponentName is the common name of Gitaly.
	GitalyComponentName = "gitaly"

	// SidekiqComponentName is the common name of Sidekiq.
	SidekiqComponentName = "sidekiq"

	// RedisComponentName is the common name of Redis.
	RedisComponentName = "redis"

	// PostgresComponentName is the common name of PostgreSQL.
	PostgresComponentName = "postgresql"
)

var (
	localUser                 int64 = 1000
	deploymentReplicasDefault int32 = 1
)

// ShellDeployment returns the Deployment of GitLab Shell component.
func ShellDeployment(adapter helpers.CustomResourceAdapter) *appsv1.Deployment {
	template, err := helpers.GetTemplate(adapter)
	if err != nil {
		return nil
		/* WARNING: This should return an error instead. */
	}

	result := template.Query().DeploymentByComponent(GitLabShellComponentName)

	return patchGitLabShellDeployment(adapter, result)
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

	return patchGitLabShellConfigMaps(adapter, result)
}

// ShellService returns the Service of GitLab Shell component.
func ShellService(adapter helpers.CustomResourceAdapter) *corev1.Service {
	template, err := helpers.GetTemplate(adapter)
	if err != nil {
		return nil
		/* WARNING: This should return an error instead. */
	}

	result := template.Query().ServiceByComponent(GitLabShellComponentName)

	return patchGitLabShellService(adapter, result)
}

// ExporterService returns the Service for the GitLab Exporter component.
func ExporterService(adapter helpers.CustomResourceAdapter) *corev1.Service {
	template, err := helpers.GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}
	result := template.Query().ServiceByComponent(GitLabExporterComponentName)

	return patchGitLabExporterService(adapter, result)
}

// ExporterDeployment returns the Deployment for the GitLab Exporter component.
func ExporterDeployment(adapter helpers.CustomResourceAdapter) *appsv1.Deployment {
	template, err := helpers.GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().DeploymentByComponent(GitLabExporterComponentName)

	return patchGitLabExporterDeployment(adapter, result)
}

// ExporterConfigMaps returns the ConfigMaps for the GitLab Exporter component.
func ExporterConfigMaps(adapter helpers.CustomResourceAdapter) []*corev1.ConfigMap {
	template, err := helpers.GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	exporterCfgMap := template.Query().ConfigMapByName(
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), GitLabExporterComponentName))

	result := []*corev1.ConfigMap{exporterCfgMap}

	return patchGitLabExporterConfigMaps(adapter, result)
}

// MigrationsConfigMap returns the ConfigMaps of Migrations component.
func MigrationsConfigMap(adapter helpers.CustomResourceAdapter) *corev1.ConfigMap {
	var result *corev1.ConfigMap
	template, err := helpers.GetTemplate(adapter)

	if err != nil {
		return result
		/* WARNING: This should return an error instead. */
	}

	result = template.Query().ConfigMapByName(
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), MigrationsComponentName))

	return patchMigrationsConfigMap(adapter, result)
}

// MigrationsJob returns the Job for Migrations component.
func MigrationsJob(adapter helpers.CustomResourceAdapter) (*batchv1.Job, error) {
	template, err := helpers.GetTemplate(adapter)

	if err != nil {
		return nil, err
	}

	job := template.Query().JobByComponent(MigrationsComponentName)

	return patchMigrationsJob(adapter, job), nil
}

// WebserviceDeployment returns the Deployment for the Webservice component.
func WebserviceDeployment(adapter helpers.CustomResourceAdapter) *appsv1.Deployment {
	template, err := helpers.GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().DeploymentByComponent(WebserviceComponentName)

	return patchWebserviceDeployment(adapter, result)
}

// WebserviceConfigMaps returns the ConfigMaps for the Webservice component.
func WebserviceConfigMaps(adapter helpers.CustomResourceAdapter) []*corev1.ConfigMap {
	template, err := helpers.GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().ConfigMapsByLabels(map[string]string{
		"app": WebserviceComponentName,
	})

	return patchWebserviceConfigMaps(adapter, result)
}

// WebserviceService returns the Service for the Webservice component.
func WebserviceService(adapter helpers.CustomResourceAdapter) *corev1.Service {
	template, err := helpers.GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().ServiceByComponent(WebserviceComponentName)

	return patchWebserviceService(adapter, result)
}

func patchGitLabShellDeployment(adapter helpers.CustomResourceAdapter, deployment *appsv1.Deployment) *appsv1.Deployment {
	updateCommonDeployments(GitLabShellComponentName, deployment)

	return deployment
}

func patchGitLabShellConfigMaps(adapter helpers.CustomResourceAdapter, configMaps []*corev1.ConfigMap) []*corev1.ConfigMap {
	for _, c := range configMaps {
		updateCommonLabels(adapter.ReleaseName(), GitLabShellComponentName, &c.ObjectMeta.Labels)
	}

	return configMaps
}

func patchGitLabShellService(adapter helpers.CustomResourceAdapter, service *corev1.Service) *corev1.Service {
	updateCommonLabels(adapter.ReleaseName(), GitLabShellComponentName, &service.ObjectMeta.Labels)
	updateCommonLabels(adapter.ReleaseName(), GitLabShellComponentName, &service.Spec.Selector)

	return service
}

func patchGitLabExporterService(adapter helpers.CustomResourceAdapter, service *corev1.Service) *corev1.Service {
	updateCommonLabels(adapter.ReleaseName(), GitLabExporterComponentName, &service.ObjectMeta.Labels)
	updateCommonLabels(adapter.ReleaseName(), GitLabExporterComponentName, &service.Spec.Selector)

	return service
}

// SharedSecretsConfigMap returns the ConfigMaps of Shared Secret component.
func SharedSecretsConfigMap(adapter helpers.CustomResourceAdapter) (*corev1.ConfigMap, error) {
	template, err := helpers.GetTemplate(adapter)

	if err != nil {
		return nil, err
	}

	cfgMap := template.Query().ConfigMapByComponent(SharedSecretsComponentName)

	return patchSharedSecretsConfigMap(adapter, cfgMap), nil
}

// SharedSecretsJob returns the Job for Shared Secret component.
func SharedSecretsJob(adapter helpers.CustomResourceAdapter) (*batchv1.Job, error) {
	template, err := helpers.GetTemplate(adapter)

	if err != nil {
		return nil, err
	}

	jobs := template.Query().JobsByLabels(map[string]string{
		"app": SharedSecretsComponentName,
	})

	return patchSharedSecretsJobs(adapter, jobs), nil
}

// SelfSignedCertsJob returns the Job for Self Signed Certificates component.
func SelfSignedCertsJob(adapter helpers.CustomResourceAdapter) (*batchv1.Job, error) {
	template, err := helpers.GetTemplate(adapter)

	if err != nil {
		return nil, err
	}

	jobs := template.Query().JobsByLabels(map[string]string{
		"app": SharedSecretsComponentName,
	})

	return patchSelfSignedCertsJob(adapter, jobs), nil
}

func patchSelfSignedCertsJob(adapter helpers.CustomResourceAdapter, jobs []*batchv1.Job) *batchv1.Job {
	for _, j := range jobs {
		if strings.HasSuffix(j.ObjectMeta.Name, "-selfsign") {
			updateCommonLabels(adapter.ReleaseName(), SelfSignedCertsComponentName, &j.ObjectMeta.Labels)
			updateCommonLabels(adapter.ReleaseName(), SelfSignedCertsComponentName, &j.Spec.Template.Labels)
			return j
		}
	}

	return nil
}

// TaskRunnerDeployment returns the Deployment of the Task Runner component.
func TaskRunnerDeployment(adapter helpers.CustomResourceAdapter) *appsv1.Deployment {
	template, err := helpers.GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().DeploymentByComponent(TaskRunnerComponentName)

	return patchTaskRunnerDeployment(adapter, result)
}

// TaskRunnerConfigMap returns the ConfigMaps of the Task Runner component.
func TaskRunnerConfigMap(adapter helpers.CustomResourceAdapter) *corev1.ConfigMap {
	var result *corev1.ConfigMap
	template, err := helpers.GetTemplate(adapter)

	if err != nil {
		return result
		/* WARNING: This should return an error instead. */
	}

	result = template.Query().ConfigMapByName(
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), TaskRunnerComponentName))

	return patchTaskRunnerConfigMap(adapter, result)
}

func patchGitLabExporterDeployment(adapter helpers.CustomResourceAdapter, deployment *appsv1.Deployment) *appsv1.Deployment {
	updateCommonDeployments(GitLabExporterComponentName, deployment)

	return deployment
}

func patchGitLabExporterConfigMaps(adapter helpers.CustomResourceAdapter, configMaps []*corev1.ConfigMap) []*corev1.ConfigMap {
	for _, c := range configMaps {
		updateCommonLabels(adapter.ReleaseName(), GitLabExporterComponentName, &c.ObjectMeta.Labels)
	}

	return configMaps
}

// GitalyStatefulSet returns the StatefulSet of Gitaly component.
func GitalyStatefulSet(adapter helpers.CustomResourceAdapter) *appsv1.StatefulSet {

	template, err := helpers.GetTemplate(adapter)

	if err != nil {
		return nil
		/* WARNING: This should return an error instead. */
	}

	result := template.Query().StatefulSetByComponent(GitalyComponentName)

	return patchGitalyStatefulSet(adapter, result)
}

// GitalyConfigMap returns the ConfigMap of Gitaly component.
func GitalyConfigMap(adapter helpers.CustomResourceAdapter) *corev1.ConfigMap {
	template, err := helpers.GetTemplate(adapter)

	if err != nil {
		return nil
		/* WARNING: This should return an error instead. */
	}

	result := template.Query().ConfigMapByComponent(GitalyComponentName)

	return patchGitalyConfigMaps(adapter, result)
}

// GitalyService returns the Service of GitLab Shell component.
func GitalyService(adapter helpers.CustomResourceAdapter) *corev1.Service {
	template, err := helpers.GetTemplate(adapter)

	if err != nil {
		return nil
		/* WARNING: This should return an error instead. */
	}

	result := template.Query().ServiceByComponent(GitalyComponentName)

	return patchGitalyService(adapter, result)
}

func patchWebserviceService(adapter helpers.CustomResourceAdapter, service *corev1.Service) *corev1.Service {
	updateCommonLabels(adapter.ReleaseName(), WebserviceComponentName, &service.ObjectMeta.Labels)
	updateCommonLabels(adapter.ReleaseName(), WebserviceComponentName, &service.Spec.Selector)

	return service
}

func patchWebserviceDeployment(adapter helpers.CustomResourceAdapter, deployment *appsv1.Deployment) *appsv1.Deployment {
	updateCommonDeployments(WebserviceComponentName, deployment)

	return deployment
}

func patchWebserviceConfigMaps(adapter helpers.CustomResourceAdapter, configMaps []*corev1.ConfigMap) []*corev1.ConfigMap {
	for _, c := range configMaps {
		updateCommonLabels(adapter.ReleaseName(), GitLabExporterComponentName, &c.ObjectMeta.Labels)
	}

	return configMaps
}

func patchTaskRunnerDeployment(adapter helpers.CustomResourceAdapter, deployment *appsv1.Deployment) *appsv1.Deployment {
	updateCommonDeployments(TaskRunnerComponentName, deployment)

	return deployment
}

func patchTaskRunnerConfigMap(adapter helpers.CustomResourceAdapter, configMap *corev1.ConfigMap) *corev1.ConfigMap {
	updateCommonLabels(adapter.ReleaseName(), TaskRunnerComponentName, &configMap.ObjectMeta.Labels)

	return configMap
}

// RegistryService returns the Service of the Registry component.
func RegistryService(adapter helpers.CustomResourceAdapter) *corev1.Service {
	template, err := helpers.GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}
	result := template.Query().ServiceByComponent(RegistryComponentName)

	return patchRegistryService(adapter, result)
}

func patchRegistryService(adapter helpers.CustomResourceAdapter, service *corev1.Service) *corev1.Service {
	updateCommonLabels(adapter.ReleaseName(), RegistryComponentName, &service.ObjectMeta.Labels)
	updateCommonLabels(adapter.ReleaseName(), RegistryComponentName, &service.Spec.Selector)

	return service
}

// RegistryDeployment returns the Deployment of the Registry component.
func RegistryDeployment(adapter helpers.CustomResourceAdapter) *appsv1.Deployment {
	template, err := helpers.GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().DeploymentByComponent(RegistryComponentName)

	return patchRegistryDeployment(result)
}

func patchRegistryDeployment(deployment *appsv1.Deployment) *appsv1.Deployment {
	updateCommonDeployments(RegistryComponentName, deployment)

	return deployment
}

// RegistryConfigMap returns the ConfigMap of the Registry component.
func RegistryConfigMap(adapter helpers.CustomResourceAdapter) *corev1.ConfigMap {
	template, err := helpers.GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	result := template.Query().ConfigMapByName(
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), RegistryComponentName))

	return patchRegistryConfigMap(adapter, result)
}

func patchRegistryConfigMap(adapter helpers.CustomResourceAdapter, configMap *corev1.ConfigMap) *corev1.ConfigMap {
	updateCommonLabels(adapter.ReleaseName(), RegistryComponentName, &configMap.ObjectMeta.Labels)

	return configMap
}

func patchSharedSecretsConfigMap(adapter helpers.CustomResourceAdapter, configMap *corev1.ConfigMap) *corev1.ConfigMap {
	updateCommonLabels(adapter.ReleaseName(), SharedSecretsComponentName, &configMap.ObjectMeta.Labels)

	return configMap
}

func patchSharedSecretsJobs(adapter helpers.CustomResourceAdapter, jobs []*batchv1.Job) *batchv1.Job {
	for _, j := range jobs {
		if !strings.HasSuffix(j.ObjectMeta.Name, "-selfsign") {
			updateCommonLabels(adapter.ReleaseName(), SharedSecretsComponentName, &j.ObjectMeta.Labels)
			updateCommonLabels(adapter.ReleaseName(), SharedSecretsComponentName, &j.Spec.Template.Labels)
			return j
		}
	}

	return nil
}

func patchGitalyStatefulSet(adapter helpers.CustomResourceAdapter, statefulSet *appsv1.StatefulSet) *appsv1.StatefulSet {
	updateCommonLabels(statefulSet.ObjectMeta.Labels["release"], GitalyComponentName,
		&statefulSet.Spec.Template.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].PodAffinityTerm.LabelSelector.MatchLabels)
	return updateCommonStatefulSets(GitalyComponentName, statefulSet)
}

func patchMigrationsConfigMap(adapter helpers.CustomResourceAdapter, configMap *corev1.ConfigMap) *corev1.ConfigMap {
	updateCommonLabels(adapter.ReleaseName(), MigrationsComponentName, &configMap.ObjectMeta.Labels)

	return configMap
}

func patchMigrationsJob(adapter helpers.CustomResourceAdapter, job *batchv1.Job) *batchv1.Job {
	updateCommonLabels(adapter.ReleaseName(), MigrationsComponentName, &job.ObjectMeta.Labels)

	job.Spec.Template.Spec.ServiceAccountName = settings.AppServiceAccount

	return job
}

func patchGitalyConfigMaps(adapter helpers.CustomResourceAdapter, cfgMap *corev1.ConfigMap) *corev1.ConfigMap {
	updateCommonLabels(adapter.ReleaseName(), GitalyComponentName, &cfgMap.ObjectMeta.Labels)

	return cfgMap
}

func patchGitalyService(adapter helpers.CustomResourceAdapter, service *corev1.Service) *corev1.Service {
	updateCommonLabels(adapter.ReleaseName(), GitalyComponentName, &service.ObjectMeta.Labels)
	updateCommonLabels(adapter.ReleaseName(), GitalyComponentName, &service.Spec.Selector)

	return service
}

// SidekiqDeployment returns the Deployment of the Sidekiq component.
func SidekiqDeployment(adapter helpers.CustomResourceAdapter) *appsv1.Deployment {
	template, err := helpers.GetTemplate(adapter)
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
func SidekiqConfigMaps(adapter helpers.CustomResourceAdapter) []*corev1.ConfigMap {
	template, err := helpers.GetTemplate(adapter)
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

func patchSidekiqConfigMaps(adapter helpers.CustomResourceAdapter, configMaps []*corev1.ConfigMap) []*corev1.ConfigMap {
	for _, c := range configMaps {
		updateCommonLabels(adapter.ReleaseName(), SidekiqComponentName, &c.ObjectMeta.Labels)
	}

	return configMaps
}

// RedisConfigMaps returns the ConfigMaps of the Redis component.
func RedisConfigMaps(adapter helpers.CustomResourceAdapter) []*corev1.ConfigMap {
	template, err := helpers.GetTemplate(adapter)
	if err != nil {
		return []*corev1.ConfigMap{} // WARNING: this should return an error instead.
	}

	result := template.Query().ConfigMapsByLabels(map[string]string{
		"app": RedisComponentName,
	})

	return patchRedisConfigMaps(adapter, result)
}

func patchRedisConfigMaps(adapter helpers.CustomResourceAdapter, configMaps []*corev1.ConfigMap) []*corev1.ConfigMap {
	for _, c := range configMaps {
		updateCommonLabels(adapter.ReleaseName(), RedisComponentName, &c.ObjectMeta.Labels)
	}

	return configMaps
}

// RedisServices returns the Services of the Redis component.
func RedisServices(adapter helpers.CustomResourceAdapter) []*corev1.Service {
	template, err := helpers.GetTemplate(adapter)
	if err != nil {
		return nil
		/* WARNING: This should return an error instead. */
	}

	results := template.Query().ServicesByLabels(map[string]string{
		"app": RedisComponentName,
	})

	return patchRedisServices(adapter, results)
}

func patchRedisServices(adapter helpers.CustomResourceAdapter, services []*corev1.Service) []*corev1.Service {
	for _, s := range services {
		updateCommonLabels(adapter.ReleaseName(), RedisComponentName, &s.ObjectMeta.Labels)
		updateCommonLabels(adapter.ReleaseName(), RedisComponentName, &s.Spec.Selector)
	}

	return services
}

// RedisStatefulSet returns the Statefulset of the Redis component.
func RedisStatefulSet(adapter helpers.CustomResourceAdapter) *appsv1.StatefulSet {
	template, err := helpers.GetTemplate(adapter)
	if err != nil {
		return nil
		/* WARNING: This should return an error instead. */
	}

	result := template.Query().StatefulSetByComponent(RedisComponentName)

	return patchRedisStatefulSet(result)
}

func patchRedisStatefulSet(statefulSet *appsv1.StatefulSet) *appsv1.StatefulSet {
	return updateCommonStatefulSets(RedisComponentName, statefulSet)
}

// PostgresServices returns the Services of the Postgres component.
func PostgresServices(adapter helpers.CustomResourceAdapter) []*corev1.Service {
	template, err := helpers.GetTemplate(adapter)
	if err != nil {
		return nil
		/* WARNING: This should return an error instead. */
	}

	results := template.Query().ServicesByLabels(map[string]string{
		"app": PostgresComponentName,
	})

	return patchPostgresServices(adapter, results)
}

func patchPostgresServices(adapter helpers.CustomResourceAdapter, services []*corev1.Service) []*corev1.Service {
	for _, s := range services {
		updateCommonLabels(adapter.ReleaseName(), PostgresComponentName, &s.ObjectMeta.Labels)
		updateCommonLabels(adapter.ReleaseName(), PostgresComponentName, &s.Spec.Selector)

		// Temporary fix: patch in the namespace because the version of the PostgreSQL chart
		// we use does not specify `namespace` in the template.
		s.ObjectMeta.Namespace = adapter.Namespace()
	}

	return services
}

// PostgresStatefulSet returns the StatefulSet of the PostgreSQL component.
func PostgresStatefulSet(adapter helpers.CustomResourceAdapter) *appsv1.StatefulSet {
	template, err := helpers.GetTemplate(adapter)
	if err != nil {
		return nil
		/* WARNING: This should return an error instead. */
	}

	result := template.Query().StatefulSetByComponent(PostgresComponentName)

	return patchPostgresStatefulSet(adapter, result)
}

func patchPostgresStatefulSet(adapter helpers.CustomResourceAdapter, statefulSet *appsv1.StatefulSet) *appsv1.StatefulSet {
	// Temporary fix: patch in the namespace because the version of the PostgreSQL chart
	// we use does not specify `namespace` in the template.
	statefulSet.ObjectMeta.Namespace = adapter.Namespace()
	return updateCommonStatefulSets(PostgresComponentName, statefulSet)
}

// PostgresConfigMap returns the ConfigMap of the PostgreSQL component.
func PostgresConfigMap(adapter helpers.CustomResourceAdapter) *corev1.ConfigMap {
	template, err := helpers.GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	initDBConfigMap := template.Query().ConfigMapByName(
		fmt.Sprintf("%s-postgresql-init-db", adapter.ReleaseName()))

	return patchPostgresConfigMap(adapter, initDBConfigMap)
}

func patchPostgresConfigMap(adapter helpers.CustomResourceAdapter, configMap *corev1.ConfigMap) *corev1.ConfigMap {
	updateCommonLabels(adapter.ReleaseName(), PostgresComponentName, &configMap.ObjectMeta.Labels)

	return configMap
}

func updateCommonDeployments(componentName string, deployment *appsv1.Deployment) {
	updateCommonLabels(deployment.ObjectMeta.Labels["release"], componentName, &deployment.ObjectMeta.Labels)
	updateCommonLabels(deployment.ObjectMeta.Labels["release"], componentName, &deployment.Spec.Selector.MatchLabels)
	updateCommonLabels(deployment.ObjectMeta.Labels["release"], componentName, &deployment.Spec.Template.ObjectMeta.Labels)

	if deployment.Spec.Template.Spec.SecurityContext == nil {
		deployment.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{}
	}

	deployment.Spec.Replicas = &deploymentReplicasDefault
	deployment.Spec.Template.Spec.SecurityContext.FSGroup = &localUser
	deployment.Spec.Template.Spec.SecurityContext.RunAsUser = &localUser
	deployment.Spec.Template.Spec.ServiceAccountName = settings.AppServiceAccount

	for _, v := range deployment.Spec.Template.Spec.Volumes {
		if v.VolumeSource.ConfigMap != nil {
			v.VolumeSource.ConfigMap.DefaultMode = &utils.ConfigMapDefaultMode
		}
	}
}

func updateCommonStatefulSets(componentName string, statefulSet *appsv1.StatefulSet) *appsv1.StatefulSet {
	updateCommonLabels(statefulSet.ObjectMeta.Labels["release"], componentName, &statefulSet.ObjectMeta.Labels)
	updateCommonLabels(statefulSet.ObjectMeta.Labels["release"], componentName, &statefulSet.Spec.Selector.MatchLabels)
	updateCommonLabels(statefulSet.ObjectMeta.Labels["release"], componentName, &statefulSet.Spec.Template.ObjectMeta.Labels)

	if statefulSet.Spec.VolumeClaimTemplates[0].Labels != nil {
		updateCommonLabels(statefulSet.ObjectMeta.Labels["release"], componentName, &statefulSet.Spec.VolumeClaimTemplates[0].Labels)
	}

	if statefulSet.Spec.Template.Spec.SecurityContext == nil {
		statefulSet.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{}
	}

	statefulSet.Spec.Template.Spec.SecurityContext.FSGroup = &localUser
	statefulSet.Spec.Template.Spec.SecurityContext.RunAsUser = &localUser
	statefulSet.Spec.Template.Spec.ServiceAccountName = settings.AppServiceAccount

	for _, v := range statefulSet.Spec.Template.Spec.Volumes {
		if v.VolumeSource.ConfigMap != nil {
			if v.VolumeSource.ConfigMap.DefaultMode == nil {
				v.VolumeSource.ConfigMap.DefaultMode = &utils.ConfigMapDefaultMode
			}
		}
	}

	return statefulSet
}

func updateCommonLabels(releaseName, componentName string, labels *map[string]string) {
	for k, v := range gitlabutils.Label(releaseName, componentName, gitlabutils.GitlabType) {
		(*labels)[k] = v
	}
}

// CfgMapFromList returns a ConfigMap by name from a list of ConfigMaps.
func CfgMapFromList(name string, cfgMaps []*corev1.ConfigMap) *corev1.ConfigMap {
	for _, cm := range cfgMaps {
		if cm.Name == name {
			return cm
		}
	}

	return nil
}

// SvcFromList returns a Service by name from a list of Services.
func SvcFromList(name string, services []*corev1.Service) *corev1.Service {
	for _, s := range services {
		if s.Name == name {
			return s
		}
	}

	return nil
}
