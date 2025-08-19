package main

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
	"github.com/tcpdump-to-jmx/internal/api"
	"github.com/tcpdump-to-jmx/internal/config"
	"github.com/tcpdump-to-jmx/internal/storage"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found")
	}

	// Initialize logger
	initLogger()

	// Load configuration
	cfg := config.Load()

	// Initialize S3 storage
	s3Storage, err := storage.NewS3Storage(cfg.AWS)
	if err != nil {
		logrus.Fatalf("Failed to initialize S3 storage: %v", err)
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
		logrus.Fatalf("Server forced to shutdown: %v", err)
	}

	logrus.Info("Server exited")
}

func initLogger() {
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})
	
	level := os.Getenv("LOG_LEVEL")
	switch level {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
}