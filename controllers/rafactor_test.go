package controllers

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/internal"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/resource"
)

/*
 * Tests for refactoring Helm Query interface and replacing concrete resource
 * types with client.Object.
 */
var _ = Describe("Refactor", func() {
	helm.GetChartVersion()

	namespace := "default"
	releaseName := "test-internal"
	mockGitLab := gitlab.CreateMockGitLab(releaseName, namespace, resource.Values{
		"global": map[string]interface{}{
			"kas": map[string]interface{}{
				"enabled": true,
			},
			"pages": map[string]interface{}{
				"enabled": true,
			},
		},
	})
	adapter := gitlab.CreateMockAdapter(mockGitLab)
	template, _ := gitlab.GetTemplate(adapter)

	type clientObjectTypeAssertion func(client.Object) bool

	Describe("Query functions", func() {
		allKinds := map[string]clientObjectTypeAssertion{
			gitlab.ConfigMapKind:             func(o client.Object) bool { _, ok := o.(*corev1.ConfigMap); return ok },
			gitlab.CronJobKind:               func(o client.Object) bool { _, ok := o.(*batchv1.CronJob); return ok },
			gitlab.DeploymentKind:            func(o client.Object) bool { _, ok := o.(*appsv1.Deployment); return ok },
			gitlab.IngressKind:               func(o client.Object) bool { _, ok := o.(*networkingv1.Ingress); return ok },
			gitlab.JobKind:                   func(o client.Object) bool { _, ok := o.(*batchv1.Job); return ok },
			gitlab.PersistentVolumeClaimKind: func(o client.Object) bool { _, ok := o.(*corev1.PersistentVolumeClaim); return ok },
			gitlab.ServiceKind:               func(o client.Object) bool { _, ok := o.(*corev1.Service); return ok },
			gitlab.StatefulSetKind:           func(o client.Object) bool { _, ok := o.(*appsv1.StatefulSet); return ok },
		}

		allComponents := map[string][]string{
			gitlab.GitLabShellComponentName:    {gitlab.DeploymentKind, gitlab.ConfigMapKind, gitlab.ServiceKind},
			gitlab.MigrationsComponentName:     {gitlab.JobKind, gitlab.ConfigMapKind},
			gitlab.GitLabExporterComponentName: {gitlab.DeploymentKind, gitlab.ConfigMapKind, gitlab.ServiceKind},
			gitlab.RegistryComponentName:       {gitlab.DeploymentKind, gitlab.ConfigMapKind, gitlab.ServiceKind, gitlab.IngressKind},
			gitlab.WebserviceComponentName:     {gitlab.DeploymentKind, gitlab.ConfigMapKind, gitlab.ServiceKind, gitlab.IngressKind},
			gitlab.GitalyComponentName:         {gitlab.StatefulSetKind, gitlab.ConfigMapKind, gitlab.ServiceKind},
			gitlab.SidekiqComponentName:        {gitlab.DeploymentKind, gitlab.ConfigMapKind},
			gitlab.RedisComponentName:          {gitlab.StatefulSetKind, gitlab.ConfigMapKind, gitlab.ServiceKind},
			gitlab.PostgresComponentName:       {gitlab.StatefulSetKind, gitlab.ServiceKind},
			gitlab.NGINXComponentName:          {gitlab.DeploymentKind, gitlab.ConfigMapKind, gitlab.ServiceKind},
			gitlab.PagesComponentName:          {gitlab.DeploymentKind, gitlab.ConfigMapKind, gitlab.ServiceKind, gitlab.IngressKind},
			gitlab.KasComponentName:            {gitlab.DeploymentKind, gitlab.ConfigMapKind, gitlab.ServiceKind, gitlab.IngressKind},
		}

		for name, kinds := range allComponents {
			for kind, check := range allKinds {
				func(name, kind string, kinds []string, check clientObjectTypeAssertion) {
					FIt(fmt.Sprintf("works for %s of %s", kind, name), func() {
						object := template.Query().ObjectByKindAndComponent(kind, name)
						if object != nil {
							Expect(check(object)).To(BeTrue())
						} else {
							Expect(kind).NotTo(BeElementOf(kinds))
						}
					})
				}(name, kind, kinds, check)
			}
		}

	})

	Describe("Labels", func() {
		It("updates on PostgreSQL StatefulSet", func() {
			object := gitlab.PostgresStatefulSet(adapter)
			statefulset, ok := object.(*appsv1.StatefulSet)

			Expect(ok).To(BeTrue())
			Expect(statefulset.Labels).To(HaveKey(MatchRegexp(`^app\.kubernetes\.io/`)))
			Expect(statefulset.Spec.Template.Labels).To(HaveKey(MatchRegexp(`^app\.kubernetes\.io/`)))
		})
	})

	Describe("GetPodTemplateSpec", func() {
		objects := map[string]client.Object{
			"gitlab-exporter": gitlab.ExporterDeployment(adapter),
			"gitlab-shell":    gitlab.ShellDeployment(adapter),
			"kas":             gitlab.KasDeployment(adapter),
			"mailroom":        gitlab.MailroomDeployment(adapter),
			"minio":           internal.MinioStatefulSet(adapter),
			"pages":           gitlab.PagesDeployment(adapter),
			"postgresql":      gitlab.PostgresStatefulSet(adapter),
			"redis":           gitlab.RedisStatefulSet(adapter),
			"registry":        gitlab.RegistryDeployment(adapter),
			"toolbox":         gitlab.ToolboxDeployment(adapter),
		}

		for i, deployment := range gitlab.SidekiqDeployments(adapter) {
			objects[fmt.Sprintf("sidekiq-%d", i)] = deployment
		}

		for i, deployment := range gitlab.WebserviceDeployments(adapter) {
			objects[fmt.Sprintf("webservice-%d", i)] = deployment
		}

		for component, object := range objects {
			if object == nil {
				continue
			}

			func(component string, object client.Object) {
				It("works for "+component, func() {
					podTemplateSpec, err := internal.GetPodTemplateSpec(object)

					Expect(err).NotTo(HaveOccurred())
					Expect(podTemplateSpec).NotTo(BeNil())
					Expect(podTemplateSpec.Spec.Containers).NotTo(BeEmpty())
				})
			}(component, object)
		}

		It("returns error when object is neither Deployment nor StatefulSet", func() {
			object := gitlab.RegistryService(adapter)
			podTemplateSpec, err := internal.GetPodTemplateSpec(object)

			Expect(err).To(HaveOccurred())
			Expect(podTemplateSpec).To(BeNil())
		})
	})
})
