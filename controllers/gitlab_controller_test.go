package controllers

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	gitlabv1beta1 "gitlab.com/gitlab-org/cloud-native/gitlab-operator/api/v1beta1"
	gitlabctl "gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/settings"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
)

var _ = Describe("GitLab controller", func() {

	fmt.Println("testing with chart version", getChartVersion())

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

	Context("Gitaly", func() {
		When("Bundled Gitaly is enabled", func() {
			releaseName := "gitaly-enabled"
			cfgMapName := fmt.Sprintf("%s-%s", releaseName, gitlabctl.GitalyComponentName)
			serviceName := fmt.Sprintf("%s-%s", releaseName, gitlabctl.GitalyComponentName)
			statefulSetName := fmt.Sprintf("%s-%s", releaseName, gitlabctl.GitalyComponentName)
			nextCfgMapName := fmt.Sprintf("%s-%s", releaseName, gitlabctl.SharedSecretsComponentName)

			chartValues := helm.EmptyValues()
			chartValues.SetValue("global.gitaly.enabled", true)

			BeforeEach(func() {
				createGitLabResource(releaseName, chartValues)
				processSharedSecretsJob(releaseName)
			})

			It("Should create Gitaly resources and continue the reconcile loop", func() {
				By("Checking Gitaly Service exists")
				Eventually(getObjectPromise(serviceName, &corev1.Service{}),
					PollTimeout, PollInterval).Should(Succeed())

				By("Checking Gitaly StatefulSet exists")
				Eventually(getObjectPromise(statefulSetName, &appsv1.StatefulSet{}),
					PollTimeout, PollInterval).Should(Succeed())

				By("Checking Gitaly ConfigMap exists")
				Eventually(getObjectPromise(cfgMapName, &corev1.ConfigMap{}),
					PollTimeout, PollInterval).Should(Succeed())

				By("Checking next resources in the reconcile loop, e.g. ConfigMaps")
				Eventually(getObjectPromise(nextCfgMapName, &corev1.ConfigMap{}),
					PollTimeout, PollInterval).Should(Succeed())
			})
		})

		When("Bundled Gitaly is disabled", func() {
			releaseName := "gitaly-disabled"
			cfgMapName := fmt.Sprintf("%s-%s", releaseName, gitlabctl.GitalyComponentName)
			serviceName := fmt.Sprintf("%s-%s", releaseName, gitlabctl.GitalyComponentName)
			statefulSetName := fmt.Sprintf("%s-%s", releaseName, gitlabctl.GitalyComponentName)
			nextCfgMapName := fmt.Sprintf("%s-%s", releaseName, gitlabctl.SharedSecretsComponentName)

			chartValues := helm.EmptyValues()
			values := `
global:
  gitaly:
    enabled: false
    external:
    - name: default
      hostname: gitaly.external.com
`
			chartValues.AddFromYAML([]byte(values))

			BeforeEach(func() {
				createGitLabResource(releaseName, chartValues)
				processSharedSecretsJob(releaseName)
			})

			It("Should not create Gitaly resources and continue the reconcile loop", func() {
				By("Checking Gitaly Service does not exist")
				Eventually(getObjectPromise(serviceName, &corev1.Service{}),
					PollTimeout, PollInterval).ShouldNot(Succeed())

				By("Checking Gitaly StatefulSet does not exist")
				Eventually(getObjectPromise(statefulSetName, &appsv1.StatefulSet{}),
					PollTimeout, PollInterval).ShouldNot(Succeed())

				By("Checking Gitaly ConfigMap does not exist")
				Eventually(getObjectPromise(cfgMapName, &corev1.ConfigMap{}),
					PollTimeout, PollInterval).ShouldNot(Succeed())

				By("Checking next resources in the reconcile loop, e.g. ConfigMaps")
				Eventually(getObjectPromise(nextCfgMapName, &corev1.ConfigMap{}),
					PollTimeout, PollInterval).Should(Succeed())
			})
		})
	})

	Context("PostgreSQL", func() {
		When("Bundled PostgreSQL is disabled", func() {
			releaseName := "postgresql-disabled"
			cfgMapName := fmt.Sprintf("%s-%s-init-db", releaseName, gitlabctl.PostgresComponentName)
			metricsServiceName := fmt.Sprintf("%s-%s-metrics", releaseName, gitlabctl.PostgresComponentName)
			headlessServiceName := fmt.Sprintf("%s-%s-headless", releaseName, gitlabctl.PostgresComponentName)
			serviceName := fmt.Sprintf("%s-%s", releaseName, gitlabctl.PostgresComponentName)
			statefulSetName := fmt.Sprintf("%s-%s", releaseName, gitlabctl.PostgresComponentName)
			nextCfgMapName := fmt.Sprintf("%s-%s", releaseName, gitlabctl.SharedSecretsComponentName)

			chartValues := helm.EmptyValues()
			chartValues.SetValue("postgresql.install", false)

			BeforeEach(func() {
				createGitLabResource(releaseName, chartValues)
				processSharedSecretsJob(releaseName)
			})

			It("Should not create PostgreSQL resources and continue the reconcile loop", func() {
				By("Checking PostgreSQL Metrics Service does not exist")
				Eventually(getObjectPromise(metricsServiceName, &corev1.Service{}),
					PollTimeout, PollInterval).ShouldNot(Succeed())

				By("Checking PostgreSQL Headless Service does not exist")
				Eventually(getObjectPromise(headlessServiceName, &corev1.Service{}),
					PollTimeout, PollInterval).ShouldNot(Succeed())

				By("Checking PostgreSQL Service does not exist")
				Eventually(getObjectPromise(serviceName, &corev1.Service{}),
					PollTimeout, PollInterval).ShouldNot(Succeed())

				By("Checking PostgreSQL StatefulSet does not exist")
				Eventually(getObjectPromise(statefulSetName, &appsv1.StatefulSet{}),
					PollTimeout, PollInterval).ShouldNot(Succeed())

				By("Checking PostgreSQL ConfigMap does not exist")
				Eventually(getObjectPromise(cfgMapName, &corev1.ConfigMap{}),
					PollTimeout, PollInterval).ShouldNot(Succeed())

				By("Checking next resources in the reconcile loop, e.g. ConfigMaps")
				Eventually(getObjectPromise(nextCfgMapName, &corev1.ConfigMap{}),
					PollTimeout, PollInterval).Should(Succeed())
			})
		})

		When("Bundled PostgreSQL is enabled", func() {
			releaseName := "postgresql-enabled"
			cfgMapName := fmt.Sprintf("%s-%s-init-db", releaseName, gitlabctl.PostgresComponentName)
			metricsServiceName := fmt.Sprintf("%s-%s-metrics", releaseName, gitlabctl.PostgresComponentName)
			headlessServiceName := fmt.Sprintf("%s-%s-headless", releaseName, gitlabctl.PostgresComponentName)
			serviceName := fmt.Sprintf("%s-%s", releaseName, gitlabctl.PostgresComponentName)
			statefulSetName := fmt.Sprintf("%s-%s", releaseName, gitlabctl.PostgresComponentName)
			nextCfgMapName := fmt.Sprintf("%s-%s", releaseName, gitlabctl.SharedSecretsComponentName)

			chartValues := helm.EmptyValues()
			chartValues.SetValue("postgresql.install", true)

			BeforeEach(func() {
				createGitLabResource(releaseName, chartValues)
				processSharedSecretsJob(releaseName)
			})

			It("Should create PostgreSQL resources and continue the reconcile loop", func() {
				By("Checking PostgreSQL Metrics Service exists")
				Eventually(getObjectPromise(metricsServiceName, &corev1.Service{}),
					PollTimeout, PollInterval).Should(Succeed())

				By("Checking PostgreSQL Headless Service exists")
				Eventually(getObjectPromise(headlessServiceName, &corev1.Service{}),
					PollTimeout, PollInterval).Should(Succeed())

				By("Checking PostgreSQL Service exists")
				Eventually(getObjectPromise(serviceName, &corev1.Service{}),
					PollTimeout, PollInterval).Should(Succeed())

				By("Checking PostgreSQL StatefulSet exists")
				Eventually(getObjectPromise(statefulSetName, &appsv1.StatefulSet{}),
					PollTimeout, PollInterval).Should(Succeed())

				By("Checking PostgreSQL ConfigMap exists")
				Eventually(getObjectPromise(cfgMapName, &corev1.ConfigMap{}),
					PollTimeout, PollInterval).Should(Succeed())

				By("Checking next resources in the reconcile loop, e.g. ConfigMaps")
				Eventually(getObjectPromise(nextCfgMapName, &corev1.ConfigMap{}),
					PollTimeout, PollInterval).Should(Succeed())
			})
		})
	})

	Context("Bundled NGINX with SSH support", func() {
		When("Bundled NGINX is disabled", func() {
			releaseName := "nginx-disabled"
			tcpCfgMapName := fmt.Sprintf("%s-%s-tcp", releaseName, gitlabctl.NGINXComponentName)
			controllerServiceName := fmt.Sprintf("%s-%s-controller", releaseName, gitlabctl.NGINXComponentName)
			controllerDeploymentName := fmt.Sprintf("%s-%s-controller", releaseName, gitlabctl.NGINXComponentName)
			defaultBackendDeploymentName := fmt.Sprintf("%s-%s-default-backend", releaseName, gitlabctl.NGINXComponentName)
			nextCfgMapName := fmt.Sprintf("%s-%s", releaseName, gitlabctl.SharedSecretsComponentName)

			chartValues := helm.EmptyValues()
			chartValues.SetValue("nginx-ingress.enabled", false)

			BeforeEach(func() {
				createGitLabResource(releaseName, chartValues)
				processSharedSecretsJob(releaseName)
			})

			It("Should not create NGINX resources and continue the reconcile loop", func() {
				By("Checking NGINX Controller Service does not exist")
				Eventually(getObjectPromise(controllerServiceName, &corev1.Service{}),
					PollTimeout, PollInterval).ShouldNot(Succeed())

				By("Checking NGINX Controller Deployment does not exist")
				Eventually(getObjectPromise(controllerDeploymentName, &appsv1.Deployment{}),
					PollTimeout, PollInterval).ShouldNot(Succeed())

				By("Checking NGINX Default Backend Deployment does not exist")
				Eventually(getObjectPromise(defaultBackendDeploymentName, &appsv1.Deployment{}),
					PollTimeout, PollInterval).ShouldNot(Succeed())

				By("Checking NGINX TCP ConfigMap from GitLab Shell does not exist")
				Eventually(getObjectPromise(tcpCfgMapName, &corev1.ConfigMap{}),
					PollTimeout, PollInterval).Should(Succeed())

				By("Checking next resources in the reconcile loop, e.g. ConfigMaps")
				Eventually(getObjectPromise(nextCfgMapName, &corev1.ConfigMap{}),
					PollTimeout, PollInterval).Should(Succeed())
			})
		})

		When("Bundled NGINX is enabled", func() {
			releaseName := "nginx-enabled"
			tcpCfgMapName := fmt.Sprintf("%s-%s-tcp", releaseName, gitlabctl.NGINXComponentName)
			controllerServiceName := fmt.Sprintf("%s-%s-controller", releaseName, gitlabctl.NGINXComponentName)
			controllerDeploymentName := fmt.Sprintf("%s-%s-controller", releaseName, gitlabctl.NGINXComponentName)
			defaultBackendDeploymentName := fmt.Sprintf("%s-%s-default-backend", releaseName, gitlabctl.NGINXComponentName)
			nextCfgMapName := fmt.Sprintf("%s-%s", releaseName, gitlabctl.SharedSecretsComponentName)

			BeforeEach(func() {
				createGitLabResource(releaseName, helm.EmptyValues())
				processSharedSecretsJob(releaseName)
			})

			It("Should create NGINX resources by default and continue the reconcile loop", func() {
				By("Checking NGINX Controller Service exists")
				Eventually(getObjectPromise(controllerServiceName, &corev1.Service{}),
					PollTimeout, PollInterval).Should(Succeed())

				By("Checking NGINX Controller Deployment exists")
				Eventually(getObjectPromise(controllerDeploymentName, &appsv1.Deployment{}),
					PollTimeout, PollInterval).Should(Succeed())

				By("Checking NGINX Default Backend Deployment exists")
				Eventually(getObjectPromise(defaultBackendDeploymentName, &appsv1.Deployment{}),
					PollTimeout, PollInterval).Should(Succeed())

				By("Checking NGINX TCP ConfigMap from GitLab Shell exists")
				Eventually(getObjectPromise(tcpCfgMapName, &corev1.ConfigMap{}),
					PollTimeout, PollInterval).Should(Succeed())

				By("Checking next resources in the reconcile loop, e.g. ConfigMaps")
				Eventually(getObjectPromise(nextCfgMapName, &corev1.ConfigMap{}),
					PollTimeout, PollInterval).Should(Succeed())
			})
		})
	})
})

func processSharedSecretsJob(releaseName string) {
	sharedSecretQuery := appLabels(releaseName, gitlabctl.SharedSecretsComponentName)

	By("Manipulating the Shared secrets Job to succeed")
	Eventually(updateJobStatusPromise(sharedSecretQuery, true),
		PollTimeout, PollInterval).Should(Succeed())
}
