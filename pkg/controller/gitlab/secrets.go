package gitlab

import (
	"bytes"
	"strings"
	"text/template"

	gitlabv1beta1 "gitlab.com/ochienged/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlabutils "gitlab.com/ochienged/gitlab-operator/pkg/controller/utils"
	corev1 "k8s.io/api/core/v1"
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

func getRegistryCertSecret(cr *gitlabv1beta1.Gitlab) *corev1.Secret {
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

	if cr.Spec.SMTP.Password != "" {
		settings.StringData["smtp_user_password"] = cr.Spec.SMTP.Password
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
	hostnames := []string{getGitlabURL(cr)}
	if !cr.Spec.Registry.Disabled {
		hostnames = append(hostnames, getRegistryURL(cr))
	}

	minio := getMinioOverrides(cr.Spec.Minio)
	if !minio.Disabled {
		hostnames = append(hostnames, getMinioURL(cr))
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

// Reconciler for secrets
func (r *ReconcileGitlab) reconcileSecrets(cr *gitlabv1beta1.Gitlab) error {
	var secrets []*corev1.Secret

	gitaly := getGitalySecret(cr)

	ca := getCertificateAuthoritySecret(cr)

	workhorse := getWorkhorseSecret(cr)

	registry := getRegistryHTTPSecret(cr)

	rails := getRailsSecret(cr)

	minio := getMinioSecret(cr)

	postgres := getPostgresSecret(cr)

	redis := getRedisSecret(cr)

	core := getGilabSecret(cr)

	smtp := getSMTPSettingsSecret(cr)

	shell := getShellSecret(cr)

	keys := getShellSSHKeysSecret(cr)

	secrets = append(secrets,
		ca,
		gitaly,
		registry,
		workhorse,
		rails,
		minio,
		postgres,
		redis,
		core,
		smtp,
		shell,
		keys,
	)

	for _, secret := range secrets {
		if err := r.createKubernetesResource(cr, secret); err != nil {
			return err
		}
	}

	return nil
}

func (r *ReconcileGitlab) reconcileGitlabCertificates(cr *gitlabv1beta1.Gitlab) error {
	var certs []*corev1.Secret

	registry := getRegistryCertSecret(cr)

	gitlab := getGitlabHTTPSCertificate(cr)

	certs = append(certs, registry, gitlab)

	for _, cert := range certs {
		if err := r.createKubernetesResource(cr, cert); err != nil {
			return err
		}
	}

	return nil
}
