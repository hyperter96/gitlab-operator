package controllers

import (
	"context"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/internal"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/settings"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab"
)

func prometheusSupported() bool {
	return settings.IsGroupVersionKindSupported("monitoring.coreos.com/v1", "Prometheus")
}

func (r *GitLabReconciler) reconcilePrometheus(ctx context.Context, adapter gitlab.Adapter) error {
	service := internal.ExposePrometheusCluster(adapter)
	if err := r.createOrPatch(ctx, service, adapter); err != nil {
		return err
	}

	prometheus := internal.PrometheusCluster(adapter)
	if err := r.createOrPatch(ctx, prometheus, adapter); err != nil {
		return err
	}

	return nil
}
