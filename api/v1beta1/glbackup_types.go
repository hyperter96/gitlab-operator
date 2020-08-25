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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GLBackupSpec defines the desired state of GLBackup
type GLBackupSpec struct {
	// Instance represents the GitLab instance to backup
	Instance string `json:"instance"`

	// Schedule defines the time and day to run backup
	// It takes cron time format
	Schedule string `json:"schedule,omitempty"`

	// Exclusions allows user to exclude components to backup
	Exclusions string `json:"skip,omitempty"`

	// Timestamp defines the prefix of the backup job
	Timestamp string `json:"timestamp,omitempty"`

	// URL defines the address of the backup job
	URL string `json:"url,omitempty"`

	// Restore when set to true the backup defined by
	// ID: will be restored to the gitlab instance
	Restore bool `json:"restore,omitempty"`
}

// BackupState informs of current backup state
type BackupState string

const (
	// BackupRunning indicates backup job is in progress
	BackupRunning BackupState = "Running"

	// BackupCompleted indicated backup finished successfully
	BackupCompleted BackupState = "Completed"

	// BackupFailed indicates an error disrupted backup process
	BackupFailed BackupState = "Failed"

	// BackupScheduled indicates backup will run at future time
	BackupScheduled BackupState = "Scheduled"
)

// GLBackupStatus defines the observed state of GLBackup
type GLBackupStatus struct {
	// +kubebuilder:validation:Enum=Running;Completed;Scheduled;Failed
	Phase BackupState `json:"phase,omitempty"`

	// StartedAt returns time when backup was initiated
	StartedAt string `json:"startedAt,omitempty"`

	// Completed returns time backup terminated or completed
	CompletedAt string `json:"completedAt,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// GLBackup is the Schema for the glbackups API
type GLBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GLBackupSpec   `json:"spec,omitempty"`
	Status GLBackupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GLBackupList contains a list of GLBackup
type GLBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GLBackup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GLBackup{}, &GLBackupList{})
}
