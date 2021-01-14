package gitlab

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/helm"
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
		"reference", adapter.Reference(),
		"hash", adapter.Hash())

	entry := store.lookup(adapter.Reference())
	if entry != nil {
		if entry.hash == adapter.Hash() {

			logger.V(1).Info("Retreiving the cached template")

			return entry.template, nil
		}

		logger.V(1).Info("Template signature has changed. Evicting it now.")

		store.evict(adapter.Reference())
	}

	logger.Info("Rendering a new template.")

	logger.V(2).Info("Rendering a new template.",
		"namespace", adapter.Namespace(),
		"releaseName", adapter.ReleaseName(),
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
	builder.DisableHooks()

	return builder.Render(adapter.Values())
}

func getChartArchive(chartVersion string) string {
	logger := templateLogger.WithValues(
		"chartVersion", chartVersion).V(1)

	chartsDir := os.Getenv("HELM_CHARTS")
	if chartsDir == "" {
		chartsDir = "/charts"
	}

	logger.Info("Looking for the designated GitLab Chart in the specified directory.", "directory", chartsDir)

	return filepath.Join(chartsDir,
		fmt.Sprintf("gitlab-%s.tgz", chartVersion))
}

// AvailableChartVersions lists the version of available GitLab Charts.
func AvailableChartVersions() []string {
	versions := []string{}

	chartsDir := os.Getenv("HELM_CHARTS")
	if chartsDir == "" {
		chartsDir = "/charts"
	}

	filepath.Walk(chartsDir, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(info.Name()) == ".tgz" {
			comp := strings.Split(info.Name()[:len(info.Name())-4], "-")
			if len(comp) == 2 {
				versions = append(versions, comp[1])
			}
		}
		return nil
	})

	sort.Sort(sort.Reverse(sort.StringSlice(versions)))

	return versions
}
