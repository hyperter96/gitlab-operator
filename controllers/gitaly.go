package controllers

import (
	"context"

	gitlabctl "gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
)

func (r *GitLabReconciler) reconcileGitaly(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	if err := r.reconcileGitalyConfigMap(ctx, adapter); err != nil {
		return err
	}

	if err := r.reconcileGitalyStatefulSet(ctx, adapter); err != nil {
		return err
	}

	if err := r.reconcileGitalyService(ctx, adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileGitalyConfigMap(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	if _, err := r.createOrPatch(ctx, gitlabctl.GitalyConfigMap(adapter), adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileGitalyStatefulSet(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	if _, err := r.createOrPatch(ctx, gitlabctl.GitalyStatefulSet(adapter), adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileGitalyService(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	if _, err := r.createOrPatch(ctx, gitlabctl.GitalyService(adapter), adapter); err != nil {
		return err
	}

	return nil
}
