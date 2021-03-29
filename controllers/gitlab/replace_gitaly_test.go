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
		It("must completely satisfy the generator function", func() {
			templated := GitalyStatefulSet(adapter)
			generated := GitalyStatefulSetDEPRECATED(mockCR)

			Expect(templated).To(
				SatisfyReplacement(generated))
		})
	})

	When("replacing ConfigMap", func() {
		It("must return two ConfigMaps with similar ObjectMeta", func() {
			templated := GitalyConfigMap(adapter)
			generated := GitalyConfigMapDEPRECATED(mockCR)

			Expect(templated.ObjectMeta).To(
				SatisfyReplacement(generated.ObjectMeta))
		})

		It("must return two ConfigMaps that contain the same Data items", func() {
			templated := GitalyConfigMap(adapter)
			generated := GitalyConfigMapDEPRECATED(mockCR)

			Expect(templated.Data).To(SatisfyReplacement(generated.Data))
		})
	})

	When("replacing Service", func() {
		It("must completely satisfy the generator function", func() {
			templated := GitalyService(adapter)
			generated := GitalyServiceDEPRECATED(mockCR)

			Expect(templated).To(
				SatisfyReplacement(generated))
		})
	})

})
