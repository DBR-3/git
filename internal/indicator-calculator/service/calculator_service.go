package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/stock-market-system/internal/indicator-calculator/calculator"
	"github.com/stock-market-system/pkg/config"
	"github.com/stock-market-system/pkg/logger"
	"github.com/stock-market-system/pkg/messaging"
	"github.com/stock-market-system/pkg/models"
	"github.com/stock-market-system/pkg/storage/clickhouse"
)

// CalculatorService handles technical indicator calculations
type CalculatorService struct {
	cfg             *config.Config
	logger          *logger.Logger
	clickhouse      *clickhouse.Client
	kafkaProducer   *messaging.Producer
	kafkaConsumer   *messaging.Consumer
	calculator      *calculator.IndicatorCalculator
	signalGenerator *calculator.SignalGenerator
	workerPool      int
	batchSize       int
}

// NewCalculatorService creates a new calculator service
func NewCalculatorService(
	cfg *config.Config,
	log *logger.Logger,
	ch *clickhouse.Client,
	producer *messaging.Producer,
	consumer *messaging.Consumer,
) *CalculatorService {
	return &CalculatorService{
		cfg:             cfg,
		logger:          log,
		clickhouse:      ch,
		kafkaProducer:   producer,
		kafkaConsumer:   consumer,
		calculator:      calculator.NewIndicatorCalculator(),
		signalGenerator: calculator.NewSignalGenerator(),
		workerPool:      50, // Number of concurrent calculation workers
		batchSize:       1000,
	}
}

// Start begins consuming data collection events and calculating indicators
func (s *CalculatorService) Start(ctx context.Context) error {
	s.logger.Info("Starting indicator calculator service")

	// Start consuming data collected events
	return s.kafkaConsumer.ConsumeDataCollected(ctx, s.handleDataCollectedEvent)
}

// handleDataCollectedEvent processes a single data collection event
func (s *CalculatorService) handleDataCollectedEvent(ctx context.Context, event models.DataCollectedEvent) error {
	startTime := time.Now()
	s.logger.WithFields(map[string]interface{}{
		"exchange":   event.Exchange,
		"num_quotes": len(event.Quotes),
	}).Info("Processing data collected event")

	// Group quotes by ticker
	tickerQuotes := s.groupQuotesByTicker(event.Quotes)

	// Calculate indicators for each ticker in parallel
	results := s.calculateIndicatorsParallel(ctx, tickerQuotes, event.Exchange, event.Timeframe)

	// Aggregate all indicators and signals
	allIndicators := make([]models.TechnicalIndicators, 0)
	allSignals := make([]models.TradingSignal, 0)

	for _, result := range results {
		if result.Error != nil {
			s.logger.WithFields(map[string]interface{}{
				"ticker": result.Ticker,
				"error":  result.Error,
			}).Error("Failed to calculate indicators for ticker")
			continue
		}

		allIndicators = append(allIndicators, result.Indicators...)
		allSignals = append(allSignals, result.Signals...)
	}

	s.logger.WithFields(map[string]interface{}{
		"num_indicators": len(allIndicators),
		"num_signals":    len(allSignals),
	}).Info("Calculated indicators")

	// Store indicators in ClickHouse
	if len(allIndicators) > 0 {
		if err := s.clickhouse.InsertTechnicalIndicators(ctx, allIndicators); err != nil {
			s.logger.WithError(err).Error("Failed to insert technical indicators")
			return fmt.Errorf("failed to insert indicators: %w", err)
		}
	}

	// Store signals in ClickHouse
	if len(allSignals) > 0 {
		if err := s.clickhouse.InsertTradingSignals(ctx, allSignals); err != nil {
			s.logger.WithError(err).Error("Failed to insert trading signals")
			return fmt.Errorf("failed to insert signals: %w", err)
		}
	}

	// Publish indicator calculated events to Kafka
	if len(allIndicators) > 0 {
		if err := s.publishIndicatorEvents(ctx, allIndicators, allSignals); err != nil {
			s.logger.WithError(err).Error("Failed to publish indicator events")
			// Don't fail the entire process if Kafka publish fails
		}
	}

	duration := time.Since(startTime)
	s.logger.WithFields(map[string]interface{}{
		"num_tickers":    len(tickerQuotes),
		"num_indicators": len(allIndicators),
		"num_signals":    len(allSignals),
		"duration_ms":    duration.Milliseconds(),
	}).Info("Completed indicator calculation event")

	return nil
}

