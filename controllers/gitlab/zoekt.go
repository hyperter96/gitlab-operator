package gitlab

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab"
)

const (
	zoektCertificateKey = "gitlab-zoekt.gateway.tls.certificate.name"
)

// ZoektStatefulSet returns the StatefulSet for the Zoekt component.
func ZoektStatefulSet(template helm.Template, adapter gitlab.Adapter) client.Object {
	sset := template.Query().ObjectByKindAndComponent(StatefulSetKind, ZoektComponentName)

	if sset == nil {
		return nil
	}

	// Namespaces are properly set on Zoekt resources with:
	//   https://gitlab.com/gitlab-org/cloud-native/charts/gitlab-zoekt/-/merge_requests/41.
	// When all of Operator's supported CHART_VERSIONS have this patch, this override
	// can be removed.
	sset.SetNamespace(adapter.Name().Namespace)

	return sset
}

// ZoektConfigMap returns the ConfigMap for the Zoekt component.
func ZoektConfigMap(template helm.Template, adapter gitlab.Adapter) client.Object {
	cm := template.Query().ObjectByKindAndComponent(ConfigMapKind, ZoektComponentName)

	if cm == nil {
		return nil
	}

	// Namespaces are properly set on Zoekt resources with:
	//   https://gitlab.com/gitlab-org/cloud-native/charts/gitlab-zoekt/-/merge_requests/41.
	// When all of Operator's supported CHART_VERSIONS have this patch, this override
	// can be removed.
	cm.SetNamespace(adapter.Name().Namespace)

	return cm
}

// ZoektIngress returns the Ingress for the Zoekt component.
func ZoektIngress(template helm.Template, adapter gitlab.Adapter) client.Object {
	ing := template.Query().ObjectByKindAndComponent(IngressKind, ZoektComponentName)

	if ing == nil {
		return nil
	}

	// Namespaces are properly set on Zoekt resources with:
	//   https://gitlab.com/gitlab-org/cloud-native/charts/gitlab-zoekt/-/merge_requests/41.
	// When all of Operator's supported CHART_VERSIONS have this patch, this override
	// can be removed.
	ing.SetNamespace(adapter.Name().Namespace)

	return ing
}

// ZoektService returns the Service for the Zoekt component.
func ZoektService(template helm.Template, adapter gitlab.Adapter) client.Object {
	svc := template.Query().ObjectByKindAndComponent(ServiceKind, ZoektComponentName)

	if svc == nil {
		return nil
	}

	// Namespaces are properly set on Zoekt resources with:
	//   https://gitlab.com/gitlab-org/cloud-native/charts/gitlab-zoekt/-/merge_requests/41.
	// When all of Operator's supported CHART_VERSIONS have this patch, this override
	// can be removed.
	svc.SetNamespace(adapter.Name().Namespace)

	return svc
}

// ZoektCertificate returns the Certificate for the Zoekt component.
func ZoektCertificate(template helm.Template, adapter gitlab.Adapter) client.Object {
	certName := adapter.Values().GetString(zoektCertificateKey, "zoekt-gateway-cert")
	// Labels will be added to the Certificate with:
	//   https://gitlab.com/gitlab-org/cloud-native/charts/gitlab-zoekt/-/merge_requests/41.
	// When all of Operator's supported CHART_VERSIONS have this patch, the certificate
	// can be selected by it's labels.
	cert := template.Query().ObjectByKindAndName(CertificateKind, certName)

	if cert == nil {
		return nil
	}

	// Namespaces are properly set on Zoekt resources with:
	//   https://gitlab.com/gitlab-org/cloud-native/charts/gitlab-zoekt/-/merge_requests/41.
	// When all of Operator's supported CHART_VERSIONS have this patch, this override
	// can be removed.
	cert.SetNamespace(adapter.Name().Namespace)

	return cert
}
