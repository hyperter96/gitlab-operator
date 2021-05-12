package controllers

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	gitlabctl "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/gitlab"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/settings"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/helm"
)

var _ = Describe("GitLab controller", func() {

	BeforeEach(func() {
		os.Setenv("GITLAB_OPERATOR_SHARED_SECRETS_JOB_TIMEOUT", "1")
	})

	AfterEach(func() {
		os.Setenv("GITLAB_OPERATOR_SHARED_SECRETS_JOB_TIMEOUT", gitlabctl.SharedSecretsJobDefaultTimeout.String())
	})

	Context("GitLab CRD", func() {
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

			By("Deleting the created GitLab resource")
			Eventually(deleteObjectPromise(releaseName, &gitlabv1beta1.GitLab{}),
				PollTimeout, PollInterval).Should(Succeed())
		})
	})

	Context("GitLab CR spec", func() {
		It("Should change the managed resources when the Chart values change", func() {
			releaseName := "cr-spec-changes"
			chartValues := helm.EmptyValues()
			cfgMapName := fmt.Sprintf("%s-%s", releaseName, gitlabctl.SharedSecretsComponentName)
			sharedSecretQuery := appLabels(releaseName, gitlabctl.SharedSecretsComponentName)

			createGitLabResource(releaseName, chartValues)

			By("Checking shared secrets ConfigMap is created")
			Eventually(getObjectPromise(cfgMapName, &corev1.ConfigMap{}),
				PollTimeout, PollInterval).Should(Succeed())

			chartValues.SetValue("shared-secrets.env", "test")
			chartValues.SetValue("shared-secrets.annotations", map[string]string{
				"foo": "FOO",
				"bar": "BAR",
			})

			updateGitLabResource(releaseName, chartValues)

			By("Checking shared secrets ConfigMap picked up the change")
			Eventually(func() error {
				cfgMap := &corev1.ConfigMap{}
				if err := getObject(cfgMapName, cfgMap); err != nil {
					return err
				}
				if !strings.Contains(cfgMap.Data["generate-secrets"], "env=test") {
					return fmt.Errorf("`generate-secrets` does not contain the changes")
				}
				return nil
			}, PollTimeout, PollInterval).Should(Succeed())

			By("Checking shared secrets Jobs picked up the change")
			Eventually(func() error {
				jobs := &batchv1.JobList{}
				if err := listObjects(sharedSecretQuery, jobs); err != nil {
					return err
				}
				if len(jobs.Items) == 0 {
					return fmt.Errorf("Job list is emptry [%s]", sharedSecretQuery)
				}
				for _, job := range jobs.Items {
					if job.Spec.Template.ObjectMeta.Annotations["foo"] == "FOO" &&
						job.Spec.Template.ObjectMeta.Annotations["bar"] == "BAR" {
						return nil
					}
				}
				return fmt.Errorf("None of the Jobs had the expected annotations")
			}, PollTimeout, PollInterval).Should(Succeed())

			By("Deleting the created GitLab resource")
			Eventually(deleteObjectPromise(releaseName, &gitlabv1beta1.GitLab{}),
				PollTimeout, PollInterval).Should(Succeed())
		})
	})

	Context("Shared secrets and Self signed certificates Jobs", func() {
		When("Both Jobs succeed", func() {
			releaseName := "jobs-succeeded"

			chartValues := helm.EmptyValues()
			chartValues.SetValue("global.ingress.configureCertmanager", false)

			BeforeEach(func() {
				createGitLabResource(releaseName, chartValues)
			})

			It("Should create resources for Jobs and continue the reconcile loop", func() {
				cfgMapName := fmt.Sprintf("%s-%s", releaseName, gitlabctl.SharedSecretsComponentName)
				sharedSecretQuery := appLabels(releaseName, gitlabctl.SharedSecretsComponentName)
				gitlabShellQuery := appLabels(releaseName, gitlabctl.GitLabShellComponentName)

				By("Checking ServiceAccount exists for Shared secrets Job")
				Eventually(getObjectPromise(settings.AppServiceAccount, &corev1.ServiceAccount{}),
					PollTimeout, PollInterval).Should(Succeed())

				By("Checking Shared secrets Job and its ConfigMap are created")
				Eventually(getObjectPromise(cfgMapName, &corev1.ConfigMap{}),
					PollTimeout, PollInterval).Should(Succeed())
				Eventually(listObjectsPromise(sharedSecretQuery, &batchv1.JobList{}, 1),
					PollTimeout, PollInterval).Should(Succeed())

				By("Manipulating the Shared secrets Job to succeed")
				Eventually(updateJobStatusPromise(sharedSecretQuery, true),
					PollTimeout, PollInterval).Should(Succeed())

				By("Checking the Self signed certificates Job is created")
				Eventually(listObjectsPromise(sharedSecretQuery, &batchv1.JobList{}, 2),
					PollTimeout, PollInterval).Should(Succeed())

				By("Manipulating the Self signed certificates Job to succeed")
				Eventually(updateJobStatusPromise(sharedSecretQuery, true),
					PollTimeout, PollInterval).Should(Succeed())

				By("Checking next resources in the reconcile loop, e.g. ConfigMaps")
				Eventually(listConfigMapsPromise(gitlabShellQuery),
					PollTimeout, PollInterval).ShouldNot(BeEmpty())
			})
		})

		When("Jobs fail", func() {
			releaseName := "jobs-fail"

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

		When("Jobs time out", func() {
			releaseName := "jobs-timeout"

			BeforeEach(func() {
				createGitLabResource(releaseName, emptyValues)
			})

			It("Should fail the reconcile loop", func() {
				gitlabShellQuery := appLabels(releaseName, gitlabctl.GitLabShellComponentName)
				sharedSecretQuery := appLabels(releaseName, gitlabctl.SharedSecretsComponentName)

				By("Checking Shared secrets Job is created")
				Eventually(listObjectsPromise(sharedSecretQuery, &batchv1.JobList{}, 1),
					PollTimeout, PollInterval).Should(Succeed())

				By("Checking next resources in the reconcile loop, e.g. ConfigMaps")
				Consistently(listConfigMapsPromise(gitlabShellQuery),
					10*time.Second, PollInterval).Should(BeEmpty())
			})
		})
	})
})
