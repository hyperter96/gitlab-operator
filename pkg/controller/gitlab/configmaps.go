package gitlab

import (
	"bytes"
	"context"
	"text/template"

	gitlabv1beta1 "gitlab.com/ochienged/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlabutils "gitlab.com/ochienged/gitlab-operator/pkg/controller/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func getGitlabConfig(cr *gitlabv1beta1.Gitlab) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "gitlab", gitlabutils.GitlabType)

	if cr.Spec.ExternalURL == "" {
		cr.Spec.ExternalURL = "http://gitlab.example.com"
	}

	var registryURL string = cr.Spec.Registry.ExternalURL
	if registryURL == "" && cr.Spec.Registry.Enabled {
		registryURL = "http://registry." + DomainNameOnly(cr.Spec.ExternalURL)
	}

	gitlab := gitlabutils.GenericConfigMap(cr.Name+"-gitlab-config", cr.Namespace, labels)
	gitlab.Data = map[string]string{
		"gitlab_external_url":   parseURL(cr.Spec.ExternalURL, hasTLS(cr)),
		"postgres_db":           "gitlab_production",
		"postgres_host":         cr.Name + "-database",
		"postgres_user":         "gitlab",
		"redis_host":            cr.Name + "-redis",
		"registry_external_url": registryURL,
		"installation_type":     labels["app.kubernetes.io/managed-by"],
	}

	return gitlab
}

func getRedisConfig(cr *gitlabv1beta1.Gitlab, s security) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "redis", gitlabutils.GitlabType)

	var redisConf bytes.Buffer

	tmpl := template.Must(template.ParseFiles("/templates/redis.conf"))
	tmpl.Execute(&redisConf, RedisConfig{
		Password: s.RedisPassword(),
		Cluster:  false,
	})

	redis := gitlabutils.GenericConfigMap(cr.Name+"-gitlab-redis", cr.Namespace, labels)
	redis.Data = map[string]string{
		"redis.conf": redisConf.String(),
	}

	return redis
}

func getGitalyConfig(cr *gitlabv1beta1.Gitlab) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "redis", gitlabutils.GitlabType)

	gitaly := gitlabutils.GenericConfigMap(cr.Name+"-gitaly-config", cr.Namespace, labels)

	var shell bytes.Buffer
	shellTemplate := template.Must(template.ParseFiles("/templates/gitaly-shell-config.yml.erb"))
	shellTemplate.Execute(&shell, gitalyShellConfigs(cr))

	gitalyConf := gitlabutils.ReadConfig("/templates/gitaly-config.toml.erb")
	configureScript := gitlabutils.ReadConfig("/templates/gitaly-configure.sh")

	gitaly.Data = map[string]string{
		"config.toml.erb":      gitalyConf,
		"configure":            configureScript,
		"shell-config.yml.erb": shell.String(),
	}

	return gitaly
}

// TODO: Get Minio/Object storage
func getUnicornConfig(cr *gitlabv1beta1.Gitlab) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "unicorn", gitlabutils.GitlabType)
	var database, gitlab, resque bytes.Buffer

	unicorn := gitlabutils.GenericConfigMap(cr.Name+"-unicorn-config", cr.Namespace, labels)

	configure := gitlabutils.ReadConfig("/templates/unicorn-configure.sh")
	unicornRB := gitlabutils.ReadConfig("/templates/unicorn.rb")
	smtpSettings := gitlabutils.ReadConfig("/templates/unicorn-smtp-settings.rb")

	options := UnicornOptions{
		ResourceName: cr.Name,
		Namespace:    cr.Namespace,
		PostgreSQL:   getName(cr.Name, "database"),
		ExternalURL:  parseURL(cr.Spec.ExternalURL, hasTLS(cr)),
		Minio:        "external-minio-instance",
		Registry:     getName(cr.Name, "registry"),
		Gitaly:       getName(cr.Name, "gitaly"),
		RedisMaster:  getName(cr.Name, "redis"),
	}

	databaseTemplate := template.Must(template.ParseFiles("/templates/unicorn-database.yml.erb"))
	databaseTemplate.Execute(&database, options)

	gitlabTemplate := template.Must(template.ParseFiles("/templates/unicorn-gitlab.yml.erb"))
	gitlabTemplate.Execute(&gitlab, options)

	resqueTemplate := template.Must(template.ParseFiles("/templates/unicorn-resque.yml.erb"))
	resqueTemplate.Execute(&resque, options)

	unicorn.Data = map[string]string{
		"configure":         configure,
		"database.yml.erb":  database.String(),
		"gitlab.yml.erb":    gitlab.String(),
		"resque.yml.erb":    resque.String(),
		"installation_type": labels["app.kubernetes.io/managed-by"],
		"smtp_settings.rb":  smtpSettings,
		"unicorn.rb":        unicornRB,
	}

	return unicorn
}

