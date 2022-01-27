package gitlab

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/resource"
)

var _ = Describe("KAS", func() {
	When("KAS is disabled", func() {
		chartValues := resource.Values{}
		mockGitLab := CreateMockGitLab(releaseName, namespace, chartValues)
		adapter := CreateMockAdapter(mockGitLab)

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
		chartValues := resource.Values{}
		_ = chartValues.SetValue("global.kas.enabled", true)
		_ = chartValues.SetValue("global.kas.service.apiExternalPort", 8153)

		mockGitLab := CreateMockGitLab(releaseName, namespace, chartValues)
		adapter := CreateMockAdapter(mockGitLab)

		It("KAS managed resources must be available", func() {
			Expect(KasEnabled(adapter)).To(BeTrue())

			cfgMap := KasConfigMap(adapter)
			deployment := KasDeployment(adapter)
			ingress := KasIngress(adapter)
			svc := KasService(adapter)

			Expect(cfgMap).NotTo(BeNil())
			Expect(deployment).NotTo(BeNil())
			Expect(ingress).NotTo(BeNil())
			Expect(svc).NotTo(BeNil())

			Expect(cfgMap.GetName()).To(Equal("test-kas"))
			Expect(deployment.GetName()).To(Equal("test-kas"))
			Expect(ingress.GetName()).To(Equal("test-kas"))
			Expect(svc.GetName()).To(Equal("test-kas"))
		})
	})

})
