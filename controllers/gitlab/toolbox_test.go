package gitlab

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/resource"
)

var _ = Describe("CustomResourceAdapter", func() {

	if namespace == "" {
		namespace = "default"
	}

	Context("Toolbox", func() {
		When("Toolbox CronJob is disabled", func() {
			chartValues := resource.Values{}

			mockGitLab := CreateMockGitLab(releaseName, namespace, chartValues)
			adapter := CreateMockAdapter(mockGitLab)
			template, err := GetTemplate(adapter)

			enabled := ToolboxCronJobEnabled(adapter)
			cronJob := ToolboxCronJob(adapter)

			It("Should render the template", func() {
				Expect(err).To(BeNil())
				Expect(template).NotTo(BeNil())
			})

			It("Should not contain Toolbox CronJob resources", func() {
				Expect(enabled).To(BeFalse())
				Expect(cronJob).To(BeNil())
			})
		})

		When("Toolbox CronJob is enabled", func() {
			key := fmt.Sprintf(gitlabToolboxCronJobEnabled, ToolboxComponentName(GetChartVersion()))

			chartValues := resource.Values{}
			_ = chartValues.SetValue(key, true)

			mockGitLab := CreateMockGitLab(releaseName, namespace, chartValues)
			adapter := CreateMockAdapter(mockGitLab)
			template, err := GetTemplate(adapter)

			enabled := ToolboxCronJobEnabled(adapter)
			cronJob := ToolboxCronJob(adapter)

			It("Should render the template", func() {
				Expect(err).To(BeNil())
				Expect(template).NotTo(BeNil())
			})

			It("Should contain Toolbox CronJob resources", func() {
				Expect(enabled).To(BeTrue())
				Expect(cronJob).NotTo(BeNil())
			})
		})
	})
})
