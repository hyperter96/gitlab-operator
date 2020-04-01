package gitlab

import (
	"context"
	"strings"

	gitlabv1beta1 "gitlab.com/ochienged/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlabutils "gitlab.com/ochienged/gitlab-operator/pkg/controller/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func getRedisHeadlessService(cr *gitlabv1beta1.Gitlab) *corev1.Service {
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

func getRedisService(cr *gitlabv1beta1.Gitlab) *corev1.Service {
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
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}

func getPostgresHeadlessService(cr *gitlabv1beta1.Gitlab) *corev1.Service {
	labels := gitlabutils.Label(cr.Name, "database", gitlabutils.GitlabType)

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

func getPostgresService(cr *gitlabv1beta1.Gitlab) *corev1.Service {
	labels := gitlabutils.Label(cr.Name, "database", gitlabutils.GitlabType)

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
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}

func getGitalyService(cr *gitlabv1beta1.Gitlab) *corev1.Service {
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

func getRegistryService(cr *gitlabv1beta1.Gitlab) *corev1.Service {
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

func getUnicornService(cr *gitlabv1beta1.Gitlab) *corev1.Service {
	labels := gitlabutils.Label(cr.Name, "unicorn", gitlabutils.GitlabType)

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
					Name:     "http-unicorn",
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

func getShellService(cr *gitlabv1beta1.Gitlab) *corev1.Service {
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
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}

func getGitlabExporterService(cr *gitlabv1beta1.Gitlab) *corev1.Service {
	labels := gitlabutils.Label(cr.Name, "gitlab-exporter", gitlabutils.GitlabType)
	labels["subsystem"] = "metrics"

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

func getRedisMetricsService(cr *gitlabv1beta1.Gitlab) *corev1.Service {
	labels := gitlabutils.Label(cr.Name, "redis", gitlabutils.GitlabType)
	labels["subsystem"] = "metrics"

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
					Port:     9121,
					Protocol: corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}

func getPostgresMetricsService(cr *gitlabv1beta1.Gitlab) *corev1.Service {
	labels := gitlabutils.Label(cr.Name, "database", gitlabutils.GitlabType)
	labels["subsystem"] = "metrics"

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
					Port:     9187,
					Protocol: corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}

func (r *ReconcileGitlab) reconcilePostgresService(cr *gitlabv1beta1.Gitlab) error {
	postgres := getPostgresService(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: postgres.Name}, postgres) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, postgres, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), postgres)
}

func (r *ReconcileGitlab) reconcilePostgresHeadlessService(cr *gitlabv1beta1.Gitlab) error {
	postgres := getPostgresHeadlessService(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: postgres.Name}, postgres) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, postgres, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), postgres)
}

func (r *ReconcileGitlab) reconcileRedisService(cr *gitlabv1beta1.Gitlab) error {
	redis := getRedisService(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: redis.Name}, redis) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, redis, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), redis)
}

func (r *ReconcileGitlab) reconcileRedisHeadlessService(cr *gitlabv1beta1.Gitlab) error {
	redis := getRedisHeadlessService(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: redis.Name}, redis) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, redis, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), redis)
}

func (r *ReconcileGitlab) reconcileGitalyService(cr *gitlabv1beta1.Gitlab) error {
	gitaly := getGitalyService(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: gitaly.Name}, gitaly) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, gitaly, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), gitaly)
}

func (r *ReconcileGitlab) reconcileRegistryService(cr *gitlabv1beta1.Gitlab) error {
	registry := getRegistryService(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: registry.Name}, registry) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, registry, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), registry)
}

func (r *ReconcileGitlab) reconcileUnicornService(cr *gitlabv1beta1.Gitlab) error {
	unicorn := getUnicornService(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: unicorn.Name}, unicorn) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, unicorn, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), unicorn)
}

func (r *ReconcileGitlab) reconcileShellService(cr *gitlabv1beta1.Gitlab) error {
	shell := getShellService(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: shell.Name}, shell) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, shell, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), shell)
}

func (r *ReconcileGitlab) reconcileGitlabExporterService(cr *gitlabv1beta1.Gitlab) error {
	exporter := getGitlabExporterService(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: exporter.Name}, exporter) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, exporter, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), exporter)
}

func (r *ReconcileGitlab) reconcilePostgresMetricsService(cr *gitlabv1beta1.Gitlab) error {
	postgres := getPostgresMetricsService(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: postgres.Name}, postgres) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, postgres, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), postgres)
}

func (r *ReconcileGitlab) reconcileRedisMetricsService(cr *gitlabv1beta1.Gitlab) error {
	redis := getRedisMetricsService(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: redis.Name}, redis) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, redis, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), redis)
}
