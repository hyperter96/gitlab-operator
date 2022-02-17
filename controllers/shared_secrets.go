package controllers

import (
	"context"

	gitlabctl "gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
)

func (r *GitLabReconciler) runSharedSecretsJob(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) error {
	cfgMap, job, err := gitlabctl.SharedSecretsResources(adapter, template)
	if err != nil {
		return err
	}

	if cfgMap == nil || job == nil {
		r.Log.Info("shared secrets job skipped, not needed per configuration", "gitlab", adapter.Reference())

		return nil
	}

	if _, err := r.createOrPatch(ctx, cfgMap, adapter); err != nil {
		return err
	}

	return r.runJobAndWait(ctx, adapter, job, gitlabctl.SharedSecretsJobTimeout())
}

func (r *GitLabReconciler) runSelfSignedCertsJob(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) error {
	job, err := gitlabctl.SelfSignedCertsJob(adapter, template)
	if err != nil {
		return err
	}

	if job == nil {
		r.Log.Info("self-signed certificates job skipped, not needed per configuration", "gitlab", adapter.Reference())

		return nil
	}

	return r.runJobAndWait(ctx, adapter, job, gitlabctl.SharedSecretsJobTimeout())
}
