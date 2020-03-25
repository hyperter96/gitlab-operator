package gitlab

import (
	"context"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	gitlabv1beta1 "gitlab.com/ochienged/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlabutils "gitlab.com/ochienged/gitlab-operator/pkg/controller/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func getServiceMonitor(cr *gitlabv1beta1.Gitlab) *monitoringv1.ServiceMonitor {
	labels := gitlabutils.Label(cr.Name, "metrics", gitlabutils.GitlabType)

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

func (r *ReconcileGitlab) reconcilePrometheusServiceMonitor(cr *gitlabv1beta1.Gitlab) error {

	servicemon := getServiceMonitor(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: servicemon.Name}, servicemon) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, servicemon, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), servicemon)
}
