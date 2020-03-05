package runner

import (
	"bytes"
	"fmt"
	"text/template"

	gitlabv1beta1 "github.com/OchiengEd/gitlab-operator/pkg/apis/gitlab/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getRunnerScriptConfig(cr *gitlabv1beta1.Runner) *corev1.ConfigMap {
	labels := getLabels(cr, "runner")
	var tomlConf, entryScript bytes.Buffer
	var gitlabURL string

	toml := template.Must(template.ParseFiles("/templates/runner-config.toml"))
	toml.Execute(&tomlConf, nil)

	entrypoint := template.Must(template.ParseFiles("/templates/runner-entrypoint.sh"))
	entrypoint.Execute(&entryScript, nil)

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
			"ci_server_url": gitlabURL,
			"entrypoint":    entryScript.String(),
			"config.toml":   tomlConf.String(),
		},
	}
}
