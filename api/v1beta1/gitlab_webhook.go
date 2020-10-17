/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	"errors"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var gitlablog = logf.Log.WithName("gitlab-resource")

// SetupWebhookWithManager adds webhook to the controller runtime Manager
func (r *GitLab) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-apps-gitlab-com-v1beta1-gitlab,mutating=false,failurePolicy=fail,groups=apps.gitlab.com,resources=gitlabs,versions=v1beta1,name=vgitlab.kb.io

var _ webhook.Validator = &GitLab{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *GitLab) ValidateCreate() error {
	gitlablog.Info("validate create", "name", r.Name)

	if r.Spec.URL == "" {
		return errors.New("spec.url is empty")
	}

	if r.Spec.AutoScaling.MinReplicas != nil &&
		*r.Spec.AutoScaling.MinReplicas > r.Spec.AutoScaling.MaxReplicas {
		return errors.New("minReplicas can not be greater than maxReplicas in spec.autoscaling")
	}

	if r.Spec.CertIssuer != nil && r.Spec.CertIssuer.Email == "" {
		return errors.New("spec.acme.email is empty")
	}

	if !r.Spec.Registry.Disabled && r.Spec.Registry.URL == "" {
		return errors.New("spec.registry.url is empty")
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *GitLab) ValidateUpdate(old runtime.Object) error {
	gitlablog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *GitLab) ValidateDelete() error {
	gitlablog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}