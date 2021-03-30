package backup

import (
	"strings"

	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/internal"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"

	corev1 "k8s.io/api/core/v1"
)

// BackupServiceAccount for GitLab backup use
const BackupServiceAccount = "gitlab-backup"

// IsOnDemandBackup returns true if no backup schedule is
// provided. This implies backup should run immediately
func IsOnDemandBackup(cr *gitlabv1beta1.GLBackup) bool {
	if len(strings.Split(cr.Spec.Schedule, " ")) == 5 {
		return false
	}

	return cr.Spec.Schedule == ""
}

// LockConfigMap is used to ensure backups do not share names
func LockConfigMap(cr *gitlabv1beta1.GLBackup) *corev1.ConfigMap {
	labels := internal.Label(cr.Name, "backup-lock", internal.BackupType)

	return internal.GenericConfigMap(labels["app.kubernetes.io/instance"], cr.Namespace, labels)
}

func backupEnvBuilder(cr *gitlabv1beta1.GLBackup) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  "GITLAB_NAME",
			Value: cr.Spec.Instance,
		},
		{
			Name:  "NAMESPACE",
			Value: cr.Namespace,
		},
		{
			Name:  "JOB_NAME",
			Value: strings.Join([]string{cr.Name, "backup", "lock"}, "-"),
		},
		{
			Name:  "REMOTE_COMMAND",
			Value: backupCommand(cr),
		},
	}
}

func backupCommand(cr *gitlabv1beta1.GLBackup) string {
	command := []string{"backup-utility"}

	if cr.Spec.Restore &&
		(cr.Spec.Timestamp != "" || cr.Spec.URL != "") {
		command = append(command, "--restore")
	}

	if cr.Spec.Timestamp != "" {
		command = append(command, "-t", cr.Spec.Timestamp)
	}

	if cr.Spec.URL != "" {
		command = append(command, "-f", cr.Spec.URL)
	}

	if cr.Spec.Exclusions != "" {
		command = append(command, "--skip", cr.Spec.Exclusions)
	}

	return strings.Join(command, " ")
}

// NewSchedule returns a CronJob with schedule for backups
func NewSchedule(cr *gitlabv1beta1.GLBackup) *batchv1beta1.CronJob {
	labels := internal.Label(cr.Name, "backup", internal.BackupType)

	backup := internal.GenericCronJob(internal.Component{
		Namespace: cr.Namespace,
		Labels:    labels,
		Containers: []corev1.Container{
			{
				Name:            "control",
				Image:           "registry.gitlab.com/ochienged/backup-control:latest",
				ImagePullPolicy: corev1.PullAlways,
				Env:             backupEnvBuilder(cr),
			},
		},
	})

	backup.Spec.Schedule = cr.Spec.Schedule
	backup.Spec.JobTemplate.Spec.Template.Spec.ServiceAccountName = BackupServiceAccount
	backup.Spec.JobTemplate.Spec.Template.Spec.RestartPolicy = corev1.RestartPolicyOnFailure

	return backup
}

// NewBackup returns a kubernetes job that will initiate a backup immediately
func NewBackup(cr *gitlabv1beta1.GLBackup) *batchv1.Job {
	labels := internal.Label(cr.Name, "backup", internal.BackupType)

	backup := internal.GenericJob(internal.Component{
		Labels:    labels,
		Namespace: cr.Namespace,
		Containers: []corev1.Container{
			{
				Name:            "control",
				Image:           "registry.gitlab.com/ochienged/backup-control:latest",
				ImagePullPolicy: corev1.PullAlways,
				Env:             backupEnvBuilder(cr),
			},
		},
	})

	backup.Spec.Template.Spec.ServiceAccountName = BackupServiceAccount

	return backup
}
