package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tcpdump-to-jmx/internal/api"
	"github.com/tcpdump-to-jmx/internal/config"
	"github.com/tcpdump-to-jmx/internal/storage"
)

var serverPort int

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the web server for API access",
	Long: `Start the tcpdump-to-jmx web server that provides REST API endpoints
for converting PCAP files to HAR and JMX formats.

The server provides the following endpoints:
- POST /api/v1/convert - Upload and convert PCAP file
- GET /api/v1/status/:jobId - Check conversion status
- GET /api/v1/download/:jobId/:fileType - Download converted file
- GET /api/v1/health - Health check endpoint

Example:
  # Start server on default port (8080)
  tcpdump-to-jmx server

  # Start server on custom port
  tcpdump-to-jmx server --port 9090`,
	RunE: runServer,
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Server flags
	serverCmd.Flags().IntVarP(&serverPort, "port", "p", 0, "Server port (overrides config)")
}

func runServer(cmd *cobra.Command, args []string) error {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found")
	}

	// Load configuration
	cfg := config.Load()

	// Override port if specified
	if serverPort > 0 {
		cfg.Server.Port = serverPort
	}

	// Initialize S3 storage
	s3Storage, err := storage.NewS3Storage(cfg.AWS)
	if err != nil {
		return fmt.Errorf("failed to initialize S3 storage: %v", err)
	}

	// Set Gin mode
	if cfg.Server.Mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.New()
	
	// Add middleware
	router.Use(gin.Recovery())
	router.Use(api.LoggerMiddleware())
	router.Use(api.CORSMiddleware())

	// Initialize API handlers
	apiHandler := api.NewHandler(s3Storage, cfg)
	
	// Register routes
	v1 := router.Group("/api/v1")
	{
		// Health check
		v1.GET("/health", apiHandler.HealthCheck)
		
		// Conversion endpoints
		v1.POST("/convert", apiHandler.ConvertTCPDump)
		v1.GET("/status/:jobId", apiHandler.GetJobStatus)
		v1.GET("/download/:jobId/:fileType", apiHandler.DownloadFile)
		
		// WebSocket for real-time progress
		v1.GET("/ws/:jobId", apiHandler.WebSocketHandler)
		
		// List conversions
		v1.GET("/conversions", apiHandler.ListConversions)
	}

	// Serve Swagger documentation
	router.Static("/docs", "./docs")
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/docs")
	})

	// Create HTTP server
	srv := &http.Server{
		Addr:           fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:        router,
		ReadTimeout:    time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(cfg.Server.WriteTimeout) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Start server in goroutine
	go func() {
		logrus.Infof("Starting server on port %d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	logrus.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %v", err)
	}

	logrus.Info("Server exited")
	return nil
}