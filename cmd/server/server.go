package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
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
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start a FastHTTP server with Kubernetes controller manager",
	Long:  "Start a high-performance HTTP server using FastHTTP with controller-runtime manager for centralized control of informers and controllers",
	PreRunE: func(c *cobra.Command, args []string) error {
		// Setup logging only, skip global Kubernetes client initialization
		// as server command handles its own authentication (including in-cluster)
		return cmd.SetupLogging()
	},
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

		// Setup controller-runtime logger
		ctrlruntime.SetLogger(zap.New(zap.UseFlagOptions(&zap.Options{
			Development: true,
		})))

		// Get Kubernetes config
		config, err := getKubernetesConfig(kubeconfig, inCluster)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get Kubernetes config")
			return err
		}
		log.Debug().Str("host", config.Host).Msg("Successfully got Kubernetes config")

		log.Debug().Msg("Creating controller-runtime manager...")
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
		log.Debug().Msg("Controller-runtime manager created successfully")

		// Add controllers to manager
		if enableController {
			log.Debug().Msg("Adding deployment controller...")
			if err := controller.AddDeploymentController(mgr); err != nil {
				log.Error().Err(err).Msg("Failed to add deployment controller")
				return err
			}
			log.Debug().Msg("Deployment controller added successfully")
		}

		log.Debug().Msg("Creating Kubernetes client...")
		// Create Kubernetes client for informers (separate from controller-runtime)
		clientset, err := kubernetes.NewKubernetesClient(inCluster, kubeconfig)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create Kubernetes client")
			return err
		}
		log.Debug().Msg("Kubernetes client created successfully")

		log.Debug().Msg("Starting controller-runtime manager...")
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

		log.Debug().Msg("Creating informer manager...")
		log.Info().Msg("Creating informer manager...")
		// Create context for informers
		ctx := context.Background()
		informerConfig := informer.InformerConfig{
			Namespace:    namespace,
			ResyncPeriod: resyncPeriod,
		}

		// Start deployment informer if enabled
		if enableDeploymentInformer {
			log.Debug().Msg("Starting deployment informer...")
			log.Info().Msg("Starting deployment informer...")
			deploymentConfig := informer.DeploymentInformerConfig{
				Namespace:    informerConfig.Namespace,
				ResyncPeriod: informerConfig.ResyncPeriod,
			}
			// Run deployment informer in background goroutine
			go func() {
				if err := informer.StartDeploymentInformer(ctx, clientset, deploymentConfig); err != nil {
					log.Error().Err(err).Msg("Deployment informer failed")
				}
			}()
			log.Info().Msg("Deployment informer started successfully")
			log.Debug().Msg("Deployment informer completed, moving to pod informer...")
		}

		log.Debug().Msg("About to check pod informer flag...")
		log.Info().Bool("enabled", enablePodInformer).Msg("Pod informer flag status")

		// Start pod informer if enabled
		if enablePodInformer {
			log.Debug().Msg("Starting pod informer...")
			log.Info().Msg("Starting pod informer...")
			podConfig := informer.PodInformerConfig{
				Namespace:    informerConfig.Namespace,
				ResyncPeriod: informerConfig.ResyncPeriod,
			}
			// Run pod informer in background goroutine
			go func() {
				if err := informer.StartPodInformer(ctx, clientset, podConfig); err != nil {
					log.Error().Err(err).Msg("Pod informer failed")
				}
			}()
			log.Info().Msg("Pod informer started successfully")
		}
		log.Debug().Msg("All informers started successfully")
		log.Info().Msg("All informers started successfully")

		log.Debug().Msg("Starting HTTP server...")
		log.Info().Msg("Creating HTTP server configuration...")
		// Create and start HTTP server
		serverConfig := server.Config{
			Port:            port,
			ShutdownTimeout: timeout,
		}

		// Start the HTTP server in background first, before waiting for informers
		log.Info().
			Int("port", port).
			Str("namespace", namespace).
			Msg("Starting HTTP server in background...")
		
		log.Debug().Msg("About to call server.Start() in background goroutine")
		log.Info().Msg("About to call server.Start() in background goroutine...")
		
		// Start HTTP server in a separate goroutine
		serverErrChan := make(chan error, 1)
		go func() {
			if err := server.Start(serverConfig); err != nil {
				log.Error().Err(err).Msg("HTTP server failed")
				serverErrChan <- err
			}
		}()
		
		// Give server a moment to start
		time.Sleep(2 * time.Second)
		log.Info().Msg("HTTP server should be running now")
		
		// Wait for either server error or completion (never happens in normal flow)
		select {
		case err := <-serverErrChan:
			log.Error().Err(err).Msg("HTTP server failed to start")
			return err
		default:
			log.Info().Msg("HTTP server started successfully in background")
		}

		// Keep the main process running - wait for interrupt signal
		log.Info().Msg("All services started successfully. Waiting for interrupt signal...")
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		
		log.Info().Msg("Received interrupt signal, shutting down gracefully...")
		return nil
	},
}

// getKubernetesConfig returns the Kubernetes configuration for controller-runtime
func getKubernetesConfig(kubeconfigPath string, inCluster bool) (*rest.Config, error) {
	if inCluster {
		return rest.InClusterConfig()
	}
	if kubeconfigPath == "" {
		kubeconfigPath = clientcmd.RecommendedHomeFile
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
