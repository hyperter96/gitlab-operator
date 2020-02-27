package runner

import (
	gitlabv1beta1 "github.com/OchiengEd/gitlab-operator/pkg/apis/gitlab/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getRunnerSecret(cr *gitlabv1beta1.Runner) *corev1.Secret {
	labels := getLabels(cr, "runner")

	// registrationToken := gitlab.GetSecretValue(cr.Namespace, ,"runner-registration-token")
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels["app.kubernetes.io/name"] + "-secret",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		StringData: map[string]string{
			"runner-registration-token": "",
			"runner-token":              "",
		},
	}
}
