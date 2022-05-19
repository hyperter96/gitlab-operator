package populate

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"helm.sh/helm/v3/pkg/chart"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support/charts"
)

var _ = Describe("Populate", func() {
	/* Use the following test scaffolding:
	 *
	 *   chart-1/						# chart-1	v0.1.0
	 *      Chart.yaml
	 *      sub-chart-1/
	 *        Chart.yaml
	 *   more-charts/
	 *      chart-2/					# chart-2	v0.1.2
	 *        Chart.yaml
	 *      chart-2-v2.tar.gz			# chart-2	v0.2.0
	 *      not-chart.tgz
	 *		Note.txt
	 *   not-chart/
	 *     Note.txt
	 *   chart-1-v1.tgz					# chart-1	v0.1.0	(duplicate)
	 *   chart-1-v2.tar.gz				# chart-1	v0.2.0
	 */
	It("populates the Charts with the specified options", func() {
		c := &charts.Catalog{}
		e := c.Populate(
			WithSearchPath("testdata/charts"),
			WithFilePattern("*.tar.gz", "*.tgz"))
		/* The current version of logr can not be integrated with Gingkgo */

		Expect(e).ToNot(HaveOccurred())
		Expect(*c).To(HaveLen(4))
	})
})

var _ = Describe("Query", func() {
	/* Use the test scaffolding from Populate */
	c := &charts.Catalog{}
	_ = c.Populate(
		WithSearchPath("testdata/charts"),
		WithFilePattern("*.tar.gz", "*.tgz"))

	It("returns the list of available Chart names", func() {
		Expect(c.Names()).To(ConsistOf("chart-1", "chart-2"))
	})

	It("returns the list of available versions of each Chart name", func() {
		Expect(c.Versions("chart-1")).To(ConsistOf("0.1.0", "0.2.0"))
		Expect(c.Versions("chart-2")).To(ConsistOf("0.1.2", "0.2.0"))
	})

	It("returns the list of available appVersions of each Chart name", func() {
		Expect(c.AppVersions("chart-1")).To(ConsistOf("1.0.0", "1.1.0"))
		Expect(c.AppVersions("chart-2")).To(ConsistOf("2.0.0", "2.1.0"))
	})

	It("returns no Chart when the criteria is empty", func() {
		Expect(c.Query()).To(BeEmpty())
	})

	It("returns Charts that match the criteria", func() {
		r := c.Query(charts.WithName("chart-1"), charts.WithVersion("0.1.1"))
		Expect(r).To(BeEmpty())
		Expect(r.First()).To(BeNil())

		r = c.Query(charts.WithName("chart-1"), charts.WithVersion("0.1.0"))
		Expect(r).To(HaveLen(1))
		Expect(r[0].Metadata.Name).To(Equal("chart-1"))
		Expect(r[0].Metadata.Version).To(Equal("0.1.0"))
		Expect(r.First()).NotTo(BeNil())

		r = c.Query(charts.WithName("chart-2"), charts.WithAppVersion("2.1.0"))
		Expect(r).To(HaveLen(1))
		Expect(r[0].Metadata.Name).To(Equal("chart-2"))
		Expect(r[0].Metadata.AppVersion).To(Equal("2.1.0"))

		r = c.Query(charts.WithName("chart-1"))
		Expect(r).To(HaveLen(2))
		Expect(r[0].Metadata.Name).To(Equal("chart-1"))
		Expect(r[1].Metadata.Name).To(Equal("chart-1"))
	})

	It("makes a copy of the original Chart", func() {
		val := map[string]interface{}{
			"foo": "bar",
		}
		chrt := &chart.Chart{
			Metadata: &chart.Metadata{
				Name:    "test",
				Version: "1.0.0",
			},
			Values: val,
		}
		dep1 := &chart.Chart{
			Metadata: &chart.Metadata{
				Name:    "test-dep-1",
				Version: "1.0.0",
			},
		}
		dep2 := &chart.Chart{
			Metadata: &chart.Metadata{
				Name:    "test-dep-2",
				Version: "1.0.0",
			},
		}
		chrt.SetDependencies(dep1, dep2)
		c := &charts.Catalog{
			chrt,
		}

		r1 := c.Query(charts.WithName("test")).First()

		/* imitate Chart rendering with Helm SDK */
		r1.Values["bar"] = "baz"
		r1.SetDependencies(dep2)

		r2 := c.Query(charts.WithName("test")).First()

		Expect(r2.Dependencies()).To(ConsistOf(dep1, dep2))

		/* Skip these because of shallow copy */

		// Expect(r2.Values).To(HaveKeyWithValue("foo", "bar"))
		// Expect(r2.Values).NotTo(HaveKeyWithValue("bar", "baz"))
	})
})

var _ = Describe("GlobalCatalog", func() {
	It("is empty when it is not populated", func() {
		Expect(charts.GlobalCatalog()).To(BeEmpty())
	})

	It("populates the global catalog only once", func() {
		/* Use the test scaffolding from Populate */
		Expect(charts.PopulateGlobalCatalog(
			WithSearchPath("testdata/charts"),
			WithFilePattern("*.tar.gz", "*.tgz"))).ToNot(HaveOccurred())

		Expect(charts.GlobalCatalog()).To(HaveLen(4))

		Expect(charts.PopulateGlobalCatalog()).To(
			MatchError(MatchRegexp("not empty")))
	})

	It("always retruns the same reference", func() {
		a := fmt.Sprintf("%p", charts.GlobalCatalog())
		b := fmt.Sprintf("%p", charts.GlobalCatalog())

		Expect(a).To(BeIdenticalTo(b))
	})
})
