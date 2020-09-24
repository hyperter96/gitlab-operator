package runner

import (
	"bytes"
	"fmt"

	"text/template"

	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	gitlabutils "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/utils"
	corev1 "k8s.io/api/core/v1"
)

// Config struct holds the values used to
// configure Runner Global options
type Config struct {
	Concurrent    int32
	CheckInterval int32
}

func userOptions(cr *gitlabv1beta1.Runner) Config {
	options := Config{Concurrent: 10, CheckInterval: 30}

	if cr.Spec.Concurrent != nil {
		options.Concurrent = *cr.Spec.Concurrent
	}

	if cr.Spec.CheckInterval != nil {
		options.CheckInterval = *cr.Spec.CheckInterval
	}

	return options
}

// GetConfigMap returns the runner configmap object
func GetConfigMap(cr *gitlabv1beta1.Runner) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "runner", gitlabutils.RunnerType)

	var gitlabURL string

	var configToml bytes.Buffer
	configTemplate := template.Must(template.ParseFiles("/templates/gitlab-runner/config.toml"))
	configTemplate.Execute(&configToml, userOptions(cr))

	entrypointScript := gitlabutils.ReadConfig("/templates/gitlab-runner/entrypoint.sh")
	configureScript := gitlabutils.ReadConfig("/templates/gitlab-runner/configure.sh")
	registrationScript := gitlabutils.ReadConfig("/templates/gitlab-runner/registration.sh")
	aliveScript := gitlabutils.ReadConfig("/templates/gitlab-runner/check-live.sh")

	// Gitlab URL should be used for Gitlab instances
	// outside k8s or the current namespace
	if cr.Spec.Gitlab.URL != "" {
		gitlabURL = cr.Spec.Gitlab.URL
	}

	// Access via k8s service is preferred if
	// name is provides
	if cr.Spec.Gitlab.Name != "" {
		service := cr.Spec.Gitlab.Name + "-gitlab"
		gitlabURL = fmt.Sprintf("http://%s:8005", service)
	}

	runnerConfigMap := gitlabutils.GenericConfigMap(labels["app.kubernetes.io/instance"]+"-config", cr.Namespace, labels)
	runnerConfigMap.Data = map[string]string{
		"ci_server_url":   gitlabURL,
		"config.toml":     configToml.String(),
		"entrypoint":      entrypointScript,
		"check-live":      aliveScript,
		"register-runner": registrationScript,
		"configure":       configureScript,
	}

	// update configmap with checksum in annotation
	gitlabutils.ConfigMapWithHash(runnerConfigMap)

	return runnerConfigMap
}
