package gitlab

import (
	gitlabv1beta1 "gitlab.com/ochienged/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlabutils "gitlab.com/ochienged/gitlab-operator/pkg/controller/utils"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func getGitlabIngress(cr *gitlabv1beta1.Gitlab) (ingress *extensionsv1beta1.Ingress) {
	labels := gitlabutils.Label(cr.Name, "ingress", gitlabutils.GitlabType)

	tls := getGitlabIngressCert(cr)

	ingress = &extensionsv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-gitlab-ingress",
			Namespace: cr.Namespace,
			Labels:    labels,
			Annotations: map[string]string{
				"kubernetes.io/ingress.class": "nginx",
				"cert-manager.io/issuer":      cr.Name + "-issuer",
			},
		},
		Spec: extensionsv1beta1.IngressSpec{
			Rules: []extensionsv1beta1.IngressRule{
				{
					// External URL for the gitlab instance
					Host: getGitlabURL(cr),
					IngressRuleValue: extensionsv1beta1.IngressRuleValue{
						HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
							Paths: []extensionsv1beta1.HTTPIngressPath{
								{
									Path: "/",
									Backend: extensionsv1beta1.IngressBackend{
										ServicePort: intstr.IntOrString{
											IntVal: 8181,
										},
										ServiceName: cr.Name + "-unicorn",
									},
								},
							},
						},
					},
				},
			},
			TLS: tls,
		},
	}

	return
}

func getRegistryIngress(cr *gitlabv1beta1.Gitlab) (ingress *extensionsv1beta1.Ingress) {
	labels := gitlabutils.Label(cr.Name, "ingress", gitlabutils.GitlabType)

	tls := getRegistryIngressCert(cr)

	return &extensionsv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-registry-ingress",
			Namespace: cr.Namespace,
			Labels:    labels,
			Annotations: map[string]string{
				"kubernetes.io/ingress.class": "nginx",
				"cert-manager.io/issuer":      cr.Name + "-issuer",
			},
		},
		Spec: extensionsv1beta1.IngressSpec{
			Rules: []extensionsv1beta1.IngressRule{
				{
					// Add Registry rule only when registry is enabled
					Host: getRegistryURL(cr),
					IngressRuleValue: extensionsv1beta1.IngressRuleValue{
						HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
							Paths: []extensionsv1beta1.HTTPIngressPath{
								{
									Path: "/",
									Backend: extensionsv1beta1.IngressBackend{
										ServicePort: intstr.IntOrString{
											IntVal: 5000,
										},
										ServiceName: cr.Name + "-registry",
									},
								},
							},
						},
					},
				},
			},
			TLS: tls,
		},
	}
}

func getMinioIngress(cr *gitlabv1beta1.Gitlab) *extensionsv1beta1.Ingress {
	labels := gitlabutils.Label(cr.Name, "minio", gitlabutils.GitlabType)

	tls := getMinioIngressCert(cr)

	return &extensionsv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-minio-ingress",
			Namespace: cr.Namespace,
			Labels:    labels,
			Annotations: map[string]string{
				"kubernetes.io/ingress.class": "nginx",
				"cert-manager.io/issuer":      cr.Name + "-issuer",
			},
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
			TLS: tls,
		},
	}
}

func (r *ReconcileGitlab) reconcileIngress(cr *gitlabv1beta1.Gitlab) error {
	var ingresses []*extensionsv1beta1.Ingress
	gitlab := getGitlabIngress(cr)

	registry := getRegistryIngress(cr)

	minio := getMinioIngress(cr)

	ingresses = append(ingresses,
		gitlab,
		registry,
		minio,
	)

	for _, ingress := range ingresses {
		if err := r.createKubernetesResource(cr, ingress); err != nil {
			return err
		}
	}

	return nil
}

func getGitlabIngressCert(cr *gitlabv1beta1.Gitlab) []extensionsv1beta1.IngressTLS {
	tlsSecret := cr.Name + "-gitlab-tls"

	if cr.Spec.TLS != "" {
		tlsSecret = cr.Spec.TLS
	}

	return []extensionsv1beta1.IngressTLS{
		{
			Hosts:      []string{getGitlabURL(cr)},
			SecretName: tlsSecret,
		},
	}
}

func getRegistryIngressCert(cr *gitlabv1beta1.Gitlab) []extensionsv1beta1.IngressTLS {
	tlsSecret := cr.Name + "-registry-tls"

	if cr.Spec.TLS != "" {
		tlsSecret = cr.Spec.Registry.TLS
	}

	return []extensionsv1beta1.IngressTLS{
		{
			Hosts:      []string{getRegistryURL(cr)},
			SecretName: tlsSecret,
		},
	}
}

func getMinioIngressCert(cr *gitlabv1beta1.Gitlab) []extensionsv1beta1.IngressTLS {
	tlsSecret := cr.Name + "-minio-tls"

	if cr.Spec.TLS != "" {
		tlsSecret = cr.Spec.Minio.TLS
	}

	return []extensionsv1beta1.IngressTLS{
		{
			Hosts:      []string{getMinioURL(cr)},
			SecretName: tlsSecret,
		},
	}
}
