package runner

import (
	"reflect"
	"strconv"

	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	gitlabutils "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// RunnerServiceAccount defines the sa for GitLab runner
const RunnerServiceAccount string = "gitlab-runner"

func getEnvironmentVars(cr *gitlabv1beta1.Runner) []corev1.EnvVar {

	envs := []corev1.EnvVar{
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
						Name: RegistrationTokenSecretName(cr),
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
						Name: RegistrationTokenSecretName(cr),
					},
					Key: "runner-registration-token",
				},
			},
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
			Name:  "CACHE_SHARED",
			Value: strconv.FormatBool(cr.Spec.CacheShared),
		},
	}

	if cr.Spec.Tags != "" {
		envs = append(envs, corev1.EnvVar{
			Name:  "RUNNER_TAG_LIST",
			Value: cr.Spec.Tags,
		})
	}

	if cr.Spec.HelperImage != "" {
		envs = append(envs, corev1.EnvVar{
			Name:  "KUBERNETES_HELPER_IMAGE",
			Value: cr.Spec.HelperImage,
		})
	}

	if cr.Spec.BuildImage != "" {
		envs = append(envs, corev1.EnvVar{
			Name:  "KUBERNETES_IMAGE",
			Value: cr.Spec.BuildImage,
		})
	}

	if cr.Spec.CloneURL != "" {
		envs = append(envs, corev1.EnvVar{
			Name:  "CLONE_URL",
			Value: cr.Spec.CloneURL,
		})
	}

	if cr.Spec.CacheType != "" {
		envs = append(envs, corev1.EnvVar{
			Name:  "CACHE_TYPE",
			Value: cr.Spec.CacheType,
		})
	}

	if cr.Spec.CachePath != "" {
		envs = append(envs, corev1.EnvVar{
			Name:  "CACHE_PATH",
			Value: cr.Spec.CachePath,
		})
	}

	cache := getRunnerCache(cr)
	if !reflect.DeepEqual(cache, []corev1.EnvVar{}) {
		envs = append(envs, cache...)
	}

	return envs
}

// GetDeployment returns the runner deployment object
func GetDeployment(cr *gitlabv1beta1.Runner) *appsv1.Deployment {
	labels := gitlabutils.Label(cr.Name, "runner", gitlabutils.RunnerType)

	runner := gitlabutils.GenericDeployment(gitlabutils.Component{
		Labels:    labels,
		Namespace: cr.Namespace,
		InitContainers: []corev1.Container{
			{
				Name:            "configure",
				Image:           GitlabRunnerImage,
				ImagePullPolicy: corev1.PullIfNotPresent,
				Command:         []string{"sh", "/config/configure"},
				Env:             getEnvironmentVars(cr),
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
				Name:            "runner",
				Image:           GitlabRunnerImage,
				ImagePullPolicy: corev1.PullIfNotPresent,
				Command:         []string{"/bin/bash", "/scripts/entrypoint"},
				Lifecycle: &corev1.Lifecycle{
					PreStop: &corev1.Handler{
						Exec: &corev1.ExecAction{
							Command: []string{"gitlab-runner", "unregister", "--all-runners"},
						},
					},
				},
				Env: getEnvironmentVars(cr),
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
					Name: RegistrationTokenSecretName(cr),
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

	if IsCacheS3(cr) {
		secrets = append(secrets,
			corev1.VolumeProjection{
				Secret: &corev1.SecretProjection{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: cr.Spec.S3.Credentials,
					},
				},
			},
		)
	}

	return secrets
}

// RegistrationTokenSecretName returns name of secret containing the
// runner-registration-token and runner-token keys
func RegistrationTokenSecretName(cr *gitlabv1beta1.Runner) string {
	var tokenSecretName string

	if cr.Spec.Gitlab.Name != "" {
		tokenSecretName = cr.Spec.Gitlab.Name + "-runner-token-secret"
	}

	if cr.Spec.RegistrationToken != "" {
		// If user provides a secret with registration token
		// set it to the gitlab secret
		tokenSecretName = cr.Spec.RegistrationToken
	}

	return tokenSecretName
}
