package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BackupSpec defines the desired state of Backup
type BackupSpec struct {
	// Instance represents the GitLab instance to backup
	Instance string `json:"instance"`

	// Schedule defines the time and day to run backup
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

// BackupStatus defines the observed state of Backup
type BackupStatus struct {
	// +kubebuilder:validation:Enum=Running;Completed;Scheduled;Failed
	Phase BackupState `json:"phase,omitempty"`

	// StartedAt returns time when backup was initiated
	StartedAt string `json:"startedAt,omitempty"`

	// Completed returns time backup terminated or completed
	CompletedAt string `json:"completedAt,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Backup is the Schema for the backups API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=backups,scope=Namespaced
type Backup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BackupSpec   `json:"spec,omitempty"`
	Status BackupStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BackupList contains a list of Backup
type BackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Backup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Backup{}, &BackupList{})
}
