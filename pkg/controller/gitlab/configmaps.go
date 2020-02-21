package gitlab

import (
	"bytes"
	"strings"
	"text/template"

	gitlabv1beta1 "github.com/OchiengEd/gitlab-operator/pkg/apis/gitlab/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getGitlabConfig(cr *gitlabv1beta1.Gitlab) *corev1.ConfigMap {
	labels := getLabels(cr, "config")
	var parsedConfig bytes.Buffer

	omnibusConfig := `
	external_url "#{ENV['GITLAB_EXTERNAL_URL']}"

	nginx['enable'] = false
	registry_nginx['enable'] = false
	mattermost_nginx['enable'] = false

	gitlab_workhorse['listen_network'] = 'tcp'
	gitlab_workhorse['listen_addr'] = '0.0.0.0:8005'

	registry['registry_http_addr'] = '0.0.0.0:8105'

	postgresql['enable'] = false
	gitlab_rails['db_host'] = ENV['POSTGRES_HOST']
	gitlab_rails['db_password'] = ENV['POSTGRES_PASSWORD']
	gitlab_rails['db_username'] = ENV['POSTGRES_USER']
	gitlab_rails['db_database'] = ENV['POSTGRES_DB']

	redis['enable'] = false
	gitlab_rails['redis_host'] = '{{ .RedisHost }}'

	manage_accounts['enable'] = true
	manage_storage_directories['manage_etc'] = false

	gitlab_shell['auth_file'] = '/gitlab-data/ssh/authorized_keys'
	git_data_dir '/gitlab-data/git-data'
	gitlab_rails['shared_path'] = '/gitlab-data/shared'
	gitlab_rails['uploads_directory'] = '/gitlab-data/uploads'
	gitlab_ci['builds_directory'] = '/gitlab-data/builds'
	gitlab_rails['trusted_proxies'] = ["10.0.0.0/8","172.16.0.0/12","192.168.0.0/16"]

	prometheus['listen_address'] = '0.0.0.0:9090'
	postgres_exporter['enable'] = true
	postgres_exporter['env'] = {
	  'DATA_SOURCE_NAME' => "user=#{ENV['POSTGRES_USER']} host=gitlab-postgresql port=5432 dbname=#{ENV['POSTGRES_DB']} password=#{ENV['POSTGRES_PASSWORD']} sslmode=disable"
	}
	redis_exporter['enable'] = true
	redis_exporter['flags'] = {
	  'redis.addr' => "{{ .RedisHost }}:6379",
	}
	`

	if cr.Spec.ExternalURL == "" {
		cr.Spec.ExternalURL = "http://gitlab.example.com"
	}

	tmpl := template.Must(template.New("omnibus").Parse(omnibusConfig))
	tmpl.Execute(&parsedConfig, OmnibusConfig{
		RedisHost: cr.Name + "-redis",
	})

	var databaseName string = cr.Name + "_gitlab_production"
	if strings.Contains(databaseName, "-") {
		databaseName = strings.ReplaceAll(cr.Name, "-", "_")
	}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-gitlab-config",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Data: map[string]string{
			"external_url":          cr.Spec.ExternalURL,
			"postgres_db":           databaseName,
			"postgres_host":         cr.Name + "-database",
			"postgres_user":         "gitlab",
			"redis_host":            cr.Name + "-redis",
			"redis_password":        "redixP@sswordx",
			"registry_external_url": "",
			"omnibus_config":        parsedConfig.String(),
		},
	}
}

func getPostgresInitdbConfig(cr *gitlabv1beta1.Gitlab) *corev1.ConfigMap {
	labels := getLabels(cr, "database")

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-postgres-initdb",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Data: map[string]string{
			"create_mattermost_production.sql": "CREATE DATABASE mattermost_production WITH OWNER gitlab;",
		},
	}
}

func getGitlabRunnerConfig(cr *gitlabv1beta1.Gitlab) *corev1.ConfigMap {
	labels := getLabels(cr, "runner")

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-gitlab-runner",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Data: map[string]string{
			"config.toml": "",
			"entrypoint":  "",
		},
	}
}
