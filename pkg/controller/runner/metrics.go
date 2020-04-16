package runner

import (
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	gitlabv1beta1 "gitlab.com/ochienged/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlabutils "gitlab.com/ochienged/gitlab-operator/pkg/controller/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func getRunnerMetricsService(cr *gitlabv1beta1.Runner) *corev1.Service {
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

func getRunnerServiceMonitorService(cr *gitlabv1beta1.Runner) *monitoringv1.ServiceMonitor {
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

func (r *ReconcileRunner) reconcileRunnerMetrics(cr *gitlabv1beta1.Runner) error {
	svc := getRunnerMetricsService(cr)

	if err := r.createKubernetesResource(cr, svc); err != nil {
		return err
	}

	if gitlabutils.IsPrometheusSupported() {
		sm := getRunnerServiceMonitorService(cr)

		if err := r.createKubernetesResource(cr, sm); err != nil {
			return err
		}
	}

	return nil
}
