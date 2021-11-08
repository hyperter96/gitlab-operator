package controllers

import (
	"context"
	"fmt"

	gitlabctl "gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
)

func (r *GitLabReconciler) runSharedSecretsJob(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	cfgMap, job, err := gitlabctl.SharedSecretsResources(adapter)
	if err != nil {
		return err
	}

	if _, err := r.createOrPatch(ctx, cfgMap, adapter); err != nil {
		return err
	}

	return r.runJobAndWait(ctx, adapter, job, gitlabctl.SharedSecretsJobTimeout())
}

func (r *GitLabReconciler) runSelfSignedCertsJob(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	job, err := gitlabctl.SelfSignedCertsJob(adapter)
	if err != nil {
		return err
	}

	if job == nil {
		return fmt.Errorf("self-signed certificate job skipped, not needed per configuration: %s", adapter.Reference())
	}

	return r.runJobAndWait(ctx, adapter, job, gitlabctl.SharedSecretsJobTimeout())
}