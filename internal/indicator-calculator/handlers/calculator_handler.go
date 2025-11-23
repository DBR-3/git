package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stock-market-system/internal/indicator-calculator/service"
	"github.com/stock-market-system/pkg/logger"
	"github.com/stock-market-system/pkg/models"
)

// CalculatorHandler handles indicator calculator HTTP requests
type CalculatorHandler struct {
	service *service.CalculatorService
	logger  *logger.Logger
}

// NewCalculatorHandler creates a new calculator handler
func NewCalculatorHandler(svc *service.CalculatorService, log *logger.Logger) *CalculatorHandler {
	return &CalculatorHandler{
		service: svc,
		logger:  log,
	}
}

// CalculateRequest represents a manual calculation request
type CalculateRequest struct {
	Exchange  string   `json:"exchange" binding:"required"`
	Tickers   []string `json:"tickers" binding:"required"`
	StartDate string   `json:"start_date" binding:"required"`
	EndDate   string   `json:"end_date" binding:"required"`
}

// CalculateResponse represents a calculation response
type CalculateResponse struct {
	JobID            string    `json:"job_id"`
	Status           string    `json:"status"`
	TickersProcessed int       `json:"tickers_processed"`
	TickersFailed    int       `json:"tickers_failed"`
	DurationSeconds  float64   `json:"duration_seconds"`
	StartedAt        time.Time `json:"started_at"`
	CompletedAt      time.Time `json:"completed_at,omitempty"`
}

// Calculate triggers manual indicator calculation
// @Summary Calculate indicators manually
// @Description Triggers indicator calculation for specified tickers
// @Tags indicators
// @Accept json
// @Produce json
// @Param request body CalculateRequest true "Calculation request"
// @Success 200 {object} CalculateResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/calculate [post]
func (h *CalculatorHandler) Calculate(c *gin.Context) {
	var req CalculateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse exchange
	exchange := models.ParseExchange(req.Exchange)
	if exchange == models.ExchangeOther {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exchange"})
		return
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format, use YYYY-MM-DD"})
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format, use YYYY-MM-DD"})
		return
	}

	// Trigger calculation
	jobLog, err := h.service.CalculateManual(c.Request.Context(), exchange, req.Tickers, startDate, endDate)
	if err != nil {
		h.logger.WithError(err).Error("Failed to calculate indicators")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Build response
	response := CalculateResponse{
		JobID:            jobLog.JobID,
		Status:           string(jobLog.Status),
		TickersProcessed: int(jobLog.TickersProcessed),
		TickersFailed:    int(jobLog.TickersFailed),
		DurationSeconds:  jobLog.CompletedAt.Sub(jobLog.StartedAt).Seconds(),
		StartedAt:        jobLog.StartedAt,
		CompletedAt:      jobLog.CompletedAt,
	}

	c.JSON(http.StatusOK, response)
}

// Health returns service health status
// @Summary Health check
// @Description Returns the health status of the service
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func (h *CalculatorHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "indicator-calculator",
		"time":    time.Now().Format(time.RFC3339),
	})
}
