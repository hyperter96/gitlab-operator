package helm_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/helm"
)

func loadTemplate() (*helm.Template, []error, error) {
	chartPath := os.Getenv("HELM_CHART")
	if chartPath == "" {
		chartPath = "testdata/chart/test"
	}

	valuesPath := os.Getenv("HELM_CHART_VALUES")
	if valuesPath == "" {
		valuesPath = "testdata/chart/values.yaml"
	}

	template := helm.NewTemplate(chartPath)

	values := helm.EmptyValues()
	values.AddFromFile(valuesPath)

	warnings, err := template.Load(values)

	return template, warnings, err
}

var _ = Describe("Template", func() {

	When("Initialized", func() {

		template := helm.NewTemplate("foo")
		helmNamespace := os.Getenv("HELM_NAMESPACE")
		if helmNamespace == "" {
			helmNamespace = "default"
		}

		It("Must be empty and use default settings", func() {
			Expect(template.ChartName()).To(Equal("foo"))
			Expect(template.Namespace()).To(Equal(helmNamespace))
			Expect(template.ReleaseName()).To(Equal(helm.DefaultReleaseName))
			Expect(template.HooksDisabled()).To(BeFalse())
			Expect(template.Objects()).To(BeEmpty())
		})

	})

	When("Uses a chart", func() {

		It("Must render the template and parse objects", func() {
			template, warnings, err := loadTemplate()

			Expect(err).To(BeNil())
			Expect(warnings).To(BeEmpty())
			Expect(template.Objects()).NotTo(BeEmpty())
		})

	})
})
