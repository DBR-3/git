# Stock Market Data Collection & Analysis System

Высокопроизводительная система для автоматического сбора, обработки и анализа данных по акциям российского (ММВБ) и американского (NASDAQ/NYSE) рынков, разработанная на Go с использованием ClickHouse, Kafka, Redis и Kubernetes.

## 🎯 Возможности

- ✅ **Автоматический сбор данных** с ММВБ и NASDAQ/NYSE
- ✅ **Высокая производительность**: обработка 3000+ тикеров за 10-15 минут
- ✅ **Расчет технических индикаторов**: RSI, AVWAP, Mansfield RS, Williams %R, MFI, RMV, VZO и др.
- ✅ **Генерация торговых сигналов** на основе анализа индикаторов
- ✅ **Микросервисная архитектура** с легким масштабированием
- ✅ **Real-time обработка** через Kafka
- ✅ **REST API и gRPC** для доступа к данным
- ✅ **Monitoring & Observability**: Prometheus, Grafana, Jaeger

## 📋 Требования

- Go 1.21+
- Docker & Docker Compose
- Kubernetes 1.28+ (для production)
- ClickHouse 23.x
- Redis 7.x
- Kafka 3.x

## 🚀 Быстрый старт

### 1. Клонирование репозитория

```bash
git clone https://github.com/yourusername/stock-market-system.git
cd stock-market-system
```

### 2. Запуск инфраструктуры

```bash
# Запуск всех сервисов через Docker Compose
docker-compose up -d

# Проверка статуса
docker-compose ps
```

Доступные сервисы:
- **ClickHouse UI**: http://localhost:8123
- **Kafka UI**: http://localhost:8080
- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090
- **Jaeger UI**: http://localhost:16686

### 3. Инициализация базы данных

```bash
# База данных автоматически инициализируется при старте ClickHouse
# Проверка структуры:
docker exec -it stock-clickhouse clickhouse-client --query "SHOW TABLES FROM stock_market"
```

### 4. Сборка и запуск сервисов

```bash
# Установка зависимостей
go mod download

# Сборка всех сервисов
make build

# Или запуск отдельных сервисов для разработки:
go run cmd/ticker-fetcher/main.go
go run cmd/data-collector/main.go
go run cmd/indicator-calculator/main.go
go run cmd/scheduler/main.go
go run cmd/api-gateway/main.go
```

## 🏗️ Архитектура

```
┌─────────────────────────────────────────────────────────────┐
│                     API Gateway (gRPC/REST)                  │
│                   Port 8000 (REST) / 50051 (gRPC)            │
└────────────────────────────┬────────────────────────────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
┌───────▼────────┐  ┌────────▼────────┐  ┌───────▼─────────┐
│  Ticker         │  │  Data           │  │  Indicator      │
│  Fetcher        │  │  Collector      │  │  Calculator     │
│  Service        │  │  Service        │  │  Service        │
│  (Port 8081)    │  │  (Port 8082)    │  │  (Port 8083)    │
└───────┬────────┘  └────────┬────────┘  └───────┬─────────┘
        │                    │                    │
        └────────────────────┼────────────────────┘
                             │
                    ┌────────▼────────┐
                    │  Kafka Cluster  │
                    │  (Port 9092)    │
                    └────────┬────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
┌───────▼────────┐  ┌────────▼────────┐  ┌───────▼─────────┐
│  ClickHouse    │  │  Redis Cache    │  │  Scheduler      │
│  (Port 9000)   │  │  (Port 6379)    │  │  (Cron Jobs)    │
└────────────────┘  └─────────────────┘  └─────────────────┘
```

### Компоненты системы

#### 1. **Ticker Fetcher Service** (cmd/ticker-fetcher)
- Получение актуальных списков тикеров с NASDAQ и ММВБ
- Кэширование в Redis (TTL: 1 день)
- REST API для запроса списков

#### 2. **Data Collector Service** (cmd/data-collector)
- Параллельный сбор исторических данных (OHLCV)
- Worker pool для высокой производительности
- Публикация событий в Kafka
- Batch insert в ClickHouse

#### 3. **Indicator Calculator Service** (cmd/indicator-calculator)
- Расчет всех технических индикаторов
- Потребление событий из Kafka
- Stream processing
- Сохранение результатов в ClickHouse

#### 4. **Scheduler Service** (cmd/scheduler)
- Автоматический запуск pipeline по расписанию
- Оркестрация задач
- Мониторинг выполнения
- Уведомления об ошибках

#### 5. **API Gateway** (cmd/api-gateway)
- REST API для внешних клиентов
- gRPC для внутренних сервисов
- Аутентификация и авторизация
- Rate limiting

## 📊 Технические индикаторы

Система рассчитывает следующие индикаторы:

### Trend Indicators
- EMA (10, 30, 50, 100, 200)
- SMA (20, 50)

### Momentum Indicators
- RSI (14)
- Williams %R (14)

### Volume Indicators
- Volume SMA (20)
- Relative Volume
- VZO (Volume Zone Oscillator)
- Up/Down Ratio (50)

### Volatility Indicators
- ATR (14)
- Volatility (15)
- RMV (Relative Market Volatility)

### Relative Strength
- Mansfield Relative Strength

### Advanced Indicators
- MFI (Money Flow Index, 58)
- Warrior Trend Indicator
- Trend Strength (TSD)
- Stock Stages (Wyckoff Method)
- AVWAP Bands (1Y, 3Y, 5Y)

## 🔌 API Endpoints

### REST API

