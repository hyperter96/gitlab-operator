package controllers

import (
	"context"

	gitlabctl "gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
)

func (r *GitLabReconciler) reconcileGitLabExporter(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	if err := r.reconcileGitLabExporterConfigMaps(ctx, adapter); err != nil {
		return err
	}

	if err := r.reconcileGitLabExporterDeployment(ctx, adapter); err != nil {
		return err
	}

	if err := r.reconcileGitLabExporterService(ctx, adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileGitLabExporterConfigMaps(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	for _, cm := range gitlabctl.ExporterConfigMaps(adapter) {
		if _, err := r.createOrPatch(ctx, cm, adapter); err != nil {
			return err
		}
	}

	return nil
}

func (r *GitLabReconciler) reconcileGitLabExporterDeployment(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	exporter := gitlabctl.ExporterDeployment(adapter)

	if err := r.annotateSecretsChecksum(ctx, adapter, exporter); err != nil {
		return err
	}

	_, err := r.createOrPatch(ctx, exporter, adapter)

	return err
}

func (r *GitLabReconciler) reconcileGitLabExporterService(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	if _, err := r.createOrPatch(ctx, gitlabctl.ExporterService(adapter), adapter); err != nil {
		return err
	}

	return nil
}
