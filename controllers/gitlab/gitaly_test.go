package gitlab

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab/component"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support"
)

const (
	globalGitalyEnabled                        = "global.gitaly.enabled"
	globalPraefectVirtualStorages              = "global.praefect.virtualStorages"
	globalPraefectReplaceInternalGitalyEnabled = "global.praefect.replaceInternalGitaly"
)

var _ = Describe("Gitaly resources", func() {
	if namespace == "" {
		namespace = testNamespace
	}

	Context("Gitaly", func() {
		When("Gitaly is enabled", func() {
			chartValues := support.Values{}
			_ = chartValues.SetValue(globalGitalyEnabled, true)

			mockGitLab := CreateMockGitLab(releaseName, namespace, chartValues)
			adapter := CreateMockAdapter(mockGitLab)
			template, err := GetTemplate(adapter)

			enabled := adapter.WantsComponent(component.Gitaly)
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
			_ = chartValues.SetValue(globalGitalyEnabled, true)
			_ = chartValues.SetValue(globalPraefectEnabled, true)

			mockGitLab := CreateMockGitLab(releaseName, namespace, chartValues)
			adapter := CreateMockAdapter(mockGitLab)
			template, err := GetTemplate(adapter)

			enabled := adapter.WantsComponent(component.Gitaly)
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
			_ = chartValues.SetValue(globalGitalyEnabled, true)
			_ = chartValues.SetValue(globalPraefectEnabled, true)
			_ = chartValues.SetValue(globalPraefectReplaceInternalGitalyEnabled, false)

			_ = chartValues.SetValue(globalPraefectVirtualStorages, []map[string]interface{}{
				{
					"name":           "virtualstorage2",
					"gitalyReplicas": 5,
					"maxUnavailable": 2,
				},
			})

			mockGitLab := CreateMockGitLab(releaseName, namespace, chartValues)
			adapter := CreateMockAdapter(mockGitLab)
			template, err := GetTemplate(adapter)

			enabled := adapter.WantsComponent(component.Gitaly)
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
