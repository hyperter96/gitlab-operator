package gitlab

import (
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	gitlabv1beta1 "gitlab.com/ochienged/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlabutils "gitlab.com/ochienged/gitlab-operator/pkg/controller/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func getGitlabMetricsServiceMonitor(cr *gitlabv1beta1.Gitlab) *monitoringv1.ServiceMonitor {
	labels := gitlabutils.Label(cr.Name, "gitlab-exporter", gitlabutils.GitlabType)

	serviceLabels := labels
	serviceLabels["subsystem"] = "metrics"

	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels["app.kubernetes.io/instance"],
			Namespace: cr.Namespace,
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: serviceLabels,
			},
			Endpoints: []monitoringv1.Endpoint{
				{
					Path: "/metrics",
					TargetPort: &intstr.IntOrString{
						IntVal: 9168,
					},
				},
			},
		},
	}

}

func getPostgresMetricsServiceMonitor(cr *gitlabv1beta1.Gitlab) *monitoringv1.ServiceMonitor {
	labels := gitlabutils.Label(cr.Name, "database", gitlabutils.GitlabType)

	serviceLabels := labels
	serviceLabels["subsystem"] = "metrics"

	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels["app.kubernetes.io/instance"],
			Namespace: cr.Namespace,
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: serviceLabels,
			},
			Endpoints: []monitoringv1.Endpoint{
				{
					Path: "/metrics",
					TargetPort: &intstr.IntOrString{
						IntVal: 9187,
					},
				},
			},
		},
	}

}

func getRedisMetricsServiceMonitor(cr *gitlabv1beta1.Gitlab) *monitoringv1.ServiceMonitor {
	labels := gitlabutils.Label(cr.Name, "redis", gitlabutils.GitlabType)

	serviceLabels := labels
	serviceLabels["subsystem"] = "metrics"

	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels["app.kubernetes.io/instance"],
			Namespace: cr.Namespace,
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: serviceLabels,
			},
			Endpoints: []monitoringv1.Endpoint{
				{
					Path: "/metrics",
					TargetPort: &intstr.IntOrString{
						IntVal: 9121,
					},
				},
			},
		},
	}

}

func (r *ReconcileGitlab) reconcileServiceMonitor(cr *gitlabv1beta1.Gitlab) error {

	var servicemonitors []*monitoringv1.ServiceMonitor

	gitlab := getGitlabMetricsServiceMonitor(cr)

	postgres := getPostgresMetricsServiceMonitor(cr)

	redis := getRedisMetricsServiceMonitor(cr)

	servicemonitors = append(servicemonitors,
		gitlab,
		postgres,
		redis,
	)

	for _, sm := range servicemonitors {
		if err := r.createKubernetesResource(cr, sm); err != nil {
			return err
		}
	}

	return nil
}
