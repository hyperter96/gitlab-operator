package gitlab

import (
	gitlabv1beta1 "github.com/OchiengEd/gitlab-operator/pkg/apis/gitlab/v1beta1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func getGitlabIngress(cr *gitlabv1beta1.Gitlab) *extensionsv1beta1.Ingress {
	labels := getLabels(cr, "ingress")

	return &extensionsv1beta1.Ingress{
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
					Host: "gitlab.baisikeli.me",
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
				{
					Host: "registry.baisikeli.me",
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
				},
				{
					Host: "mattermost.baisikeli.me",
					IngressRuleValue: extensionsv1beta1.IngressRuleValue{
						HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
							Paths: []extensionsv1beta1.HTTPIngressPath{
								{
									Path: "/",
									Backend: extensionsv1beta1.IngressBackend{
										ServicePort: intstr.IntOrString{
											IntVal: 8065,
										},
										ServiceName: cr.Name + "-gitlab",
									},
								},
							},
						},
					},
				},
				{
					Host: "prometheus.baisikeli.me",
					IngressRuleValue: extensionsv1beta1.IngressRuleValue{
						HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
							Paths: []extensionsv1beta1.HTTPIngressPath{
								{
									Path: "/",
									Backend: extensionsv1beta1.IngressBackend{
										ServicePort: intstr.IntOrString{
											IntVal: 9090,
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
}
