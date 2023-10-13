package controllers

import (
	"context"

	gitlabctl "gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab"
)

func (r *GitLabReconciler) reconcileZoekt(ctx context.Context, adapter gitlab.Adapter, template helm.Template) error {
	if err := r.reconcileZoektConfigMap(ctx, adapter, template); err != nil {
		return err
	}

	if err := r.reconcileZoektCertificate(ctx, adapter, template); err != nil {
		return err
	}

	if err := r.reconcileZoektStatefulSet(ctx, adapter, template); err != nil {
		return err
	}

	if err := r.reconcileZoektService(ctx, adapter, template); err != nil {
		return err
	}

	if err := r.reconcileZoektIngress(ctx, adapter, template); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileZoektConfigMap(ctx context.Context, adapter gitlab.Adapter, template helm.Template) error {
	return r.createOrPatch(ctx, gitlabctl.ZoektConfigMap(template, adapter), adapter)
}

func (r *GitLabReconciler) reconcileZoektStatefulSet(ctx context.Context, adapter gitlab.Adapter, template helm.Template) error {
	return r.createOrPatch(ctx, gitlabctl.ZoektStatefulSet(template, adapter), adapter)
}

func (r *GitLabReconciler) reconcileZoektService(ctx context.Context, adapter gitlab.Adapter, template helm.Template) error {
	return r.createOrPatch(ctx, gitlabctl.ZoektService(template, adapter), adapter)
}

func (r *GitLabReconciler) reconcileZoektIngress(ctx context.Context, adapter gitlab.Adapter, template helm.Template) error {
	if ing := gitlabctl.ZoektIngress(template, adapter); ing != nil {
		return r.createOrPatch(ctx, gitlabctl.ZoektIngress(template, adapter), adapter)
	}

	return nil
}

func (r *GitLabReconciler) reconcileZoektCertificate(ctx context.Context, adapter gitlab.Adapter, template helm.Template) error {
	if cert := gitlabctl.ZoektCertificate(template, adapter); cert != nil {
		return r.createOrPatch(ctx, cert, adapter)
	}

	return nil
}
