package controllers

import (
	"context"

	gitlabctl "gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
)

func (r *GitLabReconciler) reconcileGitLabShell(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	if err := r.reconcileShellConfigMaps(ctx, adapter); err != nil {
		return err
	}

	if err := r.reconcileShellDeployment(ctx, adapter); err != nil {
		return err
	}

	if err := r.reconcileShellService(ctx, adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileShellDeployment(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	shell := gitlabctl.ShellDeployment(adapter)

	if err := r.setDeploymentReplica(ctx, shell); err != nil {
		return err
	}

	if err := r.annotateSecretsChecksum(ctx, adapter, shell); err != nil {
		return err
	}

	_, err := r.createOrPatch(ctx, shell, adapter)

	return err
}

func (r *GitLabReconciler) reconcileShellConfigMaps(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	for _, cm := range gitlabctl.ShellConfigMaps(adapter) {
		if _, err := r.createOrPatch(ctx, cm, adapter); err != nil {
			return err
		}
	}

	return nil
}

func (r *GitLabReconciler) reconcileShellService(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	if _, err := r.createOrPatch(ctx, gitlabctl.ShellService(adapter), adapter); err != nil {
		return err
	}

	return nil
}
