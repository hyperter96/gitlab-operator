package gitlab

import (
	gitlabv1beta1 "github.com/OchiengEd/gitlab-operator/pkg/apis/gitlab/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GenericDeployment returns a generic deployment
func GenericDeployment(cr *gitlabv1beta1.Gitlab, component Component) *appsv1.Deployment {
	var replicas int32
	labels := component.Labels

	if component.Replicas != 0 {
		replicas = component.Replicas
	} else {
		replicas = 1
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels["app.kubernetes.io/name"],
			Namespace: cr.Namespace,
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

	return GenericDeployment(cr, Component{
		Labels:   labels,
		Replicas: cr.Spec.Replicas,
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
								Key: "external_url",
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
								Key: "omnibus_config",
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
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "config",
						MountPath: "/etc/gitlab",
					},
					{
						Name:      "data",
						MountPath: "/gitlab-data",
					},
					{
						Name:      "registry",
						MountPath: "/gitlab-registry",
					},
				},
				// LivenessProbe: &corev1.Probe{
				// 	Handler: corev1.Handler{
				// 		HTTPGet: &corev1.HTTPGetAction{
				// 			Path: "/health_check",
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
				// 			Path: "/health_check",
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
		Volumes: []corev1.Volume{
			{
				Name: "data",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: labels["app.kubernetes.io/name"] + "-data",
					},
				},
			},
			{
				Name: "registry",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: labels["app.kubernetes.io/name"] + "-registry",
					},
				},
			},
			{
				Name: "config",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: labels["app.kubernetes.io/name"] + "-config",
					},
				},
			},
		},
	})
}
