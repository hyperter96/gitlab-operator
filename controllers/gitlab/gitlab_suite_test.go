package gitlab_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
)

func GitlabMock() *gitlabv1beta1.GitLab {
	gitlab := &gitlabv1beta1.GitLab{}

	// TODO: Setup the mock object

	return gitlab
}

func TestHelm(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GitLab Suite")
}
