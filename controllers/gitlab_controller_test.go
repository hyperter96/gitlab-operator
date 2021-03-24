package controllers

import (
	"fmt"
	"os"
	"reflect"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	gitlabctl "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/gitlab"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/helm"
)

var _ = Describe("GitLab controller", func() {

	Context("CRD", func() {

		BeforeEach(func() {
			os.Setenv("GITLAB_OPERATOR_SHARED_SECRETS_JOB_TIMEOUT", "1")
		})

		AfterEach(func() {
			os.Setenv("GITLAB_OPERATOR_SHARED_SECRETS_JOB_TIMEOUT", "300")
		})

		It("Should create a CR with the specified Chart values", func() {
			releaseName := "crd-testing"

			chartValues := helm.EmptyValues()
			chartValues.SetValue("global.hosts.domain", "mydomain.com")
			chartValues.SetValue("certmanager-issuer.email", "me@mydomain.com")

			By("Creating a new GitLab resource")
			Expect(createObject(newGitLab(releaseName, chartValues), true)).Should(Succeed())

			By("Checking the created GitLab resource")
			Eventually(func() error {
				gitlab := &gitlabv1beta1.GitLab{}
				if err := getObject(releaseName, gitlab); err != nil {
					return err
				}

				if !reflect.DeepEqual(gitlab.Spec.Chart.Values.Object, chartValues.AsMap()) {
					return fmt.Errorf("The Chart values of CR are not equal to the expected values. Observed: %s",
						gitlab.Spec.Chart.Values.Object)
				}

				return nil
			}, PollTimeout, PollInterval).Should(Succeed())

			gitlab := &gitlabv1beta1.GitLab{}

			By("Deleting the created GitLab resource")
			Eventually(deleteObjectPromise(releaseName, gitlab),
				PollTimeout, PollInterval).Should(Succeed())
		})
	})

	Context("Shared secrets and Self signed certificates Jobs", func() {
		When("Both Jobs succeed", func() {
			releaseName := "jobs-succeeded"

			BeforeEach(func() {
				createGitLabResource(releaseName, emptyValues)
			})

			It("Should create resources for Jobs and continue the reconcile loop", func() {
				cfgMapName := fmt.Sprintf("%s-%s", releaseName, gitlabctl.SharedSecretsComponentName)
				sharedSecretQuery := appLabels(releaseName, gitlabctl.SharedSecretsComponentName)
				selfSignedQuery := componentLabels(releaseName, gitlabctl.SelfSignedCertsComponentName)
				gitlabShellQuery := appLabels(releaseName, gitlabctl.GitLabShellComponentName)

				By("Checking ServiceAccount exists for Shared secrets Job")
				Eventually(getObjectPromise(gitlabctl.AppServiceAccount, &corev1.ServiceAccount{}),
					PollTimeout, PollInterval).Should(Succeed())

				By("Checking Shared secrets Job and its ConfigMap are created")
				Eventually(getObjectPromise(cfgMapName, &corev1.ConfigMap{}),
					PollTimeout, PollInterval).Should(Succeed())
				Eventually(listObjectsPromise(sharedSecretQuery, &batchv1.JobList{}),
					PollTimeout, PollInterval).Should(Succeed())

				By("Manipulating the Shared secrets Job to succeed")
				Eventually(updateJobStatusPromise(sharedSecretQuery, true),
					PollTimeout, PollInterval).Should(Succeed())

				By("Checking the Self signed certificates Job is created")
				Eventually(listObjectsPromise(selfSignedQuery, &batchv1.JobList{}),
					PollTimeout, PollInterval).Should(Succeed())

				By("Manipulating the Self signed certificates Job to succeed")
				Eventually(updateJobStatusPromise(selfSignedQuery, true),
					PollTimeout, PollInterval).Should(Succeed())

				By("Checking next resources in the reconcile loop, e.g. ConfigMaps")
				Eventually(listConfigMapsPromise(gitlabShellQuery),
					PollTimeout, PollInterval).ShouldNot(BeEmpty())
			})
		})

		When("Shared secrets Job fails", func() {
			releaseName := "jobs-failed"

			BeforeEach(func() {
				createGitLabResource(releaseName, emptyValues)
			})

			It("Should fail the reconcile loop", func() {
				sharedSecretQuery := appLabels(releaseName, gitlabctl.SharedSecretsComponentName)
				gitlabShellQuery := appLabels(releaseName, gitlabctl.GitLabShellComponentName)

				By("Manipulating the Job to fail")
				Eventually(updateJobStatusPromise(sharedSecretQuery, false),
					PollTimeout, PollInterval).Should(Succeed())

				By("Checking next resources in the reconcile loop, e.g. ConfigMaps")
				Consistently(listConfigMapsPromise(gitlabShellQuery),
					10*time.Second, PollInterval).Should(BeEmpty())
			})
		})

		When("Self signed certificate Job fails", func() {
			releaseName := "self-signed-job-failed"

			BeforeEach(func() {
				createGitLabResource(releaseName, emptyValues)
			})

			It("Should fail the reconcile loop", func() {
				sharedSecretQuery := appLabels(releaseName, gitlabctl.SharedSecretsComponentName)
				selfSignedQuery := componentLabels(releaseName, gitlabctl.SelfSignedCertsComponentName)
				gitlabShellQuery := appLabels(releaseName, gitlabctl.GitLabShellComponentName)

				By("Manipulating the Self signed certificates Job to fail")
				Eventually(updateJobStatusPromise(sharedSecretQuery, true),
					PollTimeout, PollInterval).Should(Succeed())
				Eventually(updateJobStatusPromise(selfSignedQuery, false),
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
				createGitLabResource(releaseName, emptyValues)
			})

			It("Should fail the reconcile loop", func() {
				gitlabShellQuery := appLabels(releaseName, gitlabctl.GitLabShellComponentName)
				sharedSecretQuery := appLabels(releaseName, gitlabctl.SharedSecretsComponentName)

				By("Checking Shared secrets Job is created")
				Eventually(listObjectsPromise(sharedSecretQuery, &batchv1.JobList{}),
					PollTimeout, PollInterval).Should(Succeed())

				By("Checking next resources in the reconcile loop, e.g. ConfigMaps")
				Consistently(listConfigMapsPromise(gitlabShellQuery),
					10*time.Second, PollInterval).Should(BeEmpty())
			})
		})
	})
})
