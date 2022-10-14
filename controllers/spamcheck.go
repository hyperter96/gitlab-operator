package controllers

import (
	"context"

	gitlabctl "gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab"
)

func (r *GitLabReconciler) reconcileSpamcheck(ctx context.Context, adapter gitlab.Adapter, template helm.Template) error {
	if err := r.reconcileSpamcheckConfigMap(ctx, adapter, template); err != nil {
		return err
	}

	if err := r.reconcileSpamcheckDeployment(ctx, adapter, template); err != nil {
		return err
	}

	if err := r.reconcileSpamcheckService(ctx, adapter, template); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileSpamcheckConfigMap(ctx context.Context, adapter gitlab.Adapter, template helm.Template) error {
	if err := r.createOrPatch(ctx, gitlabctl.SpamcheckConfigMap(template), adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileSpamcheckDeployment(ctx context.Context, adapter gitlab.Adapter, template helm.Template) error {
	spamcheck := gitlabctl.SpamcheckDeployment(template)

	if err := r.setDeploymentReplica(ctx, spamcheck); err != nil {
		return err
	}

	if err := r.annotateSecretsChecksum(ctx, adapter, spamcheck); err != nil {
		return err
	}

	err := r.createOrPatch(ctx, spamcheck, adapter)

	return err
}

func (r *GitLabReconciler) reconcileSpamcheckService(ctx context.Context, adapter gitlab.Adapter, template helm.Template) error {
	if err := r.createOrPatch(ctx, gitlabctl.SpamcheckService(template), adapter); err != nil {
		return err
	}

	return nil
}
