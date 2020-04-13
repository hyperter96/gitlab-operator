package runner

import (
	gitlabv1beta1 "gitlab.com/ochienged/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlabutils "gitlab.com/ochienged/gitlab-operator/pkg/controller/utils"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getRunnerSecret(client client.Client, cr *gitlabv1beta1.Runner) *corev1.Secret {
	labels := gitlabutils.Label(cr.Name, "runner", gitlabutils.RunnerType)
	var gitlabSecret string

	if cr.Spec.Gitlab.Name != "" {
		gitlabSecret = cr.Spec.Gitlab.Name + "-gitlab-secrets"
	}

	if cr.Spec.RegistrationToken != "" {
		// If user provides a secret with registration token
		// set it to the gitlab secret
		gitlabSecret = cr.Spec.RegistrationToken
	}

	token, err := gitlabutils.GetSecretValue(client, cr.Namespace, gitlabSecret, "runner_registration_token")
	if err != nil {
		log.Error(err, "Secret not found!")
	}

	runnerSecret := gitlabutils.GenericSecret(labels["app.kubernetes.io/instance"]+"-secret", cr.Namespace, labels)
	runnerSecret.StringData = map[string]string{
		"runner-registration-token": token,
		"runner-token":              "",
	}

	return runnerSecret
}
