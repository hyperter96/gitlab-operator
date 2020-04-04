package gitlab

import (
	"context"

	gitlabv1beta1 "gitlab.com/ochienged/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlabutils "gitlab.com/ochienged/gitlab-operator/pkg/controller/utils"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func getGitlabIngress(cr *gitlabv1beta1.Gitlab) (ingress *extensionsv1beta1.Ingress) {
	labels := gitlabutils.Label(cr.Name, "ingress", gitlabutils.GitlabType)

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
		},
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

func getRegistryIngress(cr *gitlabv1beta1.Gitlab) (ingress *extensionsv1beta1.Ingress) {
	labels := gitlabutils.Label(cr.Name, "ingress", gitlabutils.GitlabType)

	ingress = &extensionsv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-registry-ingress",
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
					// Add Registry rule only when registry is enabled
					Host: DomainNameOnly(cr.Spec.Registry.ExternalURL),
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
		},
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

func (r *ReconcileGitlab) reconcileGitlabIngress(cr *gitlabv1beta1.Gitlab) error {
	gitlab := getGitlabIngress(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: gitlab.Name}, gitlab) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, gitlab, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), gitlab)
}

func (r *ReconcileGitlab) reconcileRegistryIngress(cr *gitlabv1beta1.Gitlab) error {
	registry := getRegistryIngress(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: registry.Name}, registry) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, registry, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), registry)
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
