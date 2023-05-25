package gitlab

import (
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support"
)

// MigrationsConfigMap returns the ConfigMaps of Migrations component.
func MigrationsConfigMap(adapter gitlab.Adapter, template helm.Template) client.Object {
	return template.Query().ObjectByKindAndName(ConfigMapKind,
		fmt.Sprintf("%s-%s", adapter.ReleaseName(), MigrationsComponentName))
}

// MigrationsJob returns the Job for Migrations component.
func MigrationsJob(adapter gitlab.Adapter, template helm.Template) (*batchv1.Job, error) {
	migrations := template.Query().ObjectByKindAndComponent(JobKind, MigrationsComponentName)
	job, ok := migrations.DeepCopyObject().(*batchv1.Job)

	if !ok {
		return nil, helm.NewTypeMistmatchError(job, migrations)
	}

	nameWithSuffix, err := support.NameWithHashSuffix(job.GetName(), adapter.Hash(), 5)
	if err != nil {
		return job, err
	}

	job.SetName(nameWithSuffix)

	return job, nil
}

// PreMigrationsJob returns for the Migrations component, running pre migrations only.
func PreMigrationsJob(adapter gitlab.Adapter, template helm.Template) (*batchv1.Job, error) {
	job, err := MigrationsJob(adapter, template)
	if err != nil {
		return nil, err
	}

	// - Append "-pre" to the Migrations Job name
	job.SetName(fmt.Sprintf("%s-pre", job.GetName()))

	// - Inject environment variable to skip post-deployment migrations.
	for i := range job.Spec.Template.Spec.Containers {
		job.Spec.Template.Spec.Containers[i].Env = append(
			job.Spec.Template.Spec.Containers[i].Env,
			corev1.EnvVar{Name: "SKIP_POST_DEPLOYMENT_MIGRATIONS", Value: "true"})
	}

	return job, nil
}
