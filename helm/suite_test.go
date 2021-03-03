package helm

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

	values := EmptyValues()
	values.AddFromFile(valuesPath)

	template, err := NewBuilder(chartPath).Render(values)

	return template, err
}

func TestHelm(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Helm Suite")
}
