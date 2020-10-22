package runner

import (
	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	gitlabutils "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// RunnerServiceAccount defines the sa for GitLab runner
const RunnerServiceAccount string = "gitlab-runner"

// Deployment returns the runner deployment object
func Deployment(cr *gitlabv1beta1.Runner) *appsv1.Deployment {
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
				Env:             getEnvironmentVariables(cr),
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
				Env: getEnvironmentVariables(cr),
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
						Sources: runnerSecretsVolume(cr),
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

	if isCacheS3(cr) {
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

	// append GCS cache if enabled
	if cr.Spec.GCS != nil {
		secrets = append(secrets, gcsCredentialsSecretProjection(cr))
	}

	if cr.Spec.CertificateAuthority != "" {
		secrets = append(secrets, getCertificateAuthoritySecretProjection(cr))
	}

	return secrets
}

// RegistrationTokenSecretName returns name of secret containing the
// runner-registration-token and runner-token keys
func RegistrationTokenSecretName(cr *gitlabv1beta1.Runner) string {
	var tokenSecretName string

	if cr.Spec.GitLab.Name != "" {
		tokenSecretName = cr.Spec.GitLab.Name + "-runner-token-secret"
	}

	if cr.Spec.RegistrationToken != "" {
		// If user provides a secret with registration token
		// set it to the gitlab secret
		tokenSecretName = cr.Spec.RegistrationToken
	}

	return tokenSecretName
}

func gcsCredentialsSecretProjection(cr *gitlabv1beta1.Runner) corev1.VolumeProjection {
	if cr.Spec.GCS.Credentials != "" {
		return corev1.VolumeProjection{
			Secret: &corev1.SecretProjection{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: cr.Spec.GCS.Credentials,
				},
				Items: []corev1.KeyToPath{
					{
						Key:  "access-id",
						Path: "gcs-access-id",
					},
					{
						Key:  "private-key",
						Path: "gcs-private-key",
					},
				},
			},
		}
	}

	return corev1.VolumeProjection{
		Secret: &corev1.SecretProjection{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: cr.Spec.GCS.CredentialsFile,
			},
			Items: []corev1.KeyToPath{
				{
					Key:  "keys.json",
					Path: "gcs-application-credentials-file",
				},
			},
		},
	}
}

func getCertificateAuthoritySecretProjection(cr *gitlabv1beta1.Runner) corev1.VolumeProjection {
	return corev1.VolumeProjection{
		Secret: &corev1.SecretProjection{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: cr.Spec.CertificateAuthority,
			},
			Items: []corev1.KeyToPath{
				{
					Key:  "tls.crt",
					Path: "hostname.crt",
				},
			},
		},
	}
}

// isCacheS3 checks if the GitLab Runner Cache is of type S3
func isCacheS3(cr *gitlabv1beta1.Runner) bool {
	return cr.Spec.S3 != nil && cr.Spec.S3.Credentials != ""
}
