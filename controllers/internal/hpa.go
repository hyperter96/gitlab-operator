package internal

import (
	gitlabv1beta1 "gitlab.com/gitlab-org/cloud-native/gitlab-operator/api/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
)

// HorizontalAutoscaler return HPA for deployment passed
func HorizontalAutoscaler(deployment *appsv1.Deployment, cr *gitlabv1beta1.GitLab) *autoscalingv1.HorizontalPodAutoscaler {
	// Since Operator-managed AutoScaling has been turned off temporarily, we can
	// safely comment the following and return nil.
	return nil

	/*
		deployment.ObjectMeta.Labels["app.kubernetes.io/component"] = "hpa"

		if cr.Spec.AutoScaling == nil {
			return nil
		}

		return &autoscalingv1.HorizontalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deployment.Name,
				Namespace: cr.Namespace,
				Labels:    labels,
			},
			Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscalingv1.CrossVersionObjectReference{
					Kind:       deployment.Kind,
					APIVersion: deployment.APIVersion,
					Name:       deployment.Name,
				},
				MinReplicas:                    cr.Spec.AutoScaling.MinReplicas,
				MaxReplicas:                    cr.Spec.AutoScaling.MaxReplicas,
				TargetCPUUtilizationPercentage: cr.Spec.AutoScaling.TargetCPU,
			},
		}
	*/
}
