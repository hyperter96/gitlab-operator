package helm_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/pkg/helm"
	apps "k8s.io/api/apps/v1"
)

var _ = Describe("Template", func() {

	When("Initialized", func() {

		template := helm.NewTemplate("foo")

		It("Must be empty and use default settings", func() {
			Expect(template.ChartName()).To(Equal("foo"))
			Expect(template.Namespace()).To(Equal("default"))
			Expect(template.ReleaseName()).To(Equal(helm.DefaultReleaseName))
			Expect(template.HooksDisabled()).To(BeFalse())
			Expect(template.Objects()).To(BeEmpty())
		})

	})

	When("Uses a chart", func() {

		loadTemplate := func() (*helm.Template, []error, error) {
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

		It("Must render the template and parse objects", func() {
			template, warnings, err := loadTemplate()

			Expect(err).To(BeNil())
			Expect(warnings).To(BeEmpty())
			Expect(template.Objects()).NotTo(BeEmpty())
		})

		It("Must return all objects when the selector matches all", func() {
			template, _, err := loadTemplate()
			Expect(err).To(BeNil())

			selectedObjects, err := template.GetObjects(helm.AnySelector)
			Expect(err).To(BeNil())
			Expect(selectedObjects).To(Equal(template.Objects()))
		})

		It("Must delete no object when the selector does not match any", func() {
			template, _, err := loadTemplate()
			Expect(err).To(BeNil())

			deletedCount, err := template.DeleteObjects(helm.NoneSelector)
			Expect(err).To(BeNil())
			Expect(deletedCount).To(BeZero())
		})

		It("Must delete the Ingress objects", func() {
			template, _, err := loadTemplate()
			Expect(err).To(BeNil())

			initialLength := len(template.Objects())
			Expect(initialLength).NotTo(BeZero())

			ingresses, err := template.GetObjects(helm.IngressSelector)
			Expect(err).To(BeNil())
			Expect(ingresses).ToNot(BeEmpty())

			deletedCount, err := template.DeleteObjects(helm.IngressSelector)
			Expect(err).To(BeNil())
			Expect(deletedCount).ToNot(BeZero())

			Expect(len(template.Objects())).To(Equal(initialLength - deletedCount))
		})

		It("Must edit Deployment objects", func() {
			template, _, err := loadTemplate()
			Expect(err).To(BeNil())

			initialLength := len(template.Objects())
			Expect(initialLength).NotTo(BeZero())

			deployments, err := template.GetObjects(helm.DeploymentSelector)
			Expect(err).To(BeNil())
			Expect(deployments).ToNot(BeEmpty())

			editedCount, err := template.EditObjects(helm.DeploymentSelector,
				helm.NewDeploymentEditor(func(d *apps.Deployment) error {
					d.Spec.Paused = true
					if d.ObjectMeta.Annotations == nil {
						d.ObjectMeta.Annotations = map[string]string{}
					}
					d.ObjectMeta.Annotations["gitlab.com/foo"] = "bar"
					return nil
				}))
			Expect(err).To(BeNil())
			Expect(editedCount).NotTo(BeZero())

			for _, o := range deployments {
				deployment, ok := (*o).(*apps.Deployment)
				Expect(ok).To(BeTrue())
				Expect(deployment.Spec.Paused).To(BeTrue())

				foo, ok := deployment.ObjectMeta.Annotations["gitlab.com/foo"]
				Expect(ok).To(BeTrue())
				Expect(foo).To(Equal("bar"))
			}
		})
	})
})
