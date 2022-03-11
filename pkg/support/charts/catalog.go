package charts

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
)

// Catalog is a list of available Helm Charts. Use different criteria with Query
// functions to select the desired Charts from the catalog.
type Catalog []*chart.Chart

// Criterion is a single criterion for querying Chart catalog. If a Chart
// matches the criterion it must return true.
type Criterion = func(*chart.Chart) bool

// PopulateConfig is the configuration used for populating available Helm
// Charts to the controller.
//
// Currently it only supports searching the local file system with a set of
// search paths and file name patterns.
type PopulateConfig struct {
	Logger       logr.Logger
	SearchPaths  []string
	FilePatterns []string

	catalog *Catalog
}

// PopulateOption represents an individual Chart population option. The
// available options are:
//
//   - WithSearchPath
//   - WithFilePattern
//   - WithLogger
//   - WithContext
//
// See each option for further details.
type PopulateOption = func(*PopulateConfig)

// Query selects any chart that matches all of the specified criteria. It can
// return an empty list when it can not find any match.
//
// Each chart must match all the criteria. Use alternative criteria builders for
// different matching requirements.
func (c Catalog) Query(criteria ...Criterion) Catalog {
	result := Catalog{}

	if len(criteria) == 0 {
		return result
	}

	for _, chart := range c {
		if All(criteria...)(chart) {
			result = append(result, clone(chart))
		}
	}

	return result
}

// Empty returns true when the catalog is empty. This is useful to check the
// results from the Query function.
func (c Catalog) Empty() bool {
	return len(c) == 0
}

// First returns the first element of the catalog or nil if the catalog is
// empty. This is useful to retrieve results from the Query function.
func (c Catalog) First() *chart.Chart {
	if len(c) > 0 {
		return c[0]
	}

	return nil
}

// Names returns the list of the names of the Charts in this catalog.
func (c Catalog) Names() []string {
	return c.collect(func(chart *chart.Chart) string {
		return chart.Metadata.Name
	})
}

// Versions returns the list of the available versions of the named Chart in
// this catalog.
func (c Catalog) Versions(name string) []string {
	return c.collect(func(chart *chart.Chart) string {
		if chart.Metadata.Name == name {
			return chart.Metadata.Version
		} else {
			return ""
		}
	})
}

// AppVersions returns the list of the available appVersions of the named Chart
// in this catalog.
func (c Catalog) AppVersions(name string) []string {
	return c.collect(func(chart *chart.Chart) string {
		if chart.Metadata.Name == name {
			return chart.Metadata.AppVersion
		} else {
			return ""
		}
	})
}

// Append adds a new chart to the catalog. It ensures that the new chart has a
// valid metadata and a chart with the same name and version does not exist in
// the catalog.
func (c *Catalog) Append(chart *chart.Chart) {
	if chart.Metadata == nil {
		return
	}

	for _, i := range *c {
		if i.Metadata.Name == chart.Metadata.Name && i.Metadata.Version == chart.Metadata.Version {
			return
		}
	}

	*c = append(*c, chart)
}

// Populate uses the provided options to populate the existing Charts into the
// catalog.
//
// Currently it can only populate Charts from the local file system using a set
// of search paths and file name patterns. If a directory or an archive file in
// the specified search paths contain a chart it loads it and appends it to the
// catalog.
func (c *Catalog) Populate(options ...PopulateOption) error {
	cfg := defaultPopulateConfig()
	cfg.applyConfig(options)

	return cfg.populate(c)
}

func (c Catalog) collect(operator func(*chart.Chart) string) []string {
	col := map[string]bool{}

	for _, chart := range c {
		out := operator(chart)
		if out != "" {
			col[out] = true
		}
	}

	i := 0
	result := make([]string, len(col))

	for k := range col {
		result[i] = k
		i++
	}

	return result
}

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
	if len(*globalCatalog) > 0 {
		return errors.New("catalog is not empty")
	}

	return globalCatalog.Populate(options...)
}

/* Query Criteria */

// WithName matches the Chart name.
func WithName(name string) Criterion {
	return func(chart *chart.Chart) bool {
		return chart.Metadata.Name == name
	}
}

// WithName matches the Chart version.
func WithVersion(version string) Criterion {
	return func(chart *chart.Chart) bool {
		return chart.Metadata.Version == version
	}
}

// WithName matches the Chart appVersion.
func WithAppVersion(appVersion string) Criterion {
	return func(chart *chart.Chart) bool {
		return chart.Metadata.AppVersion == appVersion
	}
}

// Any combines the provided Chart query criteria and succeeds when all of them
// return true.
func All(criteria ...Criterion) Criterion {
	return func(chart *chart.Chart) bool {
		for _, criterion := range criteria {
			if !criterion(chart) {
				return false
			}
		}

		return true
	}
}

