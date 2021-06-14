package internal

import (
	"strings"

	routev1 "github.com/openshift/api/route/v1"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// EndpointAnnotations generates annotation for ingresses
func EndpointAnnotations(adapter gitlab.CustomResourceAdapter, annotate bool) map[string]string {
	annotation := map[string]string{
		"kubernetes.io/ingress.class": "nginx",
	}

	if annotate {
		annotation["cert-manager.io/issuer"] = adapter.ReleaseName() + "-issuer"
	}

	return annotation
}

// MainRoute returns main GitLab application route
func MainRoute(adapter gitlab.CustomResourceAdapter) *routev1.Route {
	labels := Label(adapter.ReleaseName(), "route", GitlabType)

	return &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:        adapter.ReleaseName() + "-gitlab",
			Namespace:   adapter.Namespace(),
			Labels:      labels,
			Annotations: EndpointAnnotations(adapter, RequiresCertManagerCertificate(adapter).GitLab()),
		},
		Spec: routev1.RouteSpec{
			Host: getGitlabURL(adapter),
			Path: "/",
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: adapter.ReleaseName() + "-webservice",
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.IntOrString{
					IntVal: 8181,
				},
			},
			TLS: getRouteTLSConfig(adapter, labels["app.kubernetes.io/component:"]),
		},
	}
}

// AdminRoute returns GitLab admin route
func AdminRoute(adapter gitlab.CustomResourceAdapter) *routev1.Route {
	labels := Label(adapter.ReleaseName(), "route", GitlabType)

	return &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:        adapter.ReleaseName() + "-admin",
			Namespace:   adapter.Namespace(),
			Labels:      labels,
			Annotations: EndpointAnnotations(adapter, RequiresCertManagerCertificate(adapter).GitLab()),
		},
		Spec: routev1.RouteSpec{
			Host: getGitlabURL(adapter),
			Path: "/admin/sidekiq",
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: adapter.ReleaseName() + "-webservice",
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
func RegistryRoute(adapter gitlab.CustomResourceAdapter) *routev1.Route {
	labels := Label(adapter.ReleaseName(), "route", GitlabType)

	return &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:        adapter.ReleaseName() + "-registry",
			Namespace:   adapter.Namespace(),
			Labels:      labels,
			Annotations: EndpointAnnotations(adapter, RequiresCertManagerCertificate(adapter).Registry()),
		},
		Spec: routev1.RouteSpec{
			Host: getRegistryURL(adapter),
			Path: "/",
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: adapter.ReleaseName() + "-registry",
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.IntOrString{
					IntVal: 5000,
				},
			},
		},
	}
}

func getRouteTLSConfig(adapter gitlab.CustomResourceAdapter, _ string) *routev1.TLSConfig {
	// Ignoring target, since we only support wildcard certificate at the moment.
	tlsSecretName := strings.Join([]string{adapter.ReleaseName(), "tls"}, "-")
	var tlsCert, tlsKey, tlsCACert string

	tlsData, err := SecretData(tlsSecretName, adapter.Namespace())
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
