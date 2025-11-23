# Indicator Calculator Service

The Indicator Calculator Service is responsible for calculating technical indicators and generating trading signals based on collected stock data.

## Features

- **Event-Driven Processing**: Consumes data collection events from Kafka
- **Parallel Calculation**: Processes multiple tickers concurrently using worker pools
- **30+ Technical Indicators**: Calculates comprehensive set of indicators
  - Moving Averages (EMA 10/30/50/100/200, SMA 20/50)
  - Momentum Indicators (RSI, Williams %R)
  - Volume Indicators (VZO, MFI, Up/Down Ratio, Relative Volume)
  - Volatility Indicators (ATR, Volatility, RMV)
  - Trend Indicators (Warrior Trend, TSD)
  - Stock Stages (Wyckoff methodology)
  - Mansfield Relative Strength
  - AVWAP Bands (1Y, 3Y, 5Y)
- **Trading Signals**: Generates actionable trading signals based on indicator patterns
- **Batch Processing**: Efficient batch inserts to ClickHouse
- **Manual Triggers**: REST API for manual calculation requests

## Architecture

```
┌─────────────┐     Kafka      ┌────────────────────┐
│   Data      │───────────────▶│    Indicator       │
│  Collector  │  data.collected │   Calculator       │
└─────────────┘                 └────────────────────┘
                                         │
                                         │ Batch Insert
                                         ▼
                                ┌─────────────────┐
                                │   ClickHouse    │
                                │  - indicators   │
                                │  - signals      │
                                └─────────────────┘
                                         │
                                         │ Publish
                                         ▼
                                    Kafka Topics
                                 indicators.calculated
                                 signals.generated
```

## Configuration

Configuration file: `config.indicator-calculator.yaml`

Key settings:
- **Server Port**: 8083
- **Worker Pool**: 50 concurrent workers
- **Kafka Consumer Group**: stock-market-indicator-calculator
- **Batch Size**: 1000 indicators per batch

## API Endpoints

### Health Check

```bash
GET /health
```

Response:
```json
{
  "status": "healthy",
  "service": "indicator-calculator",
  "time": "2024-01-15T10:30:00Z"
}
```

### Manual Calculation Trigger

```bash
POST /api/v1/calculate
Content-Type: application/json

{
  "exchange": "MOEX",
  "tickers": ["SBER", "GAZP", "LKOH"],
  "start_date": "2023-01-01",
  "end_date": "2024-01-15"
}
```

Response:
```json
{
  "job_id": "manual-calc-1705315800",
  "status": "completed",
  "tickers_processed": 3,
  "tickers_failed": 0,
  "duration_seconds": 2.5,
  "started_at": "2024-01-15T10:30:00Z",
  "completed_at": "2024-01-15T10:30:02Z"
}
```

## Trading Signals Generated

The service generates the following types of trading signals:

1. **Mansfield RS Signals**
   - Zero crosses (bullish/bearish)
   - 3-day positive/negative patterns
   - Confidence: 75-85%

2. **Stage Change Signals**
   - Entry into Stage 2 (Advancing Phase) - BUY
   - Entry into Stage 4 (Declining Phase) - SELL
   - Entry into Stage 3 (Topping Phase) - WARNING
   - Confidence: 70-80%

3. **Trend Change Signals**
   - Warrior Trend switches (LONG/SHORT)
   - TSD Direction changes (UP/DOWN)
   - Confidence: 70-75%

4. **Volume Signals**
   - High relative volume (>= 2.0x average)
   - VZO extremes (>60 bullish, <-60 bearish)
   - Confidence: 60-65%

5. **RSI Signals**
   - Oversold (<30) - BUY
   - Overbought (>70) - SELL
   - Confidence: 60%

6. **Moving Average Crossovers**
   - Golden cross (EMA10 > EMA50) - BUY
   - Death cross (EMA10 < EMA50) - SELL
   - EMA200 crosses
   - Confidence: 70-75%

7. **Williams %R Signals**
   - Oversold (<-80) - BUY
   - Overbought (>-20) - SELL
   - Confidence: 55%

## Running the Service

### Locally

```bash
# Build
make build-indicator-calculator

# Run
./bin/indicator-calculator

# Or with make
make run-indicator-calculator
```

### Docker

```bash
# Build image
docker build -t indicator-calculator -f cmd/indicator-calculator/Dockerfile .

# Run container
docker run -p 8083:8083 \
  -v $(pwd)/config.indicator-calculator.yaml:/app/config.indicator-calculator.yaml \
  indicator-calculator
```

### With Docker Compose

The service is included in the main docker-compose.yml:

```bash
make docker-up
```

## Testing

Test script: `scripts/test-indicator-calculator.sh`

```bash
./scripts/test-indicator-calculator.sh
```

The script will:
1. Check service health
2. Trigger manual calculation for test tickers
3. Verify indicators were calculated
4. Check generated signals

## Monitoring

### Logs

The service logs in JSON format to stdout. Key log events:
- Data collected events received
- Indicator calculation progress
- Signals generated
- ClickHouse insertions
- Kafka publishing

### Metrics

Monitor these key metrics:
- Events consumed from Kafka
- Indicators calculated per second
- Signals generated per second
- ClickHouse insertion latency
- Worker pool utilization

### ClickHouse Queries

```sql
-- Check latest indicators for a ticker
SELECT *
FROM technical_indicators
WHERE ticker = 'SBER'
ORDER BY trade_date DESC
LIMIT 10;

-- Check generated signals
SELECT *
FROM trading_signals
WHERE ticker = 'SBER'
  AND signal_date >= today() - 30
ORDER BY signal_date DESC;

-- Count signals by type
SELECT signal_type, count(*) as count
FROM trading_signals
WHERE signal_date >= today() - 7
GROUP BY signal_type;
```

## Performance

Expected performance:
- **Throughput**: 100-200 tickers/second
- **Latency**: <50ms per ticker (with 200 data points)
- **Memory**: ~500MB for 50 workers
- **CPU**: 2-4 cores recommended

## Troubleshooting

### No indicators being calculated

1. Check Kafka consumer is running:
```bash
curl http://localhost:8083/health
```

2. Check Kafka topic has messages:
```bash
docker exec -it kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic stock.data.collected \
  --from-beginning
```

3. Check service logs for errors

### Insufficient data points error

The service requires at least 200 data points to calculate indicators. Ensure:
- Data Collector has collected enough historical data
- Date range in manual trigger includes at least 200 trading days

### Mansfield RS not calculated

Mansfield RS requires benchmark index data (IMOEX for MOEX, SPX for US exchanges). Ensure:
- Benchmark data is collected in ClickHouse
- Ticker name matches: "IMOEX" or "SPX"

## Development

### Adding New Indicators

1. Implement indicator in `pkg/indicators/`
2. Add calculation in `internal/indicator-calculator/calculator/indicator_calculator.go`
3. Update `models.TechnicalIndicators` struct
4. Add to ClickHouse schema

### Adding New Signal Types

1. Add detection logic in `internal/indicator-calculator/calculator/signal_generator.go`
2. Define signal type in `models.SignalType`
3. Set appropriate confidence level
4. Update documentation

## Dependencies

- ClickHouse: Stock quotes and indicator storage
- Kafka: Event streaming
- Redis: Caching (for benchmark data)

## Next Steps

After Phase 3, the following services will be implemented:
- **Phase 4**: Scheduler Service (automated daily execution)
- **Phase 5**: API Gateway (unified query interface)
- **Phase 6**: Monitoring & Dashboards
- **Phase 7**: Kubernetes Deployment
