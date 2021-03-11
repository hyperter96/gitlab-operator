package gitlab

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/helpers"

	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Registry replacement", func() {
	mockCR := GitLabMock()
	adapter := helpers.NewCustomResourceAdapter(mockCR)

	When("replacing Deployment", func() {
		templated := RegistryDeployment(adapter)
		generated := RegistryDeploymentDEPRECATED(mockCR)

		It("must completely satisfy the generator function", func() {
			Expect(templated).To(
				SatisfyReplacement(generated,
					IgnoreFields(corev1.ConfigMapVolumeSource{}, "Items"),
				))
		})
	})

	When("replacing ConfigMap", func() {
		templated := RegistryConfigMap(adapter)
		generated := RegistryConfigMapDEPRECATED(adapter)

		It("must return a ConfigMap with similar ObjectMeta", func() {
			Expect(templated.ObjectMeta).To(
				SatisfyReplacement(generated.ObjectMeta))
		})

		It("must return a ConfigMap that contains the same Data items", func() {
			Expect(templated.Data).To(SatisfyReplacement(generated.Data))
		})
	})

	When("replacing Service", func() {
		templated := RegistryService(adapter)
		generated := RegistryServiceDeprecated(mockCR)

		It("must completely satisfy the generator function", func() {
			Expect(templated).To(SatisfyReplacement(generated))
		})
	})
})
