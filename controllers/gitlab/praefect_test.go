package gitlab

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/resource"
)

var _ = Describe("Praefect resources", func() {
	if namespace == "" {
		namespace = "default"
	}

	Context("Praefect", func() {
		When("Praefect is enabled", func() {
			chartValues := resource.Values{}
			_ = chartValues.SetValue(GlobalPraefectEnabled, true)

			mockGitLab := CreateMockGitLab(releaseName, namespace, chartValues)
			adapter := CreateMockAdapter(mockGitLab)
			template, err := GetTemplate(adapter)

			enabled := PraefectEnabled(adapter)
			configMap := PraefectConfigMap(template)
			service := PraefectService(template)
			statefulSet := PraefectStatefulSet(template)

			It("Should render the template", func() {
				Expect(err).To(BeNil())
				Expect(template).NotTo(BeNil())
			})

			It("Should contain Praefect resources", func() {
				Expect(enabled).To(BeTrue())
				Expect(configMap).NotTo(BeNil())
				Expect(service).NotTo(BeNil())
				Expect(statefulSet).NotTo(BeNil())
			})
		})

		When("Praefect is disabled", func() {
			chartValues := resource.Values{}
			_ = chartValues.SetValue(GlobalPraefectEnabled, false)

			mockGitLab := CreateMockGitLab(releaseName, namespace, chartValues)
			adapter := CreateMockAdapter(mockGitLab)
			template, err := GetTemplate(adapter)

			enabled := PraefectEnabled(adapter)
			configMap := PraefectConfigMap(template)
			service := PraefectService(template)
			statefulSet := PraefectStatefulSet(template)

			It("Should render the template", func() {
				Expect(err).To(BeNil())
				Expect(template).NotTo(BeNil())
			})

			It("Should not contain Praefect resources", func() {
				Expect(enabled).To(BeFalse())
				Expect(configMap).To(BeNil())
				Expect(service).To(BeNil())
				Expect(statefulSet).To(BeNil())
			})
		})
	})
})
