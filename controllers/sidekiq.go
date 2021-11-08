package controllers

import (
	"context"

	gitlabctl "gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
)

func (r *GitLabReconciler) reconcileSidekiqConfigMaps(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	for _, cm := range gitlabctl.SidekiqConfigMaps(adapter) {
		if _, err := r.createOrPatch(ctx, cm, adapter); err != nil {
			return err
		}
	}

	return nil
}

func (r *GitLabReconciler) reconcileSidekiqDeployments(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, pause bool) error {
	sidekiqs := gitlabctl.SidekiqDeployments(adapter)

	for _, sidekiq := range sidekiqs {
		if err := r.setDeploymentReplica(ctx, sidekiq); err != nil {
			return err
		}

		if err := r.annotateSecretsChecksum(ctx, adapter, &sidekiq.Spec.Template); err != nil {
			return err
		}

		if pause {
			sidekiq.Spec.Paused = true
		} else {
			sidekiq.Spec.Paused = false
		}

		if _, err := r.createOrPatch(ctx, sidekiq, adapter); err != nil {
			return err
		}
	}

	return nil
}
