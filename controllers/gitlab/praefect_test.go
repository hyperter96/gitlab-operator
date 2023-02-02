package gitlab

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab/component"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support"
)

const (
	globalPraefectEnabled = "global.praefect.enabled"
)

var _ = Describe("Praefect resources", func() {
	if namespace == "" {
		namespace = testNamespace
	}

	Context("Praefect", func() {
		When("Praefect is enabled", func() {
			chartValues := support.Values{}
			_ = chartValues.SetValue(globalPraefectEnabled, true)

			mockGitLab := CreateMockGitLab(releaseName, namespace, chartValues)
			adapter := CreateMockAdapter(mockGitLab)
			template, err := GetTemplate(adapter)

			enabled := adapter.WantsComponent(component.Praefect)
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
			chartValues := support.Values{}
			_ = chartValues.SetValue(globalPraefectEnabled, false)

			mockGitLab := CreateMockGitLab(releaseName, namespace, chartValues)
			adapter := CreateMockAdapter(mockGitLab)
			template, err := GetTemplate(adapter)

			enabled := adapter.WantsComponent(component.Praefect)
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
