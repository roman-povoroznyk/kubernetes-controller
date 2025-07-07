package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/roman-povoroznyk/kubernetes-controller/k6s/pkg/config"
	"github.com/roman-povoroznyk/kubernetes-controller/k6s/pkg/logger"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/tools/clientcmd"
)

// clusterCmd represents the cluster command group
var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Manage cluster configurations for multi-cluster support",
	Long: `Cluster management commands for configuring and managing multiple 
Kubernetes clusters with the k6s controller.

The cluster configuration is stored in a unified YAML file that contains
all cluster definitions, connection settings, and operational parameters.`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

// addClusterCmd represents the add cluster command
var addClusterCmd = &cobra.Command{
	Use:   "add NAME",
	Short: "Add a new cluster to the configuration",
	Long: `Add a new cluster to the multi-cluster configuration.

Examples:
  # Add a cluster with auto-detected kubeconfig
  k6s cluster add production

  # Add a cluster with specific kubeconfig
  k6s cluster add staging --kubeconfig ~/.kube/staging-config

  # Add a cluster with specific context
  k6s cluster add dev --context dev-cluster

  # Add a cluster and set as primary
  k6s cluster add prod --primary

  # Add a cluster but keep it disabled initially
  k6s cluster add test --disabled`,
	Args: cobra.ExactArgs(1),
	RunE: addCluster,
}

// listClustersCmd represents the list clusters command
var listClustersCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured clusters",
	Long: `List all clusters in the configuration with their status and settings.

The output shows:
- Cluster name
- Enabled/Disabled status
- Primary designation
- Context information
- Namespace settings`,
	RunE: listClusters,
}

// deleteClusterCmd represents the delete cluster command
var deleteClusterCmd = &cobra.Command{
	Use:     "delete NAME",
	Aliases: []string{"remove", "rm"},
	Short:   "Delete a cluster from the configuration",
	Long: `Delete a cluster from the multi-cluster configuration.

Examples:
  # Delete a cluster
  k6s cluster delete staging

  # Delete multiple clusters
  k6s cluster delete dev test`,
	Args: cobra.MinimumNArgs(1),
	RunE: deleteCluster,
}

// enableClusterCmd represents the enable cluster command
var enableClusterCmd = &cobra.Command{
	Use:   "enable NAME",
	Short: "Enable a cluster",
	Long:  `Enable a cluster so it will be used by the multi-cluster controller.`,
	Args:  cobra.ExactArgs(1),
	RunE:  enableCluster,
}

// disableClusterCmd represents the disable cluster command
var disableClusterCmd = &cobra.Command{
	Use:   "disable NAME",
	Short: "Disable a cluster",
	Long:  `Disable a cluster so it will be ignored by the multi-cluster controller.`,
	Args:  cobra.ExactArgs(1),
	RunE:  disableCluster,
}

// setPrimaryCmd represents the set-primary command
var setPrimaryCmd = &cobra.Command{
	Use:   "set-primary NAME",
	Short: "Set a cluster as the primary cluster",
	Long: `Set a cluster as the primary cluster. Only one cluster can be primary.
The primary cluster is used for certain operations that require a single cluster context.`,
	Args: cobra.ExactArgs(1),
	RunE: setPrimaryCluster,
}

// checkConnectivityCmd represents the check-connectivity command
var checkConnectivityCmd = &cobra.Command{
	Use:     "check-connectivity [NAME]",
	Aliases: []string{"check"},
	Short:   "Check connectivity to clusters",
	Long: `Check connectivity to one or all clusters.

Examples:
  # Check connectivity to all clusters
  k6s cluster check-connectivity

  # Check connectivity to a specific cluster
  k6s cluster check-connectivity production`,
	Args: cobra.MaximumNArgs(1),
	RunE: checkConnectivity,
}

var (
	// Flags for add command
	addKubeconfig   string
	addContext      string
	addNamespace    string
	addPrimary      bool
	addDisabled     bool
	skipConnectivity bool
)

