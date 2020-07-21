package backup

import (
	"strings"

	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlabutils "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/pkg/controller/utils"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

// IsOnDemandBackup returns true if no backup schedule is
// provided. This implies backup should run immediately
func IsOnDemandBackup(cr *gitlabv1beta1.Backup) bool {
	if len(strings.Split(cr.Spec.Schedule, " ")) == 5 {
		return false
	}

	return cr.Spec.Schedule == ""
}

func backupConfigMap(cr *gitlabv1beta1.Backup) *corev1.ConfigMap {
	labels := gitlabutils.Label(cr.Name, "backup-lock", gitlabutils.BackupType)

	return gitlabutils.GenericConfigMap(labels["app.kubernetes.io/instance"], cr.Namespace, labels)
}

// NewBackupSchedule returns a CronJob with schedule for backups
func NewBackupSchedule(cr *gitlabv1beta1.Backup) *batchv1beta1.CronJob {
	labels := gitlabutils.Label(cr.Name, "backup", gitlabutils.BackupType)

	backup := gitlabutils.GenericCronJob(gitlabutils.Component{
		Namespace: cr.Namespace,
		Labels:    labels,
		Containers: []corev1.Container{
			{
				Name:            "control",
				Image:           "registry.gitlab.com/ochienged/backup-control:latest",
				ImagePullPolicy: corev1.PullIfNotPresent,
				Env: []corev1.EnvVar{
					{
						Name:  "GITLAB_NAME",
						Value: cr.Spec.Instance,
					},
					{
						Name:  "NAMESPACE",
						Value: cr.Namespace,
					},
					{
						Name:  "BACKUP_LOCK",
						Value: strings.Join([]string{cr.Name, "backup", "lock"}, "-"),
					},
				},
			},
		},
	})

	backup.Spec.Schedule = cr.Spec.Schedule
	backup.Spec.JobTemplate.Spec.Template.Spec.ServiceAccountName = "gitlab-backup"
	backup.Spec.JobTemplate.Spec.Template.Spec.RestartPolicy = corev1.RestartPolicyOnFailure

	return backup
}

// NewBackup returns a job that will initiate a backup immediately
func NewBackup(cr *gitlabv1beta1.Backup) *batchv1.Job {
	labels := gitlabutils.Label(cr.Name, "backup", gitlabutils.BackupType)

	backup := gitlabutils.GenericJob(gitlabutils.Component{
		Labels:    labels,
		Namespace: cr.Namespace,
		Containers: []corev1.Container{
			{
				Name:            "control",
				Image:           "registry.gitlab.com/ochienged/backup-control:latest",
				ImagePullPolicy: corev1.PullIfNotPresent,
				Env: []corev1.EnvVar{
					{
						Name:  "GITLAB_NAME",
						Value: cr.Spec.Instance,
					},
					{
						Name:  "NAMESPACE",
						Value: cr.Namespace,
					},
					{
						Name:  "BACKUP_LOCK",
						Value: strings.Join([]string{cr.Name, "backup", "lock"}, "-"),
					},
				},
			},
		},
	})

	backup.Spec.Template.Spec.ServiceAccountName = "gitlab-backup"

	return backup
}
