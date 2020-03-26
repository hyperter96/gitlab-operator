package gitlab

import (
	"context"

	gitlabv1beta1 "gitlab.com/ochienged/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlabutils "gitlab.com/ochienged/gitlab-operator/pkg/controller/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func getGilabSecret(cr *gitlabv1beta1.Gitlab, s security) *corev1.Secret {
	labels := gitlabutils.Label(cr.Name, "gitlab", gitlabutils.GitlabType)

	secrets := map[string]string{
		"gitlab_root_password":                      s.GitlabRootPassword(),
		"postgres_password":                         s.PostgresPassword(),
		"initial_shared_runners_registration_token": s.RunnerRegistrationToken(),
		"redis_password":                            s.RedisPassword(),
	}

	if cr.Spec.SMTP.Password != "" {
		secrets["smtp_user_password"] = cr.Spec.SMTP.Password
	}

	gitlab := gitlabutils.GenericSecret(cr.Name+"-gitlab-secrets", cr.Namespace, labels)
	gitlab.StringData = secrets

	return gitlab
}

func getShellSSHKeysSecret(cr *gitlabv1beta1.Gitlab) *corev1.Secret {
	labels := gitlabutils.Label(cr.Name, "shell", gitlabutils.GitlabType)

	privateRSAKey, publicRSAKey := gitlabutils.Keypair(gitlabutils.RSAKeyPair())
	privateDSAKey, publicDSAKey := gitlabutils.Keypair(gitlabutils.DSAKeyPair())
	privateECDSAKey, publicECDSAKey := gitlabutils.Keypair(gitlabutils.ECDSAKeyPair())
	privateED25519Key, publicED25519Key := gitlabutils.Keypair(gitlabutils.ED25519KeyPair())

	keys := gitlabutils.GenericSecret(cr.Name+"-shell-host-keys-secret", cr.Namespace, labels)
	keys.StringData = map[string]string{
		"ssh_host_dsa_key":         privateDSAKey,
		"ssh_host_dsa_key.pub":     publicDSAKey,
		"ssh_host_ecdsa_key":       privateECDSAKey,
		"ssh_host_ecdsa_key.pub":   publicECDSAKey,
		"ssh_host_ed25519_key":     privateED25519Key,
		"ssh_host_ed25519_key.pub": publicED25519Key,
		"ssh_host_rsa_key":         privateRSAKey,
		"ssh_host_rsa_key.pub":     publicRSAKey,
	}

	return keys
}

func (r *ReconcileGitlab) reconcileShellSSHKeysSecret(cr *gitlabv1beta1.Gitlab) error {
	keys := getShellSSHKeysSecret(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: keys.Name}, keys) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, keys, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), keys)
}

func getShellSecret(cr *gitlabv1beta1.Gitlab) *corev1.Secret {
	labels := gitlabutils.Label(cr.Name, "shell", gitlabutils.GitlabType)

	shellSecret := gitlabutils.Password(gitlabutils.PasswordOptions{
		EnableSpecialChars: false,
		Length:             65,
	})

	shell := gitlabutils.GenericSecret(cr.Name+"-shell-secret", cr.Namespace, labels)
	shell.StringData = map[string]string{
		"secret": shellSecret,
	}

	return shell
}

func (r *ReconcileGitlab) reconcileShellSecret(cr *gitlabv1beta1.Gitlab) error {
	shell := getShellSecret(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: shell.Name}, shell) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, shell, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), shell)
}

func getGitlabRailsSecret(cr *gitlabv1beta1.Gitlab) *corev1.Secret {
	labels := gitlabutils.Label(cr.Name, "rails", gitlabutils.GitlabType)

	rails := gitlabutils.GenericSecret(cr.Name+"-rails-secret", cr.Namespace, labels)
	rails.StringData = map[string]string{
		"secrets.yml": "",
	}

	return rails
}

func (r *ReconcileGitlab) reconcileGitlabRailsSecret(cr *gitlabv1beta1.Gitlab) error {
	shell := getGitlabRailsSecret(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: shell.Name}, shell) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, shell, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), shell)
}

