package runner

import (
	gitlabv1beta1 "github.com/OchiengEd/gitlab-operator/pkg/apis/gitlab/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getRunnerScriptConfig(cr *gitlabv1beta1.Runner) *corev1.ConfigMap {
	labels := getLabels(cr, "runner")

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels["app.kubernetes.io/name"] + "-scripts",
			Namespace: cr.Namespace,
		},
		Data: map[string]string{
			"entrypoint":  "",
			"config.toml": "",
		},
	}
}
