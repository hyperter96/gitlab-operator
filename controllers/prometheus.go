package controllers

import (
	"context"

	gitlabctl "gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab"
)

// reconcilePrometheus reconciles all Prometheus chart objects (Prometheus server, alertmanager, node exporter and pushgateway).
func (r *GitLabReconciler) reconcilePrometheus(ctx context.Context, adapter gitlab.Adapter, template helm.Template) error {
	pvcs := gitlabctl.PrometheusPersistentVolumeClaims(template)
	for _, pvc := range pvcs {
		if err := r.createOrPatch(ctx, pvc, adapter); err != nil {
			return err
		}
	}

	configmaps := gitlabctl.PrometheusConfigMaps(template)
	for _, cm := range configmaps {
		if err := r.createOrPatch(ctx, cm, adapter); err != nil {
			return err
		}
	}

	services := gitlabctl.PrometheusServices(template)
	for _, svc := range services {
		if err := r.createOrPatch(ctx, svc, adapter); err != nil {
			return err
		}
	}

	deployments := gitlabctl.PrometheusDeployments(template)
	for _, dpl := range deployments {
		if err := r.createOrPatch(ctx, dpl, adapter); err != nil {
			return err
		}
	}

	statefulSets := gitlabctl.PrometheusStatefulSets(template)
	for _, ss := range statefulSets {
		if err := r.createOrPatch(ctx, ss, adapter); err != nil {
			return err
		}
	}

	daemonSets := gitlabctl.PrometheusDaemonSets(template)
	for _, ds := range daemonSets {
		if err := r.createOrPatch(ctx, ds, adapter); err != nil {
			return err
		}
	}

	ingresses := gitlabctl.PrometheusIngresses(template)
	for _, ing := range ingresses {
		if err := r.createOrPatch(ctx, ing, adapter); err != nil {
			return err
		}
	}

	return nil
}
