package controllers

import (
	"context"

	gitlabctl "gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/internal"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab"
)

func (r *GitLabReconciler) reconcileSidekiqConfigMaps(ctx context.Context, adapter gitlab.Adapter, template helm.Template) error {
	for _, cm := range gitlabctl.SidekiqConfigMaps(template) {
		if err := r.createOrPatch(ctx, cm, adapter); err != nil {
			return err
		}
	}

	return nil
}

func (r *GitLabReconciler) reconcileSidekiqDeployments(ctx context.Context, adapter gitlab.Adapter, template helm.Template, pause bool) error {
	sidekiqs := gitlabctl.SidekiqDeployments(template)

	for _, sidekiq := range sidekiqs {
		if err := r.setDeploymentReplica(ctx, sidekiq); err != nil {
			return err
		}

		if err := r.annotateSecretsChecksum(ctx, adapter, sidekiq); err != nil {
			return err
		}

		if err := internal.ToggleDeploymentPause(sidekiq, pause); err != nil {
			return err
		}

		if err := r.createOrPatch(ctx, sidekiq, adapter); err != nil {
			return err
		}
	}

	return nil
}
