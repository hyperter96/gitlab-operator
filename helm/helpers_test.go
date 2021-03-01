package helm_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/helm"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Template", func() {

	It("must return all objects when the selector matches all", func() {
		template, err := loadTemplate()
		Expect(err).To(BeNil())

		selectedObjects, err := template.GetObjects(helm.TrueSelector)
		Expect(err).To(BeNil())
		Expect(selectedObjects).To(Equal(template.Objects()))
	})

	It("must only select ConfigMaps that match the expected label", func() {
		template, err := loadTemplate()
		Expect(err).To(BeNil())

		selector := func(configMap *corev1.ConfigMap) bool {
			return configMap.ObjectMeta.Labels["app.kubernetes.io/managed-by"] == "Helm"
		}
		selectedObjects, err := template.GetObjects(helm.NewConfigMapSelector(selector))
		Expect(err).To(BeNil())

		Expect(selectedObjects).To(HaveLen(1))
	})

	It("must delete no object when the selector does not match any", func() {
		template, err := loadTemplate()
		Expect(err).To(BeNil())

		deletedCount, err := template.DeleteObjects(helm.FalseSelector)
		Expect(err).To(BeNil())
		Expect(deletedCount).To(BeZero())
	})

	It("must delete the Ingress objects", func() {
		template, err := loadTemplate()
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

	It("must edit Deployment objects", func() {
		template, err := loadTemplate()
		Expect(err).To(BeNil())

		initialLength := len(template.Objects())
		Expect(initialLength).NotTo(BeZero())

		deployments, err := template.GetObjects(helm.DeploymentSelector)
		Expect(err).To(BeNil())
		Expect(deployments).ToNot(BeEmpty())

		editedCount, err := template.EditObjects(helm.NewDeploymentEditor(
			func(d *appsv1.Deployment) error {
				d.Spec.Paused = true
				if d.ObjectMeta.Annotations == nil {
					d.ObjectMeta.Annotations = map[string]string{}
				}
				d.ObjectMeta.Annotations["gitlab.com/foo"] = "bar"
				return nil
			}),
		)
		Expect(err).To(BeNil())
		Expect(editedCount).NotTo(BeZero())

		for _, o := range deployments {
			deployment, ok := o.(*appsv1.Deployment)
			Expect(ok).To(BeTrue())
			Expect(deployment.Spec.Paused).To(BeTrue())

			foo, ok := deployment.ObjectMeta.Annotations["gitlab.com/foo"]
			Expect(ok).To(BeTrue())
			Expect(foo).To(Equal("bar"))
		}
	})
})
