package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/stock-market-system/internal/ticker-fetcher/sources"
	"github.com/stock-market-system/pkg/logger"
	"github.com/stock-market-system/pkg/models"
	"github.com/stock-market-system/pkg/storage/redis"
)

const (
	tickerCacheTTL = 24 * time.Hour
	nasdaqCacheKey = "tickers:nasdaq"
	moexCacheKey   = "tickers:moex"
)

// TickerService handles ticker fetching and caching
type TickerService struct {
	redis  *redis.Client
	logger *logger.Logger
}

// NewTickerService creates a new TickerService
func NewTickerService(redisClient *redis.Client, log *logger.Logger) *TickerService {
	return &TickerService{
		redis:  redisClient,
		logger: log,
	}
}

// GetNASDAQTickers retrieves NASDAQ tickers (stocks or ETFs)
func (s *TickerService) GetNASDAQTickers(ctx context.Context, securityType models.SecurityType) ([]models.Ticker, error) {
	cacheKey := fmt.Sprintf("%s:%s", nasdaqCacheKey, securityType)

	// Try to get from cache
	cached, err := s.redis.Get(ctx, cacheKey)
	if err == nil && cached != "" {
		var tickers []models.Ticker
		if err := json.Unmarshal([]byte(cached), &tickers); err == nil {
			s.logger.Infof("Retrieved %d NASDAQ %s tickers from cache", len(tickers), securityType)
			return tickers, nil
		}
	}

	// Fetch from source
	s.logger.Infof("Fetching NASDAQ %s tickers from source", securityType)
	tickers, err := sources.FetchNASDAQTickers(securityType)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch NASDAQ tickers: %w", err)
	}

	// Update ticker metadata
	now := time.Now()
	for i := range tickers {
		tickers[i].Exchange = models.ExchangeNASDAQ
		tickers[i].SecurityType = securityType
		tickers[i].IsActive = true
		tickers[i].LastUpdated = now
		if tickers[i].FirstSeen.IsZero() {
			tickers[i].FirstSeen = now
		}
	}

	// Cache the result
	data, err := json.Marshal(tickers)
	if err == nil {
		if err := s.redis.Set(ctx, cacheKey, string(data), tickerCacheTTL); err != nil {
			s.logger.Warnf("Failed to cache NASDAQ tickers: %v", err)
		}
	}

	s.logger.Infof("Fetched %d NASDAQ %s tickers", len(tickers), securityType)
	return tickers, nil
}

// GetMOEXTickers retrieves MOEX tickers
func (s *TickerService) GetMOEXTickers(ctx context.Context, boardID string) ([]models.Ticker, error) {
	cacheKey := fmt.Sprintf("%s:%s", moexCacheKey, boardID)

	// Try to get from cache
	cached, err := s.redis.Get(ctx, cacheKey)
	if err == nil && cached != "" {
		var tickers []models.Ticker
		if err := json.Unmarshal([]byte(cached), &tickers); err == nil {
			s.logger.Infof("Retrieved %d MOEX tickers from cache (board: %s)", len(tickers), boardID)
			return tickers, nil
		}
	}

	// Fetch from source
	s.logger.Infof("Fetching MOEX tickers from source (board: %s)", boardID)
	tickers, err := sources.FetchMOEXTickers(boardID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch MOEX tickers: %w", err)
	}

	// Update ticker metadata
	now := time.Now()
	for i := range tickers {
		tickers[i].Exchange = models.ExchangeMOEX
		tickers[i].IsActive = true
		tickers[i].LastUpdated = now
		if tickers[i].FirstSeen.IsZero() {
			tickers[i].FirstSeen = now
		}
	}

	// Cache the result
	data, err := json.Marshal(tickers)
	if err == nil {
		if err := s.redis.Set(ctx, cacheKey, string(data), tickerCacheTTL); err != nil {
			s.logger.Warnf("Failed to cache MOEX tickers: %v", err)
		}
	}

	s.logger.Infof("Fetched %d MOEX tickers (board: %s)", len(tickers), boardID)
	return tickers, nil
}

// GetAllTickers retrieves all tickers from all exchanges
func (s *TickerService) GetAllTickers(ctx context.Context) ([]models.Ticker, error) {
	var allTickers []models.Ticker

	// Get NASDAQ stocks
	nasdaqStocks, err := s.GetNASDAQTickers(ctx, models.SecurityTypeStock)
	if err != nil {
		s.logger.Warnf("Failed to get NASDAQ stocks: %v", err)
	} else {
		allTickers = append(allTickers, nasdaqStocks...)
	}

	// Get NASDAQ ETFs
	nasdaqETFs, err := s.GetNASDAQTickers(ctx, models.SecurityTypeETF)
	if err != nil {
		s.logger.Warnf("Failed to get NASDAQ ETFs: %v", err)
	} else {
		allTickers = append(allTickers, nasdaqETFs...)
	}

	// Get MOEX tickers (main board)
	moexTickers, err := s.GetMOEXTickers(ctx, "TQBR")
	if err != nil {
		s.logger.Warnf("Failed to get MOEX tickers: %v", err)
	} else {
		allTickers = append(allTickers, moexTickers...)
	}

	s.logger.Infof("Retrieved total of %d tickers from all exchanges", len(allTickers))
	return allTickers, nil
}

// RefreshCache invalidates and refreshes the ticker cache
func (s *TickerService) RefreshCache(ctx context.Context) error {
	s.logger.Info("Refreshing ticker cache")

	// Delete all ticker cache keys
	keys := []string{
		fmt.Sprintf("%s:%s", nasdaqCacheKey, models.SecurityTypeStock),
		fmt.Sprintf("%s:%s", nasdaqCacheKey, models.SecurityTypeETF),
		fmt.Sprintf("%s:TQBR", moexCacheKey),
		fmt.Sprintf("%s:TQPI", moexCacheKey),
	}

	for _, key := range keys {
		if err := s.redis.Delete(ctx, key); err != nil {
			s.logger.Warnf("Failed to delete cache key %s: %v", key, err)
		}
	}

	// Fetch fresh data
	_, err := s.GetAllTickers(ctx)
	if err != nil {
		return fmt.Errorf("failed to refresh tickers: %w", err)
	}

	s.logger.Info("Ticker cache refreshed successfully")
	return nil
}

// GetTickerCount returns the count of tickers by exchange
func (s *TickerService) GetTickerCount(ctx context.Context) (map[models.Exchange]int, error) {
	counts := make(map[models.Exchange]int)

	// NASDAQ stocks
	nasdaqStocks, err := s.GetNASDAQTickers(ctx, models.SecurityTypeStock)
	if err == nil {
		counts[models.ExchangeNASDAQ] += len(nasdaqStocks)
	}

	// NASDAQ ETFs
	nasdaqETFs, err := s.GetNASDAQTickers(ctx, models.SecurityTypeETF)
	if err == nil {
		counts[models.ExchangeNASDAQ] += len(nasdaqETFs)
	}

	// MOEX
	moexTickers, err := s.GetMOEXTickers(ctx, "TQBR")
	if err == nil {
		counts[models.ExchangeMOEX] = len(moexTickers)
	}

	return counts, nil
}
