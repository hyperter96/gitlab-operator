package runner

import (
	"github.com/prometheus/common/log"
	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	gitlabutils "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/utils"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetSecret returns the runners secret object
func GetSecret(client client.Client, cr *gitlabv1beta1.Runner) *corev1.Secret {
	labels := gitlabutils.Label(cr.Name, "runner", gitlabutils.RunnerType)
	var gitlabSecret, token string
	var err error

	if cr.Spec.Gitlab.Name != "" {
		gitlabSecret = cr.Spec.Gitlab.Name + "-runner-token-secret"
	}

	if cr.Spec.RegistrationToken != "" {
		// If user provides a secret with registration token
		// set it to the gitlab secret
		gitlabSecret = cr.Spec.RegistrationToken
	}

	for token == "" {
		token, err = gitlabutils.GetSecretValue(client, cr.Namespace, gitlabSecret, "runner-registration-token")
		if err != nil {
			log.Error(err, "Secret not found!")
		}
	}

	runnerSecret := gitlabutils.GenericSecret(labels["app.kubernetes.io/instance"]+"-secret", cr.Namespace, labels)
	runnerSecret.StringData = map[string]string{
		"runner-registration-token": token,
		"runner-token":              "",
	}

	return runnerSecret
}
