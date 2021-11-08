package controllers

import (
	"context"

	gitlabctl "gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
)

func (r *GitLabReconciler) reconcileRegistry(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	if err := r.reconcileRegistryConfigMap(ctx, adapter); err != nil {
		return err
	}

	if err := r.reconcileRegistryDeployment(ctx, adapter); err != nil {
		return err
	}

	if err := r.reconcileRegistryService(ctx, adapter); err != nil {
		return err
	}

	if err := r.reconcileRegistryIngress(ctx, adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileRegistryConfigMap(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	if _, err := r.createOrPatch(ctx, gitlabctl.RegistryConfigMap(adapter), adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileRegistryDeployment(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	registry := gitlabctl.RegistryDeployment(adapter)

	if err := r.setDeploymentReplica(ctx, registry); err != nil {
		return err
	}

	if err := r.annotateSecretsChecksum(ctx, adapter, &registry.Spec.Template); err != nil {
		return err
	}

	_, err := r.createOrPatch(ctx, registry, adapter)

	return err
}

func (r *GitLabReconciler) reconcileRegistryService(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	if _, err := r.createOrPatch(ctx, gitlabctl.RegistryService(adapter), adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileRegistryIngress(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	if err := r.reconcileIngress(ctx, gitlabctl.RegistryIngress(adapter), adapter); err != nil {
		return err
	}

	return nil
}
