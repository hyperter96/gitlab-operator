package charts

import (
	"sync"

	"github.com/pkg/errors"
)

// GlobalCatalog returns the global Chart catalog. This catalog is created
// once and is accessible globally.
//
// Do not change the content of this catalog directly.
func GlobalCatalog() Catalog {
	return *globalCatalog
}

// PopulateGlobalCatalog uses the provided options to populate the existing
// Charts into the global Chart catalog.
//
// Call this function only once when the controller initializes.
func PopulateGlobalCatalog(options ...PopulateOption) error {
	catalogMutex.Lock()
	defer catalogMutex.Unlock()

	if len(*globalCatalog) > 0 {
		return errors.New("catalog is not empty")
	}

	return globalCatalog.Populate(options...)
}

/* Private */

var (
	catalogMutex  sync.Mutex
	globalCatalog *Catalog = &Catalog{}
)
