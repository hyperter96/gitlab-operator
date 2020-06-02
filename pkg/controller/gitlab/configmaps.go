package gitlab

import (
	"bytes"
	"text/template"

	gitlabv1beta1 "gitlab.com/ochienged/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlabutils "gitlab.com/ochienged/gitlab-operator/pkg/controller/utils"
	corev1 "k8s.io/api/core/v1"
)

func getGitlabConfig(cr *gitlabv1beta1.Gitlab) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "gitlab", gitlabutils.GitlabType)

	var registryURL string = cr.Spec.Registry.URL
	if registryURL == "" && !cr.Spec.Registry.Disabled {
		registryURL = getRegistryURL(cr)
	}

	gitlab := gitlabutils.GenericConfigMap(cr.Name+"-gitlab-config", cr.Namespace, labels)
	gitlab.Data = map[string]string{
		"gitlab_external_url":   parseURL(getGitlabURL(cr), hasTLS(cr)),
		"postgres_db":           "gitlabhq_production",
		"postgres_host":         cr.Name + "-postgresql",
		"postgres_user":         "gitlab",
		"redis_host":            cr.Name + "-redis",
		"registry_external_url": registryURL,
		"installation_type":     labels["app.kubernetes.io/managed-by"],
	}

	return gitlab
}

func getRedisConfig(cr *gitlabv1beta1.Gitlab) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "redis", gitlabutils.GitlabType)

	masterConf := gitlabutils.ReadConfig("/templates/redis/master.conf")
	replicaConf := gitlabutils.ReadConfig("/templates/redis/replica.conf")
	redisConf := gitlabutils.ReadConfig("/templates/redis/redis.conf")

	redis := gitlabutils.GenericConfigMap(cr.Name+"-redis-config", cr.Namespace, labels)
	redis.Data = map[string]string{
		"master.conf":  masterConf,
		"redis.conf":   redisConf,
		"replica.conf": replicaConf,
	}

	return redis
}

func getRedisSciptsConfig(cr *gitlabv1beta1.Gitlab) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "redis", gitlabutils.GitlabType)

	localLiveness := gitlabutils.ReadConfig("/templates/redis/liveness_local.sh")
	masterAndLocalLiveness := gitlabutils.ReadConfig("/templates/redis/liveness_local_and_master.sh")
	masterLiveness := gitlabutils.ReadConfig("/templates/redis/liveness_master.sh")
	localReadiness := gitlabutils.ReadConfig("/templates/redis/readiness_local.sh")
	masterAndLocalReadiness := gitlabutils.ReadConfig("/templates/redis/readiness_local_and_master.sh")
	masterReadiness := gitlabutils.ReadConfig("/templates/redis/readiness_master.sh")

	scripts := gitlabutils.GenericConfigMap(cr.Name+"-redis-health-config", cr.Namespace, labels)
	scripts.Data = map[string]string{
		"ping_liveness_local.sh":             localLiveness,
		"ping_liveness_local_and_master.sh":  masterAndLocalLiveness,
		"ping_liveness_master.sh":            masterLiveness,
		"ping_readiness_local.sh":            localReadiness,
		"ping_readiness_local_and_master.sh": masterAndLocalReadiness,
		"ping_readiness_master.sh":           masterReadiness,
	}

	return scripts
}

func getGitalyConfig(cr *gitlabv1beta1.Gitlab) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "redis", gitlabutils.GitlabType)

	gitaly := gitlabutils.GenericConfigMap(cr.Name+"-gitaly-config", cr.Namespace, labels)

	options := GitalyOptions{
		RedisMaster: getName(cr.Name, "redis"),
		Webservice:  getName(cr.Name, "webservice"),
	}

	var shell bytes.Buffer
	shellTemplate := template.Must(template.ParseFiles("/templates/gitaly/shell-config.yml.erb"))
	shellTemplate.Execute(&shell, options)

	gitalyConf := gitlabutils.ReadConfig("/templates/gitaly/config.toml.erb")
	configureScript := gitlabutils.ReadConfig("/templates/gitaly/configure.sh")

	gitaly.Data = map[string]string{
		"config.toml.erb":      gitalyConf,
		"configure":            configureScript,
		"shell-config.yml.erb": shell.String(),
	}

	return gitaly
}

