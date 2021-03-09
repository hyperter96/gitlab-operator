package gitlab_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"

	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/gitlab"
)

var _ = Describe("Registry replacement", func() {
	mockCR := GitLabMock()
	adapter := gitlab.NewCustomResourceAdapter(mockCR)

	When("replacing Deployment", func() {
		templated := gitlab.RegistryDeployment(adapter)
		generated := gitlab.RegistryDeploymentDEPRECATED(mockCR)

		It("must completely satisfy the generator function", func() {
			Expect(templated).To(
				SatisfyReplacement(generated,
					IgnoreFields(corev1.ConfigMapVolumeSource{}, "Items"),
				))
		})
	})

	When("replacing ConfigMap", func() {
		templated := gitlab.RegistryConfigMap(adapter)
		generated := gitlab.RegistryConfigMapDEPRECATED(adapter)

		It("must return a ConfigMap with similar ObjectMeta", func() {
			Expect(templated.ObjectMeta).To(
				SatisfyReplacement(generated.ObjectMeta))
		})

		It("must return a ConfigMap that contains the same Data items", func() {
			Expect(templated.Data).To(SatisfyReplacement(generated.Data))
		})
	})

	When("replacing Service", func() {
		templated := gitlab.RegistryService(adapter)
		generated := gitlab.RegistryServiceDeprecated(mockCR)

		It("must completely satisfy the generator function", func() {
			Expect(templated).To(SatisfyReplacement(generated))
		})
	})
})
