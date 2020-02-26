package gitlab

import (
	"bytes"
	"io/ioutil"
	"strings"
	"text/template"

	gitlabv1beta1 "github.com/OchiengEd/gitlab-operator/pkg/apis/gitlab/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getGitlabConfig(cr *gitlabv1beta1.Gitlab) *corev1.ConfigMap {
	labels := getLabels(cr, "config")

	omnibus, err := ioutil.ReadFile("templates/omnibus.conf")
	if err != nil {
		log.Error(err, "Error loading config")
	}

	if cr.Spec.ExternalURL == "" {
		cr.Spec.ExternalURL = "http://gitlab.example.com"
	}

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
			"redis_password":        "",
			"registry_external_url": "",
			"omnibus_config":        string(omnibus),
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

func getRedisConfig(cr *gitlabv1beta1.Gitlab, s security) *corev1.ConfigMap {
	labels := getLabels(cr, "redis")
	var redisConf bytes.Buffer

	tmpl := template.Must(template.ParseFiles("/templates/redis.conf"))
	err := tmpl.Execute(&redisConf, RedisConfig{
		Password: s.RedisPassword(),
		Cluster:  false,
	})
	if err != nil {
		log.Error(err, "Error creating redis.conf")
	}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-gitlab-redis",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Data: map[string]string{
			"redis.conf": redisConf.String(),
		},
	}

}
