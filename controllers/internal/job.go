package internal

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/settings"
)

// BucketCreationJob creates the buckets used by GitLab.
func BucketCreationJob(adapter gitlab.CustomResourceAdapter) *batchv1.Job {
	labels := Label(adapter.ReleaseName(), "bucket", GitlabType)
	options := SystemBuildOptions(adapter)

	buckets := GenericJob(Component{
		Namespace: adapter.Namespace(),
		Labels:    labels,
		Containers: []corev1.Container{
			{
				Name:            "mc",
				Image:           "minio/mc:RELEASE.2018-07-13T00-53-22Z",
				ImagePullPolicy: corev1.PullIfNotPresent,
				Command:         []string{"/bin/sh", "/config/initialize"},
				Env: []corev1.EnvVar{
					{
						Name:  "MINIO_ENDPOINT",
						Value: options.ObjectStore.Endpoint,
					},
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"cpu": ResourceQuantity("50m"),
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "minio-config",
						MountPath: "/config",
					},
				},
			},
		},
		Volumes: []corev1.Volume{
			{
				Name: "minio-config",
				VolumeSource: corev1.VolumeSource{
					Projected: &corev1.ProjectedVolumeSource{
						DefaultMode: &ConfigMapDefaultMode,
						Sources: []corev1.VolumeProjection{
							{
								Secret: &corev1.SecretProjection{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: options.ObjectStore.Credentials,
									},
								},
							},
							{
								ConfigMap: &corev1.ConfigMapProjection{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: adapter.ReleaseName() + "-minio-script",
									},
								},
							},
						},
					},
				},
			},
		},
	})

	var mcUser int64

	buckets.Spec.Template.Spec.ServiceAccountName = settings.AppServiceAccount
	buckets.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{
		RunAsUser: &mcUser,
		FSGroup:   &mcUser,
	}

	return buckets
}
