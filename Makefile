.PHONY: all build clean test lint

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOLINT=golangci-lint

# Binary names
BINARY_NAME=nobl9-bot
BINARY_UNIX=$(BINARY_NAME)_unix

# Build directories
BUILD_DIR=build
BIN_DIR=bin

all: clean build

build:
	@echo "Building..."
	@mkdir -p $(BUILD_DIR)
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/bot

clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf $(BIN_DIR)

test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

lint:
	@echo "Running linter..."
	$(GOLINT) run

deps:
	@echo "Downloading dependencies..."
	$(GOGET) -v ./...

# Cross compilation
build-linux:
	@echo "Building for Linux..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_UNIX) ./cmd/bot

build-darwin:
	@echo "Building for macOS..."
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin ./cmd/bot

build-windows:
	@echo "Building for Windows..."
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME).exe ./cmd/bot

# Docker
docker-build:
	@echo "Building Docker image..."
	docker build -t nobl9-bot:latest .

docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 nobl9-bot:latest 