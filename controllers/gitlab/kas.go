package gitlab

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// KasEnabled returns `true` if KAS is enabled, and `false` if not. By default it returns `false`.
func KasEnabled(adapter CustomResourceAdapter) bool {
	return adapter.Values().GetBool("global.kas.enabled", false)
}

func KasConfigMap(adapter CustomResourceAdapter) client.Object {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	return template.Query().ObjectByKindAndComponent(ConfigMapKind, KasComponentName)
}

func KasDeployment(adapter CustomResourceAdapter) client.Object {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	return template.Query().ObjectByKindAndComponent(DeploymentKind, KasComponentName)
}

func KasIngress(adapter CustomResourceAdapter) client.Object {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	return template.Query().ObjectByKindAndComponent(IngressKind, KasComponentName)
}

func KasService(adapter CustomResourceAdapter) client.Object {
	template, err := GetTemplate(adapter)
	if err != nil {
		return nil // WARNING: this should return an error
	}

	return template.Query().ObjectByKindAndComponent(ServiceKind, KasComponentName)
}
