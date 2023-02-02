package gitlab

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support"
)

var _ = Describe("Gitaly and Praefect resources", func() {
	if namespace == "" {
		namespace = testNamespace
	}

	Context("Praefect-managed Gitaly", func() {
		When("Gitaly and Praefect are enabled", func() {
			chartValues := support.Values{}
			_ = chartValues.SetValue(globalGitalyEnabled, true)
			_ = chartValues.SetValue(globalPraefectEnabled, true)

			mockGitLab := CreateMockGitLab(releaseName, namespace, chartValues)
			adapter := CreateMockAdapter(mockGitLab)
			template, err := GetTemplate(adapter)

			configMap := GitalyPraefectConfigMap(template)
			services := GitalyPraefectServices(template)
			statefulSets := GitalyPraefectStatefulSets(template)

			It("Should render the template", func() {
				Expect(err).To(BeNil())
				Expect(template).NotTo(BeNil())
			})

			It("Should contain Praefect-managed Gitaly resources", func() {
				Expect(configMap).NotTo(BeNil())
				Expect(services).NotTo(BeNil())
				Expect(statefulSets).NotTo(BeNil())
			})
		})

		When("Gitaly is enabled and Praefect is disabled", func() {
			chartValues := support.Values{}
			_ = chartValues.SetValue(globalGitalyEnabled, true)
			_ = chartValues.SetValue(globalPraefectEnabled, false)

			mockGitLab := CreateMockGitLab(releaseName, namespace, chartValues)
			adapter := CreateMockAdapter(mockGitLab)
			template, err := GetTemplate(adapter)

			configMap := GitalyPraefectConfigMap(template)
			services := GitalyPraefectServices(template)
			statefulSets := GitalyPraefectStatefulSets(template)

			It("Should render the template", func() {
				Expect(err).To(BeNil())
				Expect(template).NotTo(BeNil())
			})

			It("Should not contain Praefect-managed Gitaly resources", func() {
				Expect(configMap).To(BeNil())
				Expect(services).To(BeNil())
				Expect(statefulSets).To(BeNil())
			})
		})
	})
})
