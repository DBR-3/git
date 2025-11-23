.PHONY: help build clean test lint fmt docker-build docker-up docker-down proto-gen install-tools

# Variables
APP_NAME=stock-market-system
VERSION?=latest
DOCKER_REGISTRY?=localhost:5000

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

# Build directories
BUILD_DIR=build
CMD_DIR=cmd

# Services
SERVICES=ticker-fetcher data-collector indicator-calculator scheduler api-gateway

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Tools installed successfully"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "Dependencies downloaded"

build: deps ## Build all services
	@echo "Building services..."
	@mkdir -p $(BUILD_DIR)
	@for service in $(SERVICES); do \
		echo "Building $$service..."; \
		$(GOBUILD) -o $(BUILD_DIR)/$$service $(CMD_DIR)/$$service/main.go; \
	done
	@echo "Build complete"

build-linux: deps ## Build all services for Linux
	@echo "Building services for Linux..."
	@mkdir -p $(BUILD_DIR)
	@for service in $(SERVICES); do \
		echo "Building $$service for Linux..."; \
		GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$$service-linux $(CMD_DIR)/$$service/main.go; \
	done
	@echo "Linux build complete"

clean: ## Remove build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	@echo "Clean complete"

test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	@echo "Tests complete"

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	$(GOTEST) -v -tags=integration ./...
	@echo "Integration tests complete"

test-coverage: test ## Run tests and show coverage
	@echo "Generating coverage report..."
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

bench: ## Run benchmarks
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./pkg/indicators/
	@echo "Benchmarks complete"

lint: ## Run linter
	@echo "Running linter..."
	$(GOLINT) run --timeout=5m
	@echo "Linting complete"

fmt: ## Format code
	@echo "Formatting code..."
	$(GOFMT) -s -w .
	@echo "Formatting complete"

vet: ## Run go vet
	@echo "Running go vet..."
	$(GOCMD) vet ./...
	@echo "Vet complete"

