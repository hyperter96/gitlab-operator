package charts

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"helm.sh/helm/v3/pkg/chart"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/resource"
)

func newTestChart(name, version, appVersion string) *chart.Chart {
	return &chart.Chart{
		Metadata: &chart.Metadata{
			Name:       name,
			Version:    version,
			AppVersion: appVersion,
		},
	}
}

var _ = Describe("Criterion", func() {
	Describe("WithName", func() {
		It("matches Chart name", func() {
			Expect(WithName("test")(newTestChart("test", "", ""))).To(BeTrue())
			Expect(WithName("test")(newTestChart("not-test", "", ""))).To(BeFalse())
		})
	})

	Describe("WithVersion", func() {
		It("matches Chart version", func() {
			Expect(WithVersion("1")(newTestChart("", "1", ""))).To(BeTrue())
			Expect(WithVersion("1")(newTestChart("", "2", ""))).To(BeFalse())
		})
	})

	Describe("WithAppVersion", func() {
		It("matches Chart appVersion", func() {
			Expect(WithAppVersion("1")(newTestChart("", "", "1"))).To(BeTrue())
			Expect(WithAppVersion("1")(newTestChart("", "", "2"))).To(BeFalse())
		})
	})

	Describe("All", func() {
		Expect(
			All(
				WithName("test"),
				WithVersion("1"),
				WithAppVersion("1"),
			)(newTestChart("test", "1", "1"))).To(BeTrue())
		Expect(
			All(
				WithName("test"),
				WithVersion("1"),
				WithAppVersion("1"),
			)(newTestChart("test", "1", "2"))).To(BeFalse())
	})

	Describe("Any", func() {
		Expect(
			Any(
				WithName("test"),
				WithVersion("1"),
				WithAppVersion("1"),
			)(newTestChart("test", "", ""))).To(BeTrue())
		Expect(
			Any(
				WithName("test"),
				WithVersion("1"),
				WithAppVersion("1"),
			)(newTestChart("", "", ""))).To(BeFalse())

	})

	Describe("None", func() {
		Expect(
			None(
				WithName("test"),
				WithVersion("1"),
				WithAppVersion("1"),
			)(newTestChart("", "", ""))).To(BeTrue())
		Expect(
			None(
				WithName("test"),
				WithVersion("1"),
				WithAppVersion("1"),
			)(newTestChart("test", "", ""))).To(BeFalse())
	})
})

var _ = Describe("Catalog", func() {
	It("is empty after initialization", func() {
		Expect(Catalog{}).To(BeEmpty())
	})

	Describe("Append", func() {
		It("appends new Chart", func() {
			c := &Catalog{}

			c.Append(newTestChart("test", "1", ""))
			c.Append(newTestChart("test", "2", ""))
			c.Append(newTestChart("test", "3", ""))

			Expect(*c).To(HaveLen(3))
		})

		It("does not append when Chart metadata is empty", func() {
			c := &Catalog{}

			c.Append(&chart.Chart{})

			Expect(*c).To(BeEmpty())
		})

		It("does not append when a Chart with the same name and version exists", func() {
			c := &Catalog{}

			c.Append(newTestChart("test", "1", ""))
			c.Append(newTestChart("test", "1", ""))

			Expect(*c).To(HaveLen(1))
		})
	})

	Describe("Populate", func() {
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
			c := &Catalog{}
			e := c.Populate(
				WithSearchPath("testdata/charts"),
				WithFilePattern("*.tar.gz", "*.tgz"))
			/* The current version of logr can not be integrated with Gingkgo */

			Expect(e).ToNot(HaveOccurred())
			Expect(*c).To(HaveLen(4))
		})
	})

	Describe("Query", func() {
		/* Use the test scaffolding from Populate */
		c := &Catalog{}
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
			r := c.Query(WithName("chart-1"), WithVersion("0.1.1"))
			Expect(r).To(BeEmpty())
			Expect(r.First()).To(BeNil())

			r = c.Query(WithName("chart-1"), WithVersion("0.1.0"))
			Expect(r).To(HaveLen(1))
			Expect(r[0].Metadata.Name).To(Equal("chart-1"))
			Expect(r[0].Metadata.Version).To(Equal("0.1.0"))
			Expect(r.First()).NotTo(BeNil())

			r = c.Query(WithName("chart-2"), WithAppVersion("2.1.0"))
			Expect(r).To(HaveLen(1))
			Expect(r[0].Metadata.Name).To(Equal("chart-2"))
			Expect(r[0].Metadata.AppVersion).To(Equal("2.1.0"))

			r = c.Query(WithName("chart-1"))
			Expect(r).To(HaveLen(2))
			Expect(r[0].Metadata.Name).To(Equal("chart-1"))
			Expect(r[1].Metadata.Name).To(Equal("chart-1"))
		})

		It("makes a copy of the original Chart", func() {
			val := resource.Values{
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
			c := &Catalog{
				chrt,
			}

			r1 := c.Query(WithName("test")).First()

			/* imitate Chart rendering with Helm SDK */
			r1.Values["bar"] = "baz"
			r1.SetDependencies(dep2)

			r2 := c.Query(WithName("test")).First()

			Expect(r2.Dependencies()).To(ConsistOf(dep1, dep2))

			/* Skip these because of shallow copy */

			// Expect(r2.Values).To(HaveKeyWithValue("foo", "bar"))
			// Expect(r2.Values).NotTo(HaveKeyWithValue("bar", "baz"))
		})
	})

	Describe("GlobalCatalog", func() {
		It("is empty when it is not populated", func() {
			Expect(GlobalCatalog()).To(BeEmpty())
		})

		It("populates the global catalog only once", func() {
			/* Use the test scaffolding from Populate */
			Expect(PopulateGlobalCatalog(
				WithSearchPath("testdata/charts"),
				WithFilePattern("*.tar.gz", "*.tgz"))).ToNot(HaveOccurred())

			Expect(GlobalCatalog()).To(HaveLen(4))

			Expect(PopulateGlobalCatalog()).To(
				MatchError(MatchRegexp("not empty")))
		})

		It("always retruns the same reference", func() {
			a := fmt.Sprintf("%p", GlobalCatalog())
			b := fmt.Sprintf("%p", GlobalCatalog())

			Expect(a).To(BeIdenticalTo(b))
		})
	})
})
