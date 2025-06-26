package server

import (
	"context"
	"time"

	"github.com/roman-povoroznyk/kubernetes-controller/cmd"
	"github.com/roman-povoroznyk/kubernetes-controller/internal/informer"
	"github.com/roman-povoroznyk/kubernetes-controller/internal/kubernetes"
	"github.com/roman-povoroznyk/kubernetes-controller/internal/server"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start a FastHTTP server with Kubernetes informers",
	Long:  "Start a high-performance HTTP server using FastHTTP with configurable deployment and pod informers for monitoring Kubernetes events",
	RunE: func(c *cobra.Command, args []string) error {
		port, _ := c.Flags().GetInt("server-port")
		timeout, _ := c.Flags().GetDuration("shutdown-timeout")
		kubeconfig, _ := c.Flags().GetString("kubeconfig")
		inCluster, _ := c.Flags().GetBool("in-cluster")
		namespace, _ := c.Flags().GetString("namespace")
		resyncPeriod, _ := c.Flags().GetDuration("resync-period")
		enableDeploymentInformer, _ := c.Flags().GetBool("enable-deployment-informer")
		enablePodInformer, _ := c.Flags().GetBool("enable-pod-informer")

		// Create Kubernetes client
		clientset, err := kubernetes.NewKubernetesClient(inCluster, kubeconfig)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create Kubernetes client")
			return err
		}		// Start informers based on flags
		ctx := context.Background()
		
		if enableDeploymentInformer {
			// Start deployment informer in background
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
			// Start pod informer in background  
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

		config := server.Config{
			Port:            port,
			ShutdownTimeout: timeout,
		}

		log.Info().
			Int("port", port).
			Str("namespace", namespace).
			Bool("deployment_informer", enableDeploymentInformer).
			Bool("pod_informer", enablePodInformer).
			Msg("Starting server with informers")
		return server.Start(config)
	},
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

	cmd.RootCmd.AddCommand(serverCmd)
}
