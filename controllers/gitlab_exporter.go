package controllers

import (
	"context"

	gitlabctl "gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab"
)

func (r *GitLabReconciler) reconcileGitLabExporter(ctx context.Context, adapter gitlab.Adapter, template helm.Template) error {
	if err := r.reconcileGitLabExporterConfigMaps(ctx, adapter, template); err != nil {
		return err
	}

	if err := r.reconcileGitLabExporterDeployment(ctx, adapter, template); err != nil {
		return err
	}

	if err := r.reconcileGitLabExporterService(ctx, adapter, template); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcileGitLabExporterConfigMaps(ctx context.Context, adapter gitlab.Adapter, template helm.Template) error {
	for _, cm := range gitlabctl.ExporterConfigMaps(adapter, template) {
		if err := r.createOrPatch(ctx, cm, adapter); err != nil {
			return err
		}
	}

	return nil
}

func (r *GitLabReconciler) reconcileGitLabExporterDeployment(ctx context.Context, adapter gitlab.Adapter, template helm.Template) error {
	exporter := gitlabctl.ExporterDeployment(template)

	if err := r.annotateSecretsChecksum(ctx, adapter, exporter); err != nil {
		return err
	}

	err := r.createOrPatch(ctx, exporter, adapter)

	return err
}

func (r *GitLabReconciler) reconcileGitLabExporterService(ctx context.Context, adapter gitlab.Adapter, template helm.Template) error {
	if err := r.createOrPatch(ctx, gitlabctl.ExporterService(template), adapter); err != nil {
		return err
	}

	return nil
}
