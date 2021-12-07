package gitlab

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
)

var _ = Describe("CustomResourceAdapter", func() {

	if namespace == "" {
		namespace = "default"
	}

	Context("Toolbox", func() {
		When("Toolbox CronJob is disabled", func() {
			chartValues := helm.EmptyValues()

			adapter := createMockAdapter(namespace, chartVersions[0], chartValues)
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
			key := fmt.Sprintf(gitlabToolboxCronJobEnabled, ToolboxComponentName(chartVersions[0]))

			chartValues := helm.EmptyValues()
			_ = chartValues.SetValue(key, true)

			adapter := createMockAdapter(namespace, chartVersions[0], chartValues)
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
