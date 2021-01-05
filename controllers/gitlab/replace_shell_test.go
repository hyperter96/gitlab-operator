package gitlab_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/gitlab"
	gitlabutils "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/utils"
)

const (
	// GitLabShellComponentName is the common name of GitLab Shell.
	GitLabShellComponentName = "gitlab-shell"
)

// ShellDeployment returns the Deployment of GitLab Shell component.
func ShellDeployment(adapter gitlab.CustomResourceAdapter) *appsv1.Deployment {

	template, err := gitlab.GetTemplate(adapter)

	if err != nil {
		return nil
		/* WARNING: This should return an error instead. */
	}

	result := template.Query().DeploymentByComponent(GitLabShellComponentName)

	patchGitLabShellDeployment(adapter, result)

	return result
}

func patchGitLabShellDeployment(adapter gitlab.CustomResourceAdapter, deployment *appsv1.Deployment) {
	updateCommonLabels(adapter.ReleaseName(), GitLabShellComponentName, &deployment.ObjectMeta.Labels)

	if deployment.Spec.Template.Spec.SecurityContext == nil {
		deployment.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{}
	}

	var replicas int32 = 1
	var userID int64 = 1000

	deployment.Spec.Template.Spec.SecurityContext.FSGroup = &userID
	deployment.Spec.Template.Spec.SecurityContext.RunAsUser = &userID
	deployment.Spec.Template.Spec.ServiceAccountName = gitlab.AppServiceAccount
	deployment.Spec.Replicas = &replicas
}

func updateCommonLabels(releaseName, componentName string, labels *map[string]string) {
	for k, v := range gitlabutils.Label(releaseName, componentName, gitlabutils.GitlabType) {
		(*labels)[k] = v
	}
}

var _ = Describe("GitLab Shell replacement", func() {

	mockCR := GitLabMock()
	adapter := gitlab.NewCustomResourceAdapter(mockCR)

	When("replacing Deployment", func() {
		templated := gitlab.ShellDeployment(adapter)
		generated := gitlab.ShellDeploymentDEPRECATED(mockCR)

		It("must completely satisfy the generator function", func() {
			Expect(templated).To(
				SatisfyReplacement(generated,
					IgnoreFields(corev1.ConfigMapVolumeSource{}, "Items"),
				))
		})
	})

	When("replacing ConfigMap", func() {
		templated := gitlab.ShellConfigMaps(adapter)
		generated := gitlab.ShellConfigMapDEPRECATED(mockCR)

		It("must return two ConfigMaps with similar ObjectMeta", func() {

			Expect(templated).To(HaveLen(2))
			Expect(templated[0].ObjectMeta).To(
				SatisfyReplacement(generated.ObjectMeta))
			Expect(templated[1].ObjectMeta).To(
				SatisfyReplacement(generated.ObjectMeta))
		})

		It("must return two ConfigMaps that contain the same Data items", func() {
			Expect(templated).To(HaveLen(2))

			generatedData := map[string]string{}
			templatedData := map[string]string{}

			for k, v := range generated.Data {
				generatedData[k] = v
			}

			for _, cfgMap := range templated {
				for k, v := range cfgMap.Data {
					templatedData[k] = v
				}
			}

			Expect(templatedData).To(SatisfyReplacement(generatedData))
		})
	})

	When("replacing Service", func() {
		templated := gitlab.ShellService(adapter)
		generated := gitlab.ShellServiceDEPRECATED(mockCR)

		It("must completely satisfy the generator function", func() {
			Expect(templated).To(
				SatisfyReplacement(generated))
		})
	})

})
