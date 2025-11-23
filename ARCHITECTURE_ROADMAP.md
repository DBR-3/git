# Stock Market Data Collection & Analysis System - Architecture & Roadmap

## Обзор системы

Высокопроизводительная система для автоматического сбора, обработки и анализа данных по акциям российского (ММВБ) и американского (NASDAQ) рынков с использованием Go, ClickHouse, Kafka, Redis и Kubernetes.

## Текущее состояние

- ✅ Python скрипты для ручного сбора данных
- ✅ Расчет технических индикаторов (RSI, AVWAP, Mansfield RS, Williams %R, VZO, MFI, RMV, etc.)
- ❌ Хранение в Excel (медленно, не масштабируемо)
- ❌ Ручной запуск ежедневно
- ❌ Долгое время выполнения

## Целевая архитектура

### Компоненты системы

```
┌─────────────────────────────────────────────────────────────┐
│                     API Gateway (gRPC/REST)                  │
│                   (Go + Traefik/Envoy)                       │
└────────────────────────────┬────────────────────────────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
┌───────▼────────┐  ┌────────▼────────┐  ┌───────▼─────────┐
│  Ticker         │  │  Data           │  │  Indicator      │
│  Fetcher        │  │  Collector      │  │  Calculator     │
│  Service        │  │  Service        │  │  Service        │
│  (Go)           │  │  (Go)           │  │  (Go)           │
└───────┬────────┘  └────────┬────────┘  └───────┬─────────┘
        │                    │                    │
        └────────────────────┼────────────────────┘
                             │
                    ┌────────▼────────┐
                    │  Kafka Cluster  │
                    │  (Event Stream) │
                    └────────┬────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
┌───────▼────────┐  ┌────────▼────────┐  ┌───────▼─────────┐
│  ClickHouse    │  │  Redis Cache    │  │  Scheduler      │
│  (OLAP DB)     │  │  (Hot Data)     │  │  (Cron Jobs)    │
└────────────────┘  └─────────────────┘  └─────────────────┘
```

---

## ROADMAP - Пошаговый план реализации

### ФАЗА 1: Проектирование и подготовка (1-2 недели)

#### Шаг 1.1: Проектирование базы данных ClickHouse
**Задачи:**
- Разработать схему таблиц для хранения котировок (OHLCV)
- Создать таблицы для индикаторов
- Настроить партиционирование по датам
- Создать материализованные представления для агрегаций

**Файлы:**
- `schema/clickhouse/01_stocks.sql`
- `schema/clickhouse/02_indicators.sql`
- `schema/clickhouse/03_views.sql`

#### Шаг 1.2: Определение API контрактов
**Задачи:**
- Разработать Proto файлы для gRPC сервисов
- Определить REST API endpoints
- Создать схемы данных

**Файлы:**
- `api/proto/ticker_service.proto`
- `api/proto/data_collector.proto`
- `api/proto/indicator_calculator.proto`
- `api/openapi/swagger.yaml`

#### Шаг 1.3: Настройка инфраструктуры
**Задачи:**
- Создать Docker Compose для локальной разработки
- Настроить Kubernetes манифесты
- Настроить CI/CD pipeline

**Файлы:**
- `docker-compose.yml`
- `k8s/namespace.yaml`
- `.github/workflows/ci.yaml`

---

### ФАЗА 2: Реализация базовых сервисов (2-3 недели)

#### Шаг 2.1: Ticker Fetcher Service
**Функционал:**
- Получение списка акций с NASDAQ (nasdaqtrader.com)
- Получение списка акций с ММВБ (moex.com API)
- Фильтрация по типу (stocks/ETF)
- Кэширование в Redis

**Технологии:**
- Go 1.21+
- Redis для кэша списков
- HTTP client с retry logic

**Файлы:**
- `services/ticker-fetcher/main.go`
- `services/ticker-fetcher/handlers/nasdaq.go`
- `services/ticker-fetcher/handlers/moex.go`
- `services/ticker-fetcher/cache/redis.go`

**Endpoints:**
```
GET /api/v1/tickers/nasdaq?type=stocks
GET /api/v1/tickers/moex?boardid=TQBR
```

#### Шаг 2.2: Data Collector Service
**Функционал:**
- Получение исторических данных (OHLCV)
- Параллельный сбор данных (goroutines pool)
- Публикация событий в Kafka
- Запись в ClickHouse

**Технологии:**
- Go concurrency (worker pool)
- Kafka producer
- ClickHouse batch inserts
- Rate limiting для API

**Файлы:**
- `services/data-collector/main.go`
- `services/data-collector/sources/moex.go`
- `services/data-collector/sources/yahoo.go`
- `services/data-collector/workers/pool.go`
- `services/data-collector/storage/clickhouse.go`

**Kafka Topics:**
- `stock.data.collected` - сырые данные котировок
- `stock.data.errors` - ошибки сбора

#### Шаг 2.3: ClickHouse Integration
**Функционал:**
- Batch inserts для высокой производительности
- Connection pooling
- Retry logic
- Monitoring метрики

