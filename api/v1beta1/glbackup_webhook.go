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
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var glbackuplog = logf.Log.WithName("glbackup-resource")

// SetupWebhookWithManager adds webhook to the controller runtime manager
func (r *GLBackup) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-apps-gitlab-com-v1beta1-glbackup,mutating=true,failurePolicy=fail,groups=apps.gitlab.com,resources=glbackups,verbs=create;update,versions=v1beta1,name=mglbackup.kb.io

var _ webhook.Defaulter = &GLBackup{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *GLBackup) Default() {
	glbackuplog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-apps-gitlab-com-v1beta1-glbackup,mutating=false,failurePolicy=fail,groups=apps.gitlab.com,resources=glbackups,versions=v1beta1,name=vglbackup.kb.io

var _ webhook.Validator = &GLBackup{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *GLBackup) ValidateCreate() error {
	glbackuplog.Info("validate create", "name", r.Name)

	if r.Spec.Instance == "" {
		return errors.New("spec.instance is empty")
	}

	if r.Spec.Schedule != "" && len(strings.Split(r.Spec.Schedule, " ")) != 5 {
		return errors.New("spec.schedule is invalid cron format")
	}

	if r.Spec.Restore {
		if r.Spec.Timestamp == "" && r.Spec.URL == "" {
			return errors.New("either spec.timestamp or spec.url should be specified")
		}
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *GLBackup) ValidateUpdate(old runtime.Object) error {
	glbackuplog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *GLBackup) ValidateDelete() error {
	glbackuplog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
