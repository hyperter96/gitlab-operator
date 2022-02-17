package controllers

import (
	"context"

	gitlabctl "gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
)

func (r *GitLabReconciler) reconcileToolbox(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) error {
	if err := r.reconcileToolboxConfigMap(ctx, adapter, template); err != nil {
		return err
	}

	if gitlabctl.ToolboxCronJobEnabled(adapter) {
		if err := r.reconcileToolboxCronJob(ctx, adapter, template); err != nil {
			return err
		}
	}

	if gitlabctl.ToolboxCronJobPersistenceEnabled(adapter) {
		if err := r.reconcileToolboxPersistentVolumeClaim(ctx, adapter, template); err != nil {
			return err
		}
	}

	if err := r.reconcileToolboxDeployment(ctx, adapter, template); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileToolboxConfigMap(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) error {
	if _, err := r.createOrPatch(ctx, gitlabctl.ToolboxConfigMap(adapter, template), adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileToolboxCronJob(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) error {
	if _, err := r.createOrPatch(ctx, gitlabctl.ToolboxCronJob(adapter, template), adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileToolboxDeployment(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) error {
	deployment := gitlabctl.ToolboxDeployment(adapter, template)

	if err := r.annotateSecretsChecksum(ctx, adapter, deployment); err != nil {
		return err
	}

	_, err := r.createOrPatch(ctx, deployment, adapter)

	return err
}

func (r *GitLabReconciler) reconcileToolboxPersistentVolumeClaim(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) error {
	if _, err := r.createOrPatch(ctx, gitlabctl.ToolboxCronJobPersistentVolumeClaim(adapter, template), adapter); err != nil {
		return err
	}

	return nil
}
