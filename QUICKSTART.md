# Quick Start Guide - Stock Market System

Это краткое руководство для быстрого запуска системы.

## Предварительные требования

- Go 1.21+
- Docker & Docker Compose
- Make

## Шаг 1: Запуск инфраструктуры

```bash
# Запустить ClickHouse, Redis, Kafka и мониторинг
make infra-up

# Проверить статус
docker-compose ps
```

Дождитесь, пока все сервисы будут в статусе "healthy" (1-2 минуты).

## Шаг 2: Инициализация базы данных

```bash
# Создать таблицы в ClickHouse
make db-init
```

## Шаг 3: Создание Kafka топиков

```bash
# Создать необходимые топики
make kafka-create-topics
```

## Шаг 4: Установка зависимостей Go

```bash
# Скачать зависимости
make deps
```

## Шаг 5: Запуск Ticker Fetcher Service

```bash
# В отдельном терминале
make run-ticker-fetcher
```

Сервис будет доступен на http://localhost:8081

## Шаг 6: Тестирование API

```bash
# Получить список NASDAQ акций
curl "http://localhost:8081/api/v1/tickers/nasdaq?type=stocks"

# Получить список MOEX акций
curl "http://localhost:8081/api/v1/tickers/moex?boardid=TQBR"

# Обновить кэш
curl -X POST "http://localhost:8081/api/v1/tickers/refresh"
```

## Шаг 7: Доступ к UI сервисам

- **Kafka UI**: http://localhost:8080
- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090
- **Jaeger**: http://localhost:16686
- **ClickHouse**: http://localhost:8123

## Полезные команды

```bash
# Остановить все сервисы
make docker-down

# Просмотр логов
make docker-logs

# Подключиться к ClickHouse
make db-console

# Подключиться к Redis
make redis-cli

# Список Kafka топиков
make kafka-topics

# Запустить тесты
make test

# Сборка всех сервисов
make build
```

## Следующие шаги

1. Реализовать Data Collector для MOEX
2. Реализовать Data Collector для NASDAQ
3. Реализовать Indicator Calculator
4. Настроить Scheduler для автоматического сбора

## Устранение проблем

### ClickHouse не запускается

```bash
# Проверить логи
docker logs stock-clickhouse

# Перезапустить
docker-compose restart clickhouse
```

### Redis недоступен

```bash
# Проверить статус
docker-compose ps redis

# Перезапустить
docker-compose restart redis
```

### Kafka не принимает сообщения

```bash
# Проверить Zookeeper
docker-compose ps zookeeper

# Проверить Kafka
docker logs stock-kafka
```

## Документация

- [Архитектура и Roadmap](ARCHITECTURE_ROADMAP.md)
- [Полная документация](README_STOCK_MARKET.md)
- [Makefile команды](Makefile)
