package runner

import (
	gitlabv1beta1 "gitlab.com/ochienged/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlabutils "gitlab.com/ochienged/gitlab-operator/pkg/controller/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func getRunnerDeployment(cr *gitlabv1beta1.Runner) *appsv1.Deployment {
	labels := gitlabutils.Label(cr.Name, "runner", gitlabutils.RunnerType)

	// Add runner tags
	var tags string
	if cr.Spec.Tags != "" {
		tags = cr.Spec.Tags
	}

	runner := gitlabutils.GenericDeployment(gitlabutils.Component{
		Labels:    labels,
		Namespace: cr.Namespace,
		InitContainers: []corev1.Container{
			{
				Name:            "configure",
				Image:           gitlabutils.GitLabRunnerImage,
				ImagePullPolicy: corev1.PullIfNotPresent,
				Command:         []string{"sh", "/config/configure"},
				Env: []corev1.EnvVar{
					{
						Name: "CI_SERVER_URL",
						ValueFrom: &corev1.EnvVarSource{
							ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: cr.Name + "-runner-config",
								},
								Key: "ci_server_url",
							},
						},
					},
					{
						Name:  "CLONE_URL",
						Value: "",
					},
					{
						Name:  "RUNNER_REQUEST_CONCURRENCY",
						Value: "1",
					},
					{
						Name:  "RUNNER_EXECUTOR",
						Value: "kubernetes",
					},
					{
						Name:  "REGISTER_LOCKED",
						Value: "false",
					},
					{
						Name:  "RUNNER_TAG_LIST",
						Value: tags,
					},
					{
						Name:  "RUNNER_OUTPUT_LIMIT",
						Value: "4096",
					},
					{
						Name:  "KUBERNETES_IMAGE",
						Value: "ubuntu:16.04",
					},
					{
						Name:  "KUBERNETES_NAMESPACE",
						Value: "default",
					},
					{
						Name:  "KUBERNETES_POLL_TIMEOUT",
						Value: "180",
					},
					{
						Name:  "KUBERNETES_CPU_LIMIT",
						Value: "",
					},
					{
						Name:  "KUBERNETES_MEMORY_LIMIT",
						Value: "",
					},
					{
						Name:  "KUBERNETES_CPU_REQUEST",
						Value: "",
					},
					{
						Name:  "KUBERNETES_MEMORY_REQUEST",
						Value: "",
					},
					{
						Name:  "KUBERNETES_SERVICE_ACCOUNT",
						Value: "",
					},
					{
						Name:  "KUBERNETES_SERVICE_CPU_LIMIT",
						Value: "",
					},
					{
						Name:  "KUBERNETES_SERVICE_MEMORY_LIMIT",
						Value: "",
					},
					{
						Name:  "KUBERNETES_SERVICE_CPU_REQUEST",
						Value: "",
					},
					{
						Name:  "KUBERNETES_SERVICE_MEMORY_REQUEST",
						Value: "",
					},
					{
						Name:  "KUBERNETES_HELPER_CPU_LIMIT",
						Value: "",
					},
					{
						Name:  "KUBERNETES_HELPER_MEMORY_LIMIT",
						Value: "",
					},
					{
						Name:  "KUBERNETES_HELPER_CPU_REQUEST",
						Value: "",
					},
					{
						Name:  "KUBERNETES_HELPER_MEMORY_REQUEST",
						Value: "",
					},
					{
						Name:  "KUBERNETES_HELPER_IMAGE",
						Value: "",
					},
					{
						Name:  "KUBERNETES_PULL_POLICY",
						Value: "",
					},
					{
						Name:  "CACHE_TYPE",
						Value: "s3",
					},
					{
						Name:  "CACHE_PATH",
						Value: "gitlab-runner",
					},
					{
						Name:  "CACHE_SHARED",
						Value: "true",
					},
					{
						Name:  "CACHE_S3_SERVER_ADDRESS",
						Value: "minio.example.com",
					},
					{
						Name:  "CACHE_S3_BUCKET_NAME",
						Value: "runner-cache",
					},
					{
						Name:  "CACHE_S3_BUCKET_LOCATION",
						Value: "us-east-1",
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "runner-secrets",
						MountPath: "/secrets",
					},
					{
						Name:      "scripts",
						MountPath: "/config",
						ReadOnly:  true,
					},
					{
						Name:      "init-runner-secrets",
						MountPath: "/init-secrets",
						ReadOnly:  true,
					},
				},
			},
		},
		Containers: []corev1.Container{
			{
				Name:    "runner",
				Image:   gitlabutils.GitLabRunnerImage,
				Command: []string{"/bin/bash", "/scripts/entrypoint"},
				Lifecycle: &corev1.Lifecycle{
					PreStop: &corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: []string{"gitlab-runner", "unregister", "--all-runners"},
						},
					},
				},
				Env: []corev1.EnvVar{
					{
						Name: "CI_SERVER_URL",
						ValueFrom: &corev1.EnvVarSource{
							ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: cr.Name + "-runner-config",
								},
								Key: "ci_server_url",
							},
						},
					},
					{
						Name: "CI_SERVER_TOKEN",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: cr.Name + "-runner-secrets",
								},
								Key: "runner-token",
							},
						},
					},
					{
						Name: "REGISTRATION_TOKEN",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: cr.Name + "-runner-secrets",
								},
								Key: "runner-registration-token",
							},
						},
					},
					{
						Name:  "CLONE_URL",
						Value: "",
					},
					{
						Name:  "RUNNER_REQUEST_CONCURRENCY",
						Value: "1",
					},
					{
						Name:  "RUNNER_EXECUTOR",
						Value: "kubernetes",
					},
					{
						Name:  "REGISTER_LOCKED",
						Value: "false",
					},
					{
						Name:  "RUNNER_TAG_LIST",
						Value: tags,
					},
					{
						Name:  "RUNNER_OUTPUT_LIMIT",
						Value: "4096",
					},
					{
						Name:  "KUBERNETES_NAMESPACE",
						Value: cr.Namespace,
					},
					{
						Name:  "KUBERNETES_POLL_TIMEOUT",
						Value: "180",
					},
					{
						Name:  "KUBERNETES_CPU_LIMIT",
						Value: "",
					},
					{
						Name:  "KUBERNETES_MEMORY_LIMIT",
						Value: "",
					},
					{
						Name:  "KUBERNETES_CPU_REQUEST",
						Value: "",
					},
					{
						Name:  "KUBERNETES_MEMORY_REQUEST",
						Value: "",
					},
					{
						Name:  "KUBERNETES_SERVICE_ACCOUNT",
						Value: "",
					},
					{
						Name:  "KUBERNETES_SERVICE_CPU_LIMIT",
						Value: "",
					},
					{
						Name:  "KUBERNETES_SERVICE_MEMORY_LIMIT",
						Value: "",
					},
					{
						Name:  "KUBERNETES_SERVICE_CPU_REQUEST",
						Value: "",
					},
					{
						Name:  "KUBERNETES_SERVICE_MEMORY_REQUEST",
						Value: "",
					},
					{
						Name:  "KUBERNETES_HELPER_CPU_LIMIT",
						Value: "",
					},
					{
						Name:  "KUBERNETES_HELPER_MEMORY_LIMIT",
						Value: "",
					},
					{
						Name:  "KUBERNETES_HELPER_CPU_REQUEST",
						Value: "",
					},
					{
						Name:  "KUBERNETES_HELPER_MEMORY_REQUEST",
						Value: "",
					},
					{
						Name:  "KUBERNETES_HELPER_IMAGE",
						Value: "",
					},
					{
						Name:  "KUBERNETES_PULL_POLICY",
						Value: "",
					},
					{
						Name:  "CACHE_TYPE",
						Value: "s3",
					},
					{
						Name:  "CACHE_PATH",
						Value: "gitlab-runner",
					},
					{
						Name:  "CACHE_SHARED",
						Value: "true",
					},
					{
						Name:  "CACHE_S3_SERVER_ADDRESS",
						Value: "minio.example.com",
					},
					{
						Name:  "CACHE_S3_BUCKET_NAME",
						Value: "runner-cache",
					},
					{
						Name:  "CACHE_S3_BUCKET_LOCATION",
						Value: "us-east-1",
					},
				},
				LivenessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: []string{"/bin/bash", "/scripts/check-live"},
						},
					},
					FailureThreshold:    3,
					InitialDelaySeconds: 60,
					PeriodSeconds:       10,
					TimeoutSeconds:      1,
					SuccessThreshold:    1,
				},
				ReadinessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: []string{"/bin/bash", "/scripts/check-live"},
						},
					},
					FailureThreshold:    3,
					InitialDelaySeconds: 10,
					PeriodSeconds:       10,
					SuccessThreshold:    1,
					TimeoutSeconds:      1,
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "runner-secrets",
						MountPath: "/secrets",
					},
					{
						Name:      "etc-gitlab-runner",
						MountPath: "/home/gitlab-runner/.gitlab-runner",
					},
					{
						Name:      "scripts",
						MountPath: "/scripts",
					},
				},
			},
		},
		Volumes: []corev1.Volume{
			{
				Name: "scripts",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: cr.Name + "-runner-config",
						},
						Items: []corev1.KeyToPath{
							{
								Key:  "config.toml",
								Path: "config.toml",
							},
							{
								Key:  "entrypoint",
								Path: "entrypoint",
							},
							{
								Key:  "register-runner",
								Path: "register-runner",
							},
							{
								Key:  "check-live",
								Path: "check-live",
							},
						},
					},
				},
			},
			{
				Name: "runner-secrets",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{
						Medium: corev1.StorageMediumMemory,
					},
				},
			},
			{
				Name: "etc-gitlab-runner",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{
						Medium: corev1.StorageMediumMemory,
					},
				},
			},
			{
				Name: "init-runner-secrets",
				VolumeSource: corev1.VolumeSource{
					Projected: &corev1.ProjectedVolumeSource{
						// DefaultMode: 420,
						Sources: []corev1.VolumeProjection{
							// {
							// 	Secret: &corev1.SecretProjection{
							// 		LocalObjectReference: corev1.LocalObjectReference{
							// 			Name: cr.Name + "-minio-secret",
							// 		},
							// 	},
							// },
							{
								Secret: &corev1.SecretProjection{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: cr.Name + "-gitlab-secrets",
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "runner-registration-token",
											Path: "runner-registration-token",
										},
										{
											Key:  "runner-token",
											Path: "runner-token",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})

	// Set security context
	var fsGroup int64 = 65533
	var runUser int64 = 100
	runner.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{
		FSGroup:   &fsGroup,
		RunAsUser: &runUser,
	}

	// Set runner to use specific service account
	runner.Spec.Template.Spec.ServiceAccountName = cr.Name + "-runner"

	return runner
}
