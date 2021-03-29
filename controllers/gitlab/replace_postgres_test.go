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
		It("must completely satisfy the generator function", func() {
			templated := PostgresStatefulSet(adapter)
			generated := PostgresStatefulSetDEPRECATED(mockCR)

			Expect(templated).To(SatisfyReplacement(generated))
		})
	})

	When("replacing Services", func() {
		It("must completely satisfy the generator function", func() {
			// Only compare the services that the Operator renders.
			// - [x] test-postgresql-headless
			// - [ ] test-postgresql-metrics (unique to Chart)
			// - [x] test-postgresql
			templated := PostgresServices(adapter)
			templatedSvc := SvcFromList(fmt.Sprintf("%s-postgresql", mockCR.Name), templated)
			templatedSvcHeadless := SvcFromList(fmt.Sprintf("%s-postgresql-headless", mockCR.Name), templated)

			generatedSvc := PostgresqlServiceDEPRECATED(mockCR)               // equal to test-postgresql from Chart
			generatedSvcHeadless := PostgresHeadlessServiceDEPRECATED(mockCR) // equal to test-postgresql-headless from Chart

			Expect(templatedSvc).To(SatisfyReplacement(generatedSvc))
			Expect(templatedSvcHeadless).To(SatisfyReplacement(generatedSvcHeadless))
		})
	})

	When("replacing ConfigMap", func() {
		It("must return a ConfigMap with similar ObjectMeta", func() {
			templated := PostgresConfigMap(adapter)
			generated := PostgresInitDBConfigMapDEPRECATED(mockCR)

			Expect(templated.ObjectMeta).To(
				SatisfyReplacement(generated.ObjectMeta))
		})

		It("must return a ConfigMap that contains the same Data items", func() {
			templated := PostgresConfigMap(adapter)
			generated := PostgresInitDBConfigMapDEPRECATED(mockCR)

			Expect(templated.Data).To(
				SatisfyReplacement(generated.Data))
		})
	})

})