func getWorkhorseConfig(cr *gitlabv1beta1.Gitlab) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "workhorse", gitlabutils.GitlabType)
	var config bytes.Buffer

	workhorse := gitlabutils.GenericConfigMap(cr.Name+"-workhorse-config", cr.Namespace, labels)

	configureSh := gitlabutils.ReadConfig("/templates/workhorse-configure.sh")

	options := WorkhorseOptions{
		RedisMaster: getName(cr.Name, "redis"),
	}

	configTemplate := template.Must(template.ParseFiles("/templates/workhorse-config.toml.erb"))
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

	shellConfigs := gitlabutils.ReadConfig("/templates/shell-config.yml.erb")
	sshdConfig := gitlabutils.ReadConfig("/templates/shell-sshd-config")

	options := ShellOptions{
		Unicorn:     getName(cr.Name, "unicorn"),
		RedisMaster: getName(cr.Name, "redis"),
	}

	configureTemplate := template.Must(template.ParseFiles("/templates/shell-configure.sh"))
	configureTemplate.Execute(&script, options)

	shell := gitlabutils.GenericConfigMap(cr.Name+"-shell-config", cr.Namespace, labels)
	shell.Data = map[string]string{
		"config.yml.erb": shellConfigs,
		"configure":      script.String(),
		"sshd_config":    sshdConfig,
	}

	return shell
}

func getSidekiqConfig(cr *gitlabv1beta1.Gitlab) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "sidekiq", gitlabutils.GitlabType)

	configureScript := gitlabutils.ReadConfig("/templates/sidekiq-configure.sh")
	queuesYML := gitlabutils.ReadConfig("/templates/sidekiq_queues.yml.erb")

	options := SidekiqOptions{
		RedisMaster:    getName(cr.Name, "redis"),
		Postgres:       getName(cr.Name, "database"),
		GitlabDomain:   cr.Spec.ExternalURL,
		EnableRegistry: true,
		EmailFrom:      "gitlab.example.com",
		ReplyTo:        "noreply@example.com",
		MinioDomain:    "minio.example.com",
		Minio:          getName(cr.Name, "minio"),
	}

	var database, gitlab, resque bytes.Buffer
	databaseTemplate := template.Must(template.ParseFiles("/templates/sidekiq-database.yml.erb"))
	databaseTemplate.Execute(&database, options)

	gitlabTemplate := template.Must(template.ParseFiles("/templates/sidekiq-gitlab.yml.erb"))
	gitlabTemplate.Execute(&gitlab, options)

	redisTemplate := template.Must(template.ParseFiles("/templates/sidekiq-resque.yml.erb"))
	redisTemplate.Execute(&resque, options)

	sidekiq := gitlabutils.GenericConfigMap(cr.Name+"-sidekiq-config", cr.Namespace, labels)
	sidekiq.Data = map[string]string{
		"configure":              configureScript,
		"database.yml.erb":       database.String(),
		"resque.yml.erb":         resque.String(),
		"gitlab.yml.erb":         gitlab.String(),
		"smtp_settings.rb":       "",
		"sidekiq_queues.yml.erb": queuesYML,
	}

	return sidekiq
}

func getGitlabExporterConfig(cr *gitlabv1beta1.Gitlab) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "gitlab-exporter", gitlabutils.GitlabType)

	configure := gitlabutils.ReadConfig("/templates/gitlab-exporter-configure.sh")

	options := ExporterOptions{
		RedisMaster: getName(cr.Name, "redis"),
		Postgres:    getName(cr.Name, "database"),
	}
	var exporterYML bytes.Buffer
	exporterTemplate := template.Must(template.ParseFiles("/templates/gitlab-exporter.yml.erb"))
	exporterTemplate.Execute(&exporterYML, options)

	exporter := gitlabutils.GenericConfigMap(cr.Name+"-gitlab-exporter-config", cr.Namespace, labels)
	exporter.Data = map[string]string{
		"configure":               configure,
		"gitlab-exporter.yml.erb": exporterYML.String(),
	}

	return exporter
}

/*
	Reconcilers for all ConfigMaps come below
*/

func (r *ReconcileGitlab) reconcileShellConfigMap(cr *gitlabv1beta1.Gitlab) error {
	shell := getShellConfig(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: shell.Name}, shell) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, shell, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), shell)
}

func (r *ReconcileGitlab) reconcileGitalyConfigMap(cr *gitlabv1beta1.Gitlab) error {
	gitaly := getGitalyConfig(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: gitaly.Name}, gitaly) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, gitaly, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), gitaly)
}

func (r *ReconcileGitlab) reconcileRedisConfigMap(cr *gitlabv1beta1.Gitlab, s security) error {
	redis := getRedisConfig(cr, s)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: redis.Name}, redis) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, redis, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), redis)
}

func (r *ReconcileGitlab) reconcileUnicornConfigMap(cr *gitlabv1beta1.Gitlab) error {
	unicorn := getUnicornConfig(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: unicorn.Name}, unicorn) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, unicorn, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), unicorn)
}

func (r *ReconcileGitlab) reconcileWorkhorseConfigMap(cr *gitlabv1beta1.Gitlab) error {
	workhorse := getWorkhorseConfig(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: workhorse.Name}, workhorse) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, workhorse, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), workhorse)
}

func (r *ReconcileGitlab) reconcileGitlabConfigMap(cr *gitlabv1beta1.Gitlab) error {
	gitlab := getGitlabConfig(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: gitlab.Name}, gitlab) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, gitlab, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), gitlab)
}

func (r *ReconcileGitlab) reconcileSidekiqConfigMap(cr *gitlabv1beta1.Gitlab) error {
	sidekiq := getGitlabConfig(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: sidekiq.Name}, sidekiq) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, sidekiq, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), sidekiq)
}

func (r *ReconcileGitlab) reconcileGitlabExporterConfigMap(cr *gitlabv1beta1.Gitlab) error {
	exporter := getGitlabExporterConfig(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: exporter.Name}, exporter) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, exporter, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), exporter)
}
