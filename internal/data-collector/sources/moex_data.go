package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/stock-market-system/pkg/models"
)

const (
	moexHistoryURL = "https://iss.moex.com/iss/history/engines/stock/markets/shares/securities/%s.json"
	moexPageSize   = 100
)

// MOEXHistoryResponse represents MOEX history API response
type MOEXHistoryResponse struct {
	History struct {
		Columns []string        `json:"columns"`
		Data    [][]interface{} `json:"data"`
	} `json:"history"`
	HistoryCursor struct {
		Columns []string        `json:"columns"`
		Data    [][]interface{} `json:"data"`
	} `json:"history.cursor"`
}

// MOEXDataSource fetches historical data from MOEX
type MOEXDataSource struct {
	httpClient *http.Client
}

// NewMOEXDataSource creates a new MOEX data source
func NewMOEXDataSource() *MOEXDataSource {
	return &MOEXDataSource{
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// FetchHistoricalData fetches historical OHLCV data for a ticker
func (m *MOEXDataSource) FetchHistoricalData(ctx context.Context, ticker string, startDate, endDate time.Time) ([]models.StockQuote, error) {
	var allQuotes []models.StockQuote
	start := 0

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		quotes, total, err := m.fetchPage(ctx, ticker, startDate, endDate, start)
		if err != nil {
			return nil, err
		}

		allQuotes = append(allQuotes, quotes...)

		// Check if we've fetched all data
		if start+len(quotes) >= total || len(quotes) == 0 {
			break
		}

		start += len(quotes)

		// Small delay to avoid rate limiting
		time.Sleep(100 * time.Millisecond)
	}

	return allQuotes, nil
}

// fetchPage fetches a single page of data
func (m *MOEXDataSource) fetchPage(ctx context.Context, ticker string, startDate, endDate time.Time, start int) ([]models.StockQuote, int, error) {
	url := fmt.Sprintf(moexHistoryURL, ticker)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, 0, err
	}

	q := req.URL.Query()
	q.Add("from", startDate.Format("2006-01-02"))
	q.Add("till", endDate.Format("2006-01-02"))
	q.Add("start", fmt.Sprintf("%d", start))
	q.Add("iss.meta", "off")
	req.URL.RawQuery = q.Encode()

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch MOEX data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("MOEX API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	var moexResp MOEXHistoryResponse
	if err := json.Unmarshal(body, &moexResp); err != nil {
		return nil, 0, fmt.Errorf("failed to parse MOEX response: %w", err)
	}

	// Parse data
	quotes := m.parseQuotes(ticker, moexResp)

	// Get total count
	total := m.getTotalCount(moexResp)

	return quotes, total, nil
}

// parseQuotes parses MOEX response to StockQuote models
func (m *MOEXDataSource) parseQuotes(ticker string, resp MOEXHistoryResponse) []models.StockQuote {
	var quotes []models.StockQuote

	// Find column indices
	tradeDateIdx := findColumnIdx(resp.History.Columns, "TRADEDATE")
	openIdx := findColumnIdx(resp.History.Columns, "OPEN")
	highIdx := findColumnIdx(resp.History.Columns, "HIGH")
	lowIdx := findColumnIdx(resp.History.Columns, "LOW")
	closeIdx := findColumnIdx(resp.History.Columns, "CLOSE")
	volumeIdx := findColumnIdx(resp.History.Columns, "VOLUME")
	valueIdx := findColumnIdx(resp.History.Columns, "VALUE")
	numTradesIdx := findColumnIdx(resp.History.Columns, "NUMTRADES")
	boardIDIdx := findColumnIdx(resp.History.Columns, "BOARDID")

	if tradeDateIdx == -1 || closeIdx == -1 {
		return quotes
	}

	for _, row := range resp.History.Data {
		// Filter by BOARDID (only TQBR - main board)
		if boardIDIdx != -1 && len(row) > boardIDIdx {
			if boardID, ok := row[boardIDIdx].(string); ok && boardID != "TQBR" {
				continue
			}
		}

		// Skip rows with null close price
		if row[closeIdx] == nil {
			continue
		}

		quote := models.StockQuote{
			Ticker:    ticker,
			Exchange:  models.ExchangeMOEX,
			CreatedAt: time.Now(),
		}

		// Parse trade date
		if tradeDateStr, ok := row[tradeDateIdx].(string); ok {
			tradeDate, err := time.Parse("2006-01-02", tradeDateStr)
			if err == nil {
				quote.TradeDate = tradeDate
			}
		}

		// Parse OHLC
		if val, ok := toFloat64(row[openIdx]); ok {
			quote.Open = val
		}
		if val, ok := toFloat64(row[highIdx]); ok {
			quote.High = val
		}
		if val, ok := toFloat64(row[lowIdx]); ok {
			quote.Low = val
		}
		if val, ok := toFloat64(row[closeIdx]); ok {
			quote.Close = val
		}

		// Parse volume and value
		if val, ok := toUint64(row[volumeIdx]); ok {
			quote.Volume = val
		}
		if val, ok := toFloat64(row[valueIdx]); ok {
			quote.Value = val
		}
		if val, ok := toUint32(row[numTradesIdx]); ok {
			quote.NumTrades = val
		}

		// Skip invalid quotes
		if quote.Close > 0 && !quote.TradeDate.IsZero() {
			quotes = append(quotes, quote)
		}
	}

	return quotes
}

// getTotalCount gets total count from cursor
func (m *MOEXDataSource) getTotalCount(resp MOEXHistoryResponse) int {
	totalIdx := findColumnIdx(resp.HistoryCursor.Columns, "TOTAL")
	if totalIdx == -1 || len(resp.HistoryCursor.Data) == 0 {
		return 0
	}

	if len(resp.HistoryCursor.Data[0]) <= totalIdx {
		return 0
	}

	if total, ok := toInt(resp.HistoryCursor.Data[0][totalIdx]); ok {
		return total
	}

	return 0
}

// Helper functions

func findColumnIdx(columns []string, name string) int {
	for i, col := range columns {
		if col == name {
			return i
		}
	}
	return -1
}

func toFloat64(val interface{}) (float64, bool) {
	if val == nil {
		return 0, false
	}

	switch v := val.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

func toUint64(val interface{}) (uint64, bool) {
	if val == nil {
		return 0, false
	}

	switch v := val.(type) {
	case float64:
		return uint64(v), true
	case int:
		return uint64(v), true
	case int64:
		return uint64(v), true
	default:
		return 0, false
	}
}

func toUint32(val interface{}) (uint32, bool) {
	if val == nil {
		return 0, false
	}

	switch v := val.(type) {
	case float64:
		return uint32(v), true
	case int:
		return uint32(v), true
	case int64:
		return uint32(v), true
	default:
		return 0, false
	}
}

func toInt(val interface{}) (int, bool) {
	if val == nil {
		return 0, false
	}

	switch v := val.(type) {
	case float64:
		return int(v), true
	case int:
		return v, true
	case int64:
		return int(v), true
	default:
		return 0, false
	}
}
