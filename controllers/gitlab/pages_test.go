package gitlab

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab/component"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support"
)

const (
	globalPagesEnabled = "global.pages.enabled"
)

var _ = Describe("Pages", func() {
	When("Pages is enabled", func() {
		chartValues := support.Values{}
		_ = chartValues.SetValue(globalPagesEnabled, true)

		mockGitLab := CreateMockGitLab(releaseName, namespace, chartValues)
		adapter := CreateMockAdapter(mockGitLab)
		template, err := GetTemplate(adapter)

		enabled := adapter.WantsComponent(component.GitLabPages)
		configMap := PagesConfigMap(adapter, template)
		service := PagesService(template)
		deployment := PagesDeployment(template)
		ingress := PagesIngress(template)

		It("Should render the template", func() {
			Expect(err).To(BeNil())
			Expect(template).NotTo(BeNil())
		})

		It("Should contain Pages resources", func() {
			Expect(enabled).To(BeTrue())
			Expect(configMap).NotTo(BeNil())
			Expect(service).NotTo(BeNil())
			Expect(deployment).NotTo(BeNil())
			Expect(ingress).NotTo(BeNil())
		})
	})

	When("Pages is disabled", func() {
		chartValues := support.Values{}
		_ = chartValues.SetValue(globalPagesEnabled, false)

		mockGitLab := CreateMockGitLab(releaseName, namespace, chartValues)
		adapter := CreateMockAdapter(mockGitLab)

		template, err := GetTemplate(adapter)

		enabled := adapter.WantsComponent(component.GitLabPages)
		configMap := PagesConfigMap(adapter, template)
		service := PagesService(template)
		deployment := PagesDeployment(template)
		ingress := PagesIngress(template)

		It("Should render the template", func() {
			Expect(err).To(BeNil())
			Expect(template).NotTo(BeNil())
		})

		It("Should not contain Pages resources", func() {
			Expect(enabled).To(BeFalse())
			Expect(configMap).To(BeNil())
			Expect(service).To(BeNil())
			Expect(deployment).To(BeNil())
			Expect(ingress).To(BeNil())
		})
	})
})
