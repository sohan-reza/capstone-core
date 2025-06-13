# Makefile for Capstone-Core System

.PHONY: run build swagger test clean docker-build docker-run help

# Default variables
APP_NAME ?= capstone-core
APP_PORT ?= 8080
DOCKER_IMAGE ?= capstone-core:latest

## help: Display this help message
help:
	@echo "Usage:"
	@echo "  make <target> [OPTIONS]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## run: Build and run the application locally
run: 
	@echo "Starting application on port $(APP_PORT)..."
	@go run cmd/api/main.go

## build: Build the application binary
build: 
	@echo "Building application binary..."
	@go build -o bin/$(APP_NAME) cmd/api/main.go


## test: Run all tests
test:
	@echo "Running tests..."
	@go test -v ./...

## clean: Remove build artifacts
clean:
	@echo "Cleaning up..."
	@rm -rf bin/
	@rm -f $(APP_NAME)

## docker-build: Build the Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE) .

## docker-run: Run the application in Docker
docker-run: docker-build
	@echo "Running Docker container..."
	@docker run -p $(APP_PORT):$(APP_PORT) --env-file .env $(DOCKER_IMAGE)

## migrate: Run database migrations (example - customize for your migration tool)
migrate:
	@echo "Running database migrations..."
	@go run cmd/migrate/main.go

## lint: Run linters
lint:
	@echo "Running linters..."
	@golangci-lint run