package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/stock-market-system/internal/data-collector/sources"
	"github.com/stock-market-system/internal/data-collector/workers"
	"github.com/stock-market-system/pkg/logger"
	"github.com/stock-market-system/pkg/messaging"
	"github.com/stock-market-system/pkg/models"
	"github.com/stock-market-system/pkg/storage/clickhouse"
)

// CollectorService handles data collection
type CollectorService struct {
	clickhouse *clickhouse.Client
	kafka      *messaging.Producer
	moexSource *sources.MOEXDataSource
	logger     *logger.Logger
	workerPool *workers.WorkerPool
}

// NewCollectorService creates a new collector service
func NewCollectorService(
	ch *clickhouse.Client,
	kafka *messaging.Producer,
	logger *logger.Logger,
	numWorkers int,
) *CollectorService {
	return &CollectorService{
		clickhouse: ch,
		kafka:      kafka,
		moexSource: sources.NewMOEXDataSource(),
		logger:     logger,
		workerPool: workers.NewWorkerPool(numWorkers, numWorkers*2, logger),
	}
}

// Start starts the collector service
func (s *CollectorService) Start() {
	s.workerPool.Start()
	s.logger.Info("Collector service started")
}

// Stop stops the collector service
func (s *CollectorService) Stop() {
	s.workerPool.Stop()
	s.logger.Info("Collector service stopped")
}

// CollectMOEXData collects data for MOEX tickers
func (s *CollectorService) CollectMOEXData(ctx context.Context, tickers []string, startDate, endDate time.Time) (*models.JobExecutionLog, error) {
	jobID := uuid.New().String()
	jobStartTime := time.Now()

	s.logger.Infof("Starting MOEX data collection job %s for %d tickers", jobID, len(tickers))

	// Log job start
	jobLog := models.JobExecutionLog{
		JobID:      jobID,
		JobType:    models.JobTypeDataCollection,
		Exchange:   models.ExchangeMOEX,
		Status:     models.JobStatusStarted,
		StartTime:  jobStartTime,
		CreatedAt:  jobStartTime,
	}

	var (
		mu               sync.Mutex
		tickersProcessed uint32
		tickersFailed    uint32
		allQuotes        []models.StockQuote
	)

	// Result collector goroutine
	resultDone := make(chan struct{})
	go func() {
		defer close(resultDone)

		for result := range s.workerPool.Results() {
			if result.Success {
				tickersProcessed++
				s.logger.Infof("Collected data for %s (%d/%d)", result.Ticker, tickersProcessed, len(tickers))
			} else {
				tickersFailed++
				s.logger.Errorf("Failed to collect data for %s: %v", result.Ticker, result.Error)

				// Publish error event
				_ = s.kafka.PublishError(ctx, fmt.Sprintf("Failed to collect %s: %v", result.Ticker, result.Error), map[string]interface{}{
					"ticker":   result.Ticker,
					"job_id":   jobID,
					"exchange": models.ExchangeMOEX,
				})
			}

			// Collect quotes
			if result.Data != nil {
				if quotes, ok := result.Data.([]models.StockQuote); ok {
					mu.Lock()
					allQuotes = append(allQuotes, quotes...)
					mu.Unlock()
				}
			}
		}
	}()

	// Submit tasks to worker pool
	for _, ticker := range tickers {
		task := workers.Task{
			ID:     fmt.Sprintf("%s-%s", jobID, ticker),
			Ticker: ticker,
			Process: func(ctx context.Context) error {
				return s.collectTickerData(ctx, ticker, startDate, endDate)
			},
		}

		if err := s.workerPool.Submit(task); err != nil {
			s.logger.Errorf("Failed to submit task for %s: %v", ticker, err)
			tickersFailed++
		}
	}

	// Wait for all tasks to complete
	s.logger.Info("Waiting for all collection tasks to complete...")
	if err := s.workerPool.Wait(30 * time.Minute); err != nil {
		s.logger.Errorf("Timeout waiting for tasks: %v", err)
	}

	// Wait for result collector to finish
	<-resultDone

	// Batch insert to ClickHouse
	if len(allQuotes) > 0 {
		s.logger.Infof("Inserting %d quotes to ClickHouse", len(allQuotes))
		if err := s.insertQuotesBatch(ctx, allQuotes); err != nil {
			s.logger.Errorf("Failed to insert quotes to ClickHouse: %v", err)
		} else {
			s.logger.Infof("Successfully inserted %d quotes to ClickHouse", len(allQuotes))
		}
	}

	// Update job log
	jobEndTime := time.Now()
	duration := jobEndTime.Sub(jobStartTime)

	jobLog.EndTime = jobEndTime
	jobLog.DurationSeconds = uint32(duration.Seconds())
	jobLog.TickersProcessed = tickersProcessed
	jobLog.TickersFailed = tickersFailed

	if tickersFailed > 0 && tickersProcessed == 0 {
		jobLog.Status = models.JobStatusFailed
		jobLog.ErrorMessage = "All tickers failed"
	} else if tickersFailed > 0 {
		jobLog.Status = models.JobStatusPartial
		jobLog.ErrorMessage = fmt.Sprintf("%d tickers failed", tickersFailed)
	} else {
		jobLog.Status = models.JobStatusSuccess
	}

	// Log job execution
	if err := s.clickhouse.LogJobExecution(ctx, jobLog); err != nil {
		s.logger.Errorf("Failed to log job execution: %v", err)
	}

	s.logger.Infof("Job %s completed. Processed: %d, Failed: %d, Duration: %v",
		jobID, tickersProcessed, tickersFailed, duration)

	return &jobLog, nil
}

