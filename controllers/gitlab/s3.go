package gitlab

import (
	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	gitlabutils "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// MinioSecret returns secret containing Minio accesskey and secretkey
func MinioSecret(cr *gitlabv1beta1.GitLab) *corev1.Secret {
	labels := gitlabutils.Label(cr.Name, "minio", gitlabutils.GitlabType)
	options := SystemBuildOptions(cr)

	secretKey := gitlabutils.Password(gitlabutils.PasswordOptions{
		EnableSpecialChars: false,
		Length:             48,
	})

	minio := gitlabutils.GenericSecret(options.ObjectStore.Credentials, cr.Namespace, labels)
	minio.StringData = map[string]string{
		"accesskey": "gitlab",
		"secretkey": secretKey,
	}

	return minio
}

// MinioStatefulSet return Minio statefulset
func MinioStatefulSet(cr *gitlabv1beta1.GitLab) *appsv1.StatefulSet {
	labels := gitlabutils.Label(cr.Name, "minio", gitlabutils.GitlabType)
	options := SystemBuildOptions(cr)

	var replicas int32 = 1

	minio := gitlabutils.GenericStatefulSet(gitlabutils.Component{
		Namespace: cr.Namespace,
		Labels:    labels,
		Replicas:  replicas,
		InitContainers: []corev1.Container{
			{
				Name:            "configure",
				Image:           BuildRelease(cr).Busybox(),
				ImagePullPolicy: corev1.PullIfNotPresent,
				Command:         []string{"sh", "/config/configure"},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"cpu": gitlabutils.ResourceQuantity("50m"),
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "minio-configuration",
						MountPath: "/config",
					},
					{
						Name:      "minio-server-config",
						MountPath: "/minio",
					},
				},
			},
		},
		Containers: []corev1.Container{
			{
				Name:            "minio",
				Image:           BuildRelease(cr).Minio(),
				ImagePullPolicy: corev1.PullIfNotPresent,
				Args:            []string{"-C", "/tmp/.minio", "--quiet", "server", "/export"},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"cpu":    gitlabutils.ResourceQuantity("100m"),
						"memory": gitlabutils.ResourceQuantity("128Mi"),
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "export",
						MountPath: "/export",
					},
					{
						Name:      "minio-server-config",
						MountPath: "/tmp/.minio",
					},
					{
						Name:      "podinfo",
						MountPath: "/podinfo",
						ReadOnly:  false,
					},
				},
				Ports: []corev1.ContainerPort{
					{
						Name:          "service",
						Protocol:      corev1.ProtocolTCP,
						ContainerPort: 9000,
					},
				},
				LivenessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						TCPSocket: &corev1.TCPSocketAction{
							Port: intstr.IntOrString{
								IntVal: 9000,
							},
						},
					},
					TimeoutSeconds: 1,
				},
			},
		},
		Volumes: []corev1.Volume{
			{
				Name: "podinfo",
				VolumeSource: corev1.VolumeSource{
					DownwardAPI: &corev1.DownwardAPIVolumeSource{
						Items: []corev1.DownwardAPIVolumeFile{
							{
								Path: "labels",
								FieldRef: &corev1.ObjectFieldSelector{
									FieldPath: "metadata.labels",
								},
							},
						},
					},
				},
			},
			{
				Name: "minio-server-config",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{
						Medium: corev1.StorageMediumMemory,
					},
				},
			},
			{
				Name: "minio-configuration",
				VolumeSource: corev1.VolumeSource{
					Projected: &corev1.ProjectedVolumeSource{
						Sources: []corev1.VolumeProjection{
							{
								ConfigMap: &corev1.ConfigMapProjection{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: cr.Name + "-minio-script",
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
		},
		VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "export",
					Namespace: cr.Namespace,
					Labels:    labels,
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							"storage": gitlabutils.ResourceQuantity(options.ObjectStore.Capacity),
						},
					},
				},
			},
		},
	})

	minio.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{
		RunAsUser: &localUser,
		FSGroup:   &localUser,
	}

	minio.Spec.Template.Spec.ServiceAccountName = AppServiceAccount

	return minio
}

// MinioScriptConfigMap returns scripts used to configure Minio
func MinioScriptConfigMap(cr *gitlabv1beta1.GitLab) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "minio", gitlabutils.GitlabType)

	initScript := gitlabutils.ReadConfig("/templates/minio/initialize-buckets.sh")
	configureScript := gitlabutils.ReadConfig("/templates/minio/configure.sh")
	configJSON := gitlabutils.ReadConfig("/templates/minio/config.json")

	init := gitlabutils.GenericConfigMap(cr.Name+"-minio-script", cr.Namespace, labels)
	init.Data = map[string]string{
		"initialize":  initScript,
		"configure":   configureScript,
		"config.json": configJSON,
	}

	return init
}

// MinioService returns service that exposes Minio
func MinioService(cr *gitlabv1beta1.GitLab) *corev1.Service {
	labels := gitlabutils.Label(cr.Name, "minio", gitlabutils.GitlabType)

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels["app.kubernetes.io/instance"],
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:     "minio",
					Port:     9000,
					Protocol: corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}
