package gitlab

import (
	"context"

	gitlabv1beta1 "gitlab.com/ochienged/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlabutils "gitlab.com/ochienged/gitlab-operator/pkg/controller/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func objectNamespacedName(obj interface{}) types.NamespacedName {
	object := obj.(metav1.Object)
	return types.NamespacedName{Name: object.GetName(), Namespace: object.GetNamespace()}
}

func (r *ReconcileGitlab) createKubernetesResource(object interface{}, parent *gitlabv1beta1.Gitlab) error {

	if gitlabutils.IsObjectFound(r.client, objectNamespacedName(object), object.(runtime.Object)) {
		return nil
	}

	// If parent resource is nil, not owner reference will be set
	if parent != nil {
		if err := controllerutil.SetControllerReference(parent, object.(metav1.Object), r.scheme); err != nil {
			return err
		}
	}

	return r.client.Create(context.TODO(), object.(runtime.Object))
}

func (r *ReconcileGitlab) maskEmailPasword(cr *gitlabv1beta1.Gitlab) error {
	gitlab := &gitlabv1beta1.Gitlab{}
	r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, gitlab)

	// If password is stored in secret and is still visible in CR, update it to emty string
	emailPasswd, err := gitlabutils.GetSecretValue(r.client, cr.Namespace, cr.Name+"-smtp-settings-secret", "smtp_user_password")
	if err != nil {
		log.Error(err, "")
	}

	if gitlab.Spec.SMTP.Password == emailPasswd && cr.Spec.SMTP.Password != "" {
		// Update CR
		gitlab.Spec.SMTP.Password = ""
		if err := r.client.Update(context.TODO(), gitlab); err != nil && errors.IsResourceExpired(err) {
			return err
		}
	}

	// If stored password does not match the CR password,
	// update the secret and empty the password string in Gitlab CR

	return nil
}

func (r *ReconcileGitlab) reconcileDeployments(cr *gitlabv1beta1.Gitlab) error {

	if err := r.reconcileWebserviceDeployment(cr); err != nil {
		return err
	}

	if err := r.reconcileShellDeployment(cr); err != nil {
		return err
	}

	if err := r.reconcileSidekiqDeployment(cr); err != nil {
		return err
	}

	if err := r.reconcileRegistryDeployment(cr); err != nil {
		return err
	}

	if err := r.reconcileTaskRunnerDeployment(cr); err != nil {
		return err
	}

	if err := r.reconcileGitlabExporterDeployment(cr); err != nil {
		return err
	}

	return nil
}
