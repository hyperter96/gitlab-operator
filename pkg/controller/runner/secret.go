package runner

import (
	gitlabv1beta1 "gitlab.com/ochienged/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlabutils "gitlab.com/ochienged/gitlab-operator/pkg/controller/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getRunnerSecret(client client.Client, cr *gitlabv1beta1.Runner) *corev1.Secret {
	labels := gitlabutils.Label(cr.Name, "runner", gitlabutils.RunnerType)
	var token string

	if cr.Spec.Gitlab.Name != "" {
		gitlabSecret := cr.Spec.Gitlab.Name + "-gitlab-secrets"
		token = gitlabutils.GetSecretValue(client, cr.Namespace, gitlabSecret, "initial_shared_runners_registration_token")
	} else {
		// If the Gitlab Name is not provided, the runner will
		// register using the URL and registration token provided
		token = cr.Spec.Gitlab.RegistrationToken
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels["app.kubernetes.io/instance"] + "-secrets",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		StringData: map[string]string{
			"runner-registration-token": token,
			"runner-token":              "",
		},
	}
}
