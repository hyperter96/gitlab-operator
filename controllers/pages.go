package controllers

import (
	"context"

	gitlabctl "gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
)

func (r *GitLabReconciler) reconcilePages(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	if err := r.reconcilePagesConfigMap(ctx, adapter); err != nil {
		return err
	}

	if err := r.reconcilePagesDeployment(ctx, adapter); err != nil {
		return err
	}

	if err := r.reconcilePagesService(ctx, adapter); err != nil {
		return err
	}

	if err := r.reconcilePagesIngress(ctx, adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcilePagesConfigMap(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	if _, err := r.createOrPatch(ctx, gitlabctl.PagesConfigMap(adapter), adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcilePagesDeployment(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	pages := gitlabctl.PagesDeployment(adapter)

	if err := r.setDeploymentReplica(ctx, pages); err != nil {
		return err
	}

	if err := r.annotateSecretsChecksum(ctx, adapter, pages); err != nil {
		return err
	}

	_, err := r.createOrPatch(ctx, pages, adapter)

	return err
}

func (r *GitLabReconciler) reconcilePagesService(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	if _, err := r.createOrPatch(ctx, gitlabctl.PagesService(adapter), adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcilePagesIngress(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	if _, err := r.createOrPatch(ctx, gitlabctl.PagesIngress(adapter), adapter); err != nil {
		return err
	}

	return nil
}
