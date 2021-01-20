package gitlab_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/gitlab"
)

const (
	// GitalyComponentName is the common name of Gitaly.
	GitalyComponentName = "gitaly"
)

var _ = Describe("Gitaly replacement", func() {

	mockCR := GitLabMock()
	adapter := gitlab.NewCustomResourceAdapter(mockCR)

	When("replacing StatefulSet", func() {
		templated := gitlab.GitalyStatefulSet(adapter)
		generated := gitlab.GitalyStatefulSetDEPRECATED(mockCR)

		It("must completely satisfy the generator function", func() {
			Expect(templated).To(
				SatisfyReplacement(generated))
		})
	})

	When("replacing ConfigMap", func() {
		templated := gitlab.GitalyConfigMap(adapter)
		generated := gitlab.GitalyConfigMapDEPRECATED(mockCR)

		It("must return two ConfigMaps with similar ObjectMeta", func() {

			Expect(templated.ObjectMeta).To(
				SatisfyReplacement(generated.ObjectMeta))
		})

		It("must return two ConfigMaps that contain the same Data items", func() {
			Expect(templated.Data).To(SatisfyReplacement(generated.Data))
		})
	})

	When("replacing Service", func() {
		templated := gitlab.GitalyService(adapter)
		generated := gitlab.GitalyServiceDEPRECATED(mockCR)

		It("must completely satisfy the generator function", func() {
			Expect(templated).To(
				SatisfyReplacement(generated))
		})
	})

})
