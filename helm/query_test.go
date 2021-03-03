package helm

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Query", func() {

	template, err := loadTemplate()

	labels := map[string]string{
		"app.kubernetes.io/managed-by": "Helm",
		"app.kubernetes.io/name":       "test",
	}

	It("must cache the query", func() {
		Expect(err).To(BeNil())

		cache := CacheBackdoor(template.Query())
		saveCacheSize := len(*cache)

		deployments := template.Query().DeploymentsByLabels(labels)
		Expect(len(*cache)).To(Equal(saveCacheSize + 1))

		cachedDeployments := template.Query().DeploymentsByLabels(labels)
		Expect(len(*cache)).To(Equal(saveCacheSize + 1))

		Expect(deployments).To(Equal(cachedDeployments))
	})

	It("must return empty list for a non-existent kind specifiers", func() {
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

	It("must return ConfigMaps that match the labels", func() {
		Expect(err).To(BeNil())

		Expect(template.Query().ConfigMapsByLabels(labels)).
			To(HaveLen(1))
	})

	It("must return empty list when there is no ConfigMap that matchs the labels", func() {
		Expect(err).To(BeNil())

		Expect(template.Query().ConfigMapsByLabels(map[string]string{
			"foo": "bar",
		})).
			To(BeEmpty())
	})

	It("must return nil when there is no ConfigMap that matches the name", func() {
		Expect(err).To(BeNil())

		Expect(template.Query().ConfigMapByName("does-not-exist")).
			To(BeNil())
	})

	It("must return the ConfigMap that matches the name", func() {
		Expect(err).To(BeNil())

		Expect(template.Query().ConfigMapByName("ephemeral-test")).
			NotTo(BeNil())
	})

	It("must return Secrets that match the labels", func() {
		Expect(err).To(BeNil())

		Expect(template.Query().SecretsByLabels(labels)).
			To(HaveLen(1))
	})

	It("must return empty list when there is no Secret that matchs the labels", func() {
		Expect(err).To(BeNil())

		Expect(template.Query().SecretsByLabels(map[string]string{
			"foo": "bar",
		})).
			To(BeEmpty())
	})

	It("must return nil when there is no Secret that matches the name", func() {
		Expect(err).To(BeNil())

		Expect(template.Query().SecretByName("does-not-exist")).
			To(BeNil())
	})

	It("must return the Secret that matches the name", func() {
		Expect(err).To(BeNil())

		Expect(template.Query().SecretByName("ephemeral-test")).
			NotTo(BeNil())
	})

	It("must return Deployments that match the labels", func() {
		Expect(err).To(BeNil())

		Expect(template.Query().DeploymentsByLabels(labels)).
			To(HaveLen(1))
	})

	It("must return empty list when there is no Deployment that matchs the labels", func() {
		Expect(err).To(BeNil())

		Expect(template.Query().DeploymentsByLabels(map[string]string{
			"foo": "bar",
		})).
			To(BeEmpty())
	})

	It("must return nil when there is no Deployment that matches the name", func() {
		Expect(err).To(BeNil())

		Expect(template.Query().DeploymentByName("does-not-exist")).
			To(BeNil())
	})

	It("must return the Deployment that matches the name", func() {
		Expect(err).To(BeNil())

		Expect(template.Query().DeploymentByName("ephemeral-test")).
			NotTo(BeNil())
	})

	It("must return Services that match the labels", func() {
		Expect(err).To(BeNil())

		Expect(template.Query().ServicesByLabels(labels)).
			To(HaveLen(1))
	})

	It("must return empty list when there is no Service that matchs the labels", func() {
		Expect(err).To(BeNil())

		Expect(template.Query().ServicesByLabels(map[string]string{
			"foo": "bar",
		})).
			To(BeEmpty())
	})

	It("must return nil when there is no Service that matches the name", func() {
		Expect(err).To(BeNil())

		Expect(template.Query().ServiceByName("does-not-exist")).
			To(BeNil())
	})

	It("must return the Service that matches the name", func() {
		Expect(err).To(BeNil())

		Expect(template.Query().ServiceByName("ephemeral-test")).
			NotTo(BeNil())
	})
})
