package controllers

import (
	"context"

	gitlabctl "gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab"
)

func (r *GitLabReconciler) reconcilePages(ctx context.Context, adapter gitlab.Adapter, template helm.Template) error {
	if err := r.reconcilePagesConfigMap(ctx, adapter, template); err != nil {
		return err
	}

	if err := r.reconcilePagesDeployment(ctx, adapter, template); err != nil {
		return err
	}

	if err := r.reconcilePagesService(ctx, adapter, template); err != nil {
		return err
	}

	if err := r.reconcilePagesIngress(ctx, adapter, template); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcilePagesConfigMap(ctx context.Context, adapter gitlab.Adapter, template helm.Template) error {
	if err := r.createOrPatch(ctx, gitlabctl.PagesConfigMap(adapter, template), adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcilePagesDeployment(ctx context.Context, adapter gitlab.Adapter, template helm.Template) error {
	pages := gitlabctl.PagesDeployment(template)

	if err := r.setDeploymentReplica(ctx, pages); err != nil {
		return err
	}

	if err := r.annotateSecretsChecksum(ctx, adapter, pages); err != nil {
		return err
	}

	err := r.createOrPatch(ctx, pages, adapter)

	return err
}

func (r *GitLabReconciler) reconcilePagesService(ctx context.Context, adapter gitlab.Adapter, template helm.Template) error {
	if err := r.createOrPatch(ctx, gitlabctl.PagesService(template), adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcilePagesIngress(ctx context.Context, adapter gitlab.Adapter, template helm.Template) error {
	if err := r.reconcileIngress(ctx, gitlabctl.PagesIngress(template), adapter); err != nil {
		return err
	}

	return nil
}
