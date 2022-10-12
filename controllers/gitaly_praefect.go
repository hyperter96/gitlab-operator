package controllers

import (
	"context"

	gitlabctl "gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab"
)

func (r *GitLabReconciler) reconcileGitalyPraefect(ctx context.Context, adapter gitlab.Adapter, template helm.Template) error {
	if err := r.reconcileGitalyPraefectConfigMap(ctx, adapter, template); err != nil {
		return err
	}

	if err := r.reconcileGitalyPraefectServices(ctx, adapter, template); err != nil {
		return err
	}

	if err := r.reconcileGitalyPraefectStatefulSets(ctx, adapter, template); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileGitalyPraefectConfigMap(ctx context.Context, adapter gitlab.Adapter, template helm.Template) error {
	if err := r.createOrPatch(ctx, gitlabctl.GitalyPraefectConfigMap(template), adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileGitalyPraefectServices(ctx context.Context, adapter gitlab.Adapter, template helm.Template) error {
	gitalyPraefectServices := gitlabctl.GitalyPraefectServices(template)

	for _, gitalyPraefectService := range gitalyPraefectServices {
		if err := r.createOrPatch(ctx, gitalyPraefectService, adapter); err != nil {
			return err
		}
	}

	return nil
}

func (r *GitLabReconciler) reconcileGitalyPraefectStatefulSets(ctx context.Context, adapter gitlab.Adapter, template helm.Template) error {
	gitalyPraefectStatefulSets := gitlabctl.GitalyPraefectStatefulSets(template)

	for _, gitalyPraefectStatefulSet := range gitalyPraefectStatefulSets {
		if err := r.createOrPatch(ctx, gitalyPraefectStatefulSet, adapter); err != nil {
			return err
		}
	}

	return nil
}
