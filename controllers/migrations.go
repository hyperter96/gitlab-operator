package controllers

import (
	"context"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	gitlabctl "gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab"
)

func (r *GitLabReconciler) reconcileMigrationsConfigMap(ctx context.Context, adapter gitlab.Adapter, template helm.Template) error {
	if err := r.createOrPatch(ctx, gitlabctl.MigrationsConfigMap(adapter, template), adapter); err != nil {
		return err
	}

	return nil
}

func (r *GitLabReconciler) runMigrationsJob(ctx context.Context, adapter gitlab.Adapter, template helm.Template, skipPostMigrations bool) (bool, error) {
	migrations, err := gitlabctl.MigrationsJob(adapter, template)
	if err != nil {
		return false, err
	}

	// Attention: Type Assertion: Job.Spec is needed
	job, ok := migrations.DeepCopyObject().(*batchv1.Job)
	if !ok {
		return false, helm.NewTypeMistmatchError(job, migrations)
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

	if err := r.createOrPatch(ctx, job, adapter); err != nil {
		return false, err
	}

	return r.jobFinished(ctx, adapter, job)
}

func (r *GitLabReconciler) runPreMigrations(ctx context.Context, adapter gitlab.Adapter, template helm.Template) (bool, error) {
	return r.runMigrationsJob(ctx, adapter, template, true)
}

func (r *GitLabReconciler) runAllMigrations(ctx context.Context, adapter gitlab.Adapter, template helm.Template) (bool, error) {
	return r.runMigrationsJob(ctx, adapter, template, false)
}
