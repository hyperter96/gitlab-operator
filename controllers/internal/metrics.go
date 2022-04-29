package internal

import (
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	gitlabv1beta1 "gitlab.com/gitlab-org/cloud-native/gitlab-operator/api/v1beta1"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
)

const (
	prometheusInstall = "prometheus.install"
)

// PrometheusEnabled returns `true` if Prometheus is enabled, and `false` if not.
func PrometheusClusterEnabled(adapter gitlab.CustomResourceAdapter) bool {
	return adapter.Values().GetBool(prometheusInstall)
}

// PrometheusCluster returns a prometheus cluster object.
func PrometheusCluster(cr *gitlabv1beta1.GitLab) *monitoringv1.Prometheus {
	labels := Label(cr.Name, "prometheus", GitlabType)

	var replicas int32 = 2

	return &monitoringv1.Prometheus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gitlab-prometheus",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: monitoringv1.PrometheusSpec{
			Alerting: &monitoringv1.AlertingSpec{
				Alertmanagers: []monitoringv1.AlertmanagerEndpoints{
					{
						Name:      "alertmanager-main",
						Namespace: cr.Namespace,
						Port:      intstr.FromString("web"),
					},
				},
			},
			ServiceAccountName:     "prometheus-k8s",
			Replicas:               &replicas,
			ServiceMonitorSelector: &metav1.LabelSelector{},
		},
	}
}

// ExposePrometheusCluster creates a service for Prometheus.
func ExposePrometheusCluster(cr *gitlabv1beta1.GitLab) *corev1.Service {
	labels := Label(cr.Name, "prometheus", GitlabType)

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gitlab-prometheus",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "prometheus",
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "web",
					Port:       9090,
					TargetPort: intstr.FromInt(9090),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}
}

// ExporterServiceMonitor returns the GitLab exporter service monitor.
func ExporterServiceMonitor(cr *gitlabv1beta1.GitLab) *monitoringv1.ServiceMonitor {
	labels := Label(cr.Name, "gitlab-exporter", GitlabType)

	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels["app.kubernetes.io/instance"],
			Namespace: cr.Namespace,
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: labels,
			},
			Endpoints: []monitoringv1.Endpoint{
				{
					Path:     "/metrics",
					Port:     "gitlab-exporter",
					Interval: "30s",
				},
			},
		},
	}
}

// WebserviceServiceMonitor returns the Webservice service monitor.
func WebserviceServiceMonitor(cr *gitlabv1beta1.GitLab) *monitoringv1.ServiceMonitor {
	labels := Label(cr.Name, "webservice", GitlabType)

	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels["app.kubernetes.io/instance"],
			Namespace: cr.Namespace,
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: labels,
			},
			Endpoints: []monitoringv1.Endpoint{
				{
					Path:     "/-/metrics",
					Port:     "http-workhorse",
					Interval: "30s",
				},
			},
		},
	}
}

// PostgresqlServiceMonitor returns the Postgres service monitor.
func PostgresqlServiceMonitor(cr *gitlabv1beta1.GitLab) *monitoringv1.ServiceMonitor {
	labels := Label(cr.Name, "postgresql", GitlabType)

	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels["app.kubernetes.io/instance"],
			Namespace: cr.Namespace,
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: labels,
			},
			Endpoints: []monitoringv1.Endpoint{
				{
					Path:     "/metrics",
					Port:     "postgres-metrics",
					Interval: "30s",
				},
			},
		},
	}
}

// RedisServiceMonitor returns the Redis service monitor.
func RedisServiceMonitor(cr *gitlabv1beta1.GitLab) *monitoringv1.ServiceMonitor {
	labels := Label(cr.Name, "redis", GitlabType)

	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels["app.kubernetes.io/instance"],
			Namespace: cr.Namespace,
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: labels,
			},
			Endpoints: []monitoringv1.Endpoint{
				{
					Path:     "/metrics",
					Port:     "redis-metrics",
					Interval: "30s",
				},
			},
		},
	}
}

// GitalyServiceMonitor returns the Gitaly service monitor.
func GitalyServiceMonitor(cr *gitlabv1beta1.GitLab) *monitoringv1.ServiceMonitor {
	labels := Label(cr.Name, "gitaly", GitlabType)

	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels["app.kubernetes.io/instance"],
			Namespace: cr.Namespace,
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: labels,
			},
			Endpoints: []monitoringv1.Endpoint{
				{
					Path:     "/metrics",
					Port:     "gitaly-metrics",
					Interval: "30s",
				},
			},
		},
	}
}
