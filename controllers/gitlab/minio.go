package gitlab

import (
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/settings"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
)

const (
	globalMinioEnabled = "global.minio.enabled"
)

// MinioEnabled returns `true` if enabled, and `false` if not.
func MinioEnabled(adapter CustomResourceAdapter) bool {
	return adapter.Values().GetBool(globalMinioEnabled)
}

// MinioJob returns the Job of the Minio component.
func MinioJob(adapter CustomResourceAdapter, template helm.Template) client.Object {
	obj := template.Query().ObjectByKindAndComponent(JobKind, MinioComponentName)

	// Set ServiceAccountName and SecurityContext, as the Helm Chart does not currently
	// support setting them.
	// https://gitlab.com/gitlab-org/charts/gitlab/-/issues/3192
	var rootUser int64

	job := obj.(*batchv1.Job)
	job.Spec.Template.Spec.ServiceAccountName = settings.AppAnyUIDServiceAccount
	job.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{
		RunAsUser: &rootUser,
		FSGroup:   &rootUser,
	}

	return job
}

// MinioDeployment returns the Deployment of the Minio component.
func MinioDeployment(adapter CustomResourceAdapter, template helm.Template) client.Object {
	obj := template.Query().ObjectByKindAndComponent(DeploymentKind, MinioComponentName)

	// Set ServiceAccountName, as the Helm Chart does not currently support setting it.
	// https://gitlab.com/gitlab-org/charts/gitlab/-/issues/3192
	deployment := obj.(*appsv1.Deployment)
	deployment.Spec.Template.Spec.ServiceAccountName = settings.AppNonRootServiceAccount

	return deployment
}

// MinioConfigMap returns the ConfigMap of the Minio component.
func MinioConfigMap(adapter CustomResourceAdapter, template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(ConfigMapKind, MinioComponentName)
}

// MinioIngress returns the Ingress of the Minio component.
func MinioIngress(adapter CustomResourceAdapter, template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(IngressKind, MinioComponentName)
}

// MinioService returns the Service of the Minio component.
func MinioService(adapter CustomResourceAdapter, template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(ServiceKind, MinioComponentName)
}

// MinioPersistentVolumeClaim returns the PersistentVolumeClaim of the Minio component.
func MinioPersistentVolumeClaim(adapter CustomResourceAdapter, template helm.Template) client.Object {
	return template.Query().ObjectByKindAndComponent(PersistentVolumeClaimKind, MinioComponentName)
}
