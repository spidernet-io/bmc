package main

import (
	"flag"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	bmcv1beta1 "github.com/spidernet-io/bmc/pkg/k8s/apis/bmc.spidernet.io/v1beta1"
	controller "github.com/spidernet-io/bmc/pkg/controller/clusteragent"
	"github.com/spidernet-io/bmc/pkg/log"
	clusteragentwebhook "github.com/spidernet-io/bmc/pkg/webhook/clusteragent"
	hostendpointwebhook "github.com/spidernet-io/bmc/pkg/webhook/hostendpoint"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(bmcv1beta1.AddToScheme(scheme))
}

func main() {
	// Get log level from environment variable
	logLevel := os.Getenv("LOG_LEVEL")
	log.InitStdoutLogger(logLevel)

	// Set controller-runtime logger
	ctrl.SetLogger(zap.New())

	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var webhookPort int
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.IntVar(&webhookPort, "webhook-port", 443, "The port that the webhook server serves at.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	flag.Parse()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: metricsAddr,
		},
		WebhookServer: webhook.NewServer(webhook.Options{
			Port: webhookPort,
		}),
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "bmc-controller-lock",
	})
	if err != nil {
		log.Logger.Errorf("unable to start manager: %v", err)
		os.Exit(1)
	}

	if err = (&controller.ClusterAgentReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		log.Logger.Errorf("unable to create controller %s: %v", "ClusterAgent", err)
		os.Exit(1)
	}

	// Setup webhook
	if err = (&clusteragentwebhook.ClusterAgentWebhook{}).SetupWebhookWithManager(mgr); err != nil {
		log.Logger.Errorf("unable to create webhook %s: %v", "ClusterAgent", err)
		os.Exit(1)
	}

	// Setup HostEndpoint webhook
	if err = (&hostendpointwebhook.HostEndpointWebhook{}).SetupWebhookWithManager(mgr); err != nil {
		log.Logger.Errorf("unable to create webhook %s: %v", "HostEndpoint", err)
		os.Exit(1)
	}

	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		log.Logger.Errorf("unable to set up health check: %v", err)
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		log.Logger.Errorf("unable to set up ready check: %v", err)
		os.Exit(1)
	}

	log.Logger.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Logger.Errorf("problem running manager: %v", err)
		os.Exit(1)
	}
}
