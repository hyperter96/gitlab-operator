package gitlab

import (
	"strings"

	routev1 "github.com/openshift/api/route/v1"
	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	gitlabutils "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// MainRoute returns main GitLab application route
func MainRoute(cr *gitlabv1beta1.GitLab) *routev1.Route {
	labels := gitlabutils.Label(cr.Name, "route", gitlabutils.GitlabType)

	return &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:        cr.Name + "-gitlab",
			Namespace:   cr.Namespace,
			Labels:      labels,
			Annotations: EndpointAnnotations(cr, RequiresCertManagerCertificate(cr).GitLab()),
		},
		Spec: routev1.RouteSpec{
			Host: getGitlabURL(cr),
			Path: "/",
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: cr.Name + "-webservice",
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.IntOrString{
					IntVal: 8181,
				},
			},
			TLS: getRouteTLSConfig(cr, labels["app.kubernetes.io/component:"]),
		},
	}
}

// AdminRoute returns GitLab admin route
func AdminRoute(cr *gitlabv1beta1.GitLab) *routev1.Route {
	labels := gitlabutils.Label(cr.Name, "route", gitlabutils.GitlabType)

	return &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:        cr.Name + "-admin",
			Namespace:   cr.Namespace,
			Labels:      labels,
			Annotations: EndpointAnnotations(cr, RequiresCertManagerCertificate(cr).GitLab()),
		},
		Spec: routev1.RouteSpec{
			Host: getGitlabURL(cr),
			Path: "/admin/sidekiq",
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: cr.Name + "-webservice",
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.IntOrString{
					IntVal: 8080,
				},
			},
		},
	}
}

// RegistryRoute returns GitLab registry route
func RegistryRoute(cr *gitlabv1beta1.GitLab) *routev1.Route {
	labels := gitlabutils.Label(cr.Name, "route", gitlabutils.GitlabType)

	return &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:        cr.Name + "-registry",
			Namespace:   cr.Namespace,
			Labels:      labels,
			Annotations: EndpointAnnotations(cr, RequiresCertManagerCertificate(cr).Registry()),
		},
		Spec: routev1.RouteSpec{
			Host: getRegistryURL(cr),
			Path: "/",
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: cr.Name + "-registry",
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.IntOrString{
					IntVal: 5000,
				},
			},
		},
	}
}

func getRouteTLSConfig(cr *gitlabv1beta1.GitLab, target string) *routev1.TLSConfig {
	tlsSecretName := strings.Join([]string{cr.Name, target, "tls"}, "-")
	var tlsCert, tlsKey, tlsCACert string

	tlsData, err := gitlabutils.SecretData(tlsSecretName, cr.Namespace)
	if err != nil {
		return nil
	}

	if crt, ok := tlsData["tls.crt"]; ok {
		tlsCert = crt
	}

	if key, ok := tlsData["tls.key"]; ok {
		tlsKey = key
	}

	if cacrt, ok := tlsData["ca.crt"]; ok {
		tlsCACert = cacrt
	}

	if tlsCert != "" && tlsKey != "" {
		return &routev1.TLSConfig{
			Termination:   routev1.TLSTerminationEdge,
			Certificate:   tlsCert,
			Key:           tlsKey,
			CACertificate: tlsCACert,
		}
	}

	return nil
}
