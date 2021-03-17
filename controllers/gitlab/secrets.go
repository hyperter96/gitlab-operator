package gitlab

import (
	"bytes"
	"os"
	"strings"
	"text/template"

	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	gitlabutils "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/utils"
	corev1 "k8s.io/api/core/v1"
)

// RootUserSecretDEPRECATED returns secret containing credentials for initial root account
func RootUserSecretDEPRECATED(cr *gitlabv1beta1.GitLab) *corev1.Secret {
	labels := gitlabutils.Label(cr.Name, "root-password", gitlabutils.GitlabType)

	rootPassword := gitlabutils.Password(gitlabutils.PasswordOptions{
		EnableSpecialChars: false,
		Length:             64,
	})

	gitlab := gitlabutils.GenericSecret(cr.Name+"-initial-root-password", cr.Namespace, labels)
	gitlab.StringData = map[string]string{
		"password": rootPassword,
	}

	return gitlab
}

// RunnerRegistrationSecretDEPRECATED returns registration tokens for GitLab runner
func RunnerRegistrationSecretDEPRECATED(cr *gitlabv1beta1.GitLab) *corev1.Secret {
	labels := gitlabutils.Label(cr.Name, "runner-token", gitlabutils.GitlabType)

	registrationToken := gitlabutils.Password(gitlabutils.PasswordOptions{
		EnableSpecialChars: false,
		Length:             64,
	})

	runner := gitlabutils.GenericSecret(cr.Name+"-runner-token-secret", cr.Namespace, labels)
	runner.StringData = map[string]string{
		"runner-registration-token": registrationToken,
		"runner-token":              "",
	}

	return runner
}

// ShellSSHKeysSecretDEPRECATED returns secret containing SSH keys for GitLab shell
func ShellSSHKeysSecretDEPRECATED(cr *gitlabv1beta1.GitLab) *corev1.Secret {
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

// ShellSecretDEPRECATED returns secret for GitLab shell
func ShellSecretDEPRECATED(cr *gitlabv1beta1.GitLab) *corev1.Secret {
	labels := gitlabutils.Label(cr.Name, "shell", gitlabutils.GitlabType)

	shellSecret := gitlabutils.Password(gitlabutils.PasswordOptions{
		EnableSpecialChars: false,
		Length:             64,
	})

	shell := gitlabutils.GenericSecret(cr.Name+"-shell-secret", cr.Namespace, labels)
	shell.StringData = map[string]string{
		"secret": shellSecret,
	}

	return shell
}

// GitalySecretDEPRECATED returns secrets required by Gitaly server
func GitalySecretDEPRECATED(cr *gitlabv1beta1.GitLab) *corev1.Secret {
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

// RegistryCertSecretDEPRECATED contains certificates for the container registry
func RegistryCertSecretDEPRECATED(cr *gitlabv1beta1.GitLab) *corev1.Secret {
	labels := gitlabutils.Label(cr.Name, "registry", gitlabutils.GitlabType)

	privateKey, _ := gitlabutils.PrivateKeyRSA(4096)
	keyPEM := gitlabutils.EncodePrivateKeyToPEM(privateKey)
	certificate, _ := gitlabutils.ClientCertificate(privateKey, []string{})
	certPEM := gitlabutils.EncodeCertificateToPEM(certificate)

	registry := gitlabutils.GenericSecret(cr.Name+"-registry-secret", cr.Namespace, labels)
	registry.StringData = map[string]string{
		"registry-auth.crt": string(certPEM),
		"registry-auth.key": string(keyPEM),
	}

	return registry
}

// RegistryHTTPSecretDEPRECATED returns the HTTP secret for Registry
func RegistryHTTPSecretDEPRECATED(cr *gitlabv1beta1.GitLab) *corev1.Secret {
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

// RailsSecretDEPRECATED returns rails secrets
func RailsSecretDEPRECATED(cr *gitlabv1beta1.GitLab) *corev1.Secret {
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

	signKey, _ := gitlabutils.PrivateKeyRSA(2048)
	ciSigningKey := string(gitlabutils.EncodePrivateKeyToPEM(signKey))

	options := RailsOptions{
		SecretKey:     secretkey,
		OTPKey:        otpkey,
		DatabaseKey:   dbkey,
		RSAPrivateKey: strings.Split(privateKey, "\n"),
		JWTSigningKey: strings.Split(ciSigningKey, "\n"),
	}

	var secret bytes.Buffer
	secretsTemplate := template.Must(template.ParseFiles(os.Getenv("GITLAB_OPERATOR_ASSETS") + "/templates/rails-secret.yml"))
	secretsTemplate.Execute(&secret, options)

	rails := gitlabutils.GenericSecret(cr.Name+"-rails-secret", cr.Namespace, labels)
	rails.StringData = map[string]string{
		"secrets.yml": gitlabutils.RemoveEmptyLines(secret.String()),
	}

	return rails
}

// PostgresSecretDEPRECATED returns secrets required by Postgresql
func PostgresSecretDEPRECATED(cr *gitlabv1beta1.GitLab) *corev1.Secret {
	labels := gitlabutils.Label(cr.Name, "postgres", gitlabutils.GitlabType)

	gitlabPassword := gitlabutils.Password(gitlabutils.PasswordOptions{
		EnableSpecialChars: false,
		Length:             64,
	})

	postgresPassword := gitlabutils.Password(gitlabutils.PasswordOptions{
		EnableSpecialChars: false,
		Length:             64,
	})

	postgres := gitlabutils.GenericSecret(cr.Name+"-postgresql-secret", cr.Namespace, labels)
	postgres.StringData = map[string]string{
		"postgresql-password":          gitlabPassword,
		"postgresql-postgres-password": postgresPassword,
	}

	return postgres
}

// RedisSecretDEPRECATED returns secrets used by Redis
func RedisSecretDEPRECATED(cr *gitlabv1beta1.GitLab) *corev1.Secret {
	labels := gitlabutils.Label(cr.Name, "redis", gitlabutils.GitlabType)

	password := gitlabutils.Password(gitlabutils.PasswordOptions{
		EnableSpecialChars: false,
		Length:             64,
	})

	redis := gitlabutils.GenericSecret(cr.Name+"-redis-secret", cr.Namespace, labels)
	redis.StringData = map[string]string{
		"secret": password,
	}

	return redis
}

// WorkhorseSecretDEPRECATED returns secrets for Workhorse
func WorkhorseSecretDEPRECATED(cr *gitlabv1beta1.GitLab) *corev1.Secret {
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

// SMTPSettingsSecretDEPRECATED contains secrets used by to relay email to SMTP server
func SMTPSettingsSecretDEPRECATED(cr *gitlabv1beta1.GitLab) *corev1.Secret {
	labels := gitlabutils.Label(cr.Name, "smtp", gitlabutils.GitlabType)

	settings := gitlabutils.GenericSecret(cr.Name+"-smtp-settings-secret", cr.Namespace, labels)
	settings.StringData = map[string]string{
		// This is deprecated. We can comment it.
		// "smtp_settings.rb": getSMTPSettings(cr),
	}

	// Secrets are unused. We can safely comment the following.
	/*
		if cr.Spec.SMTP.Password != "" {
			settings.StringData["smtp_user_password"] = cr.Spec.SMTP.Password
		}
	*/

	return settings
}
