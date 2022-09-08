package v1beta1

import (
	semver "github.com/Masterminds/semver/v3"
)

/* GitLabOperation */

func (w *Adapter) IsInstall() bool {
	return w.CurrentVersion() == ""
}

func (w *Adapter) IsUpgrade() bool {
	return !w.IsInstall() && (w.compareVersions() > 0)
}

func (w *Adapter) IsDowngrade() bool {
	return !w.IsInstall() && (w.compareVersions() < 0)
}

func (w *Adapter) CurrentVersion() string {
	return w.source.Status.Version
}

func (w *Adapter) DesiredVersion() string {
	return w.source.Spec.Chart.Version
}

/* Helpers */

func (w *Adapter) chartVersion() *semver.Version {
	if ver, err := semver.NewVersion(w.DesiredVersion()); err != nil {
		return nil
	} else {
		return ver
	}
}

func (w *Adapter) statusVersion() *semver.Version {
	if ver, err := semver.NewVersion(w.CurrentVersion()); err != nil {
		return nil
	} else {
		return ver
	}
}

func (w *Adapter) compareVersions() int {
	chartVersion := w.chartVersion()
	if chartVersion == nil {
		return 0
	}

	statusVersion := w.statusVersion()
	if statusVersion == nil {
		return 0
	}

	return chartVersion.Compare(statusVersion)
}
