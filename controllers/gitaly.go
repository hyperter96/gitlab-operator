package controllers

import (
	"context"

	gitlabctl "gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab"
)

func (r *GitLabReconciler) reconcileGitaly(ctx context.Context, adapter gitlab.Adapter, template helm.Template) error {
	if err := r.reconcileGitalyConfigMap(ctx, adapter, template); err != nil {
		return err
	}

	if err := r.reconcileGitalyStatefulSet(ctx, adapter, template); err != nil {
		return err
	}

	if err := r.reconcileGitalyService(ctx, adapter, template); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileGitalyConfigMap(ctx context.Context, adapter gitlab.Adapter, template helm.Template) error {
	if err := r.createOrPatch(ctx, gitlabctl.GitalyConfigMap(template), adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileGitalyStatefulSet(ctx context.Context, adapter gitlab.Adapter, template helm.Template) error {
	if err := r.createOrPatch(ctx, gitlabctl.GitalyStatefulSet(template), adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileGitalyService(ctx context.Context, adapter gitlab.Adapter, template helm.Template) error {
	if err := r.createOrPatch(ctx, gitlabctl.GitalyService(template), adapter); err != nil {
		return err
	}

	return nil
}
