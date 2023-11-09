package gitlab

import (
	"sigs.k8s.io/controller-runtime/pkg/client"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/helm"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab/component"
)

type componentMap struct {
	component gitlab.Component
	objFn     func(template helm.Template) client.Object
}

var (
	serviceMonitorComponentMap = []componentMap{
		{component.Gitaly, GitalyServiceMonitor},
		{component.GitLabExporter, ExporterServiceMonitor},
		{component.GitLabShell, ShellServiceMonitor},
		{component.GitLabKAS, KasServiceMonitor},
		{component.GitLabPages, PagesServiceMonitor},
		{component.NginxIngress, NGINXServiceMonitor},
		{component.Praefect, PraefectServiceMonitor},
		{component.Redis, RedisServiceMonitor},
		{component.Registry, RegistryServiceMonitor},
		{component.Webservice, WebserviceServiceMonitor},
		{component.Webservice, WebserviceWorkhorseServiceMonitor},
	}

	podMonitorComponentMap = []componentMap{
		{component.Sidekiq, SidekiqPodMonitor},
	}
)

func wantedMonitoringObjects(cms []componentMap, adapter gitlab.Adapter, template helm.Template) []client.Object {
	var result []client.Object

	for _, entry := range cms {
		if adapter.WantsComponent(entry.component) {
			// Monitoring objects are disabled by default. Only try to reconcile them if present in the template.
			if obj := entry.objFn(template); obj != nil {
				result = append(result, obj)
			}
		}
	}

	return result
}

func WantedServiceMonitors(adapter gitlab.Adapter, template helm.Template) []client.Object {
	return wantedMonitoringObjects(serviceMonitorComponentMap, adapter, template)
}

func WantedPodMonitors(adapter gitlab.Adapter, template helm.Template) []client.Object {
	return wantedMonitoringObjects(podMonitorComponentMap, adapter, template)
}
