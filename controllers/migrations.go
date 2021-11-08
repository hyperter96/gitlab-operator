package controllers

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"

	gitlabctl "gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
)

func (r *GitLabReconciler) reconcileMigrationsConfigMap(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	if _, err := r.createOrPatch(ctx, gitlabctl.MigrationsConfigMap(adapter), adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) runMigrationsJob(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, skipPostMigrations bool) error {
	migrations, err := gitlabctl.MigrationsJob(adapter)
	if err != nil {
		return err
	}

	job := migrations.DeepCopy()

	// If `skipPostMigrations=true`, then:
	// - Append "-pre" to the Migrations Job name
	// - Inject environment variable to skip post-deployment migrations.
	if skipPostMigrations {
		job.Name = fmt.Sprintf("%s-pre", migrations.Name)
		for i := range job.Spec.Template.Spec.Containers {
			job.Spec.Template.Spec.Containers[i].Env = append(
				job.Spec.Template.Spec.Containers[i].Env,
				corev1.EnvVar{Name: "SKIP_POST_DEPLOYMENT_MIGRATIONS", Value: "true"})
		}
	}

	return r.runJobAndWait(ctx, adapter, job, gitlabctl.MigrationsJobTimeout())
}

func (r *GitLabReconciler) runPreMigrations(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	return r.runMigrationsJob(ctx, adapter, true)
}

func (r *GitLabReconciler) runAllMigrations(ctx context.Context, adapter gitlabctl.CustomResourceAdapter) error {
	return r.runMigrationsJob(ctx, adapter, false)
}
