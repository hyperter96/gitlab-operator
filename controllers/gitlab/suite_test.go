package gitlab

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/settings"
)

func TestGitLab(t *testing.T) {
	settings.Load()
	RegisterFailHandler(Fail)
	RunSpecs(t, "GitLab Suite")
}
