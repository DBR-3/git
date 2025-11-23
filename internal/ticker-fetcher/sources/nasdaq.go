package sources

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/stock-market-system/pkg/models"
)

const (
	nasdaqListedURL = "https://www.nasdaqtrader.com/dynamic/SymDir/nasdaqlisted.txt"
	otherListedURL  = "https://www.nasdaqtrader.com/dynamic/SymDir/otherlisted.txt"
	httpTimeout     = 60 * time.Second
)

var (
	httpClient = &http.Client{
		Timeout: httpTimeout,
	}
	userAgent = "Mozilla/5.0 (compatible; TickerFetcher/1.0)"
)

// FetchNASDAQTickers fetches tickers from NASDAQ
func FetchNASDAQTickers(securityType models.SecurityType) ([]models.Ticker, error) {
	var allTickers []models.Ticker

	// Fetch from nasdaqlisted.txt
	nasdaqTickers, err := fetchPipeTXT(nasdaqListedURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch nasdaqlisted.txt: %w", err)
	}
	allTickers = append(allTickers, nasdaqTickers...)

	// Fetch from otherlisted.txt
	otherTickers, err := fetchPipeTXT(otherListedURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch otherlisted.txt: %w", err)
	}
	allTickers = append(allTickers, otherTickers...)

	// Filter by security type
	var filtered []models.Ticker
	for _, ticker := range allTickers {
		if ticker.SecurityType == securityType {
			filtered = append(filtered, ticker)
		}
	}

	// Remove duplicates
	seen := make(map[string]bool)
	var unique []models.Ticker
	for _, ticker := range filtered {
		if !seen[ticker.Ticker] {
			seen[ticker.Ticker] = true
			unique = append(unique, ticker)
		}
	}

	return unique, nil
}

// fetchPipeTXT fetches and parses pipe-delimited text file
func fetchPipeTXT(url string) ([]models.Ticker, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	content := string(body)
	lines := strings.Split(content, "\n")

	if len(lines) < 2 {
		return nil, fmt.Errorf("invalid file format")
	}

	// Parse header
	header := strings.Split(lines[0], "|")
	symbolIdx := findColumnIndex(header, "Symbol")
	nameIdx := findColumnIndex(header, "Security Name")
	etfIdx := findColumnIndex(header, "ETF")
	testIssueIdx := findColumnIndex(header, "Test Issue")

	if symbolIdx == -1 {
		return nil, fmt.Errorf("Symbol column not found")
	}

	var tickers []models.Ticker

	// Parse data rows
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(strings.ToUpper(line), "FILE CREATION TIME") {
			continue
		}

		fields := strings.Split(line, "|")
		if len(fields) <= symbolIdx {
			continue
		}

		symbol := normalizeSymbol(fields[symbolIdx])
		if symbol == "" {
			continue
		}

		// Skip test issues
		if testIssueIdx != -1 && len(fields) > testIssueIdx {
			testIssue := strings.ToUpper(strings.TrimSpace(fields[testIssueIdx]))
			if testIssue == "Y" {
				continue
			}
		}

		// Determine security type
		secType := models.SecurityTypeStock
		if etfIdx != -1 && len(fields) > etfIdx {
			etf := strings.ToUpper(strings.TrimSpace(fields[etfIdx]))
			if etf == "Y" {
				secType = models.SecurityTypeETF
			}
		}

		// Get name
		name := ""
		if nameIdx != -1 && len(fields) > nameIdx {
			name = strings.TrimSpace(fields[nameIdx])
		}

		ticker := models.Ticker{
			Ticker:       symbol,
			Name:         name,
			Exchange:     models.ExchangeNASDAQ,
			SecurityType: secType,
			IsActive:     true,
		}

		tickers = append(tickers, ticker)
	}

	return tickers, nil
}

// findColumnIndex finds the index of a column in the header
func findColumnIndex(header []string, columnName string) int {
	for i, col := range header {
		col = strings.TrimSpace(strings.ToLower(col))
		if col == strings.ToLower(columnName) ||
			col == strings.ToLower(strings.ReplaceAll(columnName, " ", "")) {
			return i
		}
	}
	return -1
}

// normalizeSymbol normalizes a ticker symbol
func normalizeSymbol(symbol string) string {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))

	// Remove invalid characters
	re := regexp.MustCompile(`[^A-Z0-9\-]`)
	symbol = re.ReplaceAllString(symbol, "")

	return symbol
}
