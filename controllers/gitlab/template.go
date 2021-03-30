package gitlab

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"sync"

	"github.com/Masterminds/semver"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/settings"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/helm"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	store          = newTemplateStore()
	templateLogger = ctrl.Log.WithName("template")
)

// GetTemplate ensures that only one instance of Helm template exists per deployment and
// it is rendered only when needed, e.g. it has changed.
func GetTemplate(adapter CustomResourceAdapter) (helm.Template, error) {
	logger := templateLogger.WithValues(
		"namespace", adapter.Namespace(),
		"releaseName", adapter.ReleaseName(),
		"hash", adapter.Hash())

	entry := store.lookup(adapter.Reference())
	if entry != nil {
		if entry.hash == adapter.Hash() {

			logger.V(2).Info("Retrieving the cached template")

			return entry.template, nil
		}

		logger.V(1).Info("Template signature has changed. Evicting it now.")

		store.evict(adapter.Reference())
	}

	logger.Info("Rendering a new template.")

	logger.V(2).Info("Rendering a new template.",
		"values", adapter.Values())

	template, err := buildTemplate(adapter)
	if err != nil {

		logger.Error(err, "Failed to render the template")

		return nil, err
	}

	logger.V(1).Info("The template is rendered. Check the warnings (if any).",
		"warnings", len(template.Warnings()))

	if logger.V(1).Enabled() {
		for _, w := range template.Warnings() {
			logger.V(1).Info("Warning: An issue occurred while rendering the Helm template",
				"issue", w)
		}
	}

	logger.V(1).Info("Caching the template.")

	entry = store.update(adapter.Reference(), &digestedTemplate{
		hash:     adapter.Hash(),
		template: template,
	})

	return entry.template, nil
}

// AvailableChartVersions lists the version of available GitLab Charts.
// The values are sorted from newest to oldest (semantic versioning).
func AvailableChartVersions() []string {
	versions := []*semver.Version{}

	chartsDir := os.Getenv("HELM_CHARTS")
	if chartsDir == "" {
		chartsDir = "/charts"
	}

	re := regexp.MustCompile(`gitlab\-((0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?)\.tgz`)

	filepath.Walk(chartsDir, func(path string, info os.FileInfo, err error) error {
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

type digestedTemplate struct {
	hash     string
	template helm.Template
}

type internalTemplateStore struct {
	inventory map[string]*digestedTemplate
	locker    sync.Locker
}

func (s *internalTemplateStore) lookup(reference string) *digestedTemplate {
	s.locker.Lock()
	defer s.locker.Unlock()

	return s.inventory[reference]
}

func (s *internalTemplateStore) evict(reference string) {
	s.locker.Lock()
	defer s.locker.Unlock()

	s.inventory[reference] = nil
}

func (s *internalTemplateStore) update(reference string, entry *digestedTemplate) *digestedTemplate {
	s.locker.Lock()
	defer s.locker.Unlock()

	current := s.inventory[reference]
	if current != nil && current.hash == entry.hash {
		return current
	}

	s.inventory[reference] = entry
	return entry
}

func newTemplateStore() *internalTemplateStore {
	return &internalTemplateStore{
		inventory: map[string]*digestedTemplate{},
		locker:    &sync.Mutex{},
	}
}

func buildTemplate(adapter CustomResourceAdapter) (helm.Template, error) {
	builder := helm.NewBuilder(getChartArchive(adapter.ChartVersion()))

	builder.SetNamespace(adapter.Namespace())
	builder.SetReleaseName(adapter.ReleaseName())
	builder.EnableHooks()

	return builder.Render(adapter.Values())
}

func getChartArchive(chartVersion string) string {
	logger := templateLogger.WithValues(
		"chartVersion", chartVersion)

	logger.V(1).Info("Looking for the designated GitLab Chart in the specified directory.",
		"directory", settings.HelmChartsDirectory)

	return filepath.Join(settings.HelmChartsDirectory,
		fmt.Sprintf("gitlab-%s.tgz", chartVersion))
}