**Файлы:**
- `pkg/storage/clickhouse/client.go`
- `pkg/storage/clickhouse/models.go`
- `pkg/storage/clickhouse/repository.go`

---

### ФАЗА 3: Расчет индикаторов (2-3 недели)

#### Шаг 3.1: Indicator Calculator Service
**Функционал:**
- Расчет технических индикаторов:
  - RSI (Relative Strength Index)
  - AVWAP (Anchored VWAP with bands)
  - Mansfield Relative Strength
  - Williams %R
  - MFI (Money Flow Index)
  - RMV (Relative Market Volatility)
  - VZO (Volume Zone Oscillator)
  - Trend Strength Indicator
  - Stock Stages (Wyckoff)
  - EMA (10, 30, 50, 100, 200)
- Потребление событий из Kafka
- Параллельные вычисления
- Запись результатов в ClickHouse

**Технологии:**
- Go + математические библиотеки
- Kafka consumer group
- Stream processing

**Файлы:**
- `services/indicator-calculator/main.go`
- `services/indicator-calculator/indicators/rsi.go`
- `services/indicator-calculator/indicators/avwap.go`
- `services/indicator-calculator/indicators/mansfield_rs.go`
- `services/indicator-calculator/indicators/williams.go`
- `services/indicator-calculator/indicators/mfi.go`
- `services/indicator-calculator/indicators/stages.go`
- `services/indicator-calculator/processor/stream.go`

**Kafka Topics:**
- Consumer: `stock.data.collected`
- Producer: `stock.indicators.calculated`

#### Шаг 3.2: Портирование Python алгоритмов на Go
**Задачи:**
- Перевести все функции расчета индикаторов с Python на Go
- Создать unit-тесты с валидацией против Python результатов
- Оптимизировать производительность

---

### ФАЗА 4: Scheduler & Automation (1 неделя)

#### Шаг 4.1: Scheduler Service
**Функционал:**
- Ежедневный запуск сбора данных (cron)
- Оркестрация pipeline:
  1. Обновление списка тикеров
  2. Сбор котировок
  3. Расчет индикаторов
  4. Генерация отчетов
- Мониторинг статуса выполнения
- Уведомления об ошибках

**Технологии:**
- Go + robfig/cron
- State management
- Alert notifications

**Файлы:**
- `services/scheduler/main.go`
- `services/scheduler/jobs/daily_update.go`
- `services/scheduler/orchestrator/pipeline.go`

**Расписание:**
```
# После закрытия торгов ММВБ (19:00 MSK)
0 20 * * 1-5  # Сбор данных MOEX

# После закрытия торгов NYSE (22:00 EST = 06:00 MSK)
0 7 * * 2-6   # Сбор данных US
```

---

### ФАЗА 5: API Gateway & Query Layer (1-2 недели)

#### Шаг 5.1: API Gateway
**Функционал:**
- REST API для внешних клиентов
- gRPC для внутренних сервисов
- Authentication & Authorization
- Rate limiting
- Request validation

**Endpoints:**
```
GET  /api/v1/stocks/{ticker}/quotes?from=2024-01-01&to=2024-12-31
GET  /api/v1/stocks/{ticker}/indicators?indicators=RSI,AVWAP
GET  /api/v1/stocks/screener?rs_min=0&stage=Advancing
POST /api/v1/stocks/batch-query
```

**Файлы:**
- `services/api-gateway/main.go`
- `services/api-gateway/handlers/stocks.go`
- `services/api-gateway/handlers/indicators.go`
- `services/api-gateway/middleware/auth.go`
- `services/api-gateway/middleware/ratelimit.go`

#### Шаг 5.2: Query Optimization
**Задачи:**
- Создать оптимизированные запросы к ClickHouse
- Реализовать Redis caching для частых запросов
- Агрегация данных на уровне БД

---

### ФАЗА 6: Monitoring & Observability (1 неделя)

#### Шаг 6.1: Мониторинг
**Компоненты:**
- Prometheus для метрик
- Grafana для дашбордов
- Jaeger для distributed tracing
- ELK/Loki для логов

**Метрики:**
- Количество обработанных тикеров
- Время выполнения pipeline
- Ошибки сбора данных
- Latency запросов
- Database размер

**Файлы:**
- `monitoring/prometheus/prometheus.yml`
- `monitoring/grafana/dashboards/`
- `k8s/monitoring/`

---

### ФАЗА 7: Deployment & Production (1 неделя)

#### Шаг 7.1: Kubernetes Deployment
**Компоненты:**
- StatefulSet для ClickHouse
- Deployment для сервисов
- HPA (Horizontal Pod Autoscaling)
- Secrets management
- Persistent Volumes

**Файлы:**
- `k8s/clickhouse/statefulset.yaml`
- `k8s/kafka/kafka-cluster.yaml`
- `k8s/redis/deployment.yaml`
- `k8s/services/ticker-fetcher.yaml`
- `k8s/services/data-collector.yaml`
- `k8s/services/indicator-calculator.yaml`
- `k8s/services/scheduler.yaml`
- `k8s/services/api-gateway.yaml`

