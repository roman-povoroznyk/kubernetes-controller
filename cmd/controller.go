package cmd

import (
	"context"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientconfig"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"kubernetes-controller/pkg/business"
	"kubernetes-controller/pkg/controller"
	"kubernetes-controller/pkg/leader"
	"kubernetes-controller/pkg/logger"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(appsv1.AddToScheme(scheme))
}

var controllerCmd = &cobra.Command{
	Use:   "controller",
	Short: "Start controller-runtime based controller",
	Long:  "Start controller-runtime based controller for managing Kubernetes resources with leader election and metrics",
	RunE:  runController,
}

func init() {
	rootCmd.AddCommand(controllerCmd)
	
	controllerCmd.Flags().String("metrics-bind-address", ":8080", "The address the metric endpoint binds to")
	controllerCmd.Flags().String("health-probe-bind-address", ":8081", "The address the probe endpoint binds to")
	controllerCmd.Flags().Bool("leader-elect", false, "Enable leader election for controller manager")
	controllerCmd.Flags().String("leader-elect-namespace", "default", "Namespace for leader election")
}

func runController(cmd *cobra.Command, args []string) error {
	// Setup logger
	logger.SetupLogger()

	metricsAddr, _ := cmd.Flags().GetString("metrics-bind-address")
	probeAddr, _ := cmd.Flags().GetString("health-probe-bind-address")
	enableLeaderElection, _ := cmd.Flags().GetBool("leader-elect")
	leaderElectNamespace, _ := cmd.Flags().GetString("leader-elect-namespace")

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	// Load kubeconfig for leader election
	config, err := clientconfig.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load kubeconfig")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kubernetes client")
	}

	// Start metrics server
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Info().Str("address", metricsAddr).Msg("Starting metrics server")
		if err := http.ListenAndServe(metricsAddr, nil); err != nil {
			log.Error().Err(err).Msg("Failed to start metrics server")
		}
	}()

	if enableLeaderElection {
		// Setup leader election
		leaderElection := leader.NewLeaderElection(clientset, leaderElectNamespace, "kubernetes-controller")
		
		ctx := ctrl.SetupSignalHandler()
		
		return leaderElection.Run(ctx, 
			func(ctx context.Context) {
				// Start controller when becoming leader
				startController(ctx, metricsAddr, probeAddr)
			},
			func(ctx context.Context) {
				// Stop controller when losing leadership
				log.Info().Msg("Lost leadership, stopping controller")
			},
		)
	} else {
		// Start controller without leader election
		ctx := ctrl.SetupSignalHandler()
		return startController(ctx, metricsAddr, probeAddr)
	}
}

func startController(ctx context.Context, metricsAddr, probeAddr string) error {
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     "0", // Disable built-in metrics server
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         false, // Handled externally
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		return err
	}

	// Create business rule engine
	ruleEngine := business.NewRuleEngine()

	// Setup deployment reconciler
	if err = (&controller.DeploymentReconciler{
		Client:     mgr.GetClient(),
		Scheme:     mgr.GetScheme(),
		RuleEngine: ruleEngine,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Deployment")
		return err
	}

	// Add health checks
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		return err
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		return err
	}

	setupLog.Info("starting manager")
	return mgr.Start(ctx)
}
