package sources

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/stock-market-system/pkg/models"
)

const (
	moexSecuritiesURL = "https://iss.moex.com/iss/engines/stock/markets/shares/securities.json"
)

// MOEXResponse represents MOEX API response structure
type MOEXResponse struct {
	Securities struct {
		Columns []string        `json:"columns"`
		Data    [][]interface{} `json:"data"`
	} `json:"securities"`
}

// FetchMOEXTickers fetches tickers from MOEX
func FetchMOEXTickers(boardID string) ([]models.Ticker, error) {
	// Build URL with parameters
	url := moexSecuritiesURL
	if boardID == "" {
		boardID = "TQBR" // Default to main board
	}

	params := map[string]string{
		"iss.meta": "off",
	}

	// Add query parameters
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch MOEX data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("MOEX API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var moexResp MOEXResponse
	if err := json.Unmarshal(body, &moexResp); err != nil {
		return nil, fmt.Errorf("failed to parse MOEX response: %w", err)
	}

	// Find column indices
	secIDIdx := findIndex(moexResp.Securities.Columns, "SECID")
	secNameIdx := findIndex(moexResp.Securities.Columns, "SECNAME")
	secTypeIdx := findIndex(moexResp.Securities.Columns, "SECTYPE")
	boardIDIdx := findIndex(moexResp.Securities.Columns, "BOARDID")

	if secIDIdx == -1 {
		return nil, fmt.Errorf("SECID column not found in MOEX response")
	}

	var tickers []models.Ticker

	// Parse securities
	for _, row := range moexResp.Securities.Data {
		if len(row) <= secIDIdx {
			continue
		}

		// Filter by board ID
		if boardIDIdx != -1 && len(row) > boardIDIdx {
			rowBoardID, ok := row[boardIDIdx].(string)
			if ok && boardID != "" && rowBoardID != boardID {
				continue
			}
		}

		// Get ticker symbol
		secID, ok := row[secIDIdx].(string)
		if !ok || secID == "" {
			continue
		}

		// Filter by security type (only stocks)
		if secTypeIdx != -1 && len(row) > secTypeIdx {
			secType, ok := row[secTypeIdx].(string)
			if ok && secType != "1" && secType != "2" { // 1 = common stock, 2 = preferred stock
				continue
			}
		}

		// Get security name
		name := ""
		if secNameIdx != -1 && len(row) > secNameIdx {
			if secName, ok := row[secNameIdx].(string); ok {
				name = secName
			}
		}

		ticker := models.Ticker{
			Ticker:       secID,
			Name:         name,
			Exchange:     models.ExchangeMOEX,
			SecurityType: models.SecurityTypeStock,
			IsActive:     true,
		}

		tickers = append(tickers, ticker)
	}

	return tickers, nil
}

// findIndex finds the index of a column in the columns array
func findIndex(columns []string, name string) int {
	for i, col := range columns {
		if strings.EqualFold(col, name) {
			return i
		}
	}
	return -1
}
