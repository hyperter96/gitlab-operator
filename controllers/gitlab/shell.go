package gitlab

import (
	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	gitlabutils "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

var localUser int64 = 1000

// ShellDeploymentDEPRECATED returns GitLab shell deployment
func ShellDeploymentDEPRECATED(cr *gitlabv1beta1.GitLab) *appsv1.Deployment {
	labels := gitlabutils.Label(cr.Name, "gitlab-shell", gitlabutils.GitlabType)

	shell := gitlabutils.GenericDeployment(gitlabutils.Component{
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
				ImagePullPolicy: corev1.PullIfNotPresent,
				Command:         []string{"sh", "/config/configure"},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"cpu": gitlabutils.ResourceQuantity("50m"),
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "shell-config",
						MountPath: "/config",
						ReadOnly:  true,
					},
					{
						Name:      "shell-init-secrets",
						MountPath: "/init-config",
						ReadOnly:  true,
					},
					{
						Name:      "shell-secrets",
						MountPath: "/init-secrets",
					},
				},
			},
		},
		Containers: []corev1.Container{
			{
				Name:            "gitlab-shell",
				Image:           BuildRelease(cr).Shell(),
				ImagePullPolicy: corev1.PullIfNotPresent,
				Env: []corev1.EnvVar{
					{
						Name:  "GITALY_FEATURE_DEFAULT_ON",
						Value: "1",
					},
					{
						Name:  "CONFIG_TEMPLATE_DIRECTORY",
						Value: "/etc/gitlab-shell",
					},
					{
						Name:  "CONFIG_DIRECTORY",
						Value: "/srv/gitlab-shell",
					},
					{
						Name:  "KEYS_DIRECTORY",
						Value: "/etc/gitlab-secrets/ssh",
					},
				},
				Ports: []corev1.ContainerPort{
					{
						Name:          "ssh",
						ContainerPort: 2222,
						Protocol:      corev1.ProtocolTCP,
					},
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"cpu":    gitlabutils.ResourceQuantity("0"),
						"memory": gitlabutils.ResourceQuantity("6M"),
					},
				},
				LivenessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: []string{"/scripts/healthcheck"},
						},
					},
					FailureThreshold:    3,
					InitialDelaySeconds: 10,
					PeriodSeconds:       10,
					SuccessThreshold:    1,
					TimeoutSeconds:      3,
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "shell-config",
						MountPath: "/etc/gitlab-shell",
					},
					{
						Name:      "shell-secrets",
						MountPath: "/etc/gitlab-secrets",
						ReadOnly:  true,
					},
					{
						Name:      "sshd-config",
						MountPath: "/etc/ssh/sshd_config",
						SubPath:   "sshd_config",
						ReadOnly:  true,
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
				Name: "shell-config",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: cr.Name + "-gitlab-shell",
						},
						DefaultMode: &gitlabutils.ConfigMapDefaultMode,
					},
				},
			},
			{
				Name: "sshd-config",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: cr.Name + "-gitlab-shell-sshd",
						},
						Items: []corev1.KeyToPath{
							{
								Key:  "sshd_config",
								Path: "sshd_config",
							},
						},
						DefaultMode: &gitlabutils.ConfigMapDefaultMode,
					},
				},
			},
			{
				Name: "shell-init-secrets",
				VolumeSource: corev1.VolumeSource{
					Projected: &corev1.ProjectedVolumeSource{
						DefaultMode: &gitlabutils.SecretDefaultMode,
						Sources: []corev1.VolumeProjection{
							{
								Secret: &corev1.SecretProjection{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: cr.Name + "-gitlab-shell-host-keys",
									},
								},
							},
							{
								Secret: &corev1.SecretProjection{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: cr.Name + "-gitlab-shell-secret",
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
						},
					},
				},
			},
			{
				Name: "shell-secrets",
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

	shell.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{
		FSGroup:   &localUser,
		RunAsUser: &localUser,
	}

	shell.Spec.Template.Spec.ServiceAccountName = AppServiceAccount

	return shell
}
