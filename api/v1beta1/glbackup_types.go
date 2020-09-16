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
	// Name of GitLab instance to backup
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="GitLab Name",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	Instance string `json:"instance"`

	// Backup schedule in cron format.
	// Leave blank for one time on-demand backup
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Backup Schedule",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	Schedule string `json:"schedule,omitempty"`

	// Comma separated list of components to omit from backup
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Backup Exclusions",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	Exclusions string `json:"skip,omitempty"`

	// Prefix for the backup job
	// Can be used when restoring backup
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Backup Timestamp",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	Timestamp string `json:"timestamp,omitempty"`

	// The URL of the backup resource to be restored
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Backup URL",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	URL string `json:"url,omitempty"`

	// Restore when set to true the backup defined by
	// +operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Backup Restore",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
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
	// Reports status of backup task
	// +kubebuilder:validation:Enum=Running;Completed;Scheduled;Failed
	// +operator-sdk:csv:customresourcedefinitions:type=status,displayName="Backup Status",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	Phase BackupState `json:"phase,omitempty"`

	// Displays time the backup started
	// +operator-sdk:csv:customresourcedefinitions:type=status,displayName="Start Time",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	StartedAt string `json:"startedAt,omitempty"`

	// Displays time the backup completed
	// +operator-sdk:csv:customresourcedefinitions:type=status,displayName="Completion Time",xDescriptors="urn:alm:descriptor:com.tectonic.ui:text"
	CompletedAt string `json:"completedAt,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=gbk
// +kubebuilder:subresource:status

// +operator-sdk:csv:customresourcedefinitions:displayName="GitLab Backup"
// +operator-sdk:csv:customresourcedefinitions:resources={{Job,v1,""},{CronJob,v1beta1,""},{ConfigMap,v1,""}}

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
