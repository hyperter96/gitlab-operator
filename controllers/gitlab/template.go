package gitlab

import (
	"sync"

	ctrl "sigs.k8s.io/controller-runtime"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
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

	if supported, err := helm.ChartVersionSupported(adapter.ChartVersion()); !supported {
		return nil, err
	}

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
	builder := helm.NewBuilder(helm.GetChartPath(adapter.ChartVersion()))

	builder.SetNamespace(adapter.Namespace())
	builder.SetReleaseName(adapter.ReleaseName())
	builder.EnableHooks()

	return builder.Render(adapter.Values())
}
