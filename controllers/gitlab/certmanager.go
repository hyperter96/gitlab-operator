package gitlab

import (
	"fmt"

	acmev1alpha2 "github.com/jetstack/cert-manager/pkg/apis/acme/v1alpha2"
	certmanagerv1alpha2 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	certmetav1 "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	"gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/helpers"
	gitlabutils "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	specCertIssuerEmail  = "admin@example.com"
	specCertIssuerServer = "https://acme-v02.api.letsencrypt.org/directory"
	specIngressClass     = "nginx"
)

// GetIssuerConfig gets the ACME issuer to use from GitLab resource
func GetIssuerConfig(adapter helpers.CustomResourceAdapter) certmanagerv1alpha2.IssuerConfig {

	useCertIssuer, err := helpers.GetBoolValue(adapter.Values(), "global.ingress.configureCertmanager")
	if err != nil {
		useCertIssuer = true
	}

	if useCertIssuer {
		ingressClass := specIngressClass

		email, err := helpers.GetStringValue(adapter.Values(), "certmanager-issuer.email")
		if err != nil || email == "" {
			email = specCertIssuerEmail
		}

		return certmanagerv1alpha2.IssuerConfig{
			ACME: &acmev1alpha2.ACMEIssuer{
				Email:  email,
				Server: specCertIssuerServer,
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
func CertificateIssuer(adapter helpers.CustomResourceAdapter) *certmanagerv1alpha2.Issuer {
	labels := gitlabutils.Label(adapter.ReleaseName(), "issuer", gitlabutils.GitlabType)

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
func RequiresCertManagerCertificate(adapter helpers.CustomResourceAdapter) EndpointTLS {

	// This implies that Operator can only consume wildcard certificate and individual certificate
	// per service will be ignored.
	tlsSecretName, _ := helpers.GetStringValue(adapter.Values(), "global.ingress.tls.secretName")

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