// collectTickerData collects data for a single ticker
func (s *CollectorService) collectTickerData(ctx context.Context, ticker string, startDate, endDate time.Time) error {
	// Fetch historical data from MOEX
	quotes, err := s.moexSource.FetchHistoricalData(ctx, ticker, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to fetch data: %w", err)
	}

	if len(quotes) == 0 {
		s.logger.Warnf("No data found for ticker %s", ticker)
		return nil
	}

	s.logger.Debugf("Fetched %d quotes for %s", len(quotes), ticker)

	// Publish to Kafka
	event := models.DataCollectedEvent{
		Ticker:    ticker,
		Exchange:  models.ExchangeMOEX,
		TradeDate: time.Now(),
		Quotes:    quotes,
	}

	if err := s.kafka.PublishDataCollected(ctx, event); err != nil {
		s.logger.Warnf("Failed to publish data collected event for %s: %v", ticker, err)
		// Don't fail the task if Kafka publish fails
	}

	return nil
}

// insertQuotesBatch inserts quotes in batches
func (s *CollectorService) insertQuotesBatch(ctx context.Context, quotes []models.StockQuote) error {
	const batchSize = 10000

	for i := 0; i < len(quotes); i += batchSize {
		end := i + batchSize
		if end > len(quotes) {
			end = len(quotes)
		}

		batch := quotes[i:end]
		if err := s.clickhouse.InsertStockQuotes(ctx, batch); err != nil {
			return fmt.Errorf("failed to insert batch %d-%d: %w", i, end, err)
		}

		s.logger.Debugf("Inserted batch %d-%d (%d quotes)", i, end, len(batch))
	}

	return nil
}

// CollectTickerList collects data for specific tickers
func (s *CollectorService) CollectTickerList(ctx context.Context, exchange models.Exchange, tickers []string, startDate, endDate time.Time) (*models.JobExecutionLog, error) {
	switch exchange {
	case models.ExchangeMOEX:
		return s.CollectMOEXData(ctx, tickers, startDate, endDate)
	default:
		return nil, fmt.Errorf("unsupported exchange: %s", exchange)
	}
}

// GetLastCollectionDate gets the last collection date for a ticker
func (s *CollectorService) GetLastCollectionDate(ctx context.Context, ticker string, exchange models.Exchange) (time.Time, error) {
	return s.clickhouse.GetLatestQuoteDate(ctx, ticker, exchange)
}
