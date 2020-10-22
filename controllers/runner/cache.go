package runner

import (
	"strconv"

	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

func getRunnerCache(cr *gitlabv1beta1.Runner) []corev1.EnvVar {

	switch cr.Spec.CacheType {
	case "s3":
		return setupS3Cache(cr)
	case "gcs":
		return setupGCSCache(cr)
	case "azure":
		return setupAzureCache(cr)
	}

	return []corev1.EnvVar{}
}

func setupS3Cache(cr *gitlabv1beta1.Runner) []corev1.EnvVar {
	environments := []corev1.EnvVar{}

	if cr.Spec.S3 == nil {
		return environments
	}

	if cr.Spec.S3.Server != "" {
		environments = append(environments, corev1.EnvVar{
			Name:  "CACHE_S3_SERVER_ADDRESS",
			Value: cr.Spec.S3.Server,
		})
	}

	if cr.Spec.S3.BucketName != "" {
		environments = append(environments, corev1.EnvVar{
			Name:  "CACHE_S3_BUCKET_NAME",
			Value: cr.Spec.S3.BucketName,
		})
	}

	if cr.Spec.S3.BucketLocation != "" {
		environments = append(environments, corev1.EnvVar{
			Name:  "CACHE_S3_BUCKET_LOCATION",
			Value: cr.Spec.S3.BucketLocation,
		})
	}

	if cr.Spec.S3.Insecure {
		environments = append(environments, corev1.EnvVar{
			Name:  "CACHE_S3_INSECURE",
			Value: strconv.FormatBool(cr.Spec.S3.Insecure),
		})
	}

	return environments
}

func setupGCSCache(cr *gitlabv1beta1.Runner) []corev1.EnvVar {
	env := []corev1.EnvVar{}

	if cr.Spec.GCS == nil {
		return env
	}

	if cr.Spec.GCS.BucketName != "" {
		env = append(env, corev1.EnvVar{
			Name:  "CACHE_GCS_BUCKET_NAME",
			Value: cr.Spec.GCS.BucketName,
		})
	}

	return env
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

func setupAzureCache(cr *gitlabv1beta1.Runner) []corev1.EnvVar {
	env := []corev1.EnvVar{}

	if cr.Spec.Azure.Credentials != "" {
		env = append(env, corev1.EnvVar{
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

		env = append(env, corev1.EnvVar{
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
		env = append(env, corev1.EnvVar{
			Name:  "CACHE_AZURE_CONTAINER_NAME",
			Value: cr.Spec.Azure.ContainerName,
		})
	}

	if cr.Spec.Azure.StorageDomain != "" {
		env = append(env, corev1.EnvVar{
			Name:  "CACHE_AZURE_STORAGE_DOMAIN",
			Value: cr.Spec.Azure.StorageDomain,
		})
	}

	return env
}

// IsCacheS3 checks if the GitLab Runner Cache is of type S3
func isCacheS3(cr *gitlabv1beta1.Runner) bool {
	return cr.Spec.S3 != nil && cr.Spec.S3.Credentials != ""
}
