package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stock-market-system/internal/data-collector/service"
	"github.com/stock-market-system/pkg/config"
	"github.com/stock-market-system/pkg/logger"
	"github.com/stock-market-system/pkg/messaging"
	"github.com/stock-market-system/pkg/storage/clickhouse"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	appLogger := logger.NewLogger(cfg.Logger)
	appLogger.Info("Starting Data Collector Service")

	// Initialize ClickHouse client
	chClient, err := clickhouse.NewClient(&cfg.ClickHouse)
	if err != nil {
		appLogger.Fatalf("Failed to connect to ClickHouse: %v", err)
	}
	defer chClient.Close()

	appLogger.Info("Connected to ClickHouse successfully")

	// Initialize Kafka producer
	kafkaProducer, err := messaging.NewProducer(&cfg.Kafka)
	if err != nil {
		appLogger.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer kafkaProducer.Close()

	appLogger.Info("Kafka producer initialized successfully")

	// Initialize collector service
	collectorService := service.NewCollectorService(chClient, kafkaProducer, appLogger, 100)
	collectorService.Start()
	defer collectorService.Stop()

	// Initialize Gin router
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "data-collector",
			"timestamp": time.Now().Unix(),
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Collect data for specific tickers
		v1.POST("/collect", func(c *gin.Context) {
			var req struct{
				Exchange   string    `json:"exchange" binding:"required"`
				Tickers    []string  `json:"tickers" binding:"required"`
				StartDate  string    `json:"start_date"`
				EndDate    string    `json:"end_date"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// Parse dates
			startDate := time.Now().AddDate(0, -6, 0) // Default 6 months ago
			endDate := time.Now()

			if req.StartDate != "" {
				if t, err := time.Parse("2006-01-02", req.StartDate); err == nil {
					startDate = t
				}
			}

			if req.EndDate != "" {
				if t, err := time.Parse("2006-01-02", req.EndDate); err == nil {
					endDate = t
				}
			}

			// Start collection in background
			go func() {
				ctx := context.Background()
				_, err := collectorService.CollectMOEXData(ctx, req.Tickers, startDate, endDate)
				if err != nil {
					appLogger.Errorf("Collection job failed: %v", err)
				}
			}()

			c.JSON(http.StatusAccepted, gin.H{
				"message": "Collection started",
				"tickers": len(req.Tickers),
				"start_date": startDate.Format("2006-01-02"),
				"end_date": endDate.Format("2006-01-02"),
			})
		})

		// Get last collection date for ticker
		v1.GET("/last-date/:exchange/:ticker", func(c *gin.Context) {
			exchange := c.Param("exchange")
			ticker := c.Param("ticker")

			// TODO: convert exchange string to Exchange type
			lastDate, err := collectorService.GetLastCollectionDate(c.Request.Context(), ticker, "MOEX")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"ticker":     ticker,
				"exchange":   exchange,
				"last_date":  lastDate.Format("2006-01-02"),
			})
		})
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in goroutine
	go func() {
		appLogger.Infof("Data Collector Service listening on port %d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		appLogger.Fatalf("Server forced to shutdown: %v", err)
	}

	appLogger.Info("Server exited")
}
