package helm_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/helm"
)

var _ = Describe("Query", func() {

	template, err := loadTemplate()
	cache := helm.CacheBackdoor(template.Query())

	labels := map[string]string{
		"app.kubernetes.io/managed-by": "Helm",
		"app.kubernetes.io/name":       "test",
	}

	It("must return nothing for a non-existent kind specifiers", func() {
		Expect(err).To(BeNil())

		Expect(template.Query().ObjectsByKind("Ingress.v1.networking.k8s.io")).To(BeEmpty())
	})

	It("must return the same objects with different equivalent kind specifiers", func() {
		Expect(err).To(BeNil())

		list1 := template.Query().ObjectsByKind("Ingress")
		list2 := template.Query().ObjectsByKind("Ingress.networking.k8s.io")
		list3 := template.Query().ObjectsByKind("Ingress.v1beta1.networking.k8s.io")

		Expect(list1).To(Equal(list2))
		Expect(list2).To(Equal(list3))
	})

	It("must return the same object that matches the kind specifier and has the specified name", func() {
		Expect(err).To(BeNil())

		obj1 := template.Query().ObjectByKindAndName("Deployment", "ephemeral-test")
		obj2 := template.Query().ObjectByKindAndName("Deployment.apps", "ephemeral-test")
		obj3 := template.Query().ObjectByKindAndName("Deployment.v1.apps", "ephemeral-test")

		Expect(obj1).To(Equal(obj2))
		Expect(obj2).To(Equal(obj3))
	})

	It("must return the same object that matches the kind specifier and has the specified labels", func() {
		Expect(err).To(BeNil())

		list1 := template.Query().ObjectsByKindAndLabels("ConfigMap", labels)
		list2 := template.Query().ObjectsByKindAndLabels("ConfigMap.v1", labels)

		Expect(list1).To(Equal(list2))
	})

	It("must return Deployments that match the labels", func() {
		Expect(err).To(BeNil())

		saveCacheSize := len(*cache)

		deployments := template.Query().DeploymentsByLabels(labels)
		Expect(deployments).To(HaveLen(1))
		Expect(len(*cache)).To(Equal(saveCacheSize + 1))

		cachedDeployments := template.Query().DeploymentsByLabels(labels)
		Expect(deployments).To(Equal(cachedDeployments))
		Expect(len(*cache)).To(Equal(saveCacheSize + 1))
	})
})
