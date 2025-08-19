.PHONY: help build build-cli run run-server run-cli test clean docker-build docker-run docker-compose-up docker-compose-down install

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

build: ## Build the CLI application
	$(GO) build $(GOFLAGS) -o $(APP_NAME) .

build-cli: build ## Alias for build

run: run-server ## Run the API server (default)

run-server: ## Run the API server
	$(GO) run main.go server

run-cli: ## Show CLI help
	$(GO) run main.go --help

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

install: build ## Install the binary to GOPATH/bin
	cp $(APP_NAME) $(GOPATH)/bin/$(APP_NAME) || cp $(APP_NAME) ~/go/bin/$(APP_NAME)

convert-example: build ## Run example conversion (requires test.pcap)
	@if [ -f "examples/test.pcap" ]; then \
		./$(APP_NAME) convert -i examples/test.pcap -o ./output --verbose; \
	else \
		echo "Error: examples/test.pcap not found. Create it first with:" ; \
		echo "  sudo tcpdump -i any -w examples/test.pcap -c 100 port 80" ; \
	fi

.DEFAULT_GOAL := help