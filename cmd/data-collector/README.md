# Data Collector Service

High-performance service for collecting historical stock market data from MOEX and other exchanges.

## Features

- ✅ Parallel data collection (100 concurrent workers)
- ✅ MOEX API integration with pagination
- ✅ Batch inserts to ClickHouse (10,000 rows per batch)
- ✅ Kafka event publishing
- ✅ Automatic retry on failures
- ✅ Job execution logging
- ✅ REST API for manual triggers

## Architecture

```
┌──────────────┐
│   REST API   │
└──────┬───────┘
       │
┌──────▼────────────────┐
│  Collector Service    │
│  - Job Management     │
│  - Result Aggregation │
└──────┬────────────────┘
       │
┌──────▼────────────────┐
│   Worker Pool (100)   │
│  - Parallel Execution │
│  - Task Distribution  │
└──────┬────────────────┘
       │
┌──────▼────────────────┐
│  MOEX Data Source     │
│  - API Integration    │
│  - Pagination         │
│  - Rate Limiting      │
└──────┬────────────────┘
       │
┌──────▼────────────────┐     ┌─────────────────┐
│    ClickHouse         │────▶│  Kafka Producer │
│  - Batch Inserts      │     │  - Event Stream │
│  - Job Logging        │     └─────────────────┘
└───────────────────────┘
```

## API Endpoints

### POST /api/v1/collect

Starts data collection for specified tickers.

**Request:**
```json
{
  "exchange": "MOEX",
  "tickers": ["SBER", "GAZP", "LKOH"],
  "start_date": "2024-01-01",
  "end_date": "2024-12-31"
}
```

**Response:**
```json
{
  "message": "Collection started",
  "tickers": 3,
  "start_date": "2024-01-01",
  "end_date": "2024-12-31"
}
```

### GET /api/v1/last-date/:exchange/:ticker

Gets the last collection date for a ticker.

**Response:**
```json
{
  "ticker": "SBER",
  "exchange": "MOEX",
  "last_date": "2024-12-31"
}
```

### GET /health

Health check endpoint.

**Response:**
```json
{
  "status": "healthy",
  "service": "data-collector",
  "timestamp": 1234567890
}
```

## Configuration

Configuration is loaded from `config.data-collector.yaml`:

```yaml
server:
  port: 8082  # HTTP server port

clickhouse:
  addresses:
    - localhost:9000
  database: stock_market
  batchsize: 10000  # Batch size for inserts

kafka:
  brokers:
    - localhost:9092
  topicdatacollect: stock.data.collected

logger:
  level: info
```

## Running the Service

### Local Development

```bash
# With infrastructure running
make infra-up

# Run the service
make run-data-collector

# Or with custom config
go run cmd/data-collector/main.go -config config.data-collector.yaml
```

### Docker

```bash
# Build image
docker build -t stock-data-collector -f cmd/data-collector/Dockerfile .

# Run container
docker run -p 8082:8082 \
  -e STOCK_CLICKHOUSE_ADDRESSES=clickhouse:9000 \
  -e STOCK_KAFKA_BROKERS=kafka:9092 \
  stock-data-collector
```

## Testing

```bash
# Run test script
./scripts/test-data-collector.sh

# Or manually
curl -X POST http://localhost:8082/api/v1/collect \
  -H "Content-Type: application/json" \
  -d '{
    "exchange": "MOEX",
    "tickers": ["SBER", "GAZP"],
    "start_date": "2024-01-01",
    "end_date": "2024-12-31"
  }'
```

## Performance

### Collection Speed
- **Single ticker**: ~500-1000 rows/sec
- **Parallel (100 workers)**: ~50,000-100,000 rows/sec
- **3000 tickers**: ~10-15 minutes (full historical data)

### Resource Usage
- **Memory**: ~200-500 MB
- **CPU**: 2-4 cores recommended
- **Network**: ~1-5 Mbps

### Optimization
- Worker pool size: adjustable (default: 100)
- Batch size: configurable (default: 10,000)
- Rate limiting: 100ms delay between pages

## Data Flow

1. **Request received** via REST API
2. **Tasks created** for each ticker
3. **Worker pool** processes tasks in parallel
4. **MOEX API** fetched with pagination
5. **Results collected** from all workers
6. **Batch insert** to ClickHouse
7. **Events published** to Kafka
8. **Job logged** to ClickHouse

## Monitoring

### Logs
```bash
# View logs
docker logs -f stock-data-collector

# Filter for errors
docker logs stock-data-collector 2>&1 | grep ERROR
```

### Metrics
Check ClickHouse for job execution logs:
```sql
SELECT
    job_id,
    status,
    tickers_processed,
    tickers_failed,
    duration_seconds,
    start_time
FROM job_execution_log
ORDER BY start_time DESC
LIMIT 10;
```

### Kafka Events
Monitor Kafka topics:
```bash
# View topics
make kafka-topics

# Check Kafka UI
open http://localhost:8080
```

## Error Handling

### Retry Logic
- Failed API calls: 3 retries with exponential backoff
- Rate limiting: automatic delay between requests
- Network errors: task marked as failed, job continues

### Failed Tasks
- Logged to `job_execution_log` table
- Published to `stock.errors` Kafka topic
- Can be retried manually

## Troubleshooting

### Service won't start
```bash
# Check ClickHouse connection
docker exec -it stock-clickhouse clickhouse-client --query "SELECT 1"

# Check Kafka
docker exec -it stock-kafka kafka-topics --bootstrap-server localhost:9092 --list
```

### No data collected
```bash
# Check service logs
docker logs stock-data-collector

# Verify MOEX API access
curl "https://iss.moex.com/iss/history/engines/stock/markets/shares/securities/SBER.json?from=2024-01-01&till=2024-12-31"
```

### Slow performance
- Increase worker count in service initialization
- Check ClickHouse disk I/O
- Verify network connectivity to MOEX

## Development

### Adding New Data Sources

1. Create new source in `internal/data-collector/sources/`
2. Implement `FetchHistoricalData` method
3. Add to collector service
4. Update API handlers

### Testing
```bash
# Unit tests
go test ./internal/data-collector/...

# Integration tests
go test -tags=integration ./internal/data-collector/...
```

## See Also

- [Architecture Documentation](../../ARCHITECTURE_ROADMAP.md)
- [Main README](../../README_STOCK_MARKET.md)
- [Quick Start Guide](../../QUICKSTART.md)
