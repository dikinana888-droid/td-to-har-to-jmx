.PHONY: help build run test clean docker-build docker-run docker-compose-up docker-compose-down

# Variables
APP_NAME=tcpdump-to-jmx
DOCKER_IMAGE=$(APP_NAME):latest
GO=go
GOFLAGS=-v

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	$(GO) build $(GOFLAGS) -o $(APP_NAME) .

run: ## Run the application locally
	$(GO) run main.go

test: ## Run tests
	$(GO) test $(GOFLAGS) ./...

clean: ## Clean build artifacts
	rm -f $(APP_NAME)
	$(GO) clean

deps: ## Download dependencies
	$(GO) mod download
	$(GO) mod tidy

docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE) .

docker-run: docker-build ## Run application in Docker
	docker run -p 8080:8080 --env-file .env $(DOCKER_IMAGE)

docker-compose-up: ## Start services with docker-compose
	docker-compose up -d

docker-compose-down: ## Stop services with docker-compose
	docker-compose down

docker-compose-logs: ## View docker-compose logs
	docker-compose logs -f

install-tools: ## Install development tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest

lint: ## Run linter
	golangci-lint run ./...

fmt: ## Format code
	gofmt -s -w .
	goimports -w .

.DEFAULT_GOAL := help