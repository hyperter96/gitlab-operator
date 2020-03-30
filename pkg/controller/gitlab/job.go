package gitlab

import (
	"context"

	gitlabv1beta1 "gitlab.com/ochienged/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlabutils "gitlab.com/ochienged/gitlab-operator/pkg/controller/utils"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func getMigrationsJob(cr *gitlabv1beta1.Gitlab) *batchv1.Job {
	labels := gitlabutils.Label(cr.Name, "migrations", gitlabutils.GitlabType)

	migration := gitlabutils.GenericJob(gitlabutils.Component{
		Namespace: cr.Namespace,
		Labels:    labels,
		InitContainers: []corev1.Container{
			{
				Name:            "certificates",
				Image:           gitlabutils.GitLabCertificatesImage,
				ImagePullPolicy: corev1.PullIfNotPresent,
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"cpu": gitlabutils.ResourceQuantity("50m"),
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "etc-ssl-certs",
						MountPath: "/etc/ssl/certs",
					},
				},
			},
			{
				Name:            "configure",
				Image:           gitlabutils.BusyboxImage,
				ImagePullPolicy: corev1.PullAlways,
				Command:         []string{"sh", "/config/configure"},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"cpu": gitlabutils.ResourceQuantity("50m"),
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "migrations-config",
						MountPath: "/config",
						ReadOnly:  true,
					},
					{
						Name:      "init-migrations-secrets",
						MountPath: "/init-config",
						ReadOnly:  true,
					},
					{
						Name:      "migrations-secrets",
						MountPath: "/init-secrets",
					},
				},
			},
		},
		Containers: []corev1.Container{
			{
				Name:            "migrations",
				Image:           gitlabutils.GitLabTaskRunnerImage,
				ImagePullPolicy: corev1.PullIfNotPresent,
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"cpu":    gitlabutils.ResourceQuantity("250m"),
						"memory": gitlabutils.ResourceQuantity("200Mi"),
					},
				},
				Args: []string{"/scripts/wait-for-deps", "/scripts/db-migrate"},
				Env: []corev1.EnvVar{
					{
						Name: "GITLAB_SHARED_RUNNERS_REGISTRATION_TOKEN",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: cr.Name + "-gitlab-secrets",
								},
								Key: "initial_shared_runners_registration_token",
							},
						},
					},
					{
						Name:  "CONFIG_TEMPLATE_DIRECTORY",
						Value: "/var/opt/gitlab/templates",
					},
					{
						Name:  "CONFIG_DIRECTORY",
						Value: "/srv/gitlab/config",
					},
					{
						Name:  "BYPASS_SCHEMA_VERSION",
						Value: "true",
					},
					{
						Name:  "ENABLE_BOOTSNAP",
						Value: "1",
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "migrations-config",
						MountPath: "/var/opt/gitlab/templates",
					},
					{
						Name:      "migrations-secrets",
						MountPath: "/etc/gitlab",
						ReadOnly:  true,
					},
					{
						Name:      "migrations-secrets",
						MountPath: "/srv/gitlab/config/secrets.yml",
						SubPath:   "rails-secrets/secrets.yml",
					},
					{
						Name:      "migrations-secrets",
						MountPath: "/srv/gitlab/config/initial_root_password",
						SubPath:   "migrations/initial_root_password",
					},
					{
						Name:      "etc-ssl-certs",
						MountPath: "/etc/ssl/certs/",
						ReadOnly:  true,
					},
				},
			},
		},
		Volumes: []corev1.Volume{
			{
				Name: "migrations-config",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						DefaultMode: &gitlabutils.ConfigMapDefaultMode,
						LocalObjectReference: corev1.LocalObjectReference{
							Name: cr.Name + "-migrations-config",
						},
					},
				},
			},
			{
				Name: "init-migrations-secrets",
				VolumeSource: corev1.VolumeSource{
					Projected: &corev1.ProjectedVolumeSource{
						DefaultMode: &gitlabutils.ProjectedVolumeDefaultMode,
						Sources: []corev1.VolumeProjection{
							{
								Secret: &corev1.SecretProjection{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: cr.Name + "-rails-secret",
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "secrets.yml",
											Path: "rails-secrets/secrets.yml",
										},
									},
								},
							},
							{
								Secret: &corev1.SecretProjection{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: cr.Name + "-gitaly-secret",
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "token",
											Path: "gitaly/gitaly_token",
										},
									},
								},
							},
							{
								Secret: &corev1.SecretProjection{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: cr.Name + "-redis-secret",
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "secret",
											Path: "redis/password",
										},
									},
								},
							},
							{
								Secret: &corev1.SecretProjection{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: cr.Name + "-postgresql-secret",
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "postgresql-password",
											Path: "postgres/psql-password",
										},
									},
								},
							},
							{
								Secret: &corev1.SecretProjection{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: cr.Name + "-gitlab-secrets",
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "gitlab_root_password",
											Path: "migrations/initial_root_password",
										},
									},
								},
							},
						},
					},
				},
			},
			{
				Name: "migrations-secrets",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{
						Medium: corev1.StorageMediumMemory,
					},
				},
			},
			{
				Name: "etc-ssl-certs",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{
						Medium: corev1.StorageMediumMemory,
					},
				},
			},
		},
	})

	migration.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{
		RunAsUser: &runAsUser,
		FSGroup:   &fsGroup,
	}

	return migration
}

func (r *ReconcileGitlab) reconcileMigrationsJob(cr *gitlabv1beta1.Gitlab) error {
	migration := getMigrationsJob(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: migration.Name}, migration) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, migration, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), migration)
}
