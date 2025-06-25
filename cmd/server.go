package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientconfig"
	"kubernetes-controller/pkg/api"
	"kubernetes-controller/pkg/logger"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start HTTP server",
	Long:  "Start HTTP server with RESTful API for Kubernetes resources",
	RunE:  runServer,
}

func init() {
	rootCmd.AddCommand(serverCmd)
	
	serverCmd.Flags().String("host", "localhost", "Server host")
	serverCmd.Flags().Int("port", 8080, "Server port")
	viper.BindPFlag("server.host", serverCmd.Flags().Lookup("host"))
	viper.BindPFlag("server.port", serverCmd.Flags().Lookup("port"))
}

func runServer(cmd *cobra.Command, args []string) error {
	// Setup logger
	logger.SetupLogger()

	// Load kubeconfig
	config, err := clientconfig.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load kubeconfig")
	}

	// Create Kubernetes client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kubernetes client")
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			
			log.Error().Err(err).Int("status", code).Str("path", c.Path()).Msg("HTTP error")
			
			return c.Status(code).JSON(api.ErrorResponse{
				Error:   "Internal Server Error",
				Code:    code,
				Message: err.Error(),
			})
		},
	})

	// Middleware
	app.Use(recover.New())
	app.Use(cors.New())
	app.Use(loggingMiddleware())

	// API handler
	apiHandler := api.NewAPIHandler(clientset)

	// Routes
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "healthy"})
	})

	// API v1 routes
	v1 := app.Group("/api/v1")
	v1.Get("/deployments", apiHandler.GetDeployments)
	v1.Get("/deployments/:namespace/:name", apiHandler.GetDeployment)
	v1.Post("/deployments", apiHandler.CreateDeployment)
	v1.Delete("/deployments/:namespace/:name", apiHandler.DeleteDeployment)

	// Start server
	host := viper.GetString("server.host")
	port := viper.GetInt("server.port")
	addr := fmt.Sprintf("%s:%d", host, port)

	log.Info().Str("address", addr).Msg("Starting HTTP server")

	// Graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := app.Listen(addr); err != nil {
			log.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Info().Msg("Shutting down server...")
	
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server stopped")
	return nil
}

func loggingMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		log.Info().
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", c.Response().StatusCode()).
			Dur("duration", time.Since(start)).
			Str("ip", c.IP()).
			Msg("HTTP request")

		return err
	}
}
