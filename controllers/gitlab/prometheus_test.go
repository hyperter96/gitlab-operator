package gitlab

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/gitlab/component"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/pkg/support"
)

const (
	prometheusEnabled = "prometheus.install"

	serverEnabled       = "prometheus.server.enabled"
	alertmanagerEnabled = "prometheus.alertmanager.enabled"
	nodeExproterEnabled = "prometheus.nodeExporter.enabled"
	pushgatewayEnabled  = "prometheus.pushgateway.enabled"

	componentLabel        = "component"
	serverComponent       = "server"
	alertmanagerComponent = "alertmanager"
	nodeExporterComponent = "node-exporter"
	pushgatewayComponent  = "pushgateway"
)

var _ = Describe("Prometheus", func() {
	var chartValues support.Values
	var promEnabled bool
	var configMaps, services, deployments, daemonSets, statefulSets, ingresses, pvcs []client.Object

	JustBeforeEach(func() {
		mockGitLab := CreateMockGitLab(releaseName, namespace, chartValues)
		adapter := CreateMockAdapter(mockGitLab)
		template, err := GetTemplate(adapter)

		Expect(err).To(BeNil())
		Expect(template).NotTo(BeNil())

		promEnabled = adapter.WantsComponent(component.Prometheus)
		configMaps = PrometheusConfigMaps(template)
		services = PrometheusServices(template)
		deployments = PrometheusDeployments(template)
		daemonSets = PrometheusDaemonSets(template)
		statefulSets = PrometheusStatefulSets(template)
		ingresses = PrometheusIngresses(template)
		pvcs = PrometheusPersistentVolumeClaims(template)
	})

	When("All Prometheus components are enabled", func() {
		BeforeEach(func() {
			chartValues = support.Values{}
			_ = chartValues.SetValue(prometheusEnabled, true)
			_ = chartValues.SetValue(serverEnabled, true)
			_ = chartValues.SetValue("prometheus.server.ingress.enabled", true)
			_ = chartValues.SetValue(alertmanagerEnabled, true)
			_ = chartValues.SetValue(nodeExproterEnabled, true)
			_ = chartValues.SetValue(pushgatewayEnabled, true)
		})

		It("Should contain Prometheus resources", func() {
			Expect(promEnabled).To(BeTrue())
			Expect(services).To(
				matchAllPrometheusElements(Not(BeNil()), serverComponent, alertmanagerComponent, nodeExporterComponent, pushgatewayComponent),
			)
			Expect(configMaps).To(
				matchAllPrometheusElements(Not(BeNil()), serverComponent, alertmanagerComponent),
			)
			Expect(pvcs).To(
				matchAllPrometheusElements(Not(BeNil()), serverComponent, alertmanagerComponent),
			)
			Expect(deployments).To(
				matchAllPrometheusElements(Not(BeNil()), serverComponent, alertmanagerComponent, pushgatewayComponent),
			)
			Expect(daemonSets).To(
				matchAllPrometheusElements(Not(BeNil()), nodeExporterComponent),
			)
			Expect(ingresses).To(
				matchAllPrometheusElements(Not(BeNil()), serverComponent),
			)
			Expect(statefulSets).To(BeEmpty())
		})
	})

	When("Prometheus server statefulset enabled", func() {
		BeforeEach(func() {
			chartValues = support.Values{}
			_ = chartValues.SetValue(prometheusEnabled, true)
			_ = chartValues.SetValue(serverEnabled, true)
			_ = chartValues.SetValue("prometheus.server.statefulSet.enabled", true)
			_ = chartValues.SetValue(alertmanagerEnabled, false)
			_ = chartValues.SetValue(nodeExproterEnabled, false)
			_ = chartValues.SetValue(pushgatewayEnabled, false)
		})

		It("Should contain Prometheus resources", func() {
			Expect(statefulSets).To(matchAllPrometheusElements(Not(BeNil()), serverComponent))
			// check for additional headless service is created
			Expect(services).To(HaveLen(2))
			Expect(promEnabled).To(BeTrue())
			Expect(pvcs).To(BeEmpty())
			Expect(deployments).To(BeEmpty())
			Expect(ingresses).To(BeEmpty())
			Expect(daemonSets).To(BeEmpty())
		})
	})

	When("Prometheus chart is disabled", func() {
		BeforeEach(func() {
			chartValues = support.Values{}
			_ = chartValues.SetValue(prometheusEnabled, false)
		})

		It("Should not contain Prometheus resources", func() {
			Expect(promEnabled).To(BeFalse())
			Expect(configMaps).To(BeEmpty())
			Expect(services).To(BeEmpty())
			Expect(deployments).To(BeEmpty())
			Expect(pvcs).To(BeEmpty())
			Expect(ingresses).To(BeEmpty())
			Expect(daemonSets).To(BeEmpty())
			Expect(statefulSets).To(BeEmpty())
		})
	})
})

func matchAllPrometheusElements(match gomega.OmegaMatcher, components ...string) gomega.OmegaMatcher {
	return MatchAllElements(prometheusComponent, matchAllElements(match, components...))
}

func prometheusComponent(elements interface{}) string {
	if o, ok := elements.(client.Object); ok {
		return o.GetLabels()[componentLabel]
	} else {
		return ""
	}
}
