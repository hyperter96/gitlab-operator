package runner

import (
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	gitlabutils "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// MetricsService returns the kubernetes service object for metrics
func MetricsService(cr *gitlabv1beta1.Runner) *corev1.Service {
	labels := gitlabutils.Label(cr.Name, "runner", gitlabutils.RunnerType)

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels["app.kubernetes.io/instance"] + "-metrics",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:     "metrics",
					Protocol: corev1.ProtocolTCP,
					Port:     9252,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}

// ServiceMonitorService returns the prometheus service monitor object
func ServiceMonitorService(cr *gitlabv1beta1.Runner) *monitoringv1.ServiceMonitor {
	labels := gitlabutils.Label(cr.Name, "runner", gitlabutils.RunnerType)

	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels["app.kubernetes.io/instance"] + "-metrics",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: labels,
			},
			Endpoints: []monitoringv1.Endpoint{
				{
					Path: "/metrics",
					TargetPort: &intstr.IntOrString{
						IntVal: 9252,
					},
				},
			},
		},
	}
}
