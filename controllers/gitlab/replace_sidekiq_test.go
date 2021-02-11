package gitlab_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"

	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/gitlab"
)

var _ = Describe("Sidekiq replacement", func() {
	mockCR := GitLabMock()
	adapter := gitlab.NewCustomResourceAdapter(mockCR)

	When("replacing Deployment", func() {
		templated := gitlab.SidekiqDeployment(adapter)
		generated := gitlab.SidekiqDeploymentDEPRECATED(mockCR)

		It("must completely satisfy the generator function", func() {
			Expect(templated).To(
				SatisfyReplacement(generated,
					IgnoreFields(corev1.ConfigMapVolumeSource{}, "Items"),
				))
		})
	})

	When("replacing ConfigMap", func() {
		templated := gitlab.SidekiqConfigMaps(adapter)
		generated := gitlab.SidekiqConfigMapDEPRECATED(mockCR)

		It("must return two ConfigMaps with similar ObjectMeta", func() {
			Expect(templated).To(HaveLen(2))
			Expect(templated[0].ObjectMeta).To(
				SatisfyReplacement(generated.ObjectMeta))
			Expect(templated[1].ObjectMeta).To(
				SatisfyReplacement(generated.ObjectMeta))
		})

		It("must return two ConfigMaps that contain the same Data items", func() {
			templatedData := map[string]string{}

			for _, cfgMap := range templated {
				for k, v := range cfgMap.Data {
					templatedData[k] = v
				}
			}

			Expect(templatedData).To(SatisfyReplacement(generated.Data))
		})
	})
})
