package gitlab

import (
	gitlabv1beta1 "github.com/OchiengEd/gitlab-operator/pkg/apis/gitlab/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getGilabSecret(cr *gitlabv1beta1.Gitlab, s security) *corev1.Secret {
	labels := getLabels(cr, "gitlab")

	secrets := map[string]string{
		"gitlab_root_password":                      s.GitlabRootPassword(),
		"postgres_password":                         s.PostgresPassword(),
		"initial_shared_runners_registration_token": s.RunnerRegistrationToken(),
		"redis_password":                            s.RedisPassword(),
	}

	if cr.Spec.SMTP.Password != "" {
		secrets["smtp_user_password"] = cr.Spec.SMTP.Password
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-gitlab-secrets",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		StringData: secrets,
		Type:       corev1.SecretTypeOpaque,
	}
}
