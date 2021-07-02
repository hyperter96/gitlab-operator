package internal

import (
	"fmt"

	acmev1alpha2 "github.com/jetstack/cert-manager/pkg/apis/acme/v1alpha2"
	certmanagerv1alpha2 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	certmetav1 "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/gitlab"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	specCertIssuerEmail  = "admin@example.com"
	specCertIssuerServer = "https://acme-v02.api.letsencrypt.org/directory"
	specIngressClass     = "nginx"
)

// GetIssuerConfig gets the ACME issuer to use from GitLab resource
func GetIssuerConfig(adapter gitlab.CustomResourceAdapter) certmanagerv1alpha2.IssuerConfig {

	if configureCertmanager, _ := gitlab.GetBoolValue(adapter.Values(), "global.ingress.configureCertmanager", true); configureCertmanager {
		ingressClass := specIngressClass

		email, err := gitlab.GetStringValue(adapter.Values(), "certmanager-issuer.email")
		if err != nil || email == "" {
			email = specCertIssuerEmail
		}

		server, err := gitlab.GetStringValue(adapter.Values(), "certmanager-issuer.server")
		if err != nil || server == "" {
			server = specCertIssuerServer
		}

		return certmanagerv1alpha2.IssuerConfig{
			ACME: &acmev1alpha2.ACMEIssuer{
				Email:  email,
				Server: server,
				PrivateKey: certmetav1.SecretKeySelector{
					LocalObjectReference: certmetav1.LocalObjectReference{
						Name: fmt.Sprintf("%s-acme-key", adapter.ReleaseName()),
					},
				},
				Solvers: []acmev1alpha2.ACMEChallengeSolver{
					{
						Selector: &acmev1alpha2.CertificateDNSNameSelector{},
						HTTP01: &acmev1alpha2.ACMEChallengeSolverHTTP01{
							Ingress: &acmev1alpha2.ACMEChallengeSolverHTTP01Ingress{
								Class: &ingressClass,
							},
						},
					},
				},
			},
		}
	}

	return certmanagerv1alpha2.IssuerConfig{
		SelfSigned: &certmanagerv1alpha2.SelfSignedIssuer{},
	}
}

// CertificateIssuer create a certificate generator
func CertificateIssuer(adapter gitlab.CustomResourceAdapter) *certmanagerv1alpha2.Issuer {
	labels := Label(adapter.ReleaseName(), "issuer", GitlabType)

	issuer := &certmanagerv1alpha2.Issuer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels["app.kubernetes.io/instance"],
			Namespace: adapter.Namespace(),
			Labels:    labels,
		},
		Spec: certmanagerv1alpha2.IssuerSpec{
			IssuerConfig: GetIssuerConfig(adapter),
		},
	}

	return issuer
}

// EndpointTLS informs which services require
// to be secured using generated TLS certificates
type EndpointTLS struct {
	gitlab   bool
	registry bool
	minio    bool
}

// RequiresCertManagerCertificate function returns true an administrator
// did not provide a TLS ceritificate for an endpoint
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
// a cert-manager provisioned certificate
func (ep EndpointTLS) GitLab() bool {
	return ep.gitlab
}

// Registry returns true if Registry endpoint requires
// a cert-manager provisioned certificate
func (ep EndpointTLS) Registry() bool {
	return ep.registry
}

// Minio returns true if Minio endpoint requires
// a cert-manager provisioned certificate
func (ep EndpointTLS) Minio() bool {
	return ep.minio
}

// Any returns true if any ingress requires
// a cert-manager certificate
func (ep EndpointTLS) Any() bool {
	return ep.gitlab || ep.registry || ep.minio
}
