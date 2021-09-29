package gitlab

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
)

var _ = Describe("KAS", func() {

	When("KAS is disabled", func() {
		chartValues := helm.EmptyValues()

		adapter := createMockAdapter(namespace, chartVersions[0], chartValues)

		It("KasEnabled should return false", func() {
			Expect(KasEnabled(adapter)).To(BeFalse())
		})

		It("KAS managed resources must be nil", func() {
			Expect(KasConfigMap(adapter)).To(BeNil())
			Expect(KasDeployment(adapter)).To(BeNil())
			Expect(KasIngress(adapter)).To(BeNil())
			Expect(KasService(adapter)).To(BeNil())
		})
	})

	When("KAS is enabled", func() {
		chartValues := helm.EmptyValues()
		_ = chartValues.SetValue("global.kas.enabled", true)

		adapter := createMockAdapter(namespace, chartVersions[0], chartValues)

		It("KasEnabled should return true", func() {
			Expect(KasEnabled(adapter)).To(BeTrue())
		})

		It("KAS managed resources must be available", func() {
			cfgMap := KasConfigMap(adapter)
			deployment := KasDeployment(adapter)
			ingress := KasIngress(adapter)
			svc := KasService(adapter)

			Expect(cfgMap).NotTo(BeNil())
			Expect(deployment).NotTo(BeNil())
			Expect(ingress).NotTo(BeNil())
			Expect(svc).NotTo(BeNil())

			Expect(cfgMap.Name).To(Equal("test-kas"))
			Expect(deployment.Name).To(Equal("test-kas"))
			Expect(ingress.Name).To(Equal("test-kas"))
			Expect(svc.Name).To(Equal("test-kas"))
		})
	})

})
