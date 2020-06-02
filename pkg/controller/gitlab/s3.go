package gitlab

import (
	miniov1beta1 "github.com/minio/minio-operator/pkg/apis/miniocontroller/v1beta1"
	gitlabv1beta1 "gitlab.com/ochienged/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlabutils "gitlab.com/ochienged/gitlab-operator/pkg/controller/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func getMinioSecret(cr *gitlabv1beta1.Gitlab) *corev1.Secret {
	labels := gitlabutils.Label(cr.Name, "minio", gitlabutils.GitlabType)

	secretKey := gitlabutils.Password(gitlabutils.PasswordOptions{
		EnableSpecialChars: false,
		Length:             48,
	})

	minio := gitlabutils.GenericSecret(cr.Name+"-minio-secret", cr.Namespace, labels)
	minio.StringData = map[string]string{
		"accesskey": "gitlab",
		"secretkey": secretKey,
	}

	return minio
}

func getMinioInstance(cr *gitlabv1beta1.Gitlab) *miniov1beta1.MinIOInstance {
	labels := gitlabutils.Label(cr.Name, "minio", gitlabutils.GitlabType)

	minioOptions := getMinioOverrides(cr.Spec.Minio)

	minio := &miniov1beta1.MinIOInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-minio",
			Labels:    labels,
			Namespace: cr.Namespace,
		},
		Spec: miniov1beta1.MinIOInstanceSpec{
			Metadata: &metav1.ObjectMeta{
				Labels: labels,
			},
			Replicas: minioOptions.Replicas,
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					"cpu":    gitlabutils.ResourceQuantity("250m"),
					"memory": gitlabutils.ResourceQuantity("512Mi"),
				},
			},
			Image: MinioImage,
			CredsSecret: &corev1.LocalObjectReference{
				Name: cr.Name + "-minio-secret",
			},
			RequestAutoCert: false,
			Env: []corev1.EnvVar{
				{
					Name:  "MINIO_BROWSER",
					Value: "on",
				},
			},
			Mountpath: "/export",
			Liveness: &corev1.Probe{
				Handler: corev1.Handler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: "/minio/health/live",
						Port: intstr.IntOrString{
							IntVal: 9000,
						},
					},
				},
				InitialDelaySeconds: 120,
				PeriodSeconds:       20,
			},
			Readiness: &corev1.Probe{
				Handler: corev1.Handler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: "/minio/health/ready",
						Port: intstr.IntOrString{
							IntVal: 9000,
						},
					},
				},
				InitialDelaySeconds: 120,
				PeriodSeconds:       20,
			},
			PodManagementPolicy: appsv1.ParallelPodManagement,
			VolumeClaimTemplate: &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "data",
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							"storage": gitlabutils.ResourceQuantity(minioOptions.Volume.Capacity),
						},
					},
				},
			},
		},
	}

	if minioOptions.Volume.StorageClass != "" {
		minio.Spec.VolumeClaimTemplate.Spec.StorageClassName = &minioOptions.Volume.StorageClass
	}

	return minio
}

func getMinioService(cr *gitlabv1beta1.Gitlab) *corev1.Service {
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

func getMinioIngress(cr *gitlabv1beta1.Gitlab) *extensionsv1beta1.Ingress {
	labels := gitlabutils.Label(cr.Name, "minio", gitlabutils.GitlabType)

	return &extensionsv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        cr.Name + "-minio-ingress",
			Namespace:   cr.Namespace,
			Labels:      labels,
			Annotations: IngressAnnotations(cr, RequiresCertManagerCertificate(cr).Minio()),
		},
		Spec: extensionsv1beta1.IngressSpec{
			Rules: []extensionsv1beta1.IngressRule{
				{
					// Add Registry rule only when registry is enabled
					Host: getMinioURL(cr),
					IngressRuleValue: extensionsv1beta1.IngressRuleValue{
						HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
							Paths: []extensionsv1beta1.HTTPIngressPath{
								{
									Path: "/",
									Backend: extensionsv1beta1.IngressBackend{
										ServicePort: intstr.IntOrString{
											IntVal: 9000,
										},
										ServiceName: cr.Name + "-minio",
									},
								},
							},
						},
					},
				},
			},
			TLS: getMinioIngressCert(cr),
		},
	}
}

func getMinioIngressCert(cr *gitlabv1beta1.Gitlab) []extensionsv1beta1.IngressTLS {

	if RequiresCertManagerCertificate(cr).Minio() {
		return []extensionsv1beta1.IngressTLS{
			{
				Hosts:      []string{getMinioURL(cr)},
				SecretName: cr.Name + "-minio-tls",
			},
		}
	}

	return []extensionsv1beta1.IngressTLS{
		{
			Hosts:      []string{getMinioURL(cr)},
			SecretName: cr.Spec.Minio.TLS,
		},
	}
}

func (r *ReconcileGitlab) reconcileMinioInstance(cr *gitlabv1beta1.Gitlab) error {
	minio := getMinioInstance(cr)

	if !gitlabutils.IsMinioAvailable() {
		return nil
	}

	if err := r.createKubernetesResource(minio, cr); err != nil {
		return err
	}

	secret := getMinioSecret(cr)
	if err := r.createKubernetesResource(secret, cr); err != nil {
		return err
	}

	svc := getMinioService(cr)
	if err := r.createKubernetesResource(svc, cr); err != nil {
		return err
	}

	ingress := getMinioIngress(cr)
	return r.createKubernetesResource(ingress, cr)
}
