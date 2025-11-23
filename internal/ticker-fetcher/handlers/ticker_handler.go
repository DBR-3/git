package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stock-market-system/internal/ticker-fetcher/service"
	"github.com/stock-market-system/pkg/logger"
	"github.com/stock-market-system/pkg/models"
)

// TickerHandler handles HTTP requests for ticker operations
type TickerHandler struct {
	service *service.TickerService
	logger  *logger.Logger
}

// NewTickerHandler creates a new TickerHandler
func NewTickerHandler(svc *service.TickerService, log *logger.Logger) *TickerHandler {
	return &TickerHandler{
		service: svc,
		logger:  log,
	}
}

// GetNASDAQTickers handles GET /api/v1/tickers/nasdaq
func (h *TickerHandler) GetNASDAQTickers(c *gin.Context) {
	securityType := c.DefaultQuery("type", "stocks")

	var st models.SecurityType
	if securityType == "etf" || securityType == "etfs" {
		st = models.SecurityTypeETF
	} else {
		st = models.SecurityTypeStock
	}

	tickers, err := h.service.GetNASDAQTickers(c.Request.Context(), st)
	if err != nil {
		h.logger.Errorf("Failed to get NASDAQ tickers: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch NASDAQ tickers",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"exchange":      "NASDAQ",
		"security_type": st,
		"count":         len(tickers),
		"tickers":       tickers,
	})
}

// GetMOEXTickers handles GET /api/v1/tickers/moex
func (h *TickerHandler) GetMOEXTickers(c *gin.Context) {
	boardID := c.DefaultQuery("boardid", "TQBR")

	tickers, err := h.service.GetMOEXTickers(c.Request.Context(), boardID)
	if err != nil {
		h.logger.Errorf("Failed to get MOEX tickers: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch MOEX tickers",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"exchange": "MOEX",
		"board_id": boardID,
		"count":    len(tickers),
		"tickers":  tickers,
	})
}

// GetAllTickers handles GET /api/v1/tickers/all
func (h *TickerHandler) GetAllTickers(c *gin.Context) {
	tickers, err := h.service.GetAllTickers(c.Request.Context())
	if err != nil {
		h.logger.Errorf("Failed to get all tickers: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch tickers",
		})
		return
	}

	// Group by exchange
	byExchange := make(map[models.Exchange]int)
	for _, ticker := range tickers {
		byExchange[ticker.Exchange]++
	}

	c.JSON(http.StatusOK, gin.H{
		"total_count":  len(tickers),
		"by_exchange":  byExchange,
		"tickers":      tickers,
	})
}

// RefreshTickers handles POST /api/v1/tickers/refresh
func (h *TickerHandler) RefreshTickers(c *gin.Context) {
	h.logger.Info("Refreshing ticker cache via API request")

	err := h.service.RefreshCache(c.Request.Context())
	if err != nil {
		h.logger.Errorf("Failed to refresh ticker cache: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to refresh ticker cache",
		})
		return
	}

	// Get updated counts
	counts, err := h.service.GetTickerCount(c.Request.Context())
	if err != nil {
		h.logger.Errorf("Failed to get ticker counts: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Ticker cache refreshed successfully",
		"counts":  counts,
	})
}
