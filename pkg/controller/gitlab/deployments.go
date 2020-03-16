package gitlab

import (
	gitlabv1beta1 "github.com/OchiengEd/gitlab-operator/pkg/apis/gitlab/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GenericDeployment returns a generic deployment
func GenericDeployment(component Component) *appsv1.Deployment {
	var replicas int32
	labels := component.Labels

	if component.Replicas != 0 {
		replicas = component.Replicas
	} else {
		replicas = 1
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels["app.kubernetes.io/instance"],
			Namespace: component.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: component.Containers,
					Volumes:    component.Volumes,
				},
			},
		},
	}
}

func getGitlabDeployment(cr *gitlabv1beta1.Gitlab) *appsv1.Deployment {
	var containerImage string = GitlabCommunityImage
	labels := getLabels(cr, "gitlab")

	if cr.Spec.Enterprise {
		containerImage = GitlabEnterpriseImage
	}

	// Build list of volumes to used by GitLab deployment
	volumes := []corev1.Volume{}
	mounts := []corev1.VolumeMount{}

	if cr.Spec.Volumes.Data.Capacity != "" {
		volumes = append(volumes, corev1.Volume{
			Name: "data",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: labels["app.kubernetes.io/instance"] + "-data",
				},
			},
		})

		mounts = append(mounts, corev1.VolumeMount{
			Name:      "data",
			MountPath: "/gitlab-data",
		})
	}

	if cr.Spec.Volumes.Configuration.Capacity != "" {
		volumes = append(volumes, corev1.Volume{
			Name: "config",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: labels["app.kubernetes.io/instance"] + "-config",
				},
			},
		})

		mounts = append(mounts, corev1.VolumeMount{
			Name:      "config",
			MountPath: "/etc/gitlab",
		})
	}

	// Conditionally add registry volume and volume mount
	if cr.Spec.Registry.Enabled && cr.Spec.Volumes.Registry.Capacity != "" {
		volumes = append(volumes, corev1.Volume{
			Name: "registry",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: labels["app.kubernetes.io/instance"] + "-registry",
				},
			},
		})

		mounts = append(mounts, corev1.VolumeMount{
			Name:      "registry",
			MountPath: "/gitlab-registry",
		})
	}

	return GenericDeployment(Component{
		Namespace: cr.Namespace,
		Labels:    labels,
		Replicas:  cr.Spec.Replicas,
		Containers: []corev1.Container{
			{
				Name:            "gitlab",
				Image:           containerImage,
				ImagePullPolicy: corev1.PullIfNotPresent,
				Env: []corev1.EnvVar{
					{
						Name: "GITLAB_ROOT_PASSWORD",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: cr.Name + "-gitlab-secrets",
								},
								Key: "gitlab_root_password",
							},
						},
					},
					{
						Name: "GITLAB_EXTERNAL_URL",
						ValueFrom: &corev1.EnvVarSource{
							ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: cr.Name + "-gitlab-config",
								},
								Key: "gitlab_external_url",
							},
						},
					},
					{
						Name: "GITLAB_REGISTRY_EXTERNAL_URL",
						ValueFrom: &corev1.EnvVarSource{
							ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: cr.Name + "-gitlab-config",
								},
								Key: "registry_external_url",
							},
						},
					},
					{
						Name: "POSTGRES_HOST",
						ValueFrom: &corev1.EnvVarSource{
							ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: cr.Name + "-gitlab-config",
								},
								Key: "postgres_host",
							},
						},
					},
					{
						Name: "POSTGRES_USER",
						ValueFrom: &corev1.EnvVarSource{
							ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: cr.Name + "-gitlab-config",
								},
								Key: "postgres_user",
							},
						},
					},
					{
						Name: "POSTGRES_PASSWORD",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: cr.Name + "-gitlab-secrets",
								},
								Key: "postgres_password",
							},
						},
					},
					{
						Name: "POSTGRES_DB",
						ValueFrom: &corev1.EnvVarSource{
							ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: cr.Name + "-gitlab-config",
								},
								Key: "postgres_db",
							},
						},
					},
					{
						Name: "REDIS_HOST",
						ValueFrom: &corev1.EnvVarSource{
							ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: cr.Name + "-gitlab-config",
								},
								Key: "redis_host",
							},
						},
					},
					{
						Name: "REDIS_PASSWORD",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: cr.Name + "-gitlab-secrets",
								},
								Key: "redis_password",
							},
						},
					},
					{
						Name: "GITLAB_OMNIBUS_CONFIG",
						ValueFrom: &corev1.EnvVarSource{
							ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: cr.Name + "-gitlab-config",
								},
								Key: "gitlab_omnibus_config",
							},
						},
					},
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
				},
				Ports: []corev1.ContainerPort{
					{
						Name:          "registry",
						ContainerPort: 8105,
					},
					{
						Name:          "workhorse",
						ContainerPort: 8005,
					},
					{
						Name:          "ssh",
						ContainerPort: 22,
					},
				},
				VolumeMounts: mounts,
				// LivenessProbe: &corev1.Probe{
				// 	Handler: corev1.Handler{
				// 		HTTPGet: &corev1.HTTPGetAction{
				// 			Path: "/-/liveness",
				// 			Port: intstr.IntOrString{
				// 				IntVal: 8005,
				// 			},
				// 		},
				// 	},
				// 	InitialDelaySeconds: 180,
				// 	TimeoutSeconds:      15,
				// },
				// ReadinessProbe: &corev1.Probe{
				// 	Handler: corev1.Handler{
				// 		HTTPGet: &corev1.HTTPGetAction{
				// 			Path: "/-/readiness?all=1",
				// 			Port: intstr.IntOrString{
				// 				IntVal: 8005,
				// 			},
				// 		},
				// 	},
				// 	InitialDelaySeconds: 15,
				// 	TimeoutSeconds:      1,
				// },
			},
		},
		Volumes: volumes,
	})
}
