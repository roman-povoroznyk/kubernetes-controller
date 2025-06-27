package server

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/roman-povoroznyk/kubernetes-controller/cmd"
	"github.com/roman-povoroznyk/kubernetes-controller/internal/controller"
	"github.com/roman-povoroznyk/kubernetes-controller/internal/informer"
	"github.com/roman-povoroznyk/kubernetes-controller/internal/kubernetes"
	"github.com/roman-povoroznyk/kubernetes-controller/internal/server"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrlruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start a FastHTTP server with Kubernetes controller manager",
	Long:  "Start a high-performance HTTP server using FastHTTP with controller-runtime manager for centralized control of informers and controllers",
	RunE: func(c *cobra.Command, args []string) error {
		port, _ := c.Flags().GetInt("server-port")
		timeout, _ := c.Flags().GetDuration("shutdown-timeout")
		kubeconfig, _ := c.Flags().GetString("kubeconfig")
		inCluster, _ := c.Flags().GetBool("in-cluster")
		namespace, _ := c.Flags().GetString("namespace")
		resyncPeriod, _ := c.Flags().GetDuration("resync-period")
		enableDeploymentInformer, _ := c.Flags().GetBool("enable-deployment-informer")
		enablePodInformer, _ := c.Flags().GetBool("enable-pod-informer")
		enableController, _ := c.Flags().GetBool("enable-controller")
		enableLeaderElection, _ := c.Flags().GetBool("enable-leader-election")
		leaderElectionNamespace, _ := c.Flags().GetString("leader-election-namespace")
		metricsPort, _ := c.Flags().GetInt("metrics-port")

		log.Info().
			Int("server-port", port).
			Bool("leader-election", enableLeaderElection).
			Int("metrics-port", metricsPort).
			Str("namespace", namespace).
			Msg("Starting kubernetes controller server")

		// Get Kubernetes config
		config, err := getKubernetesConfig(kubeconfig, inCluster)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get Kubernetes config")
			return err
		}

		// Create controller-runtime manager with leader election and metrics
		mgrOptions := manager.Options{
			LeaderElection:          enableLeaderElection,
			LeaderElectionID:        "k8s-ctrl-leader-election",
			LeaderElectionNamespace: leaderElectionNamespace,
			Metrics: metricsserver.Options{
				BindAddress: fmt.Sprintf(":%d", metricsPort),
			},
		}

		// Disable leader election for local development if requested
		if !enableLeaderElection {
			log.Info().Msg("Leader election disabled - running without leader election")
		}

		mgr, err := ctrlruntime.NewManager(config, mgrOptions)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create controller-runtime manager")
			return err
		}
		// Add controllers to manager
		if enableController {
			if err := controller.AddDeploymentController(mgr); err != nil {
				log.Error().Err(err).Msg("Failed to add deployment controller")
				return err
			}
		}

		// Create Kubernetes client for informers (separate from controller-runtime)
		clientset, err := kubernetes.NewKubernetesClient(inCluster, kubeconfig)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create Kubernetes client")
			return err
		}

		// Start controller-runtime manager in background
		go func() {
			log.Info().
				Bool("leader-election", enableLeaderElection).
				Str("leader-election-namespace", leaderElectionNamespace).
				Msg("Starting controller-runtime manager...")
			if err := mgr.Start(context.Background()); err != nil {
				log.Error().Err(err).Msg("Manager exited with error")
				os.Exit(1)
			}
		}()

		// Add informers to manager for centralized control
		if enableDeploymentInformer || enablePodInformer {
			informerMgr := controller.NewInformerManager(clientset, namespace, resyncPeriod)
			if err := informerMgr.AddToManager(mgr, enableDeploymentInformer, enablePodInformer); err != nil {
				log.Error().Err(err).Msg("Failed to add informers to manager")
				return err
			}
			log.Info().Msg("Informers added to controller-runtime manager")
		}

		// Fallback to standalone informers if needed (backwards compatibility)
		if false { // Disable standalone informers - use manager-integrated approach
			ctx := context.Background()

			if enableDeploymentInformer {
				informerConfig := informer.DeploymentInformerConfig{
					Namespace:    namespace,
					ResyncPeriod: resyncPeriod,
				}

				go func() {
					if err := informer.StartDeploymentInformer(ctx, clientset, informerConfig); err != nil {
						log.Error().Err(err).Msg("Failed to start deployment informer")
					}
				}()
			}

			if enablePodInformer {
				podInformerConfig := informer.PodInformerConfig{
					Namespace:    namespace,
					ResyncPeriod: resyncPeriod,
				}

				go func() {
					if err := informer.StartPodInformer(ctx, clientset, podInformerConfig); err != nil {
						log.Error().Err(err).Msg("Failed to start pod informer")
					}
				}()
			}
		}

		serverConfig := server.Config{
			Port:            port,
			ShutdownTimeout: timeout,
		}

		log.Info().
			Int("port", port).
			Str("namespace", namespace).
			Bool("deployment_informer", enableDeploymentInformer).
			Bool("pod_informer", enablePodInformer).
			Bool("controller", enableController).
			Bool("leader_election", enableLeaderElection).
			Int("metrics_port", metricsPort).
			Msg("Starting server with informers and controllers")
		return server.Start(serverConfig)
	},
}

// getKubernetesConfig returns the Kubernetes configuration for controller-runtime
func getKubernetesConfig(kubeconfigPath string, inCluster bool) (*rest.Config, error) {
	if inCluster {
		return rest.InClusterConfig()
	}
	return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
}

func init() {
	serverCmd.Flags().Int("server-port", 8080, "HTTP server port")
	serverCmd.Flags().Duration("shutdown-timeout", 5*time.Second, "Server shutdown timeout")
	serverCmd.Flags().String("kubeconfig", "", "Path to kubeconfig file (default: ~/.kube/config)")
	serverCmd.Flags().Bool("in-cluster", false, "Use in-cluster Kubernetes authentication")
	serverCmd.Flags().String("namespace", "default", "Kubernetes namespace to watch")
	serverCmd.Flags().Duration("resync-period", 30*time.Second, "Informer resync period")
	serverCmd.Flags().Bool("enable-deployment-informer", true, "Enable deployment informer")
	serverCmd.Flags().Bool("enable-pod-informer", true, "Enable pod informer")
	serverCmd.Flags().Bool("enable-controller", true, "Enable controller-runtime deployment controller")
	serverCmd.Flags().Bool("enable-leader-election", true, "Enable leader election for controller manager")
	serverCmd.Flags().String("leader-election-namespace", "default", "Namespace for leader election lease resource")
	serverCmd.Flags().Int("metrics-port", 8081, "Port for controller manager metrics server")

	cmd.RootCmd.AddCommand(serverCmd)
}
