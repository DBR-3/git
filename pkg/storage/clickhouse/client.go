package clickhouse

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/stock-market-system/pkg/config"
	"github.com/stock-market-system/pkg/models"
)

// Client wraps ClickHouse connection
type Client struct {
	conn   driver.Conn
	config *config.ClickHouseConfig
}

// NewClient creates a new ClickHouse client
func NewClient(cfg *config.ClickHouseConfig) (*Client, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: cfg.Addresses,
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.Username,
			Password: cfg.Password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout:      10 * time.Second,
		MaxOpenConns:     cfg.MaxOpenConns,
		MaxIdleConns:     cfg.MaxIdleConns,
		ConnMaxLifetime:  cfg.ConnMaxLifetime,
		ConnOpenStrategy: clickhouse.ConnOpenInOrder,
		BlockBufferSize:  10,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
			Level:  cfg.CompressionLevel,
		},
		TLS: &tls.Config{
			InsecureSkipVerify: true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	return &Client{
		conn:   conn,
		config: cfg,
	}, nil
}

// Close closes the ClickHouse connection
func (c *Client) Close() error {
	return c.conn.Close()
}

// Ping tests the connection
func (c *Client) Ping(ctx context.Context) error {
	return c.conn.Ping(ctx)
}

// InsertStockQuotes inserts stock quotes in batch
func (c *Client) InsertStockQuotes(ctx context.Context, quotes []models.StockQuote) error {
	if len(quotes) == 0 {
		return nil
	}

	batch, err := c.conn.PrepareBatch(ctx, "INSERT INTO stock_quotes")
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	for _, quote := range quotes {
		err := batch.Append(
			quote.Ticker,
			quote.Exchange,
			quote.TradeDate,
			quote.Open,
			quote.High,
			quote.Low,
			quote.Close,
			quote.Volume,
			quote.Value,
			quote.NumTrades,
			quote.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to append to batch: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}

	return nil
}

// InsertBenchmarkIndex inserts benchmark index data
func (c *Client) InsertBenchmarkIndex(ctx context.Context, indices []models.BenchmarkIndex) error {
	if len(indices) == 0 {
		return nil
	}

	batch, err := c.conn.PrepareBatch(ctx, "INSERT INTO benchmark_indices")
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	for _, index := range indices {
		err := batch.Append(
			index.IndexName,
			index.TradeDate,
			index.Open,
			index.High,
			index.Low,
			index.Close,
			index.Volume,
			index.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to append to batch: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}

	return nil
}

// InsertTickers inserts tickers
func (c *Client) InsertTickers(ctx context.Context, tickers []models.Ticker) error {
	if len(tickers) == 0 {
		return nil
	}

	batch, err := c.conn.PrepareBatch(ctx, "INSERT INTO tickers")
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	for _, ticker := range tickers {
		err := batch.Append(
			ticker.Ticker,
			ticker.Name,
			ticker.Exchange,
			ticker.SecurityType,
			ticker.Sector,
			ticker.Industry,
			ticker.IsActive,
			ticker.FirstSeen,
			ticker.LastUpdated,
		)
		if err != nil {
			return fmt.Errorf("failed to append to batch: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}

	return nil
}

// InsertTechnicalIndicators inserts technical indicators
func (c *Client) InsertTechnicalIndicators(ctx context.Context, indicators []models.TechnicalIndicators) error {
	if len(indicators) == 0 {
		return nil
	}

	batch, err := c.conn.PrepareBatch(ctx, "INSERT INTO technical_indicators")
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	for _, ind := range indicators {
		err := batch.Append(
			ind.Ticker,
			ind.Exchange,
			ind.TradeDate,
			ind.Timeframe,
			ind.EMA10,
			ind.EMA30,
			ind.EMA50,
			ind.EMA100,
			ind.EMA200,
			ind.SMA20,
			ind.SMA50,
			ind.RSI14,
			ind.WilliamsR14,
			ind.VolumeSMA20,
			ind.RelativeVolume,
			ind.VZO14,
			ind.UpDownRatio50,
			ind.ATR14,
			ind.Volatility15,
			ind.RMV,
			ind.MansfieldRS,
			ind.MansfieldRSMA,
			ind.MFI58,
			ind.WarriorTrend,
			ind.WarriorEMARSI,
			ind.WarriorTrendLength,
			ind.TSDDirection,
			ind.TSDLevel,
			ind.TSDLength,
			ind.StockStage,
			ind.AVWAP1Y,
			ind.AVWAP1YStd,
			ind.AVWAP3Y,
			ind.AVWAP3YStd,
			ind.AVWAP5Y,
			ind.AVWAP5YStd,
			ind.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to append to batch: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}

	return nil
}

// InsertTradingSignals inserts trading signals
func (c *Client) InsertTradingSignals(ctx context.Context, signals []models.TradingSignal) error {
	if len(signals) == 0 {
		return nil
	}

	batch, err := c.conn.PrepareBatch(ctx, "INSERT INTO trading_signals")
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	for _, signal := range signals {
		err := batch.Append(
			signal.SignalID,
			signal.Ticker,
			signal.Exchange,
			signal.SignalDate,
			signal.SignalType,
			signal.SignalStrength,
			signal.Description,
			signal.Metadata,
			signal.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to append to batch: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}

	return nil
}

// LogJobExecution logs job execution
func (c *Client) LogJobExecution(ctx context.Context, log models.JobExecutionLog) error {
	return c.conn.Exec(ctx, `
		INSERT INTO job_execution_log (
			job_id, job_type, exchange, status, start_time, end_time,
			duration_seconds, tickers_processed, tickers_failed,
			error_message, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		log.JobID,
		log.JobType,
		log.Exchange,
		log.Status,
		log.StartTime,
		log.EndTime,
		log.DurationSeconds,
		log.TickersProcessed,
		log.TickersFailed,
		log.ErrorMessage,
		log.CreatedAt,
	)
}

// GetLatestQuoteDate gets the latest quote date for a ticker
func (c *Client) GetLatestQuoteDate(ctx context.Context, ticker string, exchange models.Exchange) (time.Time, error) {
	var tradeDate time.Time
	err := c.conn.QueryRow(ctx, `
		SELECT max(trade_date)
		FROM stock_quotes
		WHERE ticker = ? AND exchange = ?
	`, ticker, exchange).Scan(&tradeDate)

	if err != nil {
		return time.Time{}, err
	}

	return tradeDate, nil
}

// GetStockQuotes retrieves stock quotes for a ticker within a date range
func (c *Client) GetStockQuotes(ctx context.Context, ticker string, exchange models.Exchange, startDate, endDate time.Time) ([]models.StockQuote, error) {
	rows, err := c.conn.Query(ctx, `
		SELECT ticker, exchange, trade_date, open, high, low, close, volume, value, num_trades, created_at
		FROM stock_quotes
		WHERE ticker = ? AND exchange = ? AND trade_date >= ? AND trade_date <= ?
		ORDER BY trade_date ASC
	`, ticker, exchange, startDate, endDate)

	if err != nil {
		return nil, fmt.Errorf("failed to query stock quotes: %w", err)
	}
	defer rows.Close()

	quotes := make([]models.StockQuote, 0)
	for rows.Next() {
		var quote models.StockQuote
		err := rows.Scan(
			&quote.Ticker,
			&quote.Exchange,
			&quote.TradeDate,
			&quote.Open,
			&quote.High,
			&quote.Low,
			&quote.Close,
			&quote.Volume,
			&quote.Value,
			&quote.NumTrades,
			&quote.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		quotes = append(quotes, quote)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return quotes, nil
}

// Execute executes a query
func (c *Client) Execute(ctx context.Context, query string, args ...interface{}) error {
	return c.conn.Exec(ctx, query, args...)
}

// Query executes a query and returns rows
func (c *Client) Query(ctx context.Context, query string, args ...interface{}) (driver.Rows, error) {
	return c.conn.Query(ctx, query, args...)
}
