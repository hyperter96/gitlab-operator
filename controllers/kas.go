package controllers

import (
	"context"

	gitlabctl "gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
)

func (r *GitLabReconciler) reconcileKas(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) error {
	if err := r.reconcileKasConfigMap(ctx, adapter, template); err != nil {
		return err
	}

	if err := r.reconcileKasDeployment(ctx, adapter, template); err != nil {
		return err
	}

	if err := r.reconcileKasService(ctx, adapter, template); err != nil {
		return err
	}

	if err := r.reconcileKasIngress(ctx, adapter, template); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileKasConfigMap(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) error {
	if _, err := r.createOrPatch(ctx, gitlabctl.KasConfigMap(template), adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileKasDeployment(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) error {
	kas := gitlabctl.KasDeployment(template)

	if err := r.setDeploymentReplica(ctx, kas); err != nil {
		return err
	}

	if err := r.annotateSecretsChecksum(ctx, adapter, kas); err != nil {
		return err
	}

	_, err := r.createOrPatch(ctx, kas, adapter)

	return err
}

func (r *GitLabReconciler) reconcileKasService(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) error {
	if _, err := r.createOrPatch(ctx, gitlabctl.KasService(template), adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileKasIngress(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) error {
	if _, err := r.createOrPatch(ctx, gitlabctl.KasIngress(template), adapter); err != nil {
		return err
	}

	return nil
}
