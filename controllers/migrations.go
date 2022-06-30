package controllers

import (
	"context"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	gitlabctl "gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
)

func (r *GitLabReconciler) reconcileMigrationsConfigMap(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) error {
	if err := r.createOrPatch(ctx, gitlabctl.MigrationsConfigMap(adapter, template), adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) runMigrationsJob(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template, skipPostMigrations bool) error {
	migrations, err := gitlabctl.MigrationsJob(adapter, template)
	if err != nil {
		return err
	}

	// Attention: Type Assertion: Job.Spec is needed
	job, ok := migrations.DeepCopyObject().(*batchv1.Job)
	if !ok {
		return helm.NewTypeMistmatchError(job, migrations)
	}

	// If `skipPostMigrations=true`, then:
	// - Append "-pre" to the Migrations Job name
	// - Inject environment variable to skip post-deployment migrations.
	if skipPostMigrations {
		job.SetName(fmt.Sprintf("%s-pre", migrations.GetName()))

		for i := range job.Spec.Template.Spec.Containers {
			job.Spec.Template.Spec.Containers[i].Env = append(
				job.Spec.Template.Spec.Containers[i].Env,
				corev1.EnvVar{Name: "SKIP_POST_DEPLOYMENT_MIGRATIONS", Value: "true"})
		}
	}

	return r.runJobAndWait(ctx, adapter, job, gitlabctl.MigrationsJobTimeout())
}

func (r *GitLabReconciler) runPreMigrations(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) error {
	return r.runMigrationsJob(ctx, adapter, template, true)
}

func (r *GitLabReconciler) runAllMigrations(ctx context.Context, adapter gitlabctl.CustomResourceAdapter, template helm.Template) error {
	return r.runMigrationsJob(ctx, adapter, template, false)
}