// Any combines the provided Chart query criteria and succeeds when any of them
// returns true.
func Any(criteria ...Criterion) Criterion {
	return func(chart *chart.Chart) bool {
		for _, criterion := range criteria {
			if criterion(chart) {
				return true
			}
		}

		return false
	}
}

// None combines the provided Chart query criteria and succeeds when none of
// them return true.
func None(criteria ...Criterion) Criterion {
	return func(chart *chart.Chart) bool {
		for _, criterion := range criteria {
			if criterion(chart) {
				return false
			}
		}

		return true
	}
}

/* Chart Populate Options */

// WithSearchPath configures Chart population with the provided search paths.
//
// By default Chart population uses the current directory unless this option is
// used.
//
// Note that passing an empty list of search paths will not override the current
// search path.
func WithSearchPath(paths ...string) PopulateOption {
	return func(cfg *PopulateConfig) {
		if len(paths) > 0 {
			cfg.SearchPaths = paths
		}
	}
}

// WithFilePattern configures Chart population with the provided file name
// patterns. Each pattern must be a valid glob pattern. These patterns are
// only applied on file names not directories.
//
// By default Chart population uses `*.tgz` unless this option is used.
//
// Note that passing an empty list of patterns will not override the current
// file name patterns.
func WithFilePattern(patterns ...string) PopulateOption {
	return func(cfg *PopulateConfig) {
		if len(patterns) > 0 {
			cfg.FilePatterns = patterns
		}
	}
}

// WithContext configures Chart population with the logger from the context.
func WithContext(ctx context.Context) PopulateOption {
	return func(cfg *PopulateConfig) {
		cfg.Logger = logr.FromContextOrDiscard(ctx)
	}
}

// WithLogger configures Chart population with the specified logger.
func WithLogger(logger logr.Logger) PopulateOption {
	return func(cfg *PopulateConfig) {
		cfg.Logger = logger
	}
}

/* PopulateConfig */

func defaultPopulateConfig() *PopulateConfig {
	defaultSearchPath, err := os.Getwd()
	if err != nil {
		defaultSearchPath = "."
	}

	return &PopulateConfig{
		Logger:       logr.Discard(),
		SearchPaths:  []string{defaultSearchPath},
		FilePatterns: []string{"*.tgz"},
	}
}

func (c *PopulateConfig) applyConfig(options []PopulateOption) {
	for _, option := range options {
		option(c)
	}

	c.Logger.WithName("Charts")
}

func (c *PopulateConfig) populate(catalog *Catalog) error {
	/* Attach to the catalog */
	c.catalog = catalog

	c.Logger.V(2).Info("searching directories for matching charts",
		"searchPaths", c.SearchPaths,
		"filePatterns", c.FilePatterns)

	for _, path := range c.SearchPaths {
		_ = filepath.WalkDir(path, c.processDirEntry)
	}

	return nil
}

func (c *PopulateConfig) processDirEntry(path string, d fs.DirEntry, e error) error {
	if e != nil {
		c.Logger.V(2).Error(e,
			"error occurred while searching directory",
			"path", path)

		return nil
	}

	if d.IsDir() {
		return c.tryEntryAsChart(path, true)
	} else {
		for _, pattern := range c.FilePatterns {
			matched, err := filepath.Match(pattern, d.Name())

			if matched && err == nil {
				c.Logger.V(2).Info("found a matching file",
					"path", path)

				return c.tryEntryAsChart(path, false)
			}

			if err != nil {
				c.Logger.V(2).Error(err,
					"error occurred while matching file name to the pattern",
					"path", path,
					"pattern", pattern)
			}
		}
	}

	return nil
}

func (c *PopulateConfig) tryEntryAsChart(path string, isDir bool) error {
	c.Logger.V(2).Info("trying entry as chart",
		"path", path,
		"isDirectory", isDir)

	chart, err := loader.Load(path)
	if err != nil {
		c.Logger.V(2).Error(err,
			"entry does not contain a chart",
			"path", path,
			"isDirectory", isDir)
	} else {
		c.Logger.V(2).Info("entry contains a chart",
			"path", path,
			"isDirectory", isDir)

		c.catalog.Append(chart)

		c.Logger.Info("chart added to the catalog",
			"chartName", chart.Metadata.Name,
			"chartVersion", chart.Metadata.Version)

		if isDir {
			return filepath.SkipDir
		}
	}

	return nil
}

func clone(in *chart.Chart) *chart.Chart {
	/*
	 *  CAUTION:
	 *
	 *   This does a shallow copy of the incoming Chart structure and a true
	 *   clone of it. But for our purpose this is enough and we do not use
	 *   a reflection-based deep copy.
	 *
	 */
	out := *in
	return &out
}

/* Private */

var (
	globalCatalog *Catalog = &Catalog{}
)
