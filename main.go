/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"errors"
	"flag"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	certmanagerv1alpha2 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha2"

	// miniov1 "github.com/minio/operator/pkg/apis/minio.min.io/v1"
	routev1 "github.com/openshift/api/route/v1"

	appsv1beta1 "gitlab.com/gitlab-org/cloud-native/gitlab-operator/api/v1beta1"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers"
	"gitlab.com/gitlab-org/cloud-native/gitlab-operator/controllers/settings"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	settings.Load()

	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(appsv1beta1.AddToScheme(scheme))

	utilruntime.Must(monitoringv1.AddToScheme(scheme))

	utilruntime.Must(certmanagerv1alpha2.AddToScheme(scheme))

	utilruntime.Must(routev1.AddToScheme(scheme))

	// utilruntime.Must(miniov1.AddToScheme(scheme))

	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	operatorScope := "namespace"
	watchedNamespace, err := getWatchedNamespace()
	if err != nil {
		operatorScope = "cluster"
	}

	setupLog.Info("setting operator scope", "scope", operatorScope)
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "852d23b0.gitlab.com",
		Namespace:              watchedNamespace,
		HealthProbeBindAddress: settings.HealthProbeBindAddress,
		ReadinessEndpointName:  settings.ReadinessEndpointName,
		LivenessEndpointName:   settings.LivenessEndpointName,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	mgr.AddHealthzCheck("healthz", settings.HealthzCheck)
	mgr.AddReadyzCheck("readyz", settings.ReadyzCheck)

	if err = (&controllers.GitLabReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("GitLab"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GitLab")
		os.Exit(1)
	}
	if err = (&appsv1beta1.GitLab{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "GitLab")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	// Report Operator as "alive" to probe.
	// TODO: put more thought into when the Operator should report liveness.
	settings.AliveStatus = nil

	// Report Operator as "ready" to probe.
	// TODO: put more thought into when the Operator should report readiness.
	settings.ReadyStatus = nil

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func getWatchedNamespace() (string, error) {
	ns, ok := os.LookupEnv("WATCH_NAMESPACE")
	if !ok {
		return "", errors.New("WATCH_NAMESPACE env required")
	}

	return ns, nil
}
