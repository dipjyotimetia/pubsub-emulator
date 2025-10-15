.PHONY: build test lint run docker clean help

# Variables
BINARY_NAME=pubsub-emulator
DOCKER_IMAGE=dipjyotimetia/pubsub-emulator
VERSION?=latest
GO_VERSION=1.24.0

# Colors for output
GREEN=\033[0;32m
YELLOW=\033[0;33m
NC=\033[0m # No Color

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  ${GREEN}%-15s${NC} %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary
	@echo "${GREEN}Building ${BINARY_NAME}...${NC}"
	@go build -o ${BINARY_NAME} .
	@echo "${GREEN}Build complete!${NC}"

build-cmd: ## Build the binary from cmd/server (for future refactored structure)
	@echo "${GREEN}Building ${BINARY_NAME} from cmd/server...${NC}"
	@go build -o ${BINARY_NAME} ./cmd/server
	@echo "${GREEN}Build complete!${NC}"

test: ## Run tests
	@echo "${GREEN}Running tests...${NC}"
	@go test -v -race -coverprofile=coverage.out ./...
	@echo "${GREEN}Tests complete!${NC}"

test-coverage: test ## Run tests with coverage report
	@go tool cover -html=coverage.out -o coverage.html
	@echo "${GREEN}Coverage report: coverage.html${NC}"

lint: ## Run linter
	@echo "${GREEN}Running linter...${NC}"
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "${YELLOW}golangci-lint not installed. Install: https://golangci-lint.run/usage/install/${NC}"; \
	fi

fmt: ## Format code
	@echo "${GREEN}Formatting code...${NC}"
	@go fmt ./...
	@echo "${GREEN}Formatting complete!${NC}"

vet: ## Run go vet
	@echo "${GREEN}Running go vet...${NC}"
	@go vet ./...
	@echo "${GREEN}Vet complete!${NC}"

run: ## Run the application locally
	@echo "${GREEN}Running ${BINARY_NAME}...${NC}"
	@PUBSUB_PROJECT=test-project \
	 PUBSUB_TOPIC=test-topic \
	 PUBSUB_SUBSCRIPTION=test-sub \
	 DASHBOARD_PORT=8080 \
	 go run .

docker-build: ## Build Docker image
	@echo "${GREEN}Building Docker image ${DOCKER_IMAGE}:${VERSION}...${NC}"
	@docker build -t ${DOCKER_IMAGE}:${VERSION} .
	@echo "${GREEN}Docker build complete!${NC}"

docker-run: ## Run Docker container
	@echo "${GREEN}Running Docker container...${NC}"
	@docker run --rm -it \
		-p 8085:8681 \
		-p 8080:8080 \
		-e PUBSUB_PROJECT=test-project \
		-e PUBSUB_TOPIC=test-topic1,test-topic2 \
		-e PUBSUB_SUBSCRIPTION=test-sub1,test-sub2 \
		-e DASHBOARD_PORT=8080 \
		${DOCKER_IMAGE}:${VERSION}

docker-push: docker-build ## Push Docker image to registry
	@echo "${GREEN}Pushing Docker image ${DOCKER_IMAGE}:${VERSION}...${NC}"
	@docker push ${DOCKER_IMAGE}:${VERSION}
	@echo "${GREEN}Docker push complete!${NC}"

clean: ## Clean build artifacts
	@echo "${GREEN}Cleaning build artifacts...${NC}"
	@rm -f ${BINARY_NAME}
	@rm -f coverage.out coverage.html
	@go clean
	@echo "${GREEN}Clean complete!${NC}"

deps: ## Download dependencies
	@echo "${GREEN}Downloading dependencies...${NC}"
	@go mod download
	@go mod tidy
	@echo "${GREEN}Dependencies downloaded!${NC}"

verify: lint vet test ## Run all verification checks
	@echo "${GREEN}All verification checks passed!${NC}"

install: build ## Install binary to $GOPATH/bin
	@echo "${GREEN}Installing ${BINARY_NAME} to $$GOPATH/bin...${NC}"
	@go install
	@echo "${GREEN}Installation complete!${NC}"

# Development helpers
dev: ## Run with hot reload (requires air)
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "${YELLOW}Air not installed. Install: go install github.com/cosmtrek/air@latest${NC}"; \
		echo "${YELLOW}Running without hot reload...${NC}"; \
		make run; \
	fi

check-deps: ## Check for outdated dependencies
	@echo "${GREEN}Checking for outdated dependencies...${NC}"
	@go list -u -m all

update-deps: ## Update dependencies
	@echo "${GREEN}Updating dependencies...${NC}"
	@go get -u ./...
	@go mod tidy
	@echo "${GREEN}Dependencies updated!${NC}"

# Docker Compose helpers
compose-up: ## Start services with docker-compose
	@echo "${GREEN}Starting services with docker-compose...${NC}"
	@docker-compose up -d
	@echo "${GREEN}Services started!${NC}"

compose-down: ## Stop services
	@echo "${GREEN}Stopping services...${NC}"
	@docker-compose down
	@echo "${GREEN}Services stopped!${NC}"

compose-logs: ## View logs
	@docker-compose logs -f

# Show current version info
version: ## Show version information
	@echo "${GREEN}Go Version:${NC} $(shell go version)"
	@echo "${GREEN}Binary:${NC} ${BINARY_NAME}"
	@echo "${GREEN}Docker Image:${NC} ${DOCKER_IMAGE}:${VERSION}"

.DEFAULT_GOAL := help
