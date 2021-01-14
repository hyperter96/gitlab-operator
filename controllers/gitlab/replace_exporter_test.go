package gitlab_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"

	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/gitlab"
)

var _ = Describe("GitLab Exporter replacement", func() {
	mockCR := GitLabMock()
	adapter := gitlab.NewCustomResourceAdapter(mockCR)

	When("replacing Service", func() {
		templated := gitlab.ExporterService(adapter)
		generated := gitlab.ExporterServiceDEPRECATED(mockCR)

		It("must completely satisfy the genrator function", func() {
			Expect(templated).To(
				SatisfyReplacement(generated))
		})
	})

	When("replacing Deployment", func() {
		templated := gitlab.ExporterDeployment(adapter)
		generated := gitlab.ExporterDeploymentDEPRECATED(mockCR)

		It("must completely satisfy the generator function", func() {
			Expect(templated).To(
				SatisfyReplacement(generated,
					IgnoreFields(corev1.ConfigMapVolumeSource{}, "Items"),
				))
		})
	})

	When("replacing ConfigMap", func() {
		templated := gitlab.ExporterConfigMaps(adapter)
		generated := gitlab.ExporterConfigMapDEPRECATED(mockCR)

		It("must return one ConfigMap with similar ObjectMeta", func() {
			Expect(templated).To(HaveLen(1))
			Expect(templated[0].ObjectMeta).To(
				SatisfyReplacement(generated.ObjectMeta),
			)
		})

		It("must return one ConfigMap that contains the same Data items", func() {
			Expect(templated).To(HaveLen(1))

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
})