func init() {
	// Add subcommands
	clusterCmd.AddCommand(addClusterCmd)
	clusterCmd.AddCommand(listClustersCmd)
	clusterCmd.AddCommand(deleteClusterCmd)
	clusterCmd.AddCommand(enableClusterCmd)
	clusterCmd.AddCommand(disableClusterCmd)
	clusterCmd.AddCommand(setPrimaryCmd)
	clusterCmd.AddCommand(checkConnectivityCmd)

	// Flags for add command
	addClusterCmd.Flags().StringVar(&addKubeconfig, "kubeconfig", "", "path to kubeconfig file (default: auto-detect)")
	addClusterCmd.Flags().StringVar(&addContext, "context", "", "kubeconfig context to use (default: current-context)")
	addClusterCmd.Flags().StringVar(&addNamespace, "namespace", "", "default namespace for the cluster (default: default)")
	addClusterCmd.Flags().BoolVar(&addPrimary, "primary", false, "set this cluster as primary")
	addClusterCmd.Flags().BoolVar(&addDisabled, "disabled", false, "add cluster in disabled state")
	addClusterCmd.Flags().BoolVar(&skipConnectivity, "skip-connectivity", false, "skip connectivity test when adding cluster")
}

func addCluster(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Load or create configuration
	cfg, err := loadMultiClusterConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Check if cluster already exists
	for _, cluster := range cfg.MultiCluster.Clusters {
		if cluster.Name == name {
			return fmt.Errorf("cluster '%s' already exists", name)
		}
	}

	// Determine kubeconfig path
	kubeconfigPath := addKubeconfig
	if kubeconfigPath == "" {
		if env := os.Getenv("KUBECONFIG"); env != "" {
			kubeconfigPath = env
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			kubeconfigPath = filepath.Join(home, ".kube", "config")
		}
	}

	// Test connectivity unless skipped
	if !skipConnectivity {
		logger.Info("Testing connectivity to cluster", map[string]interface{}{
			"cluster":    name,
			"kubeconfig": kubeconfigPath,
			"context":    addContext,
		})

		if err := testClusterConnectivity(kubeconfigPath, addContext); err != nil {
			return fmt.Errorf("connectivity test failed for cluster '%s': %w", name, err)
		}
	}

	// If setting as primary, unset other primary clusters
	if addPrimary {
		for i := range cfg.MultiCluster.Clusters {
			cfg.MultiCluster.Clusters[i].Primary = false
		}
	}

	// Create new cluster config
	clusterConfig := config.ClusterConfig{
		Name:       name,
		KubeConfig: kubeconfigPath,
		Context:    addContext,
		Namespace:  addNamespace,
		Enabled:    !addDisabled,
		Primary:    addPrimary,
	}

	// Add to configuration
	cfg.MultiCluster.Clusters = append(cfg.MultiCluster.Clusters, clusterConfig)

	// Save configuration
	if err := saveMultiClusterConfig(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Output confirmation in kubectl style
	fmt.Printf("cluster/%s created\n", name)

	return nil
}

func listClusters(cmd *cobra.Command, args []string) error {
	cfg, err := loadMultiClusterConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if len(cfg.MultiCluster.Clusters) == 0 {
		fmt.Println("No clusters configured")
		return nil
	}

	// Print header
	fmt.Printf("%-20s %-10s %-10s %-30s %-15s\n", "NAME", "STATUS", "PRIMARY", "CONTEXT", "NAMESPACE")
	fmt.Printf("%-20s %-10s %-10s %-30s %-15s\n", "----", "------", "-------", "-------", "---------")

	// Print clusters
	for _, cluster := range cfg.MultiCluster.Clusters {
		status := "Disabled"
		if cluster.Enabled {
			status = "Enabled"
		}

		primary := ""
		if cluster.Primary {
			primary = "Yes"
		}

		context := cluster.Context
		if context == "" {
			context = "<current>"
		}

		namespace := cluster.Namespace
		if namespace == "" {
			namespace = "default"
		}

		fmt.Printf("%-20s %-10s %-10s %-30s %-15s\n", cluster.Name, status, primary, context, namespace)
	}

	return nil
}

func deleteCluster(cmd *cobra.Command, args []string) error {
	cfg, err := loadMultiClusterConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	var deleted []string
	var notFound []string

	for _, name := range args {
		found := false
		for i, cluster := range cfg.MultiCluster.Clusters {
			if cluster.Name == name {
				// Remove cluster from slice
				cfg.MultiCluster.Clusters = append(cfg.MultiCluster.Clusters[:i], cfg.MultiCluster.Clusters[i+1:]...)
				deleted = append(deleted, name)
				found = true
				break
			}
		}
		if !found {
			notFound = append(notFound, name)
		}
	}

	if len(deleted) > 0 {
		// Save configuration
		if err := saveMultiClusterConfig(cfg); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		// Output confirmation
		for _, name := range deleted {
			fmt.Printf("cluster/%s deleted\n", name)
		}
	}

	if len(notFound) > 0 {
		for _, name := range notFound {
			fmt.Printf("cluster/%s not found\n", name)
		}
	}

	return nil
}

func enableCluster(cmd *cobra.Command, args []string) error {
	return updateClusterStatus(args[0], true, "enabled")
}

func disableCluster(cmd *cobra.Command, args []string) error {
	return updateClusterStatus(args[0], false, "disabled")
}

func updateClusterStatus(name string, enabled bool, action string) error {
	cfg, err := loadMultiClusterConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	found := false
	for i, cluster := range cfg.MultiCluster.Clusters {
		if cluster.Name == name {
			cfg.MultiCluster.Clusters[i].Enabled = enabled
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("cluster '%s' not found", name)
	}

	// Save configuration
	if err := saveMultiClusterConfig(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("cluster/%s %s\n", name, action)
	return nil
}

func setPrimaryCluster(cmd *cobra.Command, args []string) error {
	name := args[0]
	cfg, err := loadMultiClusterConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	found := false
	for i, cluster := range cfg.MultiCluster.Clusters {
		if cluster.Name == name {
			// Unset all other primary flags
			for j := range cfg.MultiCluster.Clusters {
				cfg.MultiCluster.Clusters[j].Primary = false
			}
			// Set this cluster as primary
			cfg.MultiCluster.Clusters[i].Primary = true
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("cluster '%s' not found", name)
	}

	// Save configuration
	if err := saveMultiClusterConfig(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("cluster/%s set as primary\n", name)
	return nil
}

func checkConnectivity(cmd *cobra.Command, args []string) error {
	cfg, err := loadMultiClusterConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	var clustersToCheck []config.ClusterConfig

	if len(args) == 1 {
		// Check specific cluster
		name := args[0]
		found := false
		for _, cluster := range cfg.MultiCluster.Clusters {
			if cluster.Name == name {
				clustersToCheck = append(clustersToCheck, cluster)
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("cluster '%s' not found", name)
		}
	} else {
		// Check all clusters
		clustersToCheck = cfg.MultiCluster.Clusters
	}

	if len(clustersToCheck) == 0 {
		fmt.Println("No clusters to check")
		return nil
	}

	fmt.Printf("%-20s %-15s %-50s\n", "NAME", "STATUS", "MESSAGE")
	fmt.Printf("%-20s %-15s %-50s\n", "----", "------", "-------")

	for _, cluster := range clustersToCheck {
		status := "Reachable"
		message := "Connection successful"

		err := testClusterConnectivity(cluster.KubeConfig, cluster.Context)
		if err != nil {
			status = "Unreachable"
			message = err.Error()
			if len(message) > 47 {
				message = message[:47] + "..."
			}
		}

		fmt.Printf("%-20s %-15s %-50s\n", cluster.Name, status, message)
	}

	return nil
}

func testClusterConnectivity(kubeconfigPath, contextName string) error {
	// Build config from kubeconfig
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{CurrentContext: contextName}).ClientConfig()
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Set a reasonable timeout
	config.Timeout = 5 * time.Second

	return nil
}

func loadMultiClusterConfig() (*config.Config, error) {
	// Default config path
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(home, ".k6s", "k6s.yaml")

	// Check if config exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config
		cfg := config.DefaultConfig()
		
		// Ensure directory exists
		if err := os.MkdirAll(filepath.Dir(configPath), 0750); err != nil {
			return nil, fmt.Errorf("failed to create config directory: %w", err)
		}

		return cfg, nil
	}

	// Load existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

func saveMultiClusterConfig(cfg *config.Config) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(home, ".k6s", "k6s.yaml")

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write file
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
