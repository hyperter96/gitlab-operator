package helm_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/helm"
)

func loadTemplate() (helm.Template, error) {
	chartPath := os.Getenv("HELM_CHART")
	if chartPath == "" {
		chartPath = "testdata/chart/test"
	}

	valuesPath := os.Getenv("HELM_VALUES")
	if valuesPath == "" {
		valuesPath = "testdata/chart/values.yaml"
	}

	values := helm.EmptyValues()
	values.AddFromFile(valuesPath)

	template, err := helm.NewBuilder(chartPath).Render(values)

	return template, err
}

func TestHelm(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Helm Suite")
}
