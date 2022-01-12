package internal

import (
	"fmt"

	acmev1 "github.com/jetstack/cert-manager/pkg/apis/acme/v1"
	certmanagerv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	certmetav1 "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
)

const (
	globalIngressConfigureCertmanager = "global.ingress.configureCertmanager"
	configureCertmanagerDefault       = true
	specCertIssuerEmail               = "admin@example.com"
	specCertIssuerServer              = "https://acme-v02.api.letsencrypt.org/directory"
)

// CertManager returns `true` if CertManager is enabled, and `false` if not.
func CertManagerEnabled(adapter gitlab.CustomResourceAdapter) bool {
	configureCertmanager, _ := gitlab.GetBoolValue(adapter.Values(), globalIngressConfigureCertmanager, configureCertmanagerDefault)

	return configureCertmanager
}

// GetIssuerConfig gets the ACME issuer to use from GitLab resource.
func GetIssuerConfig(adapter gitlab.CustomResourceAdapter) certmanagerv1.IssuerConfig {
	if CertManagerEnabled(adapter) {
		email, err := gitlab.GetStringValue(adapter.Values(), "certmanager-issuer.email")
		if err != nil || email == "" {
			email = specCertIssuerEmail
		}

		server, err := gitlab.GetStringValue(adapter.Values(), "certmanager-issuer.server")
		if err != nil || server == "" {
			server = specCertIssuerServer
		}

		ingressClass, err := gitlab.GetStringValue(adapter.Values(), "global.ingress.class")
		if err != nil || ingressClass == "" {
			ingressClass = fmt.Sprintf("%s-nginx", adapter.ReleaseName())
		}

		return certmanagerv1.IssuerConfig{
			ACME: &acmev1.ACMEIssuer{
				Email:  email,
				Server: server,
				PrivateKey: certmetav1.SecretKeySelector{
					LocalObjectReference: certmetav1.LocalObjectReference{
						Name: fmt.Sprintf("%s-acme-key", adapter.ReleaseName()),
					},
				},
				Solvers: []acmev1.ACMEChallengeSolver{
					{
						Selector: &acmev1.CertificateDNSNameSelector{},
						HTTP01: &acmev1.ACMEChallengeSolverHTTP01{
							Ingress: &acmev1.ACMEChallengeSolverHTTP01Ingress{
								Class: &ingressClass,
							},
						},
					},
				},
			},
		}
	}

	return certmanagerv1.IssuerConfig{
		SelfSigned: &certmanagerv1.SelfSignedIssuer{},
	}
}

// CertificateIssuer create a certificate generator.
func CertificateIssuer(adapter gitlab.CustomResourceAdapter) *certmanagerv1.Issuer {
	labels := Label(adapter.ReleaseName(), "issuer", GitlabType)

	issuer := &certmanagerv1.Issuer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels["app.kubernetes.io/instance"],
			Namespace: adapter.Namespace(),
			Labels:    labels,
		},
		Spec: certmanagerv1.IssuerSpec{
			IssuerConfig: GetIssuerConfig(adapter),
		},
	}

	return issuer
}

// EndpointTLS informs which services require
// to be secured using generated TLS certificates.
type EndpointTLS struct {
	gitlab   bool
	registry bool
	minio    bool
}

// RequiresCertManagerCertificate function returns true an administrator
// did not provide a TLS ceritificate for an endpoint.
func RequiresCertManagerCertificate(adapter gitlab.CustomResourceAdapter) EndpointTLS {
	// This implies that Operator can only consume wildcard certificate and individual certificate
	// per service will be ignored.
	tlsSecretName, _ := gitlab.GetStringValue(adapter.Values(), "global.ingress.tls.secretName")

	return EndpointTLS{
		gitlab:   tlsSecretName == "",
		registry: tlsSecretName == "",
	}
}

// GitLab returns true if GitLab endpoint requires
// a cert-manager provisioned certificate.
func (ep EndpointTLS) GitLab() bool {
	return ep.gitlab
}

// Registry returns true if Registry endpoint requires
// a cert-manager provisioned certificate.
func (ep EndpointTLS) Registry() bool {
	return ep.registry
}

// Minio returns true if Minio endpoint requires
// a cert-manager provisioned certificate.
func (ep EndpointTLS) Minio() bool {
	return ep.minio
}

// Any returns true if any ingress requires
// a cert-manager certificate.
func (ep EndpointTLS) Any() bool {
	return ep.gitlab || ep.registry || ep.minio
}