// TODO: Get Minio/Object storage
// TODO 2: Expose .MinioURL
func getWebserviceConfig(cr *gitlabv1beta1.Gitlab) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "webservice", gitlabutils.GitlabType)

	webservice := gitlabutils.GenericConfigMap(cr.Name+"-webservice-config", cr.Namespace, labels)

	configure := gitlabutils.ReadConfig("/templates/webservice/configure.sh")

	options := WebserviceOptions{
		Namespace:   cr.Namespace,
		GitlabURL:   cr.Spec.URL,
		Minio:       getName(cr.Name, "minio"),
		MinioURL:    getMinioURL(cr),
		Registry:    getName(cr.Name, "registry"),
		RegistryURL: getRegistryURL(cr),
		Gitaly:      getName(cr.Name, "gitaly"),
		RedisMaster: getName(cr.Name, "redis"),
		PostgreSQL:  getName(cr.Name, "postgresql"),
	}

	if IsEmailConfigured(cr) {
		options.EmailFrom, options.ReplyTo = setupSMTPOptions(cr)
	}

	var gitlab bytes.Buffer
	gitlabTemplate := template.Must(template.ParseFiles("/templates/webservice/gitlab.yml.erb"))
	gitlabTemplate.Execute(&gitlab, options)

	webservice.Data = map[string]string{
		"configure":         configure,
		"gitlab.yml.erb":    gitlab.String(),
		"database.yml.erb":  getDatabaseConfiguration(options.PostgreSQL),
		"resque.yml.erb":    getRedisConfiguration(options.RedisMaster),
		"cable.yml.erb":     getCableConfiguration(options.RedisMaster),
		"installation_type": labels["app.kubernetes.io/managed-by"],
	}

	return webservice
}

func getWorkhorseConfig(cr *gitlabv1beta1.Gitlab) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "workhorse", gitlabutils.GitlabType)
	var config bytes.Buffer

	workhorse := gitlabutils.GenericConfigMap(cr.Name+"-workhorse-config", cr.Namespace, labels)

	configureSh := gitlabutils.ReadConfig("/templates/workhorse/configure.sh")

	options := WorkhorseOptions{
		RedisMaster: getName(cr.Name, "redis"),
	}

	configTemplate := template.Must(template.ParseFiles("/templates/workhorse/workhorse-config.toml.erb"))
	configTemplate.Execute(&config, options)

	workhorse.Data = map[string]string{
		"configure":                 configureSh,
		"workhorse-config.toml.erb": config.String(),
		"installation_type":         labels["app.kubernetes.io/managed-by"],
	}

	return workhorse
}

func getShellConfig(cr *gitlabv1beta1.Gitlab) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "shell", gitlabutils.GitlabType)
	var script bytes.Buffer

	configureScript := gitlabutils.ReadConfig("/templates/shell/configure.sh")
	sshdConfig := gitlabutils.ReadConfig("/templates/shell/sshd-config")

	options := ShellOptions{
		Webservice:  getName(cr.Name, "webservice"),
		RedisMaster: getName(cr.Name, "redis"),
	}

	configureTemplate := template.Must(template.ParseFiles("/templates/shell/config.yml.erb"))
	configureTemplate.Execute(&script, options)

	shell := gitlabutils.GenericConfigMap(cr.Name+"-shell-config", cr.Namespace, labels)
	shell.Data = map[string]string{
		"configure":      configureScript,
		"config.yml.erb": script.String(),
		"sshd_config":    sshdConfig,
	}

	return shell
}

