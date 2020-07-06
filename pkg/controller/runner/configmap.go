package runner

import (
	"fmt"

	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlabutils "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/pkg/controller/utils"
	corev1 "k8s.io/api/core/v1"
)

func getRunnerConfigMap(cr *gitlabv1beta1.Runner) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "runner", gitlabutils.RunnerType)

	var gitlabURL string

	configToml := gitlabutils.ReadConfig("/templates/gitlab-runner/config.toml")
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
		"config.toml":     configToml,
		"entrypoint":      entrypointScript,
		"check-live":      aliveScript,
		"register-runner": registrationScript,
		"configure":       configureScript,
	}

	return runnerConfigMap
}
