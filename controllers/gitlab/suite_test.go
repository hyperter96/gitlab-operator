package gitlab

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/settings"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support/charts"
)

func TestGitLab(t *testing.T) {
	settings.Load()
	_ = charts.PopulateGlobalCatalog(
		charts.WithSearchPath(settings.HelmChartsDirectory))

	RegisterFailHandler(Fail)
	RunSpecs(t, "GitLab Suite")
}
