package gitlab_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
)

func GitLabMock() *gitlabv1beta1.GitLab {
	namespace := os.Getenv("HELM_NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}

	gitlab := &gitlabv1beta1.GitLab{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: namespace,
			Labels: map[string]string{
				"chart": "gitlab-4.6.5",
			},
		},
		Spec: gitlabv1beta1.GitLabSpec{
			Release: "13.6.3",
		},
	}

	return gitlab
}

func TestGitLab(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GitLab Suite")
}
