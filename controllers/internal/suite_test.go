package internal

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGitlabOperator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "controllers/internal")
}
