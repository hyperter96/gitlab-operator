package helm

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support/charts"
)

func loadTemplate() (Template, error) {
	values := support.Values{}
	_ = values.AddFromYAMLFile("testdata/chart/values.yaml")

	builder, err := NewBuilder("test", "0.1.0")
	if err != nil {
		return nil, err
	}

	return builder.Render(values)
}

func TestHelm(t *testing.T) {
	_ = charts.PopulateGlobalCatalog(
		charts.WithSearchPath("testdata/chart"))

	RegisterFailHandler(Fail)
	RunSpecs(t, "Helm Suite")
}
