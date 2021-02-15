package controllers

import (
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	gitlabctl "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/gitlab"
)

var _ = Describe("GitLab controller", func() {

	Context("Shared secrets job", func() {
		When("Shared secrets Job succeed", func() {
			releaseName := "shared-secrets-job-succeeded"

			BeforeEach(func() {
				createGitLabResource(releaseName)
			})

			It("Should create resources for shared secrets Job and continue the reconcile loop", func() {
				cfgMapName := fmt.Sprintf("%s-%s", releaseName, gitlabctl.SharedSecretsComponentName)
				sharedSecretQuery := appLabels(releaseName, gitlabctl.SharedSecretsComponentName)
				gitlabShellQuery := appLabels(releaseName, gitlabctl.GitLabShellComponentName)

				By("Checking ServiceAccount exists for shared secret Job")
				Eventually(getObjectPromise(gitlabctl.AppServiceAccount, &corev1.ServiceAccount{}),
					PollTimeout, PollInterval).Should(Succeed())

				By("Checking shared secret Job and its ConfigMap are created")
				Eventually(getObjectPromise(cfgMapName, &corev1.ConfigMap{}),
					PollTimeout, PollInterval).Should(Succeed())
				Eventually(listObjectsPromise(sharedSecretQuery, &batchv1.JobList{}),
					PollTimeout, PollInterval).Should(Succeed())

				By("Manipulating the Job to succeed")
				Eventually(updateJobStatusPromise(sharedSecretQuery, true),
					PollTimeout, PollInterval).Should(Succeed())

				By("Checking next resources in the reconcile loop, e.g. ConfigMaps")
				Eventually(listConfigMapsPromise(gitlabShellQuery),
					PollTimeout, PollInterval).ShouldNot(BeEmpty())
			})
		})

		When("Shared secrets Job fails", func() {
			releaseName := "shared-secrets-job-failed"

			BeforeEach(func() {
				createGitLabResource(releaseName)
			})

			It("Should fail the reconcile loop", func() {
				sharedSecretQuery := appLabels(releaseName, gitlabctl.SharedSecretsComponentName)
				gitlabShellQuery := appLabels(releaseName, gitlabctl.GitLabShellComponentName)

				By("Manipulating the Job to succeed")
				Eventually(updateJobStatusPromise(sharedSecretQuery, false),
					PollTimeout, PollInterval).Should(Succeed())

				By("Checking next resources in the reconcile loop, e.g. ConfigMaps")
				Consistently(listConfigMapsPromise(gitlabShellQuery),
					10*time.Second, PollInterval).Should(BeEmpty())
			})
		})

		When("Shared secrets Job times out", func() {
			releaseName := "shared-secrets-job-timedout"

			BeforeEach(func() {
				os.Setenv("GITLAB_OPERATOR_SHARED_SECRETS_JOB_TIMEOUT", "1")
				createGitLabResource(releaseName)
			})

			It("Should fail the reconcile loop", func() {
				gitlabShellQuery := appLabels(releaseName, gitlabctl.GitLabShellComponentName)
				sharedSecretQuery := appLabels(releaseName, gitlabctl.SharedSecretsComponentName)

				By("Checking shared secret Job is created")
				Eventually(listObjectsPromise(sharedSecretQuery, &batchv1.JobList{}),
					PollTimeout, PollInterval).Should(Succeed())

				By("Checking next resources in the reconcile loop, e.g. ConfigMaps")
				Consistently(listConfigMapsPromise(gitlabShellQuery),
					10*time.Second, PollInterval).Should(BeEmpty())
			})
		})
	})
})
