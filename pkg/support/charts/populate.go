package charts

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/go-logr/logr"
	"helm.sh/helm/v3/pkg/chart/loader"
)

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

/* PopulateConfig */

func defaultPopulateConfig(catalog *Catalog) *PopulateConfig {
	defaultSearchPath, err := os.Getwd()
	if err != nil {
		defaultSearchPath = "."
	}

	return &PopulateConfig{
		Logger:       logr.Discard(),
		SearchPaths:  []string{defaultSearchPath},
		FilePatterns: []string{"*.tgz"},

		/* Attach to the catalog */
		catalog: catalog,
	}
}

func (c *PopulateConfig) applyConfig(options []PopulateOption) {
	for _, option := range options {
		option(c)
	}

	c.Logger.WithName("Charts")
}

func (c *PopulateConfig) populate() error {
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
