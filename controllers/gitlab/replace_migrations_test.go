package gitlab

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/helpers"
)

var _ = Describe("Migrations replacement", func() {

	mockCR := GitLabMock()
	adapter := helpers.NewCustomResourceAdapter(mockCR)

	When("replacing Job", func() {
		It("must completely satisfy the generator function", func() {
			templated, _ := MigrationsJob(adapter)
			generated := MigrationsJobDEPRECATED(mockCR)

			Expect(templated).To(
				SatisfyReplacement(generated))
		})
	})

	When("replacing ConfigMap", func() {
		It("must return one ConfigMap with similar ObjectMeta", func() {
			templated := MigrationsConfigMap(adapter)
			generated := MigrationsConfigMapDEPRECATED(mockCR)

			Expect(templated.ObjectMeta).To(
				SatisfyReplacement(generated.ObjectMeta))
		})

		It("must return one ConfigMap that contains the same Data items", func() {
			templated := MigrationsConfigMap(adapter)
			generated := MigrationsConfigMapDEPRECATED(mockCR)

			Expect(templated.Data).To(SatisfyReplacement(generated.Data))
		})
	})
})
