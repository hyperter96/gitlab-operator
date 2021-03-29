package gitlab

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/helpers"

	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("GitLab Shell replacement", func() {

	mockCR := GitLabMock()
	adapter := helpers.NewCustomResourceAdapter(mockCR)

	When("replacing Deployment", func() {
		It("must completely satisfy the generator function", func() {
			templated := ShellDeployment(adapter)
			generated := ShellDeploymentDEPRECATED(mockCR)

			Expect(templated).To(
				SatisfyReplacement(generated,
					IgnoreFields(corev1.ConfigMapVolumeSource{}, "Items"),
				))
		})
	})

	When("replacing ConfigMap", func() {
		It("must return two ConfigMaps with similar ObjectMeta", func() {
			templated := ShellConfigMaps(adapter)
			generated := ShellConfigMapDEPRECATED(mockCR)

			Expect(templated).To(HaveLen(2))
			Expect(templated[0].ObjectMeta).To(
				SatisfyReplacement(generated.ObjectMeta))
			Expect(templated[1].ObjectMeta).To(
				SatisfyReplacement(generated.ObjectMeta))
		})

		It("must return two ConfigMaps that contain the same Data items", func() {
			templated := ShellConfigMaps(adapter)
			generated := ShellConfigMapDEPRECATED(mockCR)

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
		It("must completely satisfy the generator function", func() {
			templated := ShellService(adapter)
			generated := ShellServiceDEPRECATED(mockCR)

			Expect(templated).To(
				SatisfyReplacement(generated))
		})
	})

})
