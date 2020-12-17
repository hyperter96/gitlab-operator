package gitlab_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGitLab(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GitLab Suite")
}
