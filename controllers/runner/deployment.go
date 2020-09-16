package runner

import (
	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	gitlabutils "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// RunnerServiceAccount defines the sa for GitLab runner
const RunnerServiceAccount string = "gitlab-runner"

// GetDeployment returns the runner deployment object
func GetDeployment(cr *gitlabv1beta1.Runner) *appsv1.Deployment {
	labels := gitlabutils.Label(cr.Name, "runner", gitlabutils.RunnerType)

	config := runnerConfig(cr)

	runner := gitlabutils.GenericDeployment(gitlabutils.Component{
		Labels:    labels,
		Namespace: cr.Namespace,
		InitContainers: []corev1.Container{
			{
				Name:            "configure",
				Image:           GitlabRunnerImage,
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
						Value: config.Tags,
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
						Value: config.Cache,
					},
					{
						Name:  "CACHE_SHARED",
						Value: "true",
					},
					{
						Name:  "CACHE_S3_SERVER_ADDRESS",
						Value: config.Server,
					},
					{
						Name:  "CACHE_S3_BUCKET_NAME",
						Value: config.Bucket,
					},
					{
						Name:  "CACHE_S3_BUCKET_LOCATION",
						Value: config.Region,
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
				Image:   GitlabRunnerImage,
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
									Name: cr.Name + "-runner-secret",
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
									Name: cr.Name + "-runner-secret",
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
						Value: config.Tags,
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
						Value: config.Cache,
					},
					{
						Name:  "CACHE_SHARED",
						Value: "true",
					},
					{
						Name:  "CACHE_S3_SERVER_ADDRESS",
						Value: config.Server,
					},
					{
						Name:  "CACHE_S3_BUCKET_NAME",
						Value: config.Bucket,
					},
					{
						Name:  "CACHE_S3_BUCKET_LOCATION",
						Value: config.Region,
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
						Name:      "scripts",
						MountPath: "/scripts",
					},
				},
			},
		},
		Volumes: []corev1.Volume{
			{
				Name: "runner-secrets",
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
						DefaultMode: &gitlabutils.ConfigMapDefaultMode,
						Sources:     runnerSecretsVolume(cr),
					},
				},
			},
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
							{
								Key:  "configure",
								Path: "configure",
							},
						},
					},
				},
			},
		},
	})

	// Use certified image if running on Openshift
	if gitlabutils.IsOpenshift() {
		runner.Spec.Template.Spec.InitContainers[0].Image = CertifiedRunnerImage
		runner.Spec.Template.Spec.Containers[0].Image = CertifiedRunnerImage
	}

	// Set runner to use specific service account
	runner.Spec.Template.Spec.ServiceAccountName = RunnerServiceAccount

	return runner
}

func runnerSecretsVolume(cr *gitlabv1beta1.Runner) []corev1.VolumeProjection {
	secrets := []corev1.VolumeProjection{
		{
			Secret: &corev1.SecretProjection{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: cr.Name + "-runner-secret",
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
	}

	if cr.Spec.Cache != nil && cr.Spec.Cache.Credentials != "" {
		secrets = append(secrets,
			corev1.VolumeProjection{
				Secret: &corev1.SecretProjection{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: cr.Spec.Cache.Credentials,
					},
				},
			},
		)
	}

	return secrets
}

type runnerOptions struct {
	Tags        string
	Server      string
	Region      string
	Bucket      string
	Cache       string
	Credentials string
}

func runnerConfig(cr *gitlabv1beta1.Runner) (options runnerOptions) {

	if cr.Spec.Tags != "" {
		options.Tags = cr.Spec.Tags
	}

	if cr.Spec.Cache != nil {
		options.Server = cr.Spec.Cache.Server
		options.Region = cr.Spec.Cache.Region

		if cr.Spec.Cache.Path != "" {
			options.Cache = cr.Spec.Cache.Path
		} else {
			options.Cache = "gitlab-runner"
		}

		if cr.Spec.Cache.Bucket != "" {
			options.Bucket = cr.Spec.Cache.Bucket
		} else {
			options.Bucket = "runner-cache"
		}

		if cr.Spec.Cache.Credentials != "" {
			options.Credentials = cr.Spec.Cache.Credentials
		}
	}

	return
}
