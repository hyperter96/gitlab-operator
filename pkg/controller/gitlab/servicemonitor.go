package gitlab

import (
	gitlabv1beta1 "gitlab.com/ochienged/gitlab-operator/pkg/apis/gitlab/v1beta1"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func getServiceMonitor(cr *gitlabv1beta1.Gitlab) *monitoringv1.ServiceMonitor {
	labels := getLabels(cr, "metrics")

	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-gitlab-metrics",
			Namespace: cr.Namespace,
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: labels,
			},
			Endpoints: []monitoringv1.Endpoint{
				{
					Path: "/metrics",
					TargetPort: &intstr.IntOrString{
						IntVal: 9168,
					},
				},
				{
					Path: "/metrics",
					TargetPort: &intstr.IntOrString{
						IntVal: 9121,
					},
				},
				{
					Path: "/metrics",
					TargetPort: &intstr.IntOrString{
						IntVal: 9187,
					},
				},
				{
					Path: "/metrics",
					TargetPort: &intstr.IntOrString{
						IntVal: 9236,
					},
				},
				{
					Path: "/metrics",
					TargetPort: &intstr.IntOrString{
						IntVal: 9229,
					},
				},
			},
		},
	}

}
