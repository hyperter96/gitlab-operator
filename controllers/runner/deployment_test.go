package runner

import (
	"testing"

	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getTestRunner() *gitlabv1beta1.Runner {
	return &gitlabv1beta1.Runner{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-runner",
			Namespace: "default",
			Labels: map[string]string{
				"purpose": "test",
			},
		},
		Spec: gitlabv1beta1.RunnerSpec{
			Gitlab: gitlabv1beta1.GitlabInstanceSpec{
				URL: "https://gitlab.com",
			},
			RegistrationToken: "runner-token-secret",
			Tags:              "openshift, test",
			HelperImage:       "gitlab.com/gitlab-org/gitlab-runner/gitlab-runner-helper-ubi:latest",
			BuildImage:        "ubuntu:20.04",
		},
	}
}
func TestGetEnvironmentVars(t *testing.T) {
	runner := getTestRunner()
	var tags, helperImg string

	vars := getEnvironmentVars(runner)

	if len(vars) == 0 {
		t.Errorf("Error generating GitLab Runner environment variables")
	}

	for _, envvar := range vars {

		if envvar.Name == "KUBERNETES_HELPER_IMAGE" {
			helperImg = envvar.Value
		}

		if envvar.Name == "RUNNER_TAG_LIST" {
			tags = envvar.Value
		}
	}

	if tags != "openshift, test" {
		t.Log("Error setting Runner tags")
	}

	if helperImg != "gitlab.com/gitlab-org/gitlab-runner/gitlab-runner-helper-ubi:latest" {
		t.Log("Error setting Runner Helper image")
	}
}

func TestGetDeployment(t *testing.T) {

	runner := getTestRunner()

	deployment := GetDeployment(runner)

	if deployment != nil {
		if deployment.Namespace != "default" {
			t.Errorf("Wrong namespace was found")
		}

		// check service account is set for the init container
		if len(deployment.Spec.Template.Spec.InitContainers[0].Env) == 0 {
			t.Errorf("Error setting ENVs for init container")
		}

		// check service account is set for the runner container
		if len(deployment.Spec.Template.Spec.Containers[0].Env) == 0 {
			t.Errorf("Error setting ENVs for Runner container")
		}

		// check that the runner service account is used
		if deployment.Spec.Template.Spec.ServiceAccountName != RunnerServiceAccount {
			t.Errorf("The %s service account was not used", RunnerServiceAccount)
		}
	}
}
