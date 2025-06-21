package cmd

import (
	"github.com/roman-povoroznyk/k6s/pkg/config"
	"github.com/roman-povoroznyk/k6s/pkg/logger"
	"github.com/roman-povoroznyk/k6s/pkg/server"
	"github.com/roman-povoroznyk/k6s/pkg/version"
	"github.com/spf13/cobra"
)

var serverPort int

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the HTTP server",
	Long: `Start the HTTP server for handling API requests.
	
The server provides REST API endpoints for managing Kubernetes resources
and serves as the foundation for the controller functionality.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load configuration
		cfg, err := config.Load()
		if err != nil {
			logger.Fatal("Failed to load configuration", err, nil)
		}

		// Initialize logger
		logger.Init()

		// Use port from flag or config
		port := serverPort
		if port == 0 {
			port = 8080 // default port
		}

		logger.Info("Starting HTTP server", map[string]interface{}{
			"port":        port,
			"environment": cfg.App.Env,
			"version":     version.GetVersion(),
		})

		// Start server
		srv := server.New(port)
		if err := srv.Start(); err != nil {
			logger.Fatal("Server failed to start", err, map[string]interface{}{
				"port": port,
			})
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	
	// Add server-specific flags
	serverCmd.Flags().IntVarP(&serverPort, "port", "p", 8080, "Server port")
}
