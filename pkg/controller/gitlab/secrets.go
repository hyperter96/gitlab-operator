package gitlab

import (
	gitlabv1beta1 "github.com/OchiengEd/gitlab-operator/pkg/apis/gitlab/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getGilabSecret(cr *gitlabv1beta1.Gitlab) *corev1.Secret {
	labels := getLabels(cr, "gitlab")

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-gitlab-secrets",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		StringData: map[string]string{
			"gitlab_root_password":                      "gitlab123",
			"postgres_password":                         "postgres123",
			"initial_shared_runners_registration_token": GeneratePassword(PasswordStrengthMedium),
			"redis_password":                            "redis123",
		},
		Type: corev1.SecretTypeOpaque,
	}
}

func getGilabRunnerSecret(cr *gitlabv1beta1.Gitlab) *corev1.Secret {
	labels := getLabels(cr, "runner")

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-runner-secrets",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		StringData: map[string]string{
			"runner-registration-token": "",
			"runner-token":              "",
		},
		Type: corev1.SecretTypeOpaque,
	}
}