```bash
# Health check
GET /health

# Get tickers
GET /api/v1/tickers/nasdaq?type=stocks
GET /api/v1/tickers/moex?boardid=TQBR

# Get stock quotes
GET /api/v1/stocks/{ticker}/quotes?from=2024-01-01&to=2024-12-31&exchange=MOEX

# Get technical indicators
GET /api/v1/stocks/{ticker}/indicators?timeframe=DAILY&exchange=MOEX

# Stock screener
GET /api/v1/stocks/screener?rs_min=0&stage=Advancing&exchange=MOEX&limit=100

# Get trading signals
GET /api/v1/signals?ticker={ticker}&from=2024-01-01&to=2024-12-31

# Trigger manual data collection
POST /api/v1/jobs/collect
{
  "exchange": "MOEX",
  "tickers": ["SBER", "GAZP", "LKOH"]
}

# Get job status
GET /api/v1/jobs/{job_id}
```

### gRPC Services

См. proto файлы в `api/proto/`:
- `ticker_service.proto`
- `data_collector.proto`
- `indicator_calculator.proto`

## 📈 База данных

### Основные таблицы ClickHouse

- `tickers` - список всех тикеров
- `stock_quotes` - котировки (OHLCV)
- `stock_quotes_weekly` - недельные агрегаты
- `benchmark_indices` - индексы (IMOEX, S&P500)
- `technical_indicators` - технические индикаторы
- `trading_signals` - торговые сигналы
- `job_execution_log` - логи выполнения задач

### Партиционирование

- По месяцам: `toYYYYMM(trade_date)`
- ORDER BY: `(ticker, exchange, trade_date)`
- TTL: intraday данные хранятся 30 дней

## ⚙️ Конфигурация

Конфигурация через файл `config.yaml` или переменные окружения:

```yaml
server:
  port: 8080
  grpcport: 50051

clickhouse:
  addresses:
    - localhost:9000
  database: stock_market
  username: stock_app
  password: stock_password
  batchsize: 10000

redis:
  address: localhost:6379
  password: redis_password
  database: 0

kafka:
  brokers:
    - localhost:9092
  groupid: stock-market-consumer
  topicdatacollect: stock.data.collected
  topicindicators: stock.indicators.calculated

logger:
  level: info
  format: json
```

Переменные окружения (префикс `STOCK_`):
```bash
STOCK_SERVER_PORT=8080
STOCK_CLICKHOUSE_ADDRESSES=localhost:9000
STOCK_REDIS_ADDRESS=localhost:6379
STOCK_KAFKA_BROKERS=localhost:9092
```

## 🔄 Data Flow

```
1. Scheduler запускает ежедневную задачу
   ↓
2. Ticker Fetcher получает список акций
   ↓
3. Data Collector собирает котировки параллельно
   ↓
4. Публикация в Kafka (topic: stock.data.collected)
   ↓
5. ClickHouse Consumer записывает в БД
   ↓
6. Indicator Calculator читает из Kafka
   ↓
7. Расчет индикаторов и публикация результатов
   ↓
8. Сохранение в ClickHouse
   ↓
9. Генерация торговых сигналов
```

## 📦 Deployment

### Docker Compose (для разработки)

```bash
docker-compose up -d
```

### Kubernetes (production)

```bash
# Создание namespace
kubectl create namespace stock-market

# Применение конфигураций
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/clickhouse/
kubectl apply -f k8s/kafka/
kubectl apply -f k8s/redis/
kubectl apply -f k8s/services/

# Проверка статуса
kubectl get pods -n stock-market
kubectl get services -n stock-market
```

## 📊 Мониторинг

### Prometheus Metrics

Каждый сервис экспортирует метрики:
- `http_requests_total` - количество HTTP запросов
- `http_request_duration_seconds` - latency запросов
- `kafka_messages_produced_total` - сообщения в Kafka
- `kafka_messages_consumed_total` - обработанные сообщения
- `clickhouse_batch_insert_duration_seconds` - время batch insert
- `job_execution_duration_seconds` - время выполнения задач
- `tickers_processed_total` - обработанные тикеры

### Grafana Dashboards

Преднастроенные дашборды в `monitoring/grafana/dashboards/`:
- Service Overview
- ClickHouse Performance
- Kafka Lag Monitoring
- Job Execution Tracking

### Distributed Tracing

Jaeger UI доступен по адресу http://localhost:16686

## 🧪 Тестирование

```bash
# Unit tests
go test ./...

# Integration tests
go test -tags=integration ./...

# Benchmark tests
go test -bench=. ./pkg/indicators/

# Coverage
go test -cover ./...
```

## 🔧 Development

```bash
# Установка инструментов разработки
make install-tools

# Линтинг
make lint

# Форматирование кода
make fmt

# Генерация proto файлов
make proto-gen

# Сборка Docker образов
make docker-build

# Запуск в dev режиме с hot-reload
make dev
```

## 📝 TODO

- [ ] Поддержка real-time данных (WebSocket)
- [ ] Machine Learning модели для предсказаний
- [ ] Backtesting engine
- [ ] Web UI для визуализации
- [ ] Mobile приложение
- [ ] Интеграция с брокерами для автоматической торговли
- [ ] Поддержка криптовалют
- [ ] Алерты через Telegram/Email
- [ ] Portfolio tracking

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

---

**⭐ Если проект полезен, поставьте звезду на GitHub!**

---

## 📚 Дополнительная документация

- [Architecture & Roadmap](ARCHITECTURE_ROADMAP.md) - Детальная архитектура и план разработки
- [Database Schema](schema/clickhouse/) - Схема ClickHouse
- [API Documentation](api/) - gRPC и REST API спецификации
