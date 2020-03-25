package runner

import (
	"fmt"

	gitlabv1beta1 "gitlab.com/ochienged/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlabutils "gitlab.com/ochienged/gitlab-operator/pkg/controller/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getRunnerScriptConfig(cr *gitlabv1beta1.Runner) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "runner", gitlabutils.RunnerType)

	var gitlabURL string

	configToml := gitlabutils.ReadConfig("/templates/runner-config.toml")
	entrypointScript := gitlabutils.ReadConfig("/templates/runner-entrypoint.sh")
	configureScript := gitlabutils.ReadConfig("/templates/runner-configure.sh")
	registrationScript := gitlabutils.ReadConfig("/templates/runner-registration.sh")
	aliveScript := gitlabutils.ReadConfig("/templates/runner-check-live.sh")

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

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels["app.kubernetes.io/instance"] + "-config",
			Namespace: cr.Namespace,
		},
		Data: map[string]string{
			"ci_server_url":   gitlabURL,
			"config.toml":     configToml,
			"entrypoint":      entrypointScript,
			"check-live":      aliveScript,
			"register-runner": registrationScript,
			"configure":       configureScript,
		},
	}
}
