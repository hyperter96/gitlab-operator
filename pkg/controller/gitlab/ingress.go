package gitlab

import (
	gitlabv1beta1 "github.com/OchiengEd/gitlab-operator/pkg/apis/gitlab/v1beta1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func getGitlabIngress(cr *gitlabv1beta1.Gitlab) (ingress *extensionsv1beta1.Ingress) {
	labels := getLabels(cr, "ingress")

	ingress = &extensionsv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-gitlab-ingress",
			Namespace: cr.Namespace,
			Labels:    labels,
			Annotations: map[string]string{
				"kubernetes.io/tls-acme":      "true",
				"kubernetes.io/ingress.class": "nginx",
			},
		},
		Spec: extensionsv1beta1.IngressSpec{
			Rules: []extensionsv1beta1.IngressRule{
				{
					// External URL for the gitlab instance
					Host: DomainNameOnly(cr.Spec.ExternalURL),
					IngressRuleValue: extensionsv1beta1.IngressRuleValue{
						HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
							Paths: []extensionsv1beta1.HTTPIngressPath{
								{
									Path: "/",
									Backend: extensionsv1beta1.IngressBackend{
										ServicePort: intstr.IntOrString{
											IntVal: 8005,
										},
										ServiceName: cr.Name + "-gitlab",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Add Registry rule only when registry is enabled
	if cr.Spec.Registry.Enabled {
		ingress.Spec.Rules = append(ingress.Spec.Rules, extensionsv1beta1.IngressRule{
			Host: DomainNameOnly(cr.Spec.Registry.ExternalURL),
			IngressRuleValue: extensionsv1beta1.IngressRuleValue{
				HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
					Paths: []extensionsv1beta1.HTTPIngressPath{
						{
							Path: "/",
							Backend: extensionsv1beta1.IngressBackend{
								ServicePort: intstr.IntOrString{
									IntVal: 8105,
								},
								ServiceName: cr.Name + "-gitlab",
							},
						},
					},
				},
			},
		})
	}

	// Add TLS certificate if TLS secret is provided
	if cr.Spec.TLSCertificate != "" {
		ingress.Spec.TLS = []extensionsv1beta1.IngressTLS{
			{
				SecretName: cr.Spec.TLSCertificate,
				Hosts:      getExternalURLs(cr),
			},
		}
	}

	return
}

func getExternalURLs(cr *gitlabv1beta1.Gitlab) []string {
	var hosts []string
	if cr.Spec.ExternalURL != "" {
		hosts = append(hosts, DomainNameOnly(cr.Spec.ExternalURL))
	}

	if cr.Spec.Registry.ExternalURL != "" {
		hosts = append(hosts, DomainNameOnly(cr.Spec.Registry.ExternalURL))
	}

	return hosts
}
