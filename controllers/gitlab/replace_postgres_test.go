package gitlab

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/helpers"
)

var _ = Describe("Postgres replacement", func() {

	mockCR := GitLabMock()
	adapter := helpers.NewCustomResourceAdapter(mockCR)

	When("replacing StatefulSet", func() {
		templated := PostgresStatefulSet(adapter)
		generated := PostgresStatefulSetDEPRECATED(mockCR)

		It("must completely satisfy the generator function", func() {
			Expect(templated).To(SatisfyReplacement(generated))
		})
	})

	When("replacing Services", func() {
		// Only compare the services that the Operator renders.
		// - [x] test-postgresql-headless
		// - [ ] test-postgresql-metrics (unique to Chart)
		// - [x] test-postgresql
		templated := PostgresServices(adapter)
		templatedSvc := SvcFromList(fmt.Sprintf("%s-postgresql", mockCR.Name), templated)
		templatedSvcHeadless := SvcFromList(fmt.Sprintf("%s-postgresql-headless", mockCR.Name), templated)

		generatedSvc := PostgresqlServiceDEPRECATED(mockCR)               // equal to test-postgresql from Chart
		generatedSvcHeadless := PostgresHeadlessServiceDEPRECATED(mockCR) // equal to test-postgresql-headless from Chart

		It("must completely satisfy the generator function", func() {
			Expect(templatedSvc).To(SatisfyReplacement(generatedSvc))
			Expect(templatedSvcHeadless).To(SatisfyReplacement(generatedSvcHeadless))
		})
	})

	When("replacing ConfigMap", func() {
		templated := PostgresConfigMap(adapter)
		generated := PostgresInitDBConfigMapDEPRECATED(mockCR)

		It("must return a ConfigMap with similar ObjectMeta", func() {
			Expect(templated.ObjectMeta).To(
				SatisfyReplacement(generated.ObjectMeta))
		})

		It("must return a ConfigMap that contains the same Data items", func() {
			Expect(templated.Data).To(
				SatisfyReplacement(generated.Data))
		})
	})

})