func getGitalySecret(cr *gitlabv1beta1.Gitlab) *corev1.Secret {
	labels := gitlabutils.Label(cr.Name, "gitaly", gitlabutils.GitlabType)

	gitaly := gitlabutils.GenericSecret(cr.Name+"-gitaly-secret", cr.Namespace, labels)
	gitaly.StringData = map[string]string{
		"token": "",
	}

	return gitaly
}

func (r *ReconcileGitlab) reconcileGitalySecret(cr *gitlabv1beta1.Gitlab) error {
	gitaly := getGitalySecret(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: gitaly.Name}, gitaly) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, gitaly, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), gitaly)
}

func getRegistrySecret(cr *gitlabv1beta1.Gitlab) *corev1.Secret {
	labels := gitlabutils.Label(cr.Name, "registry", gitlabutils.GitlabType)

	key, cert := gitlabutils.Keypair(gitlabutils.KeyCertificate())
	registry := gitlabutils.GenericSecret(cr.Name+"-registry-secret", cr.Namespace, labels)
	registry.StringData = map[string]string{
		"registry-auth.crt": cert,
		"registry-auth.key": key,
	}

	return registry
}

func getRegistryHTTPSecret(cr *gitlabv1beta1.Gitlab) *corev1.Secret {
	labels := gitlabutils.Label(cr.Name, "registry-http", gitlabutils.GitlabType)

	secret := gitlabutils.Password(gitlabutils.PasswordOptions{
		EnableSpecialChars: false,
		Length:             172,
	})

	registry := gitlabutils.GenericSecret(cr.Name+"-registry-http-secret", cr.Namespace, labels)
	registry.StringData = map[string]string{
		"secret": secret,
	}

	return registry
}

func getMinioSecret(cr *gitlabv1beta1.Gitlab) *corev1.Secret {
	labels := gitlabutils.Label(cr.Name, "minio", gitlabutils.GitlabType)

	accesskey := gitlabutils.Password(gitlabutils.PasswordOptions{
		EnableSpecialChars: false,
		Length:             64,
	})

	secretkey := gitlabutils.Password(gitlabutils.PasswordOptions{
		EnableSpecialChars: false,
		Length:             10,
	})

	registry := gitlabutils.GenericSecret(cr.Name+"-minio-secret", cr.Namespace, labels)
	registry.StringData = map[string]string{
		"accesskey": accesskey,
		"secretkey": secretkey,
	}

	return registry
}

/***************************************************************
 * Reconcilers for the different secrets begin below this line *
 ***************************************************************/
func (r *ReconcileGitlab) reconcileRegistrySecret(cr *gitlabv1beta1.Gitlab) error {
	registry := getRegistrySecret(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: registry.Name}, registry) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, registry, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), registry)
}

func getWorkhorseSecret(cr *gitlabv1beta1.Gitlab) *corev1.Secret {
	labels := gitlabutils.Label(cr.Name, "workhorse", gitlabutils.GitlabType)

	sharedSecret := gitlabutils.Password(gitlabutils.PasswordOptions{
		EnableSpecialChars: false,
		Length:             45,
	})

	workhorse := gitlabutils.GenericSecret(cr.Name+"-workhorse-secret", cr.Namespace, labels)
	workhorse.StringData = map[string]string{
		"shared_secret": sharedSecret,
	}

	return workhorse
}

func (r *ReconcileGitlab) reconcileWorkhorseSecret(cr *gitlabv1beta1.Gitlab) error {
	workhorse := getWorkhorseSecret(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: workhorse.Name}, workhorse) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, workhorse, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), workhorse)
}

func (r *ReconcileGitlab) reconcileRegistryHTTPSecret(cr *gitlabv1beta1.Gitlab) error {
	registry := getRegistryHTTPSecret(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: registry.Name}, registry) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, registry, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), registry)
}

func (r *ReconcileGitlab) reconcileMinioSecret(cr *gitlabv1beta1.Gitlab) error {
	minio := getMinioSecret(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: minio.Name}, minio) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, minio, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), minio)
}