func getSidekiqConfig(cr *gitlabv1beta1.Gitlab) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "sidekiq", gitlabutils.GitlabType)

	configureScript := gitlabutils.ReadConfig("/templates/sidekiq/configure.sh")
	queuesYML := gitlabutils.ReadConfig("/templates/sidekiq/sidekiq_queues.yml.erb")

	options := SidekiqOptions{
		RedisMaster:    getName(cr.Name, "redis"),
		PostgreSQL:     getName(cr.Name, "postgresql"),
		GitlabURL:      getGitlabURL(cr),
		EnableRegistry: true,
		Registry:       getName(cr.Name, "registry"),
		RegistryURL:    getRegistryURL(cr),
		Gitaly:         getName(cr.Name, "gitaly"),
		Namespace:      cr.Namespace,
		MinioURL:       getMinioURL(cr),
		Minio:          getName(cr.Name, "minio"),
	}

	if IsEmailConfigured(cr) {
		options.EmailFrom, options.ReplyTo = setupSMTPOptions(cr)
	}

	var gitlab bytes.Buffer
	gitlabTemplate := template.Must(template.ParseFiles("/templates/sidekiq/gitlab.yml.erb"))
	gitlabTemplate.Execute(&gitlab, options)

	sidekiq := gitlabutils.GenericConfigMap(cr.Name+"-sidekiq-config", cr.Namespace, labels)
	sidekiq.Data = map[string]string{
		"configure":              configureScript,
		"database.yml.erb":       getDatabaseConfiguration(options.PostgreSQL),
		"resque.yml.erb":         getRedisConfiguration(options.RedisMaster),
		"cable.yml.erb":          getCableConfiguration(options.RedisMaster),
		"gitlab.yml.erb":         gitlab.String(),
		"installation_type":      "gitlab-operator",
		"sidekiq_queues.yml.erb": queuesYML,
	}

	return sidekiq
}

func getGitlabExporterConfig(cr *gitlabv1beta1.Gitlab) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "gitlab-exporter", gitlabutils.GitlabType)

	configure := gitlabutils.ReadConfig("/templates/gitlab-exporter/configure.sh")

	options := ExporterOptions{
		RedisMaster: getName(cr.Name, "redis"),
		Postgres:    getName(cr.Name, "postgresql"),
	}
	var exporterYML bytes.Buffer
	exporterTemplate := template.Must(template.ParseFiles("/templates/gitlab-exporter/gitlab-exporter.yml.erb"))
	exporterTemplate.Execute(&exporterYML, options)

	exporter := gitlabutils.GenericConfigMap(cr.Name+"-gitlab-exporter-config", cr.Namespace, labels)
	exporter.Data = map[string]string{
		"configure":               configure,
		"gitlab-exporter.yml.erb": exporterYML.String(),
	}

	return exporter
}

func getRegistryConfig(cr *gitlabv1beta1.Gitlab) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "registry", gitlabutils.GitlabType)

	configure := gitlabutils.ReadConfig("/templates/registry/configure.sh")

	options := RegistryOptions{
		GitlabURL: getGitlabURL(cr),
		Minio:     getName(cr.Name, "minio"),
	}
	var configYML bytes.Buffer
	registryTemplate := template.Must(template.ParseFiles("/templates/registry/config.yml"))
	registryTemplate.Execute(&configYML, options)

	registry := gitlabutils.GenericConfigMap(cr.Name+"-registry-config", cr.Namespace, labels)
	registry.Data = map[string]string{
		"configure":  configure,
		"config.yml": configYML.String(),
	}

	return registry
}

func getTaskRunnerConfig(cr *gitlabv1beta1.Gitlab) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "task-runner", gitlabutils.GitlabType)

	options := TaskRunnerOptions{
		Namespace:   cr.Namespace,
		GitlabURL:   getGitlabURL(cr),
		Minio:       getName(cr.Name, "minio"),
		RedisMaster: getName(cr.Name, "redis"),
		PostgreSQL:  getName(cr.Name, "postgresql"),
		MinioURL:    getMinioURL(cr),
		Registry:    getName(cr.Name, "registry"),
		RegistryURL: getRegistryURL(cr),
		Gitaly:      getName(cr.Name, "gitaly"),
	}

	if IsEmailConfigured(cr) {
		options.EmailFrom, options.ReplyTo = setupSMTPOptions(cr)
	}

	gsutilconf := gitlabutils.ReadConfig("/templates/task-runner/configure-gsutil.sh")

	var configure, gitlab bytes.Buffer
	configureTemplate := template.Must(template.ParseFiles("/templates/task-runner/configure.sh"))
	configureTemplate.Execute(&configure, options)

	gitlabTemplate := template.Must(template.ParseFiles("/templates/task-runner/gitlab.yml.erb"))
	gitlabTemplate.Execute(&gitlab, options)

	tasker := gitlabutils.GenericConfigMap(cr.Name+"-task-runner-config", cr.Namespace, labels)
	tasker.Data = map[string]string{
		"configure":        configure.String(),
		"configure-gsutil": gsutilconf,
		"gitlab.yml.erb":   gitlab.String(),
		"database.yml.erb": getDatabaseConfiguration(options.PostgreSQL),
		"resque.yml.erb":   getRedisConfiguration(options.RedisMaster),
		"cable.yml.erb":    getCableConfiguration(options.RedisMaster),
	}

	return tasker
}

