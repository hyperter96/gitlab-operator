package gitlab

import (
	routev1 "github.com/openshift/api/route/v1"
	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	gitlabutils "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func getGitlabRoute(cr *gitlabv1beta1.GitLab) *routev1.Route {
	labels := gitlabutils.Label(cr.Name, "route", gitlabutils.GitlabType)

	return &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-gitlab-route",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: routev1.RouteSpec{
			Path: "/",
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: cr.Name + "-gitlab",
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.IntOrString{
					IntVal: 8005,
				},
			},
		},
	}
}

// Gitlab registry route
func getRegistryRoute(cr *gitlabv1beta1.GitLab) *routev1.Route {
	labels := gitlabutils.Label(cr.Name, "route", gitlabutils.GitlabType)

	return &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-registry-route",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: routev1.RouteSpec{
			Path: "/",
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: cr.Name + "-gitlab",
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.IntOrString{
					IntVal: 8105,
				},
			},
		},
	}
}

// SSH route
func getSecureShellRoute(cr *gitlabv1beta1.GitLab) *routev1.Route {
	labels := gitlabutils.Label(cr.Name, "route", gitlabutils.GitlabType)

	return &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-ssh-route",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: routev1.RouteSpec{
			Path: "/",
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: cr.Name + "-gitlab",
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.IntOrString{
					IntVal: 22,
				},
			},
		},
	}
}
