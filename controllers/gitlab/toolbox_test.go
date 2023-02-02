package gitlab

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	feature "gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab/features"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support"
)

const (
	gitlabToolboxCronJobEnabled            = "gitlab.toolbox.backups.cron.enabled"
	gitlabToolboxCronJobPersistenceEnabled = "gitlab.toolbox.backups.cron.persistence.enabled"
)

var _ = Describe("CustomResourceAdapter", func() {

	if namespace == "" {
		namespace = testNamespace
	}

	Context("Toolbox", func() {
		When("Toolbox CronJob is disabled", func() {
			chartValues := support.Values{}

			mockGitLab := CreateMockGitLab(releaseName, namespace, chartValues)
			adapter := CreateMockAdapter(mockGitLab)
			template, err := GetTemplate(adapter)

			enabled := adapter.WantsFeature(feature.BackupCronJob)
			cronJob := ToolboxCronJob(adapter, template)

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
			key := gitlabToolboxCronJobEnabled

			chartValues := support.Values{}
			_ = chartValues.SetValue(key, true)

			mockGitLab := CreateMockGitLab(releaseName, namespace, chartValues)
			adapter := CreateMockAdapter(mockGitLab)
			template, err := GetTemplate(adapter)

			enabled := adapter.WantsFeature(feature.BackupCronJob)
			cronJob := ToolboxCronJob(adapter, template)

			persistenceEnabled := adapter.WantsFeature(feature.BackupCronJobPersistence)
			cronJobPersistentVolumeClaim := ToolboxCronJobPersistentVolumeClaim(adapter, template)

			It("Should render the template", func() {
				Expect(err).To(BeNil())
				Expect(template).NotTo(BeNil())
			})

			It("Should contain Toolbox CronJob resources", func() {
				Expect(enabled).To(BeTrue())
				Expect(cronJob).NotTo(BeNil())
			})

			It("Should not contain Toolbox CronJob Persistence resources", func() {
				Expect(persistenceEnabled).To(BeFalse())
				Expect(cronJobPersistentVolumeClaim).To(BeNil())
			})
		})

		When("Toolbox CronJob and CronJob Persistence is enabled", func() {
			gitlabToolboxCronJobEnabled := gitlabToolboxCronJobEnabled
			gitlabToolboxCronJobPersistenceEnabled := gitlabToolboxCronJobPersistenceEnabled

			chartValues := support.Values{}
			_ = chartValues.SetValue(gitlabToolboxCronJobEnabled, true)
			_ = chartValues.SetValue(gitlabToolboxCronJobPersistenceEnabled, true)

			mockGitLab := CreateMockGitLab(releaseName, namespace, chartValues)
			adapter := CreateMockAdapter(mockGitLab)
			template, err := GetTemplate(adapter)

			enabled := adapter.WantsFeature(feature.BackupCronJob)
			cronJob := ToolboxCronJob(adapter, template)

			persistenceEnabled := adapter.WantsFeature(feature.BackupCronJobPersistence)
			cronJobPersistentVolumeClaim := ToolboxCronJobPersistentVolumeClaim(adapter, template)

			It("Should render the template", func() {
				Expect(err).To(BeNil())
				Expect(template).NotTo(BeNil())
			})

			It("Should contain Toolbox CronJob resources", func() {
				Expect(enabled).To(BeTrue())
				Expect(cronJob).NotTo(BeNil())
			})

			It("Should contain Toolbox CronJob Persistence resources", func() {
				Expect(persistenceEnabled).To(BeTrue())
				Expect(cronJobPersistentVolumeClaim).NotTo(BeNil())
			})
		})
	})
})
