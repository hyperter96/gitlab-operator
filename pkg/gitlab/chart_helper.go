package gitlab

import (
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support/charts"
)

/*
 * NOTICE: These functions are for test purposes. They may change or removed
 *         without notice. Do not use them directly.
 */

func SetChartCatalog(catalog charts.Catalog) {
	*internalChartCatalog = catalog
}

func GetChartCatalog() charts.Catalog {
	if internalChartCatalog == nil {
		return charts.GlobalCatalog()
	} else {
		return *internalChartCatalog
	}
}

var (
	internalChartCatalog *charts.Catalog = nil
)
