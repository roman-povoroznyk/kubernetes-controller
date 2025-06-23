package server

import (
	"time"

	"github.com/roman-povoroznyk/kubernetes-controller/cmd"
	"github.com/roman-povoroznyk/kubernetes-controller/internal/server"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start a FastHTTP server",
	Long:  "Start a high-performance HTTP server using FastHTTP",
	RunE: func(c *cobra.Command, args []string) error {
		port, _ := c.Flags().GetInt("server-port")
		timeout, _ := c.Flags().GetDuration("shutdown-timeout")

		config := server.Config{
			Port:            port,
			ShutdownTimeout: timeout,
		}

		log.Info().Int("port", port).Msg("Starting server")
		return server.Start(config)
	},
}

func init() {
	serverCmd.Flags().Int("server-port", 8080, "HTTP server port")
	serverCmd.Flags().Duration("shutdown-timeout", 5*time.Second, "Server shutdown timeout")

	cmd.RootCmd.AddCommand(serverCmd)
}
