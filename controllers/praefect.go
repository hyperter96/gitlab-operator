package controllers

import (
	"context"

	gitlabctl "gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab"
)

func (r *GitLabReconciler) reconcilePraefect(ctx context.Context, adapter gitlab.Adapter, template helm.Template) error {
	if err := r.reconcilePraefectConfigMap(ctx, adapter, template); err != nil {
		return err
	}

	if err := r.reconcilePraefectStatefulSet(ctx, adapter, template); err != nil {
		return err
	}

	if err := r.reconcilePraefectService(ctx, adapter, template); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcilePraefectConfigMap(ctx context.Context, adapter gitlab.Adapter, template helm.Template) error {
	if err := r.createOrPatch(ctx, gitlabctl.PraefectConfigMap(template), adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcilePraefectStatefulSet(ctx context.Context, adapter gitlab.Adapter, template helm.Template) error {
	if err := r.createOrPatch(ctx, gitlabctl.PraefectStatefulSet(template), adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcilePraefectService(ctx context.Context, adapter gitlab.Adapter, template helm.Template) error {
	if err := r.createOrPatch(ctx, gitlabctl.PraefectService(template), adapter); err != nil {
		return err
	}

	return nil
}
