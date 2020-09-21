package gitlab

import (
	"strings"

	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	gitlabutils "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// RedisHeadlessService returns the headless service for Postgres statefulset
func RedisHeadlessService(cr *gitlabv1beta1.GitLab) *corev1.Service {
	labels := gitlabutils.Label(cr.Name, "redis", gitlabutils.GitlabType)

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      strings.Join([]string{labels["app.kubernetes.io/instance"], "headless"}, "-"),
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:     "redis",
					Port:     6379,
					Protocol: corev1.ProtocolTCP,
				},
			},
			Type:      corev1.ServiceTypeClusterIP,
			ClusterIP: corev1.ClusterIPNone,
		},
	}
}

// RedisService returns service to expose Redis
func RedisService(cr *gitlabv1beta1.GitLab) *corev1.Service {
	labels := gitlabutils.Label(cr.Name, "redis", gitlabutils.GitlabType)

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
					Name:     "redis",
					Port:     6379,
					Protocol: corev1.ProtocolTCP,
				},
				{
					Name:     "redis-metrics",
					Port:     9121,
					Protocol: corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}

// PostgresHeadlessService returns headless service for Postgresql statefulset
func PostgresHeadlessService(cr *gitlabv1beta1.GitLab) *corev1.Service {
	labels := gitlabutils.Label(cr.Name, "postgresql", gitlabutils.GitlabType)

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      strings.Join([]string{labels["app.kubernetes.io/instance"], "headless"}, "-"),
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:     "postgres",
					Port:     5432,
					Protocol: corev1.ProtocolTCP,
				},
			},
			Type:      corev1.ServiceTypeClusterIP,
			ClusterIP: corev1.ClusterIPNone,
		},
	}
}

// PostgresqlService returns Postgres service
func PostgresqlService(cr *gitlabv1beta1.GitLab) *corev1.Service {
	labels := gitlabutils.Label(cr.Name, "postgresql", gitlabutils.GitlabType)

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
					Name:     "postgres",
					Port:     5432,
					Protocol: corev1.ProtocolTCP,
				},
				{
					Name:     "postgres-metrics",
					Port:     9187,
					Protocol: corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}

// GitalyService returns service to expose the Gitaly service
func GitalyService(cr *gitlabv1beta1.GitLab) *corev1.Service {
	labels := gitlabutils.Label(cr.Name, "gitaly", gitlabutils.GitlabType)

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
					Name:     "gitaly",
					Port:     8075,
					Protocol: corev1.ProtocolTCP,
				},
				{
					Name:     "gitaly-metrics",
					Port:     9236,
					Protocol: corev1.ProtocolTCP,
				},
			},
			Type:      corev1.ServiceTypeClusterIP,
			ClusterIP: corev1.ClusterIPNone,
		},
	}
}

// RegistryService returns the service to expose GitLab container registry
func RegistryService(cr *gitlabv1beta1.GitLab) *corev1.Service {
	labels := gitlabutils.Label(cr.Name, "registry", gitlabutils.GitlabType)

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
					Name:     "registry",
					Port:     5000,
					Protocol: corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}

// WebserviceService returns service to expose Webservice
func WebserviceService(cr *gitlabv1beta1.GitLab) *corev1.Service {
	labels := gitlabutils.Label(cr.Name, "webservice", gitlabutils.GitlabType)

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
					Name:     "http-webservice",
					Port:     8080,
					Protocol: corev1.ProtocolTCP,
				},
				{
					Name:     "http-workhorse",
					Port:     8181,
					Protocol: corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}

// ShellService returns service to export GitLab shell
func ShellService(cr *gitlabv1beta1.GitLab) *corev1.Service {
	labels := gitlabutils.Label(cr.Name, "shell", gitlabutils.GitlabType)

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
					Name: "ssh",
					Port: 22,
					TargetPort: intstr.IntOrString{
						IntVal: 2222,
					},
					Protocol: corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeNodePort,
		},
	}
}

// ExporterService returns service that exposes GitLab exporter
func ExporterService(cr *gitlabv1beta1.GitLab) *corev1.Service {
	labels := gitlabutils.Label(cr.Name, "gitlab-exporter", gitlabutils.GitlabType)

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
					Name:     "gitlab-exporter",
					Port:     9168,
					Protocol: corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}
