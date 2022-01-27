package controllers

import (
	"context"

	gitlabctl "gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
)

func (r *GitLabReconciler) reconcileKas(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	if err := r.reconcileKasConfigMap(ctx, adapter); err != nil {
		return err
	}

	if err := r.reconcileKasDeployment(ctx, adapter); err != nil {
		return err
	}

	if err := r.reconcileKasService(ctx, adapter); err != nil {
		return err
	}

	if err := r.reconcileKasIngress(ctx, adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileKasConfigMap(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	if _, err := r.createOrPatch(ctx, gitlabctl.KasConfigMap(adapter), adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileKasDeployment(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	kas := gitlabctl.KasDeployment(adapter)

	if err := r.setDeploymentReplica(ctx, kas); err != nil {
		return err
	}

	if err := r.annotateSecretsChecksum(ctx, adapter, kas); err != nil {
		return err
	}

	_, err := r.createOrPatch(ctx, kas, adapter)

	return err
}

func (r *GitLabReconciler) reconcileKasService(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	if _, err := r.createOrPatch(ctx, gitlabctl.KasService(adapter), adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileKasIngress(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	if _, err := r.createOrPatch(ctx, gitlabctl.KasIngress(adapter), adapter); err != nil {
		return err
	}

	return nil
}
