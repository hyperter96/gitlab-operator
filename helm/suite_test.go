package helm

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support/charts"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support/charts/populate"
)

func loadTemplate() (Template, error) {
	values := support.Values{}
	_ = values.AddFromYAMLFile("testdata/chart/values.yaml")

	builder, err := NewBuilder(charts.GlobalCatalog())
	if err != nil {
		return nil, err
	}

	return builder.Render(values)
}

func TestHelm(t *testing.T) {
	_ = charts.PopulateGlobalCatalog(
		populate.WithSearchPath("testdata/chart"))

	RegisterFailHandler(Fail)
	RunSpecs(t, "Helm Suite")
}
