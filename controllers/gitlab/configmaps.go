package gitlab

import (
	"bytes"
	"os"
	"text/template"

	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	gitlabutils "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/utils"
	corev1 "k8s.io/api/core/v1"
)

// GetGitLabConfigMap returns the configmap object for GitLab resources
func GetGitLabConfigMap(cr *gitlabv1beta1.GitLab) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "gitlab", gitlabutils.GitlabType)

	var registryURL string = cr.Spec.Registry.URL
	if registryURL == "" && !cr.Spec.Registry.Disabled {
		registryURL = getRegistryURL(cr)
	}

	gitlab := gitlabutils.GenericConfigMap(cr.Name+"-gitlab-config", cr.Namespace, labels)
	options := SystemBuildOptions(cr)
	gitlab.Data = map[string]string{
		"gitlab_external_url":   parseURL(getGitlabURL(cr), hasTLS(cr)),
		"postgres_db":           "gitlabhq_production",
		"postgres_host":         options.PostgreSQL,
		"postgres_user":         "gitlab",
		"redis_host":            options.RedisMaster,
		"registry_external_url": registryURL,
		"installation_type":     labels["app.kubernetes.io/managed-by"],
	}

	gitlabutils.ConfigMapWithHash(gitlab)

	return gitlab
}

// RedisConfigMap returns the configmap object containing Redis config
func RedisConfigMap(cr *gitlabv1beta1.GitLab) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "redis", gitlabutils.GitlabType)

	masterConf := gitlabutils.ReadConfig(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/redis/master.conf")
	replicaConf := gitlabutils.ReadConfig(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/redis/replica.conf")
	redisConf := gitlabutils.ReadConfig(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/redis/redis.conf")

	redis := gitlabutils.GenericConfigMap(cr.Name+"-redis-config", cr.Namespace, labels)
	redis.Data = map[string]string{
		"master.conf":  masterConf,
		"redis.conf":   redisConf,
		"replica.conf": replicaConf,
	}

	gitlabutils.ConfigMapWithHash(redis)

	return redis
}

// RedisSciptsConfigMap returns the configmap object containing Redis scripts
func RedisSciptsConfigMap(cr *gitlabv1beta1.GitLab) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "redis", gitlabutils.GitlabType)

	localLiveness := gitlabutils.ReadConfig(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/redis/liveness_local.sh")
	masterAndLocalLiveness := gitlabutils.ReadConfig(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/redis/liveness_local_and_master.sh")
	masterLiveness := gitlabutils.ReadConfig(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/redis/liveness_master.sh")
	localReadiness := gitlabutils.ReadConfig(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/redis/readiness_local.sh")
	masterAndLocalReadiness := gitlabutils.ReadConfig(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/redis/readiness_local_and_master.sh")
	masterReadiness := gitlabutils.ReadConfig(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/redis/readiness_master.sh")

	scripts := gitlabutils.GenericConfigMap(cr.Name+"-redis-health-config", cr.Namespace, labels)
	scripts.Data = map[string]string{
		"ping_liveness_local.sh":             localLiveness,
		"ping_liveness_local_and_master.sh":  masterAndLocalLiveness,
		"ping_liveness_master.sh":            masterLiveness,
		"ping_readiness_local.sh":            localReadiness,
		"ping_readiness_local_and_master.sh": masterAndLocalReadiness,
		"ping_readiness_master.sh":           masterReadiness,
	}

	gitlabutils.ConfigMapWithHash(scripts)

	return scripts
}

// GitalyConfigMapDEPRECATED returns the configmap object for Gitaly
func GitalyConfigMapDEPRECATED(cr *gitlabv1beta1.GitLab) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "redis", gitlabutils.GitlabType)

	gitaly := gitlabutils.GenericConfigMap(cr.Name+"-gitaly-config", cr.Namespace, labels)

	options := SystemBuildOptions(cr)

	var shell bytes.Buffer
	shellTemplate := template.Must(template.ParseFiles(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/gitaly/shell-config.yml.erb"))
	shellTemplate.Execute(&shell, options)

	gitalyConf := gitlabutils.ReadConfig(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/gitaly/config.toml.erb")
	configureScript := gitlabutils.ReadConfig(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/gitaly/configure.sh")

	gitaly.Data = map[string]string{
		"config.toml.erb":      gitalyConf,
		"configure":            configureScript,
		"shell-config.yml.erb": shell.String(),
	}

	gitlabutils.ConfigMapWithHash(gitaly)

	return gitaly
}

// WebserviceConfigMapDEPRECATED returns the configmap object for GitLab webservice
func WebserviceConfigMapDEPRECATED(cr *gitlabv1beta1.GitLab) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "webservice", gitlabutils.GitlabType)

	webservice := gitlabutils.GenericConfigMap(cr.Name+"-webservice-config", cr.Namespace, labels)

	configure := gitlabutils.ReadConfig(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/webservice/configure.sh")

	options := SystemBuildOptions(cr)

	var gitlab bytes.Buffer
	gitlabTemplate := template.Must(template.ParseFiles(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/webservice/gitlab.yml.erb"))
	gitlabTemplate.Execute(&gitlab, options)

	webservice.Data = map[string]string{
		"configure":         configure,
		"gitlab.yml.erb":    gitlab.String(),
		"database.yml.erb":  getDatabaseConfiguration(cr),
		"resque.yml.erb":    getRedisConfiguration(cr),
		"cable.yml.erb":     getCableConfiguration(cr),
		"installation_type": labels["app.kubernetes.io/managed-by"],
	}

	gitlabutils.ConfigMapWithHash(webservice)

	return webservice
}

// WorkhorseConfigMap returns the configmap object for GitLab workhorse
func WorkhorseConfigMap(cr *gitlabv1beta1.GitLab) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "workhorse", gitlabutils.GitlabType)
	var config bytes.Buffer

	workhorse := gitlabutils.GenericConfigMap(cr.Name+"-workhorse-config", cr.Namespace, labels)

	configureSh := gitlabutils.ReadConfig(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/workhorse/configure.sh")

	options := SystemBuildOptions(cr)

	configTemplate := template.Must(template.ParseFiles(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/workhorse/workhorse-config.toml.erb"))
	configTemplate.Execute(&config, options)

	workhorse.Data = map[string]string{
		"configure":                 configureSh,
		"workhorse-config.toml.erb": config.String(),
		"installation_type":         labels["app.kubernetes.io/managed-by"],
	}

	gitlabutils.ConfigMapWithHash(workhorse)

	return workhorse
}

// ShellConfigMapDEPRECATED returns the configmap object for GitLab shell
func ShellConfigMapDEPRECATED(cr *gitlabv1beta1.GitLab) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "gitlab-shell", gitlabutils.GitlabType)
	var script bytes.Buffer

	configureScript := gitlabutils.ReadConfig(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/shell/configure.sh")
	sshdConfig := gitlabutils.ReadConfig(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/shell/sshd-config")

	options := SystemBuildOptions(cr)

	configureTemplate := template.Must(template.ParseFiles(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/shell/config.yml.erb"))
	configureTemplate.Execute(&script, options)

	shell := gitlabutils.GenericConfigMap(cr.Name+"-gitlab-shell", cr.Namespace, labels)
	shell.Data = map[string]string{
		"configure":      configureScript,
		"config.yml.erb": script.String(),
		"sshd_config":    sshdConfig,
	}

	gitlabutils.ConfigMapWithHash(shell)

	return shell
}

// SidekiqConfigMap returns the configmap object for GitLab sidekiq
func SidekiqConfigMap(cr *gitlabv1beta1.GitLab) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "sidekiq", gitlabutils.GitlabType)

	configureScript := gitlabutils.ReadConfig(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/sidekiq/configure.sh")
	queuesYML := gitlabutils.ReadConfig(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/sidekiq/sidekiq_queues.yml.erb")

	options := SystemBuildOptions(cr)

	var gitlab bytes.Buffer
	gitlabTemplate := template.Must(template.ParseFiles(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/sidekiq/gitlab.yml.erb"))
	gitlabTemplate.Execute(&gitlab, options)

	sidekiq := gitlabutils.GenericConfigMap(cr.Name+"-sidekiq-config", cr.Namespace, labels)
	sidekiq.Data = map[string]string{
		"configure":              configureScript,
		"database.yml.erb":       getDatabaseConfiguration(cr),
		"resque.yml.erb":         getRedisConfiguration(cr),
		"cable.yml.erb":          getCableConfiguration(cr),
		"gitlab.yml.erb":         gitlab.String(),
		"installation_type":      "gitlab-operator",
		"sidekiq_queues.yml.erb": queuesYML,
	}

	gitlabutils.ConfigMapWithHash(sidekiq)

	return sidekiq
}

// ExporterConfigMapDEPRECATED returns the configmap object for the GitLab Exporter
func ExporterConfigMapDEPRECATED(cr *gitlabv1beta1.GitLab) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "gitlab-exporter", gitlabutils.GitlabType)

	configure := gitlabutils.ReadConfig(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/gitlab-exporter/configure.sh")

	options := SystemBuildOptions(cr)
	var exporterYML bytes.Buffer
	exporterTemplate := template.Must(template.ParseFiles(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/gitlab-exporter/gitlab-exporter.yml.erb"))
	exporterTemplate.Execute(&exporterYML, options)

	exporter := gitlabutils.GenericConfigMap(cr.Name+"-gitlab-exporter-config", cr.Namespace, labels)
	exporter.Data = map[string]string{
		"configure":               configure,
		"gitlab-exporter.yml.erb": exporterYML.String(),
	}

	gitlabutils.ConfigMapWithHash(exporter)

	return exporter
}

// RegistryConfigMap returns configmap object for container Registry
func RegistryConfigMap(cr *gitlabv1beta1.GitLab) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "registry", gitlabutils.GitlabType)

	options := SystemBuildOptions(cr)
	configure := gitlabutils.ReadConfig(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/registry/configure.sh")

	var configYML bytes.Buffer
	registryTemplate := template.Must(template.ParseFiles(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/registry/config.yml"))
	registryTemplate.Execute(&configYML, options)

	registry := gitlabutils.GenericConfigMap(cr.Name+"-registry-config", cr.Namespace, labels)
	registry.Data = map[string]string{
		"configure":  configure,
		"config.yml": configYML.String(),
	}

	gitlabutils.ConfigMapWithHash(registry)

	return registry
}

// TaskRunnerConfigMapDEPRECATED returns configmap object for the TaskRunner deployment
func TaskRunnerConfigMapDEPRECATED(cr *gitlabv1beta1.GitLab) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "task-runner", gitlabutils.GitlabType)

	options := SystemBuildOptions(cr)
	gsutilconf := gitlabutils.ReadConfig(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/task-runner/configure-gsutil.sh")

	var configure, gitlab bytes.Buffer
	configureTemplate := template.Must(template.ParseFiles(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/task-runner/configure.sh"))
	configureTemplate.Execute(&configure, options)

	gitlabTemplate := template.Must(template.ParseFiles(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/task-runner/gitlab.yml.erb"))
	gitlabTemplate.Execute(&gitlab, options)

	tasker := gitlabutils.GenericConfigMap(cr.Name+"-task-runner-config", cr.Namespace, labels)
	tasker.Data = map[string]string{
		"configure":        configure.String(),
		"configure-gsutil": gsutilconf,
		"gitlab.yml.erb":   gitlab.String(),
		"database.yml.erb": getDatabaseConfiguration(cr),
		"resque.yml.erb":   getRedisConfiguration(cr),
		"cable.yml.erb":    getCableConfiguration(cr),
	}

	gitlabutils.ConfigMapWithHash(tasker)

	return tasker
}

// MigrationsConfigMap returns configmap object for the Migration job
func MigrationsConfigMap(cr *gitlabv1beta1.GitLab) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "migrations", gitlabutils.GitlabType)

	options := SystemBuildOptions(cr)
	configure := gitlabutils.ReadConfig(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/migration/configure.sh")

	var gitlab bytes.Buffer
	gitlabTemplate := template.Must(template.ParseFiles(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/migration/gitlab.yml.erb"))
	gitlabTemplate.Execute(&gitlab, options)

	migrations := gitlabutils.GenericConfigMap(cr.Name+"-migrations-config", cr.Namespace, labels)
	migrations.Data = map[string]string{
		"configure":        configure,
		"gitlab.yml.erb":   gitlab.String(),
		"database.yml.erb": getDatabaseConfiguration(cr),
		"resque.yml.erb":   getRedisConfiguration(cr),
		"cable.yml.erb":    getCableConfiguration(cr),
	}

	gitlabutils.ConfigMapWithHash(migrations)

	return migrations
}

// PostgresInitDBConfigMap returns configmap object containing Postgresql init scripts
func PostgresInitDBConfigMap(cr *gitlabv1beta1.GitLab) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "postgres", gitlabutils.GitlabType)

	script := gitlabutils.ReadConfig(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/postgresql/postgresql-pgtrm.sh")

	postgres := gitlabutils.GenericConfigMap(cr.Name+"-postgresql-initdb-config", cr.Namespace, labels)
	postgres.Data = map[string]string{
		"enable_extensions.sh": script,
	}

	gitlabutils.ConfigMapWithHash(postgres)

	return postgres
}
