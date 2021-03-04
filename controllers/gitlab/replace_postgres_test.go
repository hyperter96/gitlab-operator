package gitlab_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/gitlab"
)

var _ = Describe("Postgres replacement", func() {

	mockCR := GitLabMock()
	adapter := gitlab.NewCustomResourceAdapter(mockCR)

	When("replacing StatefulSet", func() {
		templated := gitlab.PostgresStatefulSet(adapter)
		generated := gitlab.PostgresStatefulSetDEPRECATED(mockCR)

		It("must completely satisfy the generator function", func() {
			Expect(templated).To(SatisfyReplacement(generated))
		})
	})

	When("replacing Services", func() {
		// Only compare the services that the Operator renders.
		// - [x] test-postgresql-headless
		// - [ ] test-postgresql-metrics (unique to Chart)
		// - [x] test-postgresql
		templated := gitlab.PostgresServices(adapter)
		templatedSvc := gitlab.SvcFromList(fmt.Sprintf("%s-postgresql", mockCR.Name), templated)
		templatedSvcHeadless := gitlab.SvcFromList(fmt.Sprintf("%s-postgresql-headless", mockCR.Name), templated)

		generatedSvc := gitlab.PostgresqlServiceDEPRECATED(mockCR)               // equal to test-postgresql from Chart
		generatedSvcHeadless := gitlab.PostgresHeadlessServiceDEPRECATED(mockCR) // equal to test-postgresql-headless from Chart

		It("must completely satisfy the generator function", func() {
			Expect(templatedSvc).To(SatisfyReplacement(generatedSvc))
			Expect(templatedSvcHeadless).To(SatisfyReplacement(generatedSvcHeadless))
		})
	})

	When("replacing ConfigMap", func() {
		templated := gitlab.PostgresConfigMap(adapter)
		generated := gitlab.PostgresInitDBConfigMapDEPRECATED(mockCR)

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
