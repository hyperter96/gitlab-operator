package gitlab

import (
	"bytes"
	"context"
	"strings"
	"text/template"

	gitlabv1beta1 "gitlab.com/ochienged/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlabutils "gitlab.com/ochienged/gitlab-operator/pkg/controller/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func getGilabSecret(cr *gitlabv1beta1.Gitlab) *corev1.Secret {
	labels := gitlabutils.Label(cr.Name, "gitlab", gitlabutils.GitlabType)

	registrationToken := gitlabutils.Password(gitlabutils.PasswordOptions{
		EnableSpecialChars: false,
		Length:             32,
	})

	rootPassword := gitlabutils.Password(gitlabutils.PasswordOptions{
		EnableSpecialChars: false,
		Length:             36,
	})

	secrets := map[string]string{
		"gitlab_root_password":                      rootPassword,
		"initial_shared_runners_registration_token": registrationToken,
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

func getGitalySecret(cr *gitlabv1beta1.Gitlab) *corev1.Secret {
	labels := gitlabutils.Label(cr.Name, "gitaly", gitlabutils.GitlabType)

	token := gitlabutils.Password(gitlabutils.PasswordOptions{
		EnableSpecialChars: false,
		Length:             64,
	})

	gitaly := gitlabutils.GenericSecret(cr.Name+"-gitaly-secret", cr.Namespace, labels)
	gitaly.StringData = map[string]string{
		"token": token,
	}

	return gitaly
}

func getRegistrySecret(cr *gitlabv1beta1.Gitlab) *corev1.Secret {
	labels := gitlabutils.Label(cr.Name, "registry", gitlabutils.GitlabType)

	// Retrieve CA cert amd key
	caSecret := gitlabutils.SecretData("gitlab-ca-cert", cr.Namespace)
	caKey, _ := gitlabutils.ParsePEMPrivateKey(caSecret["tls.key"])
	caCert, _ := gitlabutils.ParsePEMCertificate(caSecret["tls.crt"])
	caCertPEM := gitlabutils.EncodeCertificateToPEM(caCert)

	hostnames := []string{}
	privateKey, _ := gitlabutils.PrivateKeyRSA(4096)
	keyPEM := gitlabutils.EncodePrivateKeyToPEM(privateKey)
	certificate, _ := gitlabutils.ClientCertificate(privateKey, caKey, caCert, hostnames)
	certPEM := gitlabutils.EncodeCertificateToPEM(certificate)

	registry := gitlabutils.GenericSecret(cr.Name+"-registry-secret", cr.Namespace, labels)
	registry.StringData = map[string]string{
		"registry-auth.crt": string(certPEM),
		"registry-auth.key": string(keyPEM),
		"ca.crt":            string(caCertPEM),
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

func getRailsSecret(cr *gitlabv1beta1.Gitlab) *corev1.Secret {
	labels := gitlabutils.Label(cr.Name, "rails", gitlabutils.GitlabType)

	secretkey := gitlabutils.Password(gitlabutils.PasswordOptions{
		EnableSpecialChars: false,
		Length:             129,
	})

	otpkey := gitlabutils.Password(gitlabutils.PasswordOptions{
		EnableSpecialChars: false,
		Length:             129,
	})

	dbkey := gitlabutils.Password(gitlabutils.PasswordOptions{
		EnableSpecialChars: false,
		Length:             129,
	})

	key, _ := gitlabutils.PrivateKeyRSA(2048)
	privateKey := string(gitlabutils.EncodePrivateKeyToPEM(key))

	options := RailsOptions{
		SecretKey:     secretkey,
		OTPKey:        otpkey,
		DatabaseKey:   dbkey,
		RSAPrivateKey: strings.Split(privateKey, "\n"),
	}

	var secret bytes.Buffer
	secretsTemplate := template.Must(template.ParseFiles("/templates/rails-secret.yml"))
	secretsTemplate.Execute(&secret, options)

	rails := gitlabutils.GenericSecret(cr.Name+"-rails-secret", cr.Namespace, labels)
	rails.StringData = map[string]string{
		"secrets.yml": gitlabutils.RemoveEmptyLines(secret.String()),
	}

	return rails
}

func getMinioSecret(cr *gitlabv1beta1.Gitlab) *corev1.Secret {
	labels := gitlabutils.Label(cr.Name, "minio", gitlabutils.GitlabType)

	secretKey := gitlabutils.Password(gitlabutils.PasswordOptions{
		EnableSpecialChars: false,
		Length:             48,
	})

	registry := gitlabutils.GenericSecret(cr.Name+"-minio-secret", cr.Namespace, labels)
	registry.StringData = map[string]string{
		"accesskey": "gitlab",
		"secretkey": secretKey,
	}

	return registry
}

func getPostgresSecret(cr *gitlabv1beta1.Gitlab) *corev1.Secret {
	labels := gitlabutils.Label(cr.Name, "postgres", gitlabutils.GitlabType)

	gitlabPassword := gitlabutils.Password(gitlabutils.PasswordOptions{
		EnableSpecialChars: false,
		Length:             36,
	})

	postgresPassword := gitlabutils.Password(gitlabutils.PasswordOptions{
		EnableSpecialChars: false,
		Length:             36,
	})

	postgres := gitlabutils.GenericSecret(cr.Name+"-postgresql-secret", cr.Namespace, labels)
	postgres.StringData = map[string]string{
		"postgresql-password":          gitlabPassword,
		"postgresql-postgres-password": postgresPassword,
	}

	return postgres
}

func getRedisSecret(cr *gitlabv1beta1.Gitlab) *corev1.Secret {
	labels := gitlabutils.Label(cr.Name, "redis", gitlabutils.GitlabType)

	password := gitlabutils.Password(gitlabutils.PasswordOptions{
		EnableSpecialChars: false,
		Length:             36,
	})

	redis := gitlabutils.GenericSecret(cr.Name+"-redis-secret", cr.Namespace, labels)
	redis.StringData = map[string]string{
		"secret": password,
	}

	return redis
}

func getWorkhorseSecret(cr *gitlabv1beta1.Gitlab) *corev1.Secret {
	labels := gitlabutils.Label(cr.Name, "workhorse", gitlabutils.GitlabType)

	sharedSecret := gitlabutils.Password(gitlabutils.PasswordOptions{
		EnableSpecialChars: false,
		Length:             32,
	})
	encodedSecret := gitlabutils.EncodeString(sharedSecret)

	workhorse := gitlabutils.GenericSecret(cr.Name+"-workhorse-secret", cr.Namespace, labels)
	workhorse.StringData = map[string]string{
		"shared_secret": encodedSecret,
	}

	return workhorse
}

func getSMTPSettingsSecret(cr *gitlabv1beta1.Gitlab) *corev1.Secret {
	labels := gitlabutils.Label(cr.Name, "smtp", gitlabutils.GitlabType)

	settings := gitlabutils.GenericSecret(cr.Name+"-smtp-settings-secret", cr.Namespace, labels)
	settings.StringData = map[string]string{
		"smtp_settings.rb": getSMTPSettings(cr),
	}

	return settings
}

func getGitlabHTTPSCertificate(cr *gitlabv1beta1.Gitlab) *corev1.Secret {
	labels := gitlabutils.Label(cr.Name, "gitlab", gitlabutils.GitlabType)

	caSecret := gitlabutils.SecretData("gitlab-ca-cert", cr.Namespace)
	caKey, _ := gitlabutils.ParsePEMPrivateKey(caSecret["tls.key"])
	caCert, _ := gitlabutils.ParsePEMCertificate(caSecret["tls.crt"])
	caCertPEM := gitlabutils.EncodeCertificateToPEM(caCert)

	// Add hostnames to be secured by cert
	hostnames := []string{cr.Spec.ExternalURL}
	if cr.Spec.Registry.Enabled {
		hostnames = append(hostnames, cr.Spec.Registry.ExternalURL)
	}

	privateKey, _ := gitlabutils.PrivateKeyRSA(4096)
	keyPEM := gitlabutils.EncodePrivateKeyToPEM(privateKey)
	certificate, err := gitlabutils.ClientCertificate(privateKey, caKey, caCert, hostnames)
	if err != nil {
		log.Error(err, "Error creating client cert")
	}
	certPEM := gitlabutils.EncodeCertificateToPEM(certificate)

	cert := gitlabutils.GenericSecret(cr.Name+"-gitlab-cert", cr.Namespace, labels)
	cert.Type = corev1.SecretTypeTLS
	cert.StringData = map[string]string{
		"tls.crt": string(certPEM),
		"tls.key": string(keyPEM),
		"ca.crt":  string(caCertPEM),
	}

	return cert
}

func getCertificateAuthoritySecret(cr *gitlabv1beta1.Gitlab) *corev1.Secret {
	labels := gitlabutils.Label(cr.Name, "gitlab-ca", gitlabutils.GitlabType)

	key, _ := gitlabutils.PrivateKeyRSA(4096)
	pemKey := gitlabutils.EncodePrivateKeyToPEM(key)
	certificate, _ := gitlabutils.CACertificate(key)
	caCert := gitlabutils.EncodeCertificateToPEM(certificate)

	cert := gitlabutils.GenericSecret("gitlab-ca-cert", cr.Namespace, labels)
	cert.Type = corev1.SecretTypeTLS
	cert.StringData = map[string]string{
		"tls.crt": string(caCert),
		"tls.key": string(pemKey),
	}

	return cert
}

// func getRegistryCertificate(cr *gitlabv1beta1.Gitlab) *corev1.Secret {
// 	labels := gitlabutils.Label(cr.Name, "registry", gitlabutils.GitlabType)

// 	certificate, key, err := gitlabutils.KeyCertificate()
// 	if err != nil {
// 		log.Error(err, "")
// 	}

// 	cert := gitlabutils.GenericSecret(cr.Name+"-registry-cert", cr.Namespace, labels)
// 	cert.Type = corev1.SecretTypeTLS
// 	cert.StringData = map[string]string{
// 		"tls.crt": string(certificate),
// 		"tls.key": string(key),
// 	}

// 	return cert
// }

/***************************************************************
 * Reconcilers for the different secrets begin below this line *
 ***************************************************************/

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

func (r *ReconcileGitlab) reconcileGitlabCASecret(cr *gitlabv1beta1.Gitlab) error {
	ca := getCertificateAuthoritySecret(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: ca.Name}, ca) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, ca, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), ca)
}

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

func (r *ReconcileGitlab) reconcileRailsSecret(cr *gitlabv1beta1.Gitlab) error {
	rails := getRailsSecret(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: rails.Name}, rails) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, rails, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), rails)
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