func getMigrationsConfig(cr *gitlabv1beta1.Gitlab) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "migrations", gitlabutils.GitlabType)

	configure := gitlabutils.ReadConfig("/templates/migration/configure.sh")

	options := MigrationOptions{
		Namespace:   cr.Namespace,
		RedisMaster: getName(cr.Name, "redis"),
		PostgreSQL:  getName(cr.Name, "postgresql"),
		Gitaly:      getName(cr.Name, "gitaly"),
		GitlabURL:   getGitlabURL(cr),
	}

	var gitlab bytes.Buffer
	gitlabTemplate := template.Must(template.ParseFiles("/templates/migration/gitlab.yml.erb"))
	gitlabTemplate.Execute(&gitlab, options)

	migrations := gitlabutils.GenericConfigMap(cr.Name+"-migrations-config", cr.Namespace, labels)
	migrations.Data = map[string]string{
		"configure":        configure,
		"gitlab.yml.erb":   gitlab.String(),
		"database.yml.erb": getDatabaseConfiguration(options.PostgreSQL),
		"resque.yml.erb":   getRedisConfiguration(options.RedisMaster),
		"cable.yml.erb":    getCableConfiguration(options.RedisMaster),
	}

	return migrations
}

func getPostgresInitDBConfig(cr *gitlabv1beta1.Gitlab) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "postgres", gitlabutils.GitlabType)

	script := gitlabutils.ReadConfig("/templates/postgresql/postgresql-pgtrm.sh")

	postgres := gitlabutils.GenericConfigMap(cr.Name+"-postgresql-initdb-config", cr.Namespace, labels)
	postgres.Data = map[string]string{
		"enable_extensions.sh": script,
	}

	return postgres
}

func getMinioScriptConfig(cr *gitlabv1beta1.Gitlab) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "minio", gitlabutils.GitlabType)

	script := gitlabutils.ReadConfig("/templates/jobs/initialize-buckets.sh")

	init := gitlabutils.GenericConfigMap(cr.Name+"-minio-script", cr.Namespace, labels)
	init.Data = map[string]string{
		"initialize": script,
	}

	return init
}

//	Reconciler for all ConfigMaps come below
func (r *ReconcileGitlab) reconcileConfigMaps(cr *gitlabv1beta1.Gitlab) error {
	var configmaps []*corev1.ConfigMap

	shell := getShellConfig(cr)

	gitaly := getGitalyConfig(cr)

	redis := getRedisConfig(cr)

	redisScripts := getRedisSciptsConfig(cr)

	webservice := getWebserviceConfig(cr)

	workhorse := getWorkhorseConfig(cr)

	gitlab := getGitlabConfig(cr)

	sidekiq := getSidekiqConfig(cr)

	exporter := getGitlabExporterConfig(cr)

	registry := getRegistryConfig(cr)

	taskRunner := getTaskRunnerConfig(cr)

	migration := getMigrationsConfig(cr)

	initdb := getPostgresInitDBConfig(cr)

	minio := getMinioScriptConfig(cr)

	configmaps = append(configmaps,
		shell,
		gitaly,
		redis,
		redisScripts,
		webservice,
		workhorse,
		initdb,
		gitlab,
		sidekiq,
		exporter,
		registry,
		taskRunner,
		migration,
		minio,
	)

	for _, cm := range configmaps {
		if err := r.createKubernetesResource(cm, cr); err != nil {
			return err
		}
	}

	return nil
}
