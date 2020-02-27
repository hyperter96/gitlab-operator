package gitlab

import (
	"bytes"
	"text/template"

	gitlabv1beta1 "github.com/OchiengEd/gitlab-operator/pkg/apis/gitlab/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getGitlabConfig(cr *gitlabv1beta1.Gitlab) *corev1.ConfigMap {
	labels := getLabels(cr, "config")
	var omnibus bytes.Buffer

	if cr.Spec.ExternalURL == "" {
		cr.Spec.ExternalURL = "http://gitlab.example.com"
	}

	var registryURL string = cr.Spec.Registry.ExternalURL
	if registryURL == "" && cr.Spec.Registry.Enabled {
		registryURL = "http://registry." + GetDomainNameOnly(cr.Spec.ExternalURL)
	}

	omnibusConf := OmnibusOptions{
		RegistryEnabled:     cr.Spec.Registry.Enabled,
		RegistryExternalURL: registryURL,
	}

	tmpl := template.Must(template.ParseFiles("templates/omnibus.conf"))
	tmpl.Execute(&omnibus, omnibusConf)

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-gitlab-config",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Data: map[string]string{
			"gitlab_external_url":   cr.Spec.ExternalURL,
			"postgres_db":           "gitlab_production",
			"postgres_host":         cr.Name + "-database",
			"postgres_user":         "gitlab",
			"redis_host":            cr.Name + "-redis",
			"registry_external_url": registryURL,
			"gitlab_omnibus_config": omnibus.String(),
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
