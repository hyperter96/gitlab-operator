package runner

import (
	"strconv"

	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

func getEnvironmentVariables(cr *gitlabv1beta1.Runner) []corev1.EnvVar {

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

	if cr.Spec.S3 != nil {
		// setup S3 block storage
		if cr.Spec.S3.Server != "" {
			envs = append(envs, corev1.EnvVar{
				Name:  "CACHE_S3_SERVER_ADDRESS",
				Value: cr.Spec.S3.Server,
			})
		}

		if cr.Spec.S3.BucketName != "" {
			envs = append(envs, corev1.EnvVar{
				Name:  "CACHE_S3_BUCKET_NAME",
				Value: cr.Spec.S3.BucketName,
			})
		}

		if cr.Spec.S3.BucketLocation != "" {
			envs = append(envs, corev1.EnvVar{
				Name:  "CACHE_S3_BUCKET_LOCATION",
				Value: cr.Spec.S3.BucketLocation,
			})
		}

		if cr.Spec.S3.Insecure {
			envs = append(envs, corev1.EnvVar{
				Name:  "CACHE_S3_INSECURE",
				Value: strconv.FormatBool(cr.Spec.S3.Insecure),
			})
		}
	}

	if cr.Spec.GCS != nil {
		// GCS cloud storage
		if cr.Spec.GCS.BucketName != "" {
			envs = append(envs, corev1.EnvVar{
				Name:  "CACHE_GCS_BUCKET_NAME",
				Value: cr.Spec.GCS.BucketName,
			})
		}
	}

	if cr.Spec.Azure != nil {
		// setup Azure blob storage
		if cr.Spec.Azure.Credentials != "" {
			envs = append(envs, corev1.EnvVar{
				Name: "CACHE_AZURE_ACCOUNT_NAME",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: cr.Spec.Azure.Credentials,
						},
						Key: "accountName",
					},
				},
			})

			envs = append(envs, corev1.EnvVar{
				Name: "CACHE_AZURE_ACCOUNT_KEY",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: cr.Spec.Azure.Credentials,
						},
						Key: "privateKey",
					},
				},
			})
		}

		if cr.Spec.Azure.ContainerName != "" {
			envs = append(envs, corev1.EnvVar{
				Name:  "CACHE_AZURE_CONTAINER_NAME",
				Value: cr.Spec.Azure.ContainerName,
			})
		}

		if cr.Spec.Azure.StorageDomain != "" {
			envs = append(envs, corev1.EnvVar{
				Name:  "CACHE_AZURE_STORAGE_DOMAIN",
				Value: cr.Spec.Azure.StorageDomain,
			})
		}
	}

	return envs
}
