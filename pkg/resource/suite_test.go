package resource

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGitlabOperator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GitlabOperator Suite")
}