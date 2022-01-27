package helm

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"k8s.io/apimachinery/pkg/runtime"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/settings"
)

// AvailableChartVersions lists the version of available GitLab Charts.
// The values are sorted from newest to oldest (semantic versioning).
func AvailableChartVersions() []string {
	versions := []*semver.Version{}

	chartsDir := os.Getenv("HELM_CHARTS")
	if chartsDir == "" {
		chartsDir = "/charts"
	}

	re := regexp.MustCompile(`gitlab\-((0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?)\.tgz`)

	_ = filepath.Walk(chartsDir, func(path string, info os.FileInfo, err error) error {
		submatches := re.FindStringSubmatch(info.Name())

		if len(submatches) > 1 {
			semver, err := semver.NewVersion(submatches[1])
			if err != nil {
				return err
			}

			versions = append(versions, semver)
		}

		return nil
	})

	// Sort versions from newest to oldest.
	sort.Sort(sort.Reverse(semver.Collection(versions)))

	// Convert list back to strings for compatibility with rest of codebase.
	// NOTE: We can consider returning SemVer objects if we want to do comparisons.
	result := make([]string, len(versions))
	for i, v := range versions {
		result[i] = v.String()
	}

	return result
}

func ChartVersionSupported(version string) (bool, error) {
	for _, v := range AvailableChartVersions() {
		if v == version {
			return true, nil
		}
	}

	return false, fmt.Errorf("chart version %s not supported; please use one of the following: %s", version, strings.Join(AvailableChartVersions(), ", "))
}

func GetChartPath(chartVersion string) string {
	return filepath.Join(settings.HelmChartsDirectory,
		fmt.Sprintf("gitlab-%s.tgz", chartVersion))
}

func GetChartVersion() string {
	version, found := os.LookupEnv("CHART_VERSION")
	if !found {
		version = AvailableChartVersions()[0]
		os.Setenv("CHART_VERSION", version)
	}

	return version
}

// TrueSelector is an ObjectSelector that selects all objects.
var TrueSelector = func(_ runtime.Object) bool {
	return true
}

// FalseSelector is an ObjectSelector that selects no object.
var FalseSelector = func(_ runtime.Object) bool {
	return false
}

type typeMistmatchError struct {
	expected interface{}
	observed interface{}
}

func (e *typeMistmatchError) Error() string {
	return fmt.Sprintf("expected %T, got %T", e.expected, e.observed)
}

func NewTypeMistmatchError(expected, observed interface{}) error {
	return &typeMistmatchError{
		expected: expected,
		observed: observed,
	}
}

// IsTypeMistmatchError returns true if the error is raised because of type mistmatch.
func IsTypeMistmatchError(err error) bool {
	_, ok := err.(*typeMistmatchError)
	return ok
}
