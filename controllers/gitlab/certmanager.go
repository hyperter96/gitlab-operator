package gitlab

import (
	acmev1beta1 "github.com/jetstack/cert-manager/pkg/apis/acme/v1beta1"
	certmanagerv1beta1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1beta1"
	certmetav1 "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	gitlabv1beta1 "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/api/v1beta1"
	gitlabutils "gitlab.com/gitlab-org/gl-openshift/gitlab-operator/controllers/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetIssuerConfig gets the ACME issuer to use from GitLab resource
func GetIssuerConfig(cr *gitlabv1beta1.GitLab) certmanagerv1beta1.IssuerConfig {

	if cr.Spec.CertIssuer != nil {
		var ingressClass string = "nginx"

		if cr.Spec.CertIssuer.Server == "" {
			cr.Spec.CertIssuer.Server = "https://acme-v02.api.letsencrypt.org/directory"
		}

		var solvers []acmev1beta1.ACMEChallengeSolver = cr.Spec.CertIssuer.Solvers
		if len(solvers) == 0 {
			solvers = []acmev1beta1.ACMEChallengeSolver{
				{
					Selector: &acmev1beta1.CertificateDNSNameSelector{},
					HTTP01: &acmev1beta1.ACMEChallengeSolverHTTP01{
						Ingress: &acmev1beta1.ACMEChallengeSolverHTTP01Ingress{
							Class: &ingressClass,
						},
					},
				},
			}
		}

		return certmanagerv1beta1.IssuerConfig{
			ACME: &acmev1beta1.ACMEIssuer{
				Email:                  cr.Spec.CertIssuer.Email,
				Server:                 cr.Spec.CertIssuer.Server,
				SkipTLSVerify:          cr.Spec.CertIssuer.SkipTLSVerify,
				ExternalAccountBinding: cr.Spec.CertIssuer.ExternalAccountBinding,
				PrivateKey: certmetav1.SecretKeySelector{
					LocalObjectReference: certmetav1.LocalObjectReference{
						Name: cr.Name + "-issuer-key",
					},
				},
				Solvers: solvers,
			},
		}
	}

	return certmanagerv1beta1.IssuerConfig{
		SelfSigned: &certmanagerv1beta1.SelfSignedIssuer{},
	}
}

// CertificateIssuer create a certificate generator
func CertificateIssuer(cr *gitlabv1beta1.GitLab) *certmanagerv1beta1.Issuer {
	labels := gitlabutils.Label(cr.Name, "issuer", gitlabutils.GitlabType)

	issuer := &certmanagerv1beta1.Issuer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels["app.kubernetes.io/instance"],
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: certmanagerv1beta1.IssuerSpec{
			IssuerConfig: GetIssuerConfig(cr),
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
func RequiresCertManagerCertificate(cr *gitlabv1beta1.GitLab) EndpointTLS {
	return EndpointTLS{
		gitlab:   cr.Spec.TLS == "",
		registry: cr.Spec.Registry.TLS == "",
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

// All returns true if all ingresses require
// a cert-manager certificate
func (ep EndpointTLS) All() bool {
	return ep.gitlab || ep.registry || ep.minio
}
