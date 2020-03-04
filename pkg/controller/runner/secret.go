package runner

import (
	gitlabv1beta1 "github.com/OchiengEd/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlab "github.com/OchiengEd/gitlab-operator/pkg/controller/gitlab"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getRunnerSecret(cr *gitlabv1beta1.Runner) *corev1.Secret {
	labels := getLabels(cr, "runner")
	var token string

	if cr.Spec.Gitlab.Name != "" {
		gitlabSecret := cr.Spec.Gitlab.Name + "-gitlab-secrets"
		token = string(gitlab.GetSecretValue(cr.Namespace, gitlabSecret, "initial_shared_runners_registration_token"))
	} else {
		// If the Gitlab Name is not provided, the runner will
		// register using the URL and registration token provided
		token = cr.Spec.Gitlab.RegistrationToken
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels["app.kubernetes.io/name"] + "-secrets",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		StringData: map[string]string{
			"runner-registration-token": token,
			"runner-token":              "",
		},
	}
}
