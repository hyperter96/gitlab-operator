package gitlab_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"

	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/gitlab"
)

var _ = Describe("Task Runner replacement", func() {

	mockCR := GitLabMock()
	adapter := gitlab.NewCustomResourceAdapter(mockCR)

	When("replacing Deployment", func() {
		templated := gitlab.TaskRunnerDeployment(adapter)
		generated := gitlab.TaskRunnerDeploymentDEPRECATED(mockCR)

		It("must completely satisfy the generator function", func() {
			Expect(templated).To(
				SatisfyReplacement(generated,
					IgnoreFields(corev1.ConfigMapVolumeSource{}, "Items"),
				))
		})
	})

	When("replacing ConfigMap", func() {
		templated := gitlab.TaskRunnerConfigMap(adapter)
		generated := gitlab.TaskRunnerConfigMapDEPRECATED(mockCR)

		It("must return one ConfigMap with similar ObjectMeta", func() {
			Expect(templated.ObjectMeta).To(
				SatisfyReplacement(generated.ObjectMeta))
		})

		It("must return one ConfigMap that contains the same Data items", func() {
			Expect(templated.Data).To(SatisfyReplacement(generated.Data))
		})
	})
})