// CalculationResult holds the result of indicator calculation for a ticker
type CalculationResult struct {
	Ticker     string
	Indicators []models.TechnicalIndicators
	Signals    []models.TradingSignal
	Error      error
}

// calculateIndicatorsParallel calculates indicators for multiple tickers in parallel
func (s *CalculatorService) calculateIndicatorsParallel(
	ctx context.Context,
	tickerQuotes map[string][]models.StockQuote,
	exchange models.Exchange,
	timeframe models.Timeframe,
) []CalculationResult {
	// Create channels for work distribution
	type job struct {
		ticker string
		quotes []models.StockQuote
	}

	jobs := make(chan job, len(tickerQuotes))
	results := make(chan CalculationResult, len(tickerQuotes))

	// Start workers
	var wg sync.WaitGroup
	numWorkers := s.workerPool
	if numWorkers > len(tickerQuotes) {
		numWorkers = len(tickerQuotes)
	}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				select {
				case <-ctx.Done():
					results <- CalculationResult{
						Ticker: j.ticker,
						Error:  ctx.Err(),
					}
					return
				default:
					result := s.calculateIndicatorsForTicker(ctx, j.ticker, j.quotes, exchange, timeframe)
					results <- result
				}
			}
		}()
	}

	// Send jobs
	for ticker, quotes := range tickerQuotes {
		jobs <- job{ticker: ticker, quotes: quotes}
	}
	close(jobs)

	// Wait for all workers to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	allResults := make([]CalculationResult, 0, len(tickerQuotes))
	for result := range results {
		allResults = append(allResults, result)
	}

	return allResults
}

// calculateIndicatorsForTicker calculates indicators for a single ticker
func (s *CalculatorService) calculateIndicatorsForTicker(
	ctx context.Context,
	ticker string,
	quotes []models.StockQuote,
	exchange models.Exchange,
	timeframe models.Timeframe,
) CalculationResult {
	// Check minimum data points
	if len(quotes) < 200 {
		return CalculationResult{
			Ticker: ticker,
			Error:  fmt.Errorf("insufficient data points: need 200, got %d", len(quotes)),
		}
	}

	// Load benchmark data if needed for Mansfield RS
	if err := s.loadBenchmarkData(ctx, exchange); err != nil {
		s.logger.WithFields(map[string]interface{}{
			"ticker": ticker,
			"error":  err,
		}).Warn("Failed to load benchmark data, Mansfield RS will be empty")
	}

	// Calculate all technical indicators
	indicators, err := s.calculator.CalculateAll(quotes, exchange, timeframe)
	if err != nil {
		return CalculationResult{
			Ticker: ticker,
			Error:  fmt.Errorf("failed to calculate indicators: %w", err),
		}
	}

	// Generate trading signals
	signals := s.signalGenerator.GenerateSignals(indicators)

	return CalculationResult{
		Ticker:     ticker,
		Indicators: indicators,
		Signals:    signals,
		Error:      nil,
	}
}

// loadBenchmarkData loads benchmark index data for Mansfield RS calculation
func (s *CalculatorService) loadBenchmarkData(ctx context.Context, exchange models.Exchange) error {
	var benchmarkTicker string
	switch exchange {
	case models.ExchangeMOEX:
		benchmarkTicker = "IMOEX"
	case models.ExchangeNASDAQ, models.ExchangeNYSE:
		benchmarkTicker = "SPX"
	default:
		benchmarkTicker = "IMOEX"
	}

	// Load last 252 days (1 year of trading days) for the benchmark
	endDate := time.Now()
	startDate := endDate.AddDate(-1, 0, 0)

	quotes, err := s.clickhouse.GetStockQuotes(ctx, benchmarkTicker, exchange, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to load benchmark data: %w", err)
	}

	if len(quotes) > 0 {
		s.calculator.SetBenchmarkData(benchmarkTicker, quotes)
	}

	return nil
}

// groupQuotesByTicker groups stock quotes by ticker symbol
func (s *CalculatorService) groupQuotesByTicker(quotes []models.StockQuote) map[string][]models.StockQuote {
	grouped := make(map[string][]models.StockQuote)

	for _, quote := range quotes {
		grouped[quote.Ticker] = append(grouped[quote.Ticker], quote)
	}

	// Sort quotes by date for each ticker (important for indicator calculation)
	for ticker := range grouped {
		// Quotes should already be sorted from data collector, but just to be safe
		// we rely on the order they come in
		_ = ticker
	}

	return grouped
}

