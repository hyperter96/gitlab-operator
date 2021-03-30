package runner

import (
	"bytes"
	"fmt"
	"os"

	"text/template"

	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/internal"
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

// ConfigMap returns the runner configmap object
func ConfigMap(cr *gitlabv1beta1.Runner) *corev1.ConfigMap {
	labels := internal.Label(cr.Name, "runner", internal.RunnerType)

	var gitlabURL string

	var configToml bytes.Buffer
	configTemplate := template.Must(template.ParseFiles(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/gitlab-runner/config.toml"))
	configTemplate.Execute(&configToml, userOptions(cr))

	entrypointScript := internal.ReadConfig(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/gitlab-runner/entrypoint.sh")
	configureScript := internal.ReadConfig(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/gitlab-runner/configure.sh")
	registrationScript := internal.ReadConfig(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/gitlab-runner/registration.sh")
	aliveScript := internal.ReadConfig(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/gitlab-runner/check-live.sh")

	// Gitlab URL should be used for Gitlab instances
	// outside k8s or the current namespace
	if cr.Spec.GitLab.URL != "" {
		gitlabURL = cr.Spec.GitLab.URL
	}

	// Access via k8s service is preferred if
	// name is provides
	if cr.Spec.GitLab.Name != "" {
		service := cr.Spec.GitLab.Name + "-gitlab"
		gitlabURL = fmt.Sprintf("http://%s:8005", service)
	}

	runnerConfigMap := internal.GenericConfigMap(labels["app.kubernetes.io/instance"]+"-config", cr.Namespace, labels)
	runnerConfigMap.Data = map[string]string{
		"ci_server_url":   gitlabURL,
		"config.toml":     configToml.String(),
		"entrypoint":      entrypointScript,
		"check-live":      aliveScript,
		"register-runner": registrationScript,
		"configure":       configureScript,
	}

	// update configmap with checksum in annotation
	internal.ConfigMapWithHash(runnerConfigMap)

	return runnerConfigMap
}
