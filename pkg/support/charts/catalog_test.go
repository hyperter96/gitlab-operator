package charts

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"helm.sh/helm/v3/pkg/chart"
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
})