// publishIndicatorEvents publishes indicator calculated events to Kafka
func (s *CalculatorService) publishIndicatorEvents(
	ctx context.Context,
	indicators []models.TechnicalIndicators,
	signals []models.TradingSignal,
) error {
	// Group indicators by ticker for efficient publishing
	tickerIndicators := make(map[string][]models.TechnicalIndicators)
	for _, ind := range indicators {
		tickerIndicators[ind.Ticker] = append(tickerIndicators[ind.Ticker], ind)
	}

	// Group signals by ticker
	tickerSignals := make(map[string][]models.TradingSignal)
	for _, sig := range signals {
		tickerSignals[sig.Ticker] = append(tickerSignals[sig.Ticker], sig)
	}

	// Publish events for each ticker
	for ticker, inds := range tickerIndicators {
		event := models.IndicatorCalculatedEvent{
			Ticker:     ticker,
			Exchange:   inds[0].Exchange,
			Timeframe:  inds[0].Timeframe,
			Indicators: inds,
			Signals:    tickerSignals[ticker],
			Timestamp:  time.Now(),
		}

		if err := s.kafkaProducer.PublishIndicatorCalculated(ctx, event); err != nil {
			s.logger.WithFields(map[string]interface{}{
				"ticker": ticker,
				"error":  err,
			}).Error("Failed to publish indicator event")
			// Continue with other tickers
		}
	}

	return nil
}

// CalculateManual manually triggers indicator calculation for specific tickers
func (s *CalculatorService) CalculateManual(
	ctx context.Context,
	exchange models.Exchange,
	tickers []string,
	startDate, endDate time.Time,
) (*models.JobExecutionLog, error) {
	jobLog := &models.JobExecutionLog{
		JobID:     fmt.Sprintf("manual-calc-%d", time.Now().Unix()),
		JobType:   models.JobTypeIndicatorCalculation,
		Status:    models.JobStatusRunning,
		StartedAt: time.Now(),
	}

	s.logger.WithFields(map[string]interface{}{
		"exchange":    exchange,
		"num_tickers": len(tickers),
		"start_date":  startDate.Format("2006-01-02"),
		"end_date":    endDate.Format("2006-01-02"),
	}).Info("Starting manual indicator calculation")

	successCount := 0
	failedCount := 0

	// Load quotes for all tickers
	allQuotes := make([]models.StockQuote, 0)
	for _, ticker := range tickers {
		quotes, err := s.clickhouse.GetStockQuotes(ctx, ticker, exchange, startDate, endDate)
		if err != nil {
			s.logger.WithFields(map[string]interface{}{
				"ticker": ticker,
				"error":  err,
			}).Error("Failed to load quotes for ticker")
			failedCount++
			continue
		}

		allQuotes = append(allQuotes, quotes...)
	}

	// Create a fake data collected event
	event := models.DataCollectedEvent{
		Exchange:  exchange,
		Timeframe: models.TimeframeDaily,
		Quotes:    allQuotes,
		Timestamp: time.Now(),
	}

	// Process the event
	if err := s.handleDataCollectedEvent(ctx, event); err != nil {
		jobLog.Status = models.JobStatusFailed
		jobLog.ErrorMessage = err.Error()
		jobLog.CompletedAt = time.Now()
		s.clickhouse.LogJobExecution(ctx, *jobLog)
		return jobLog, err
	}

	successCount = len(tickers) - failedCount

	jobLog.Status = models.JobStatusCompleted
	jobLog.TickersProcessed = uint32(successCount)
	jobLog.TickersFailed = uint32(failedCount)
	jobLog.CompletedAt = time.Now()
	s.clickhouse.LogJobExecution(ctx, *jobLog)

	s.logger.WithFields(map[string]interface{}{
		"job_id":            jobLog.JobID,
		"tickers_processed": successCount,
		"tickers_failed":    failedCount,
		"duration_sec":      jobLog.CompletedAt.Sub(jobLog.StartedAt).Seconds(),
	}).Info("Completed manual indicator calculation")

	return jobLog, nil
}