proto-gen: ## Generate protobuf code
	@echo "Generating protobuf code..."
	@mkdir -p api/generated
	protoc --go_out=api/generated --go_opt=paths=source_relative \
		--go-grpc_out=api/generated --go-grpc_opt=paths=source_relative \
		api/proto/*.proto
	@echo "Protobuf generation complete"

docker-build: ## Build Docker images for all services
	@echo "Building Docker images..."
	@for service in $(SERVICES); do \
		echo "Building Docker image for $$service..."; \
		docker build -t $(DOCKER_REGISTRY)/$(APP_NAME)-$$service:$(VERSION) \
			-f $(CMD_DIR)/$$service/Dockerfile .; \
	done
	@echo "Docker images built"

docker-push: ## Push Docker images to registry
	@echo "Pushing Docker images..."
	@for service in $(SERVICES); do \
		echo "Pushing $$service..."; \
		docker push $(DOCKER_REGISTRY)/$(APP_NAME)-$$service:$(VERSION); \
	done
	@echo "Docker images pushed"

docker-up: ## Start all services with docker-compose
	@echo "Starting services..."
	docker-compose up -d
	@echo "Services started"
	@echo "ClickHouse UI: http://localhost:8123"
	@echo "Kafka UI: http://localhost:8080"
	@echo "Grafana: http://localhost:3000"
	@echo "Prometheus: http://localhost:9090"
	@echo "Jaeger: http://localhost:16686"

docker-down: ## Stop all services
	@echo "Stopping services..."
	docker-compose down
	@echo "Services stopped"

docker-logs: ## Show docker-compose logs
	docker-compose logs -f

docker-ps: ## Show running containers
	docker-compose ps

infra-up: ## Start only infrastructure (no app services)
	@echo "Starting infrastructure..."
	docker-compose up -d clickhouse redis zookeeper kafka kafka-ui prometheus grafana jaeger
	@echo "Infrastructure started"

db-init: ## Initialize ClickHouse database
	@echo "Initializing database..."
	docker exec -it stock-clickhouse clickhouse-client --query "CREATE DATABASE IF NOT EXISTS stock_market"
	docker exec -it stock-clickhouse clickhouse-client --database stock_market < schema/clickhouse/01_init.sql
	@echo "Database initialized"

db-console: ## Open ClickHouse console
	docker exec -it stock-clickhouse clickhouse-client --database stock_market

db-reset: ## Drop and recreate database
	@echo "Resetting database..."
	docker exec -it stock-clickhouse clickhouse-client --query "DROP DATABASE IF EXISTS stock_market"
	$(MAKE) db-init
	@echo "Database reset complete"

redis-cli: ## Open Redis CLI
	docker exec -it stock-redis redis-cli -a redis_password

kafka-topics: ## List Kafka topics
	docker exec -it stock-kafka kafka-topics --bootstrap-server localhost:9092 --list

kafka-create-topics: ## Create required Kafka topics
	@echo "Creating Kafka topics..."
	docker exec -it stock-kafka kafka-topics --create --if-not-exists \
		--bootstrap-server localhost:9092 \
		--topic stock.data.collected \
		--partitions 10 \
		--replication-factor 1
	docker exec -it stock-kafka kafka-topics --create --if-not-exists \
		--bootstrap-server localhost:9092 \
		--topic stock.indicators.calculated \
		--partitions 10 \
		--replication-factor 1
	docker exec -it stock-kafka kafka-topics --create --if-not-exists \
		--bootstrap-server localhost:9092 \
		--topic stock.signals.generated \
		--partitions 10 \
		--replication-factor 1
	docker exec -it stock-kafka kafka-topics --create --if-not-exists \
		--bootstrap-server localhost:9092 \
		--topic stock.errors \
		--partitions 5 \
		--replication-factor 1
	@echo "Kafka topics created"

run-ticker-fetcher: ## Run ticker-fetcher service locally
	$(GOCMD) run $(CMD_DIR)/ticker-fetcher/main.go

run-data-collector: ## Run data-collector service locally
	$(GOCMD) run $(CMD_DIR)/data-collector/main.go

run-indicator-calculator: ## Run indicator-calculator service locally
	$(GOCMD) run $(CMD_DIR)/indicator-calculator/main.go

run-scheduler: ## Run scheduler service locally
	$(GOCMD) run $(CMD_DIR)/scheduler/main.go

run-api-gateway: ## Run api-gateway service locally
	$(GOCMD) run $(CMD_DIR)/api-gateway/main.go

test-ticker-fetcher: ## Test ticker-fetcher service
	@./scripts/test-ticker-fetcher.sh

test-data-collector: ## Test data-collector service
	@./scripts/test-data-collector.sh

test-indicator-calculator: ## Test indicator-calculator service
	@./scripts/test-indicator-calculator.sh

k8s-deploy: ## Deploy to Kubernetes
	@echo "Deploying to Kubernetes..."
	kubectl apply -f k8s/namespace.yaml
	kubectl apply -f k8s/clickhouse/
	kubectl apply -f k8s/kafka/
	kubectl apply -f k8s/redis/
	kubectl apply -f k8s/services/
	@echo "Deployment complete"

k8s-delete: ## Delete Kubernetes deployment
	@echo "Deleting Kubernetes resources..."
	kubectl delete -f k8s/services/
	kubectl delete -f k8s/redis/
	kubectl delete -f k8s/kafka/
	kubectl delete -f k8s/clickhouse/
	kubectl delete -f k8s/namespace.yaml
	@echo "Deletion complete"

k8s-status: ## Check Kubernetes deployment status
	kubectl get pods -n stock-market
	kubectl get services -n stock-market

k8s-logs: ## Show logs from Kubernetes pods
	kubectl logs -n stock-market -l app=stock-market --tail=100 -f

dev: infra-up ## Start infrastructure and run services locally
	@echo "Development environment ready"
	@echo "Start services manually with: make run-<service-name>"

all: clean deps lint test build ## Run all checks and build

ci: deps lint test build ## CI pipeline

.DEFAULT_GOAL := help
