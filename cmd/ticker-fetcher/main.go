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
	"github.com/stock-market-system/pkg/config"
	"github.com/stock-market-system/pkg/logger"
	"github.com/stock-market-system/pkg/storage/redis"
	"github.com/stock-market-system/internal/ticker-fetcher/handlers"
	"github.com/stock-market-system/internal/ticker-fetcher/service"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	appLogger := logger.NewLogger(cfg.Logger)
	appLogger.Info("Starting Ticker Fetcher Service")

	// Initialize Redis client
	redisClient, err := redis.NewClient(&cfg.Redis)
	if err != nil {
		appLogger.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	appLogger.Info("Connected to Redis successfully")

	// Initialize ticker service
	tickerService := service.NewTickerService(redisClient, appLogger)

	// Initialize Gin router
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"service": "ticker-fetcher",
			"timestamp": time.Now().Unix(),
		})
	})

	// Initialize handlers
	tickerHandler := handlers.NewTickerHandler(tickerService, appLogger)

	// API routes
	v1 := router.Group("/api/v1")
	{
		v1.GET("/tickers/nasdaq", tickerHandler.GetNASDAQTickers)
		v1.GET("/tickers/moex", tickerHandler.GetMOEXTickers)
		v1.GET("/tickers/all", tickerHandler.GetAllTickers)
		v1.POST("/tickers/refresh", tickerHandler.RefreshTickers)
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
		appLogger.Infof("Ticker Fetcher Service listening on port %d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		appLogger.Fatalf("Server forced to shutdown: %v", err)
	}

	appLogger.Info("Server exited")
}
