package helm

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/resource"
)

func loadTemplate() (Template, error) {
	chartPath := os.Getenv("HELM_CHART")
	if chartPath == "" {
		chartPath = "testdata/chart/test"
	}

	valuesPath := os.Getenv("HELM_VALUES")
	if valuesPath == "" {
		valuesPath = "testdata/chart/values.yaml"
	}

	values := resource.Values{}
	_ = values.AddFromYAMLFile(valuesPath)

	builder, err := NewBuilder(chartPath)
	if err != nil {
		return nil, err
	}

	return builder.Render(values)
}

func TestHelm(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Helm Suite")
}
