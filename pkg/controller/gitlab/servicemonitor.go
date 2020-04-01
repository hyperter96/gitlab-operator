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

func (r *ReconcileGitlab) reconcileGitlabMetricsServiceMonitor(cr *gitlabv1beta1.Gitlab) error {

	gitlab := getGitlabMetricsServiceMonitor(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: gitlab.Name}, gitlab) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, gitlab, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), gitlab)
}

func (r *ReconcileGitlab) reconcilePostgresMetricsServiceMonitor(cr *gitlabv1beta1.Gitlab) error {

	postgres := getPostgresMetricsServiceMonitor(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: postgres.Name}, postgres) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, postgres, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), postgres)
}

func (r *ReconcileGitlab) reconcileRedisMetricsServiceMonitor(cr *gitlabv1beta1.Gitlab) error {

	redis := getRedisMetricsServiceMonitor(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: redis.Name}, redis) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, redis, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), redis)
}
