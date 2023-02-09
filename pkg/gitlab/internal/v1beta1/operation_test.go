package v1beta1

import (
	"context"
	"sort"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	semver "github.com/Masterminds/semver/v3"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support/charts"
)

func TestIsInstall(t *testing.T) {
	When("testing IsInstall", func() {
		testCases := []struct {
			name               string
			statusVersionIndex int
			expected           bool
		}{
			{
				name:               "returns true for new install",
				statusVersionIndex: -1,
				expected:           true,
			},
			{
				name:               "returns false for existing install",
				statusVersionIndex: 0,
				expected:           false,
			},
		}

		for _, tc := range testCases {
			It(tc.name, func() {
				availableVersions := getAvailableVersions(t)
				a := createAdapter(0, tc.statusVersionIndex, availableVersions)
				got := a.IsInstall()

				Expect(tc.expected).To(Equal(got))
			})
		}
	})
}

func TestIsUpgrade(t *testing.T) {
	When("testing IsUpgrade", func() {
		testCases := []struct {
			name                                  string
			statusVersionIndex, chartVersionIndex int
			expected                              bool
		}{
			{
				name:               "new install",
				chartVersionIndex:  0,
				statusVersionIndex: -1,
				expected:           false,
			},
			{
				name:               "wanting an older version",
				chartVersionIndex:  1,
				statusVersionIndex: 2,
				expected:           false,
			},
			{
				name:               "wanting an equal version",
				chartVersionIndex:  2,
				statusVersionIndex: 2,
				expected:           false,
			},
			{
				name:               "wanting a newer version",
				chartVersionIndex:  1,
				statusVersionIndex: 0,
				expected:           true,
			},
		}

		availableVersions := getAvailableVersions(t)

		for _, tc := range testCases {
			It(tc.name, func() {
				a := createAdapter(tc.chartVersionIndex, tc.statusVersionIndex, availableVersions)
				got := a.IsUpgrade()

				Expect(tc.expected).To(Equal(got))
			})
		}
	})
}

func TestIsDowngrade(t *testing.T) {
	When("testing IsDowngrade", func() {
		testCases := []struct {
			name                                  string
			statusVersionIndex, chartVersionIndex int
			expected                              bool
		}{
			{
				name:               "new install",
				chartVersionIndex:  0,
				statusVersionIndex: -1,
				expected:           false,
			},
			{
				name:               "wanting an older version",
				chartVersionIndex:  1,
				statusVersionIndex: 2,
				expected:           true,
			},
			{
				name:               "wanting an equal version",
				chartVersionIndex:  2,
				statusVersionIndex: 2,
				expected:           false,
			},
			{
				name:               "wanting a newer version",
				chartVersionIndex:  1,
				statusVersionIndex: 0,
				expected:           false,
			},
		}

		availableVersions := getAvailableVersions(t)

		for _, tc := range testCases {
			It(tc.name, func() {
				a := createAdapter(tc.chartVersionIndex, tc.statusVersionIndex, availableVersions)
				got := a.IsDowngrade()

				Expect(tc.expected).To(Equal(got))
			})
		}
	})
}

func TestCompareVersions(t *testing.T) {
	When("comparing versions", func() {
		testCases := []struct {
			name                                  string
			chartVersionIndex, statusVersionIndex int
			expected                              int
		}{
			{
				name:               "finds no difference for matching versions (index 0)",
				chartVersionIndex:  0,
				statusVersionIndex: 0,
				expected:           0,
			},
			{
				name:               "finds no difference for matching versions (index 1)",
				chartVersionIndex:  1,
				statusVersionIndex: 1,
				expected:           0,
			},
			{
				name:               "finds a difference for mismatching versions (greater)",
				chartVersionIndex:  2,
				statusVersionIndex: 1,
				expected:           1,
			},
			{
				name:               "finds a difference for mismatching versions (lesser)",
				chartVersionIndex:  1,
				statusVersionIndex: 2,
				expected:           -1,
			},
		}

		availableVersions := getAvailableVersions(t)

		for _, tc := range testCases {
			It(tc.name, func() {
				a := createAdapter(tc.chartVersionIndex, tc.statusVersionIndex, availableVersions)
				got := a.compareVersions()

				Expect(tc.expected).To(Equal(got))
			})
		}
	})
}

func TestChartVersion(t *testing.T) {
	When("testing ChartVersion", func() {
		availableVersions := getAvailableVersions(t)
		chartVersionIndices := []int{0, 1, 2}

		for _, tc := range chartVersionIndices {
			It("returns a valid chartVersion", func() {
				a := createAdapter(tc, 0, availableVersions)

				want := availableVersions[tc].String()
				got := a.chartVersion().String()

				Expect(want).To(Equal(got))
			})
		}
	})
}

func TestStatusVersion(t *testing.T) {
	When("testing statusVersion", func() {
		availableVersions := getAvailableVersions(t)
		chartVersionIndices := []int{0, 1, 2}

		for _, tc := range chartVersionIndices {
			It("returns a valid chartVersion", func() {
				a := createAdapter(0, tc, availableVersions)

				want := availableVersions[tc].String()
				got := a.statusVersion().String()

				Expect(want).To(Equal(got))
			})
		}
	})
}

func getAvailableVersions(t *testing.T) []*semver.Version {
	availableVersions := charts.GlobalCatalog().Versions("gitlab")

	vs := make([]*semver.Version, len(availableVersions))

	for i, r := range availableVersions {
		v, err := semver.NewVersion(r)
		if err != nil {
			t.Errorf("error parsing version: %s", err)
		}

		vs[i] = v
	}

	sort.Sort(semver.Collection(vs))

	return vs
}

func createAdapter(chartVersionIndex, statusVersionIndex int, availableVersions []*semver.Version) *Adapter {
	chartVersion := availableVersions[chartVersionIndex]

	gl := newGitLabResource(chartVersion.String(), support.Values{})

	if statusVersionIndex != -1 {
		statusVersion := availableVersions[statusVersionIndex]
		gl.Status.Version = statusVersion.String()
	}

	a, err := NewAdapter(context.TODO(), gl)
	Expect(err).NotTo(HaveOccurred())

	return a
}
