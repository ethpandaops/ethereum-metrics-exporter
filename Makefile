# ethereum-metrics-exporter Makefile

PROJECT_NAME := ethereum-metrics-exporter
GO := go
GOTEST := $(GO) test
GOVET := $(GO) vet
GOLINT := golangci-lint
GORELEASER := goreleaser

# Build variables
VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT ?= $(shell git rev-parse --short HEAD)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-s -w -X github.com/ethpandaops/ethereum-metrics-exporter/pkg/exporter/version.Release=$(VERSION) -X github.com/ethpandaops/ethereum-metrics-exporter/pkg/exporter/version.GitCommit=$(COMMIT)"

# Output directory
BUILD_DIR := ./build
BIN_NAME := $(PROJECT_NAME)

.PHONY: help
help: ## Display this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

.PHONY: all
all: clean lint test build ## Run clean, lint, test, and build

.PHONY: build
build: ## Build the binary for the current platform
	@echo "Building $(BIN_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BIN_NAME) .

.PHONY: build-linux
build-linux: ## Build the binary for Linux amd64
	@echo "Building $(BIN_NAME) for Linux amd64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BIN_NAME)-linux-amd64 .

.PHONY: build-all
build-all: ## Build binaries for all platforms using goreleaser
	@echo "Building for all platforms..."
	$(GORELEASER) build --clean --snapshot

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR) dist/

.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.txt -covermode=atomic ./...

.PHONY: test-short
test-short: ## Run short tests
	@echo "Running short tests..."
	$(GOTEST) -short ./...

.PHONY: coverage
coverage: test ## Generate test coverage report
	@echo "Generating coverage report..."
	$(GO) tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	$(GOVET) ./...

.PHONY: lint
lint: ## Run golangci-lint
	@echo "Running golangci-lint..."
	$(GOLINT) run --new-from-rev="origin/main"

.PHONY: lint-fix
lint-fix: ## Run golangci-lint with auto-fix
	@echo "Running golangci-lint with fixes..."
	$(GOLINT) run --fix --new-from-rev="origin/main"

.PHONY: fmt
fmt: ## Format code using go fmt
	@echo "Formatting code..."
	$(GO) fmt ./...

.PHONY: tidy
tidy: ## Run go mod tidy
	@echo "Running go mod tidy..."
	$(GO) mod tidy

.PHONY: deps
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GO) mod download

.PHONY: verify
verify: ## Verify dependencies
	@echo "Verifying dependencies..."
	$(GO) mod verify

.PHONY: run
run: ## Run the application
	@echo "Running $(BIN_NAME)..."
	$(GO) run .

.PHONY: install
install: ## Install the binary to GOPATH/bin
	@echo "Installing $(BIN_NAME)..."
	CGO_ENABLED=0 $(GO) install $(LDFLAGS) .

.PHONY: release
release: ## Create a new release using goreleaser
	@echo "Creating release..."
	@docker context use default
	RELEASE_SUFFIX="" $(GORELEASER) release --clean

.PHONY: release-snapshot
release-snapshot: ## Create a snapshot release
	@echo "Creating snapshot release..."
	@docker context use default
	RELEASE_SUFFIX="" $(GORELEASER) release --snapshot --clean

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(PROJECT_NAME):$(VERSION) .

.PHONY: docker-push
docker-push: ## Push Docker image
	@echo "Pushing Docker image..."
	docker push $(PROJECT_NAME):$(VERSION)

.PHONY: check-tools
check-tools: ## Check if required tools are installed
	@echo "Checking required tools..."
	@which $(GO) > /dev/null || (echo "go not found. Please install Go." && exit 1)
	@which $(GOLINT) > /dev/null || (echo "golangci-lint not found. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	@which $(GORELEASER) > /dev/null || (echo "goreleaser not found. Visit https://goreleaser.com/install/" && exit 1)
	@echo "All required tools are installed."


.PHONY: devnet
devnet:
	.hack/devnet/run.sh

.PHONY: devnet-run
devnet-run: devnet
	go run main.go --config .hack/devnet/generated-ethereum-metrics-exporter-config.yaml

.PHONY: devnet-clean
devnet-clean:
	.hack/devnet/cleanup.sh
