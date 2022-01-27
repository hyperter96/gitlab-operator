package controllers

import (
	"context"

	gitlabctl "gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
)

func (r *GitLabReconciler) reconcileToolbox(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	if err := r.reconcileToolboxConfigMap(ctx, adapter); err != nil {
		return err
	}

	if gitlabctl.ToolboxCronJobEnabled(adapter) {
		if err := r.reconcileToolboxCronJob(ctx, adapter); err != nil {
			return err
		}
	}

	if gitlabctl.ToolboxCronJobPersistenceEnabled(adapter) {
		if err := r.reconcileToolboxPersistentVolumeClaim(ctx, adapter); err != nil {
			return err
		}
	}

	if err := r.reconcileToolboxDeployment(ctx, adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileToolboxConfigMap(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	if _, err := r.createOrPatch(ctx, gitlabctl.ToolboxConfigMap(adapter), adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileToolboxCronJob(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	if _, err := r.createOrPatch(ctx, gitlabctl.ToolboxCronJob(adapter), adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileToolboxDeployment(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	deployment := gitlabctl.ToolboxDeployment(adapter)

	if err := r.annotateSecretsChecksum(ctx, adapter, deployment); err != nil {
		return err
	}

	_, err := r.createOrPatch(ctx, deployment, adapter)

	return err
}

func (r *GitLabReconciler) reconcileToolboxPersistentVolumeClaim(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	if _, err := r.createOrPatch(ctx, gitlabctl.ToolboxCronJobPersistentVolumeClaim(adapter), adapter); err != nil {
		return err
	}

	return nil
}
