package gitlab

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support"
)

var _ = Describe("Gitaly resources", func() {
	if namespace == "" {
		namespace = "default"
	}

	Context("Gitaly", func() {
		When("Gitaly is enabled", func() {
			chartValues := support.Values{}
			_ = chartValues.SetValue(GlobalGitalyEnabled, true)

			mockGitLab := CreateMockGitLab(releaseName, namespace, chartValues)
			adapter := CreateMockAdapter(mockGitLab)
			template, err := GetTemplate(adapter)

			enabled := GitalyEnabled(adapter)
			configMap := GitalyConfigMap(template)
			service := GitalyService(template)
			statefulSet := GitalyStatefulSet(template)

			It("Should render the template", func() {
				Expect(err).To(BeNil())
				Expect(template).NotTo(BeNil())
			})

			It("Should contain Gitaly resources", func() {
				Expect(enabled).To(BeTrue())
				Expect(configMap).NotTo(BeNil())
				Expect(service).NotTo(BeNil())
				Expect(statefulSet).NotTo(BeNil())
			})
		})

		When("Gitaly and Praefect are enabled", func() {
			chartValues := support.Values{}
			_ = chartValues.SetValue(GlobalGitalyEnabled, true)
			_ = chartValues.SetValue(GlobalPraefectEnabled, true)

			mockGitLab := CreateMockGitLab(releaseName, namespace, chartValues)
			adapter := CreateMockAdapter(mockGitLab)
			template, err := GetTemplate(adapter)

			enabled := GitalyEnabled(adapter)
			configMap := GitalyConfigMap(template)
			service := GitalyService(template)
			statefulSet := GitalyStatefulSet(template)

			It("Should render the template", func() {
				Expect(err).To(BeNil())
				Expect(template).NotTo(BeNil())
			})

			It("Should not contain Gitaly resources", func() {
				Expect(enabled).To(BeTrue())
				Expect(configMap).To(BeNil())
				Expect(service).To(BeNil())
				Expect(statefulSet).To(BeNil())
			})
		})

		When("Gitaly and Praefect is enabled and replaceInternalGitaly is false", func() {
			chartValues := support.Values{}
			_ = chartValues.SetValue(GlobalGitalyEnabled, true)
			_ = chartValues.SetValue(GlobalPraefectEnabled, true)
			_ = chartValues.SetValue(GlobalPraefectReplaceInternalGitalyEnabled, false)

			_ = chartValues.SetValue(GlobalPraefectVirtualStorages, []map[string]interface{}{
				{
					"name":           "virtualstorage2",
					"gitalyReplicas": 5,
					"maxUnavailable": 2,
				},
			})

			mockGitLab := CreateMockGitLab(releaseName, namespace, chartValues)
			adapter := CreateMockAdapter(mockGitLab)
			template, err := GetTemplate(adapter)

			enabled := GitalyEnabled(adapter)
			configMap := GitalyConfigMap(template)
			service := GitalyService(template)
			statefulSet := GitalyStatefulSet(template)

			It("Should render the template", func() {
				Expect(err).To(BeNil())
				Expect(template).NotTo(BeNil())
			})

			It("Should contain Gitaly resources", func() {
				Expect(enabled).To(BeTrue())
				Expect(configMap).NotTo(BeNil())
				Expect(service).NotTo(BeNil())
				Expect(statefulSet).NotTo(BeNil())
			})
		})
	})
})
