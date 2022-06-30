package controllers

import (
	"context"

	gitlabctl "gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
)

func (r *GitLabReconciler) reconcilePostgres(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) error {
	if err := r.reconcilePostgresConfigMap(ctx, adapter, template); err != nil {
		return err
	}

	if err := r.reconcilePostgresStatefulSet(ctx, adapter, template); err != nil {
		return err
	}

	if err := r.reconcilePostgresServices(ctx, adapter, template); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcilePostgresConfigMap(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) error {
	if err := r.createOrPatch(ctx, gitlabctl.PostgresConfigMap(adapter, template), adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) reconcilePostgresStatefulSet(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) error {
	ss := gitlabctl.PostgresStatefulSet(adapter, template)

	if err := r.annotateSecretsChecksum(ctx, adapter, ss); err != nil {
		return err
	}

	if err := r.createOrPatch(ctx, ss, adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) validateExternalPostgresConfiguration(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	// Ensure that the PostgreSQL password Secret was created.
	pgSecretName := adapter.Values().GetString("global.psql.password.secret")
	if err := r.ensureSecret(ctx, adapter, pgSecretName); err != nil {
		return err
	}

	// If set, ensure that the PostgreSQL SSL Secret was created.
	pgSecretNameSSL := adapter.Values().GetString("global.psql.ssl.secret", "unset")
	if pgSecretNameSSL != "unset" {
		if err := r.ensureSecret(ctx, adapter, pgSecretNameSSL); err != nil {
			return err
		}
	}

	return nil
}

func (r *GitLabReconciler) reconcilePostgresServices(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) error {
	for _, svc := range gitlabctl.PostgresServices(adapter, template) {
		if err := r.createOrPatch(ctx, svc, adapter); err != nil {
			return err
		}
	}

	return nil
}
