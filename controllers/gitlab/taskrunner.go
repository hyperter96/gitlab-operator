package gitlab

import (
	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	gitlabutils "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// TaskRunnerDeployment returns deployment for TaskRunner
func TaskRunnerDeployment(cr *gitlabv1beta1.GitLab) *appsv1.Deployment {
	labels := gitlabutils.Label(cr.Name, "task-runner", gitlabutils.GitlabType)
	options := SystemBuildOptions(cr)

	taskRunner := gitlabutils.GenericDeployment(gitlabutils.Component{
		Namespace: cr.Namespace,
		Labels:    labels,
		Replicas:  1,
		InitContainers: []corev1.Container{
			{
				Name:            "certificates",
				Image:           BuildRelease(cr).Certificates(),
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
				Image:           BuildRelease(cr).Busybox(),
				ImagePullPolicy: corev1.PullAlways,
				Command:         []string{"sh", "/config/configure"},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"cpu": gitlabutils.ResourceQuantity("50m"),
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "task-runner-config",
						MountPath: "/config",
						ReadOnly:  true,
					},
					{
						Name:      "init-task-runner-secrets",
						MountPath: "/init-config",
						ReadOnly:  true,
					},
					{
						Name:      "task-runner-secrets",
						MountPath: "/init-secrets",
					},
				},
			},
		},
		Containers: []corev1.Container{
			{
				Name:            "task-runner",
				Image:           BuildRelease(cr).TaskRunner(),
				ImagePullPolicy: corev1.PullIfNotPresent,
				Args: []string{
					"/bin/bash",
					"-c",
					"cp -v -r -L /etc/gitlab/.s3cfg $HOME/.s3cfg && while sleep 3600; do :; done",
				},
				Env: []corev1.EnvVar{
					{
						Name:  "ARTIFACTS_BUCKET_NAME",
						Value: "gitlab-artifacts",
					},
					{
						Name:  "REGISTRY_BUCKET_NAME",
						Value: "registry",
					},
					{
						Name:  "LFS_BUCKET_NAME",
						Value: "git-lfs",
					},
					{
						Name:  "UPLOADS_BUCKET_NAME",
						Value: "gitlab-uploads",
					},
					{
						Name:  "PACKAGES_BUCKET_NAME",
						Value: "gitlab-packages",
					},
					{
						Name:  "BACKUP_BUCKET_NAME",
						Value: "gitlab-backups",
					},
					{
						Name:  "BACKUP_BACKEND",
						Value: "s3",
					},
					{
						Name:  "TMP_BUCKET_NAME",
						Value: "tmp",
					},
					{
						Name:  "GITALY_FEATURE_DEFAULT_ON",
						Value: "1",
					},
					{
						Name:  "ENABLE_BOOTSNAP",
						Value: "1",
					},
					{
						Name:  "CONFIG_TEMPLATE_DIRECTORY",
						Value: "/var/opt/gitlab/templates",
					},
					{
						Name:  "CONFIG_DIRECTORY",
						Value: "/srv/gitlab/config",
					},
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"cpu":    gitlabutils.ResourceQuantity("50m"),
						"memory": gitlabutils.ResourceQuantity("350M"),
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "task-runner-config",
						MountPath: "/var/opt/gitlab/templates",
					},
					{
						Name:      "task-runner-config",
						MountPath: "/srv/gitlab/config/initializers/smtp_settings.rb",
						SubPath:   "smtp_settings.rb",
					},
					{
						Name:      "task-runner-secrets",
						MountPath: "/etc/gitlab",
						ReadOnly:  true,
					},
					{
						Name:      "task-runner-secrets",
						MountPath: "/srv/gitlab/config/secrets.yml",
						SubPath:   "rails-secrets/secrets.yml",
					},
					{
						Name:      "task-runner-tmp",
						MountPath: "/srv/gitlab/tmp",
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
				Name: "task-runner-config",
				VolumeSource: corev1.VolumeSource{
					Projected: &corev1.ProjectedVolumeSource{
						DefaultMode: &gitlabutils.ConfigMapDefaultMode,
						Sources: []corev1.VolumeProjection{
							{
								ConfigMap: &corev1.ConfigMapProjection{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: cr.Name + "-task-runner-config",
									},
								},
							},
							{
								Secret: &corev1.SecretProjection{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: cr.Name + "-smtp-settings-secret",
									},
								},
							},
						},
					},
				},
			},
			{
				Name: "task-runner-tmp",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
			{
				Name: "init-task-runner-secrets",
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
										Name: cr.Name + "-shell-secret",
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "secret",
											Path: "shell/.gitlab_shell_secret",
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
											Path: "redis/redis-password",
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
										Name: cr.Name + "-registry-secret",
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "registry-auth.key",
											Path: "registry/gitlab-registry.key",
										},
									},
								},
							},
							{
								Secret: &corev1.SecretProjection{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: options.ObjectStore.Credentials,
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "accesskey",
											Path: "minio/accesskey",
										},
										{
											Key:  "secretkey",
											Path: "minio/secretkey",
										},
									},
								},
							},
						},
					},
				},
			},
			{
				Name: "task-runner-secrets",
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

	taskRunner.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{
		RunAsUser: &localUser,
		FSGroup:   &localUser,
	}

	taskRunner.Spec.Template.Spec.ServiceAccountName = "gitlab"

	return taskRunner
}
