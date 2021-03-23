package gitlab

import (
	nginxv1alpha1 "github.com/nginxinc/nginx-ingress-operator/pkg/apis/k8s/v1alpha1"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/helpers"
	gitlabutils "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/utils"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// EndpointAnnotations generates annotation for ingresses
func EndpointAnnotations(adapter helpers.CustomResourceAdapter, annotate bool) map[string]string {
	annotation := map[string]string{
		"kubernetes.io/ingress.class": "nginx",
	}

	if annotate {
		annotation["cert-manager.io/issuer"] = adapter.ReleaseName() + "-issuer"
	}

	return annotation
}

// IngressDEPRECATED returns Ingress object used for GitLab
func IngressDEPRECATED(adapter helpers.CustomResourceAdapter) *extensionsv1beta1.Ingress {
	labels := gitlabutils.Label(adapter.ReleaseName(), "ingress", gitlabutils.GitlabType)

	return &extensionsv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        adapter.ReleaseName() + "-gitlab-ingress",
			Namespace:   adapter.Namespace(),
			Labels:      labels,
			Annotations: EndpointAnnotations(adapter, RequiresCertManagerCertificate(adapter).GitLab()),
		},
		Spec: extensionsv1beta1.IngressSpec{
			Rules: []extensionsv1beta1.IngressRule{
				{
					// External URL for the gitlab instance
					Host: getGitlabURL(adapter),
					IngressRuleValue: extensionsv1beta1.IngressRuleValue{
						HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
							Paths: []extensionsv1beta1.HTTPIngressPath{
								{
									Path: "/",
									Backend: extensionsv1beta1.IngressBackend{
										ServicePort: intstr.IntOrString{
											IntVal: 8181,
										},
										ServiceName: adapter.ReleaseName() + "-webservice-default",
									},
								},
								{
									Path: "/admin/sidekiq",
									Backend: extensionsv1beta1.IngressBackend{
										ServicePort: intstr.IntOrString{
											IntVal: 8080,
										},
										ServiceName: adapter.ReleaseName() + "-webservice-default",
									},
								},
							},
						},
					},
				},
			},
			TLS: getGitlabIngressCert(adapter),
		},
	}
}

// RegistryIngressDEPRECATED returns ingress object for GitLab registry
func RegistryIngressDEPRECATED(adapter helpers.CustomResourceAdapter) *extensionsv1beta1.Ingress {
	labels := gitlabutils.Label(adapter.ReleaseName(), "ingress", gitlabutils.GitlabType)

	return &extensionsv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        adapter.ReleaseName() + "-registry-ingress",
			Namespace:   adapter.Namespace(),
			Labels:      labels,
			Annotations: EndpointAnnotations(adapter, RequiresCertManagerCertificate(adapter).Registry()),
		},
		Spec: extensionsv1beta1.IngressSpec{
			Rules: []extensionsv1beta1.IngressRule{
				{
					// Add Registry rule only when registry is enabled
					Host: getRegistryURL(adapter),
					IngressRuleValue: extensionsv1beta1.IngressRuleValue{
						HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
							Paths: []extensionsv1beta1.HTTPIngressPath{
								{
									Path: "/",
									Backend: extensionsv1beta1.IngressBackend{
										ServicePort: intstr.IntOrString{
											IntVal: 5000,
										},
										ServiceName: adapter.ReleaseName() + "-registry",
									},
								},
							},
						},
					},
				},
			},
			TLS: getRegistryIngressCert(adapter),
		},
	}
}

func getGitlabIngressCert(adapter helpers.CustomResourceAdapter) []extensionsv1beta1.IngressTLS {
	if RequiresCertManagerCertificate(adapter).GitLab() {
		return []extensionsv1beta1.IngressTLS{
			{
				Hosts:      []string{getGitlabURL(adapter)},
				SecretName: adapter.ReleaseName() + "-tls",
			},
		}
	}

	// This implies that Operator can only consume wildcard certificate and individual certificate
	// per service will be ignored.
	tlsSecretName, _ := helpers.GetStringValue(adapter.Values(), "global.ingress.tls.secretName")

	return []extensionsv1beta1.IngressTLS{
		{
			Hosts:      []string{getGitlabURL(adapter)},
			SecretName: tlsSecretName,
		},
	}
}

func getRegistryIngressCert(adapter helpers.CustomResourceAdapter) []extensionsv1beta1.IngressTLS {

	if RequiresCertManagerCertificate(adapter).Registry() {
		return []extensionsv1beta1.IngressTLS{
			{
				Hosts:      []string{getRegistryURL(adapter)},
				SecretName: adapter.ReleaseName() + "-tls",
			},
		}
	}

	// This implies that Operator can only consume wildcard certificate and individual certificate
	// per service will be ignored.
	tlsSecretName, _ := helpers.GetStringValue(adapter.Values(), "global.ingress.tls.secretName")

	return []extensionsv1beta1.IngressTLS{
		{
			Hosts:      []string{getRegistryURL(adapter)},
			SecretName: tlsSecretName,
		},
	}
}

// IngressController is a GitLab controller for exposing GitLab instances
func IngressController(adapter helpers.CustomResourceAdapter) *nginxv1alpha1.NginxIngressController {
	labels := gitlabutils.Label(adapter.ReleaseName(), "ingress-controller", gitlabutils.GitlabType)

	var replicas int32 = 1

	return &nginxv1alpha1.NginxIngressController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gitlab-ingress-controller",
			Namespace: adapter.Namespace(),
			Labels:    labels,
		},
		Spec: nginxv1alpha1.NginxIngressControllerSpec{
			EnableCRDs: true,
			Image: nginxv1alpha1.Image{
				Repository: "docker.io/nginx/nginx-ingress",
				Tag:        "1.10.1-ubi",
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
