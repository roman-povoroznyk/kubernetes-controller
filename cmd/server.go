package cmd

import (
	"github.com/roman-povoroznyk/kubernetes-controller/k6s/pkg/logger"
	"github.com/roman-povoroznyk/kubernetes-controller/k6s/pkg/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serverPort int

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "start the HTTP server",
	Long: `Start the HTTP server for handling API requests.
	
The server provides REST API endpoints for health checks, version info,
and serves as the foundation for the controller functionality.

Examples:
  k6s server                        # start server on default port 8080
  k6s server --port 9090            # start server on port 9090
  K6S_SERVER_PORT=8081 k6s server   # start server using env var`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get port from viper (supports env vars, config files, and flags)
		port := viper.GetInt("server.port")
		if port == 0 {
			port = serverPort // fallback to flag value
		}
		
		logger.Info("Starting k6s server", map[string]interface{}{
			"component": "server",
			"port":      port,
			"version":   Version,
		})

		// Create and start server
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
	serverCmd.Flags().IntVarP(&serverPort, "port", "p", 8080, "server port")
	
	// Bind flag to viper for environment variable support
	if err := viper.BindPFlag("server.port", serverCmd.Flags().Lookup("port")); err != nil {
		logger.Error("Failed to bind port flag", err, nil)
	}
	
	// Allow environment variables like K6S_SERVER_PORT
	if err := viper.BindEnv("server.port", "K6S_SERVER_PORT"); err != nil {
		logger.Error("Failed to bind server port env", err, nil)
	}
}
