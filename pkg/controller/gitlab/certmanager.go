package gitlab

import (
	acmev1alpha2 "github.com/jetstack/cert-manager/pkg/apis/acme/v1alpha2"
	certmanagerv1alpha2 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"
	certmetav1 "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	gitlabv1beta1 "gitlab.com/ochienged/gitlab-operator/pkg/apis/gitlab/v1beta1"
	gitlabutils "gitlab.com/ochienged/gitlab-operator/pkg/controller/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetIssuerConfig gets the ACME issuer to use from GitLab resource
func GetIssuerConfig(cr *gitlabv1beta1.Gitlab) certmanagerv1alpha2.IssuerConfig {

	if cr.Spec.CertIssuer != nil {
		var ingressClass string = "nginx"

		if cr.Spec.CertIssuer.Server == "" {
			cr.Spec.CertIssuer.Server = "https://acme-v02.api.letsencrypt.org/directory"
		}

		return certmanagerv1alpha2.IssuerConfig{
			ACME: &acmev1alpha2.ACMEIssuer{
				Email:                  cr.Spec.CertIssuer.Email,
				Server:                 cr.Spec.CertIssuer.Server,
				SkipTLSVerify:          cr.Spec.CertIssuer.SkipTLSVerify,
				ExternalAccountBinding: cr.Spec.CertIssuer.ExternalAccountBinding,
				PrivateKey: certmetav1.SecretKeySelector{
					LocalObjectReference: certmetav1.LocalObjectReference{
						Name: cr.Name + "-issuer-key",
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
func CertificateIssuer(cr *gitlabv1beta1.Gitlab) *certmanagerv1alpha2.Issuer {
	labels := gitlabutils.Label(cr.Name, "issuer", gitlabutils.GitlabType)

	issuer := &certmanagerv1alpha2.Issuer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labels["app.kubernetes.io/instance"],
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: certmanagerv1alpha2.IssuerSpec{
			IssuerConfig: GetIssuerConfig(cr),
		},
	}

	return issuer
}

func (r *ReconcileGitlab) reconcileCertManagerCertificates(cr *gitlabv1beta1.Gitlab) error {

	issuer := CertificateIssuer(cr)

	return r.createKubernetesResource(cr, issuer)
}
