package gitlab

import (
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	gitlabv1beta1 "gitlab.com/ochienged/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlabutils "gitlab.com/ochienged/gitlab-operator/pkg/controller/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// TODO: receive replicas from user input
func createPrometheusCluster(cr *gitlabv1beta1.Gitlab) *monitoringv1.Prometheus {
	labels := gitlabutils.Label(cr.Name, "prometheus", gitlabutils.GitlabType)

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

func exposePrometheusCluster(cr *gitlabv1beta1.Gitlab) *corev1.Service {
	labels := gitlabutils.Label(cr.Name, "prometheus", gitlabutils.GitlabType)

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
					Name:     "web",
					Port:     9090,
					Protocol: corev1.ProtocolTCP,
				},
			},
		},
	}
}

func getGitlabExporterServiceMonitor(cr *gitlabv1beta1.Gitlab) *monitoringv1.ServiceMonitor {
	labels := gitlabutils.Label(cr.Name, "gitlab-exporter", gitlabutils.GitlabType)

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

func getUnicornServiceMonitor(cr *gitlabv1beta1.Gitlab) *monitoringv1.ServiceMonitor {
	labels := gitlabutils.Label(cr.Name, "unicorn", gitlabutils.GitlabType)

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

func getPostgresMetricsServiceMonitor(cr *gitlabv1beta1.Gitlab) *monitoringv1.ServiceMonitor {
	labels := gitlabutils.Label(cr.Name, "postgresql", gitlabutils.GitlabType)

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

func getRedisServiceMonitor(cr *gitlabv1beta1.Gitlab) *monitoringv1.ServiceMonitor {
	labels := gitlabutils.Label(cr.Name, "redis", gitlabutils.GitlabType)

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

func getGitalyServiceMonitor(cr *gitlabv1beta1.Gitlab) *monitoringv1.ServiceMonitor {
	labels := gitlabutils.Label(cr.Name, "gitaly", gitlabutils.GitlabType)

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

func (r *ReconcileGitlab) reconcileServiceMonitor(cr *gitlabv1beta1.Gitlab) error {
	var servicemonitors []*monitoringv1.ServiceMonitor

	gitaly := getGitalyServiceMonitor(cr)

	gitlab := getGitlabExporterServiceMonitor(cr)

	postgres := getPostgresMetricsServiceMonitor(cr)

	redis := getRedisServiceMonitor(cr)

	workhorse := getUnicornServiceMonitor(cr)

	servicemonitors = append(servicemonitors,
		gitlab,
		gitaly,
		postgres,
		redis,
		workhorse,
	)

	for _, sm := range servicemonitors {
		if err := r.createKubernetesResource(sm, cr); err != nil {
			return err
		}
	}

	service := exposePrometheusCluster(cr)
	if err := r.createKubernetesResource(service, nil); err != nil {
		return err
	}

	prometheus := createPrometheusCluster(cr)
	return r.createKubernetesResource(prometheus, nil)
}
