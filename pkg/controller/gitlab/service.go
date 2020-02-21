package gitlab

import (
	gitlabv1beta1 "github.com/OchiengEd/gitlab-operator/pkg/apis/gitlab/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getGitlabService(cr *gitlabv1beta1.Gitlab) *corev1.Service {
	labels := getLabels(cr, "gitlab")

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels["app.kubernetes.io/name"],
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:     "ssh",
					Port:     22,
					Protocol: corev1.ProtocolTCP,
				},
				{
					Name:     "mattermost",
					Port:     8065,
					Protocol: corev1.ProtocolTCP,
				},
				{
					Name:     "registry",
					Port:     8105,
					Protocol: corev1.ProtocolTCP,
				},
				{
					Name:     "workhorse",
					Port:     8005,
					Protocol: corev1.ProtocolTCP,
				},
				{
					Name:     "prometheus",
					Port:     9090,
					Protocol: corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}

func getRedisService(cr *gitlabv1beta1.Gitlab) corev1.Service {
	labels := getLabels(cr, "redis")

	return corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels["app.kubernetes.io/name"],
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

func getPostgresService(cr *gitlabv1beta1.Gitlab) corev1.Service {
	labels := getLabels(cr, "database")

	return corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels["app.kubernetes.io/name"],
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