#### Шаг 7.2: CI/CD Pipeline
**Этапы:**
1. Build Docker images
2. Run unit tests
3. Run integration tests
4. Push to registry
5. Deploy to staging
6. Deploy to production (manual approval)

---

## Технический стек

### Backend Services
- **Язык:** Go 1.21+
- **Frameworks:**
  - gRPC (google.golang.org/grpc)
  - Gin/Echo для REST API
  - robfig/cron для scheduler

### Storage
- **OLAP:** ClickHouse 23.x
- **Cache:** Redis 7.x
- **Message Queue:** Kafka 3.x

### Инфраструктура
- **Containerization:** Docker
- **Orchestration:** Kubernetes 1.28+
- **Service Mesh:** Istio (опционально)

### Monitoring
- **Metrics:** Prometheus + Grafana
- **Tracing:** Jaeger
- **Logging:** Loki или ELK

### CI/CD
- **Version Control:** Git
- **CI/CD:** GitHub Actions
- **Registry:** Docker Hub или private registry

---

## Структура проекта

```
stock-market-system/
├── api/
│   ├── proto/                    # gRPC proto definitions
│   └── openapi/                  # REST API specs
├── cmd/
│   ├── ticker-fetcher/          # Сервис получения списка тикеров
│   ├── data-collector/          # Сервис сбора котировок
│   ├── indicator-calculator/    # Сервис расчета индикаторов
│   ├── scheduler/               # Планировщик задач
│   └── api-gateway/             # API Gateway
├── pkg/
│   ├── storage/                 # Storage abstractions
│   │   ├── clickhouse/
│   │   └── redis/
│   ├── messaging/               # Kafka producers/consumers
│   ├── models/                  # Data models
│   ├── indicators/              # Indicator calculation library
│   └── utils/
├── schema/
│   └── clickhouse/              # Database schemas
├── k8s/                         # Kubernetes manifests
├── monitoring/                  # Monitoring configs
├── docker-compose.yml           # Local development
└── README.md
```

---

## Производительность и масштабируемость

### Оптимизации

1. **Параллельная обработка:**
   - Worker pool для сбора данных (100+ concurrent requests)
   - Batch processing для индикаторов

2. **ClickHouse оптимизации:**
   - Партиционирование по месяцам
   - MergeTree engine с ORDER BY (ticker, date)
   - Материализованные представления для агрегаций
   - Batch inserts (10000+ rows)

3. **Кэширование:**
   - Redis для списков тикеров (TTL: 1 день)
   - Redis для последних котировок (TTL: 5 минут)
   - Application-level caching для индикаторов

4. **Rate Limiting:**
   - Token bucket для API requests
   - Backoff strategy для failed requests

### Ожидаемая производительность

- **Сбор данных:** ~3000 тикеров за 10-15 минут
- **Расчет индикаторов:** ~3000 тикеров за 5-10 минут
- **Query latency:** <100ms для простых запросов, <500ms для сложных
- **Throughput:** 1000+ RPS на API Gateway

---

## Безопасность

1. **API Authentication:** JWT tokens
2. **Secrets Management:** Kubernetes Secrets или HashiCorp Vault
3. **Network Policies:** Изоляция сервисов в K8s
4. **Rate Limiting:** Защита от DDoS
5. **Data Encryption:** TLS для всех коммуникаций

---

## Миграция с текущего решения

### Этап 1: Параллельный запуск
- Продолжать использовать Python скрипты
- Запустить Go систему в параллель
- Сравнение результатов

### Этап 2: Валидация
- Проверка корректности расчетов
- Performance testing
- Bug fixing

### Этап 3: Переход
- Отключение Python скриптов
- Полный переход на Go систему
- Мониторинг стабильности

---

## Стоимость и ресурсы

### Development Time: 8-12 недель (1 разработчик)

### Infrastructure (примерная стоимость/месяц):
- **ClickHouse:** 3 nodes, 16GB RAM each → ~$300-500
- **Kafka:** 3 brokers → ~$200-300
- **Redis:** 1 instance, 8GB RAM → ~$50-100
- **Application Services:** 5-10 pods → ~$200-400
- **Monitoring:** Prometheus, Grafana → ~$100
- **Total:** ~$850-1400/месяц (можно снизить с помощью spot instances)

---

## Next Steps

1. ✅ Утвердить архитектуру и roadmap
2. 📝 Создать Git репозиторий и структуру проекта
3. 🗄️ Разработать схему ClickHouse
4. 🔨 Начать имплементацию с Ticker Fetcher Service
5. 📊 Настроить локальное окружение (Docker Compose)

---

## Вопросы для обсуждения

1. Приоритеты: начать с ММВБ или NASDAQ?
2. Нужна ли поддержка real-time данных или только end-of-day?
3. Требования к SLA (uptime, latency)?
4. Budget constraints для инфраструктуры?
5. Нужен ли web UI для визуализации или только API?
