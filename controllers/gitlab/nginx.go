package gitlab

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	nginxEnabled        = "nginx-ingress.enabled"
	nginxEnabledDefault = true
)

// NGINXEnabled returns `true` if NGINX is enabled and `false` if not.
func NGINXEnabled(adapter CustomResourceAdapter) bool {
	return adapter.Values().GetBool(nginxEnabled, nginxEnabledDefault)
}

// NGINXConfigMaps returns the ConfigMaps of the NGINX component.
func NGINXConfigMaps(adapter CustomResourceAdapter) []client.Object {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil
	}

	result := template.Query().ObjectsByKindAndLabels(ConfigMapKind, map[string]string{
		"app": NGINXComponentName,
	})

	// Namespaces are properly set on NGINX objects in Chart version 5.6.0.
	// When all of Operator's supported CHART_VERSIONS are at or above 5.6.0,
	// we can remove this override.
	for _, cm := range result {
		cm.SetNamespace(adapter.Namespace())
	}

	return result
}

// NGINXServices returns the Services of the NGINX Component.
func NGINXServices(adapter CustomResourceAdapter) []client.Object {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil
	}

	result := template.Query().ObjectsByKindAndLabels(ServiceKind, map[string]string{
		"app": NGINXComponentName,
	})

	// Namespaces are properly set on NGINX objects in Chart version 5.6.0.
	// When all of Operator's supported CHART_VERSIONS are at or above 5.6.0,
	// we can remove this override.
	for _, svc := range result {
		svc.SetNamespace(adapter.Namespace())
	}

	return result
}

// NGINXDeployments returns the Deployments of the NGINX Component.
func NGINXDeployments(adapter CustomResourceAdapter) []client.Object {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil
	}

	result := template.Query().ObjectsByKindAndLabels(DeploymentKind, map[string]string{
		"app": NGINXComponentName,
	})

	// Namespaces are properly set on NGINX objects in Chart version 5.6.0.
	// When all of Operator's supported CHART_VERSIONS are at or above 5.6.0,
	// we can remove this override.
	for _, dep := range result {
		dep.SetNamespace(adapter.Namespace())
	}

	return result
}

// NGINXAnnotations returns the annotations for Ingress objects.
func NGINXAnnotations() map[string]string {
	return map[string]string{
		"nginx.ingress.kubernetes.io/proxy-body-size":         "0",
		"nginx.ingress.kubernetes.io/proxy-buffering":         "off",
		"nginx.ingress.kubernetes.io/proxy-read-timeout":      "900",
		"nginx.ingress.kubernetes.io/proxy-request-buffering": "off",
	}
}
