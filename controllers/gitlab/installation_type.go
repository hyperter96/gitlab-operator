package gitlab

import (
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// InstallationType defines the installation type for usage pings (if enabled).
	installationType = "gitlab-operator"
)

func setInstallationType(obj client.Object) {
	// Attention: Type Assertion: ConfigMap.Data is needed
	cm := obj.(*corev1.ConfigMap)
	for k, v := range cm.Data {
		if k == "installation_type" && v != installationType {
			cm.Data[k] = installationType
		}
	}
}
