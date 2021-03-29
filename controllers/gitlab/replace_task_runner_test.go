package gitlab

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/helpers"

	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Task Runner replacement", func() {

	mockCR := GitLabMock()
	adapter := helpers.NewCustomResourceAdapter(mockCR)

	When("replacing Deployment", func() {
		It("must completely satisfy the generator function", func() {
			templated := TaskRunnerDeployment(adapter)
			generated := TaskRunnerDeploymentDEPRECATED(mockCR)

			Expect(templated).To(
				SatisfyReplacement(generated,
					IgnoreFields(corev1.ConfigMapVolumeSource{}, "Items"),
				))
		})
	})

	When("replacing ConfigMap", func() {
		It("must return one ConfigMap with similar ObjectMeta", func() {
			templated := TaskRunnerConfigMap(adapter)
			generated := TaskRunnerConfigMapDEPRECATED(mockCR)

			Expect(templated.ObjectMeta).To(
				SatisfyReplacement(generated.ObjectMeta))
		})

		It("must return one ConfigMap that contains the same Data items", func() {
			templated := TaskRunnerConfigMap(adapter)
			generated := TaskRunnerConfigMapDEPRECATED(mockCR)

			Expect(templated.Data).To(SatisfyReplacement(generated.Data))
		})
	})
})
