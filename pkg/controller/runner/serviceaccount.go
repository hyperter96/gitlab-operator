package runner

import (
	gitlabv1beta1 "gitlab.com/ochienged/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlabutils "gitlab.com/ochienged/gitlab-operator/pkg/controller/utils"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getRunnerServiceAccount(cr *gitlabv1beta1.Runner) *corev1.ServiceAccount {
	labels := gitlabutils.Label(cr.Name, "runner", gitlabutils.RunnerType)

	sa := gitlabutils.ServiceAccount(cr.Namespace, labels)
	return sa
}

// TODO: Create fine grained RBAC rules
func getRunnerRoles(cr *gitlabv1beta1.Runner) *rbacv1.Role {
	labels := gitlabutils.Label(cr.Name, "runner", gitlabutils.RunnerType)
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-runner",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"secrets", "pods", "pods/exec", "services", "endpoints", "configmaps"},
				Verbs:     []string{"create", "list", "get", "watch", "delete"},
			},
		},
	}
}

func getRunnerRoleBinding(cr *gitlabv1beta1.Runner) *rbacv1.RoleBinding {
	labels := gitlabutils.Label(cr.Name, "runner", gitlabutils.RunnerType)
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-runner",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind: "ServiceAccount",
				Name: cr.Name + "-runner",
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "Role",
			Name: cr.Name + "-runner",
		},
	}
}
