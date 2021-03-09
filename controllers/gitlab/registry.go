package gitlab

import (
	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	gitlabutils "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// RegistryDeploymentDEPRECATED returns container registry deployment
func RegistryDeploymentDEPRECATED(cr *gitlabv1beta1.GitLab) *appsv1.Deployment {
	labels := gitlabutils.Label(cr.Name, "registry", gitlabutils.GitlabType)
	options := SystemBuildOptions(cr)

	return gitlabutils.GenericDeployment(gitlabutils.Component{
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
				ImagePullPolicy: corev1.PullAlways,
				Command: []string{
					"sh",
					"/config/configure",
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"cpu": gitlabutils.ResourceQuantity("50m"),
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "registry-secrets",
						MountPath: "/config",
					},
					{
						Name:      "registry-server-config",
						MountPath: "/registry",
					},
				},
			},
		},
		Containers: []corev1.Container{
			{
				Name:            "registry",
				Image:           BuildRelease(cr).Registry(),
				ImagePullPolicy: corev1.PullIfNotPresent,
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"cpu":    gitlabutils.ResourceQuantity("50m"),
						"memory": gitlabutils.ResourceQuantity("32Mi"),
					},
				},
				LivenessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						HTTPGet: &corev1.HTTPGetAction{
							Path: "/debug/health",
							Port: intstr.IntOrString{
								IntVal: 5001,
							},
							Scheme: corev1.URISchemeHTTP,
						},
					},
					FailureThreshold:    3,
					InitialDelaySeconds: 5,
					PeriodSeconds:       10,
					SuccessThreshold:    1,
					TimeoutSeconds:      1,
				},
				ReadinessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						HTTPGet: &corev1.HTTPGetAction{
							Path: "/debug/health",
							Port: intstr.IntOrString{
								IntVal: 5001,
							},
							Scheme: corev1.URISchemeHTTP,
						},
					},
					FailureThreshold:    3,
					InitialDelaySeconds: 5,
					PeriodSeconds:       5,
					SuccessThreshold:    1,
					TimeoutSeconds:      1,
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "registry-server-config",
						MountPath: "/etc/docker/registry/",
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
				Name: "registry-server-config",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{
						Medium: corev1.StorageMediumMemory,
					},
				},
			},
			{
				Name: "registry-secrets",
				VolumeSource: corev1.VolumeSource{
					Projected: &corev1.ProjectedVolumeSource{
						DefaultMode: &gitlabutils.ConfigMapDefaultMode,
						Sources: []corev1.VolumeProjection{
							{
								ConfigMap: &corev1.ConfigMapProjection{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: cr.Name + "-registry-config",
									},
								},
							},
							{
								Secret: &corev1.SecretProjection{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: cr.Name + "-registry-secret",
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "registry-auth.crt",
											Path: "certificate.crt",
										},
									},
								},
							},
							{
								Secret: &corev1.SecretProjection{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: cr.Name + "-registry-http-secret",
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "secret",
											Path: "httpSecret",
										},
									},
								},
							},
							{
								Secret: &corev1.SecretProjection{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: options.ObjectStore.Credentials,
									},
								},
							},
						},
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
}