func (r *ReconcileGitlab) reconcilePostgresSecret(cr *gitlabv1beta1.Gitlab) error {
	postgres := getPostgresSecret(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: postgres.Name}, postgres) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, postgres, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), postgres)
}

func (r *ReconcileGitlab) reconcileRedisSecret(cr *gitlabv1beta1.Gitlab) error {
	redis := getRedisSecret(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: redis.Name}, redis) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, redis, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), redis)
}

func (r *ReconcileGitlab) reconcileGitlabSecret(cr *gitlabv1beta1.Gitlab) error {
	core := getGilabSecret(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: core.Name}, core) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, core, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), core)
}

func (r *ReconcileGitlab) reconcileSMTPSettingsSecret(cr *gitlabv1beta1.Gitlab) error {
	smtp := getSMTPSettingsSecret(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: smtp.Name}, smtp) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, smtp, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), smtp)
}

func (r *ReconcileGitlab) reconcileGitlabCertSecret(cr *gitlabv1beta1.Gitlab) error {
	gitlab := getGitlabHTTPSCertificate(cr)

	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: gitlab.Name}, gitlab) {
		return nil
	}

	if err := controllerutil.SetControllerReference(cr, gitlab, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), gitlab)
}

// func (r *ReconcileGitlab) reconcileRegistryCertSecret(cr *gitlabv1beta1.Gitlab) error {
// 	registry := getRegistryCertificate(cr)

// 	if gitlabutils.IsObjectFound(r.client, types.NamespacedName{Namespace: cr.Namespace, Name: registry.Name}, registry) {
// 		return nil
// 	}

// 	if err := controllerutil.SetControllerReference(cr, registry, r.scheme); err != nil {
// 		return err
// 	}

// 	return r.client.Create(context.TODO(), registry)
// }
