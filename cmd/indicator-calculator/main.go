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
	"github.com/stock-market-system/internal/indicator-calculator/handlers"
	"github.com/stock-market-system/internal/indicator-calculator/service"
	"github.com/stock-market-system/pkg/config"
	"github.com/stock-market-system/pkg/logger"
	"github.com/stock-market-system/pkg/messaging"
	"github.com/stock-market-system/pkg/storage/clickhouse"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config.indicator-calculator.yaml")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.NewLogger(cfg.Logger)
	log.Info("Starting Indicator Calculator Service")

	// Initialize ClickHouse client
	chClient, err := clickhouse.NewClient(&cfg.ClickHouse)
	if err != nil {
		log.WithError(err).Fatal("Failed to create ClickHouse client")
	}
	defer chClient.Close()
	log.Info("Connected to ClickHouse")

	// Initialize Kafka producer
	producer, err := messaging.NewProducer(&cfg.Kafka)
	if err != nil {
		log.WithError(err).Fatal("Failed to create Kafka producer")
	}
	defer producer.Close()
	log.Info("Connected to Kafka producer")

	// Initialize Kafka consumer
	consumer, err := messaging.NewConsumer(&cfg.Kafka)
	if err != nil {
		log.WithError(err).Fatal("Failed to create Kafka consumer")
	}
	defer consumer.Close()
	log.Info("Connected to Kafka consumer")

	// Initialize calculator service
	calculatorSvc := service.NewCalculatorService(cfg, log, chClient, producer, consumer)

	// Start Kafka consumer in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		log.Info("Starting Kafka consumer for data collected events")
		if err := calculatorSvc.Start(ctx); err != nil {
			log.WithError(err).Error("Kafka consumer error")
		}
	}()

	// Initialize HTTP server
	router := setupRouter(calculatorSvc, log)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start HTTP server in background
	go func() {
		log.WithField("port", cfg.Server.Port).Info("Starting HTTP server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("HTTP server error")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Cancel context to stop Kafka consumer
	cancel()

	// Graceful shutdown of HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.WithError(err).Error("Server forced to shutdown")
	}

	log.Info("Server exited")
}

// setupRouter configures the HTTP router
func setupRouter(calculatorSvc *service.CalculatorService, log *logger.Logger) *gin.Engine {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Recovery())

	// Custom logger middleware
	router.Use(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		duration := time.Since(start)
		log.WithFields(map[string]interface{}{
			"method":   c.Request.Method,
			"path":     path,
			"status":   c.Writer.Status(),
			"duration": duration.Milliseconds(),
		}).Info("HTTP request")
	})

	// Initialize handler
	calculatorHandler := handlers.NewCalculatorHandler(calculatorSvc, log)

	// Health check
	router.GET("/health", calculatorHandler.Health)

	// API routes
	api := router.Group("/api/v1")
	{
		api.POST("/calculate", calculatorHandler.Calculate)
	}

	return router
}
