package gitlab

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/helpers"
)

var _ = Describe("Gitaly replacement", func() {

	mockCR := GitLabMock()
	adapter := helpers.NewCustomResourceAdapter(mockCR)

	When("replacing StatefulSet", func() {
		templated := GitalyStatefulSet(adapter)
		generated := GitalyStatefulSetDEPRECATED(mockCR)

		It("must completely satisfy the generator function", func() {
			Expect(templated).To(
				SatisfyReplacement(generated))
		})
	})

	When("replacing ConfigMap", func() {
		templated := GitalyConfigMap(adapter)
		generated := GitalyConfigMapDEPRECATED(mockCR)

		It("must return two ConfigMaps with similar ObjectMeta", func() {

			Expect(templated.ObjectMeta).To(
				SatisfyReplacement(generated.ObjectMeta))
		})

		It("must return two ConfigMaps that contain the same Data items", func() {
			Expect(templated.Data).To(SatisfyReplacement(generated.Data))
		})
	})

	When("replacing Service", func() {
		templated := GitalyService(adapter)
		generated := GitalyServiceDEPRECATED(mockCR)

		It("must completely satisfy the generator function", func() {
			Expect(templated).To(
				SatisfyReplacement(generated))
		})
	})

})
