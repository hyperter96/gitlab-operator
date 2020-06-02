package gitlab

import (
	nginxv1alpha1 "github.com/nginxinc/nginx-ingress-operator/pkg/apis/k8s/v1alpha1"
	gitlabv1beta1 "gitlab.com/ochienged/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlabutils "gitlab.com/ochienged/gitlab-operator/pkg/controller/utils"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// IngressAnnotations generates annotation for ingresses
func IngressAnnotations(cr *gitlabv1beta1.Gitlab, annotate bool) map[string]string {
	annotation := map[string]string{
		"kubernetes.io/ingress.class": "nginx",
	}

	if annotate {
		annotation["cert-manager.io/issuer"] = cr.Name + "-issuer"
	}

	return annotation
}

func getIngressController(cr *gitlabv1beta1.Gitlab) *nginxv1alpha1.NginxIngressController {
	labels := gitlabutils.Label(cr.Name, "ingress-controller", gitlabutils.GitlabType)

	var replicas int32 = 1

	return &nginxv1alpha1.NginxIngressController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gitlab-ingress-controller",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: nginxv1alpha1.NginxIngressControllerSpec{
			EnableCRDs: true,
			Image: nginxv1alpha1.Image{
				Repository: "docker.io/nginx/nginx-ingress",
				Tag:        "1.6.3-ubi",
				PullPolicy: "Always",
			},
			// IngressClass: "gitlab",
			// UseIngressClassOnly: true,
			NginxPlus:   false,
			Replicas:    &replicas,
			ServiceType: "NodePort",
			Type:        "deployment",
		},
	}
}

func getGitlabIngress(cr *gitlabv1beta1.Gitlab) *extensionsv1beta1.Ingress {
	labels := gitlabutils.Label(cr.Name, "ingress", gitlabutils.GitlabType)

	return &extensionsv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        cr.Name + "-gitlab-ingress",
			Namespace:   cr.Namespace,
			Labels:      labels,
			Annotations: IngressAnnotations(cr, RequiresCertManagerCertificate(cr).GitLab()),
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
										ServiceName: cr.Name + "-webservice",
									},
								},
								{
									Path: "/admin/sidekiq",
									Backend: extensionsv1beta1.IngressBackend{
										ServicePort: intstr.IntOrString{
											IntVal: 8080,
										},
										ServiceName: cr.Name + "-webservice",
									},
								},
							},
						},
					},
				},
			},
			TLS: getGitlabIngressCert(cr),
		},
	}
}

func getRegistryIngress(cr *gitlabv1beta1.Gitlab) *extensionsv1beta1.Ingress {
	labels := gitlabutils.Label(cr.Name, "ingress", gitlabutils.GitlabType)

	return &extensionsv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        cr.Name + "-registry-ingress",
			Namespace:   cr.Namespace,
			Labels:      labels,
			Annotations: IngressAnnotations(cr, RequiresCertManagerCertificate(cr).Registry()),
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
			TLS: getRegistryIngressCert(cr),
		},
	}
}

func getGitlabIngressCert(cr *gitlabv1beta1.Gitlab) []extensionsv1beta1.IngressTLS {
	if RequiresCertManagerCertificate(cr).GitLab() {
		return []extensionsv1beta1.IngressTLS{
			{
				Hosts:      []string{getGitlabURL(cr)},
				SecretName: cr.Name + "-gitlab-tls",
			},
		}
	}

	return []extensionsv1beta1.IngressTLS{
		{
			Hosts:      []string{getGitlabURL(cr)},
			SecretName: cr.Spec.TLS,
		},
	}
}

func getRegistryIngressCert(cr *gitlabv1beta1.Gitlab) []extensionsv1beta1.IngressTLS {

	if RequiresCertManagerCertificate(cr).Registry() {
		return []extensionsv1beta1.IngressTLS{
			{
				Hosts:      []string{getRegistryURL(cr)},
				SecretName: cr.Name + "-registry-tls",
			},
		}
	}

	return []extensionsv1beta1.IngressTLS{
		{
			Hosts:      []string{getRegistryURL(cr)},
			SecretName: cr.Spec.Registry.TLS,
		},
	}
}

func (r *ReconcileGitlab) reconcileIngress(cr *gitlabv1beta1.Gitlab) error {
	controller := getIngressController(cr)
	if err := r.createKubernetesResource(controller, nil); err != nil {
		return err
	}

	var ingresses []*extensionsv1beta1.Ingress
	gitlab := getGitlabIngress(cr)

	registry := getRegistryIngress(cr)

	ingresses = append(ingresses,
		gitlab,
		registry,
	)

	for _, ingress := range ingresses {
		if err := r.createKubernetesResource(ingress, cr); err != nil {
			return err
		}
	}

	return nil
}
