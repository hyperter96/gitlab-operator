package populate

import (
	"context"

	"github.com/go-logr/logr"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support/charts"
)

/* Chart Populate Options */

// WithSearchPath configures Chart population with the provided search paths.
//
// By default Chart population uses the current directory unless this option is
// used.
//
// Note that passing an empty list of search paths will not override the current
// search path.
func WithSearchPath(paths ...string) charts.PopulateOption {
	return func(cfg *charts.PopulateConfig) {
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
func WithFilePattern(patterns ...string) charts.PopulateOption {
	return func(cfg *charts.PopulateConfig) {
		if len(patterns) > 0 {
			cfg.FilePatterns = patterns
		}
	}
}

// WithContext configures Chart population with the logger from the context.
func WithContext(ctx context.Context) charts.PopulateOption {
	return func(cfg *charts.PopulateConfig) {
		cfg.Logger = logr.FromContextOrDiscard(ctx)
	}
}

// WithLogger configures Chart population with the specified logger.
func WithLogger(logger logr.Logger) charts.PopulateOption {
	return func(cfg *charts.PopulateConfig) {
		cfg.Logger = logger
	}
}
