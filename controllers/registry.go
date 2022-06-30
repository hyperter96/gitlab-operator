package controllers

import (
	"context"

	gitlabctl "gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
)

func (r *GitLabReconciler) reconcileRegistry(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) error {
	if err := r.reconcileRegistryConfigMap(ctx, adapter, template); err != nil {
		return err
	}

	if err := r.reconcileRegistryDeployment(ctx, adapter, template); err != nil {
		return err
	}

	if err := r.reconcileRegistryService(ctx, adapter, template); err != nil {
		return err
	}

	if err := r.reconcileRegistryIngress(ctx, adapter, template); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileRegistryConfigMap(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) error {
	if err := r.createOrPatch(ctx, gitlabctl.RegistryConfigMap(adapter, template), adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileRegistryDeployment(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) error {
	registry := gitlabctl.RegistryDeployment(template)

	if err := r.setDeploymentReplica(ctx, registry); err != nil {
		return err
	}

	if err := r.annotateSecretsChecksum(ctx, adapter, registry); err != nil {
		return err
	}

	err := r.createOrPatch(ctx, registry, adapter)

	return err
}

func (r *GitLabReconciler) reconcileRegistryService(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) error {
	if err := r.createOrPatch(ctx, gitlabctl.RegistryService(template), adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileRegistryIngress(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) error {
	if err := r.reconcileIngress(ctx, gitlabctl.RegistryIngress(template), adapter); err != nil {
		return err
	}

	return nil
}
