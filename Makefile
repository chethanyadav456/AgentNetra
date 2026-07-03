# =============================================================================
# AgentNetra Makefile
# =============================================================================

BINARY     := agentnetra
MODULE     := github.com/chethanyadav456/agentnetra
CMD_PATH   := ./cmd/agentnetra
BUILD_DIR  := ./build
VERSION    := 0.1.0
COMMIT     := $(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -ldflags "-s -w \
	-X $(MODULE)/pkg/constants.AppVersion=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.buildTime=$(BUILD_TIME)"

GO        := go
GOTEST    := $(GO) test
GOBUILD   := $(GO) build
GOVET     := $(GO) vet
GOFMT     := gofmt

.PHONY: all build clean test test-coverage lint fmt vet tidy run help install

## all: Build the binary (default target)
all: build

## build: Compile the agentnetra binary
build:
	@echo "→ Building $(BINARY) $(VERSION)"
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) $(CMD_PATH)
	@echo "✓ Binary written to $(BUILD_DIR)/$(BINARY)"

## install: Install agentnetra to GOPATH/bin
install:
	@echo "→ Installing $(BINARY)"
	$(GO) install $(LDFLAGS) $(CMD_PATH)
	@echo "✓ Installed to $(shell go env GOPATH)/bin/$(BINARY)"

## run: Build and run a scan
run: build
	$(BUILD_DIR)/$(BINARY) scan

## test: Run all unit and integration tests
test:
	@echo "→ Running tests"
	$(GOTEST) -race -count=1 ./...

## test-coverage: Run tests and produce HTML coverage report
test-coverage:
	@echo "→ Running tests with coverage"
	$(GOTEST) -race -count=1 -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report: coverage.html"

## lint: Run go vet
lint: vet

## vet: Run go vet on all packages
vet:
	@echo "→ Running go vet"
	$(GOVET) ./...

## fmt: Format all Go source files
fmt:
	@echo "→ Running gofmt"
	$(GOFMT) -w .

## tidy: Tidy Go module dependencies
tidy:
	@echo "→ Running go mod tidy"
	$(GO) mod tidy

## clean: Remove build artifacts
clean:
	@echo "→ Cleaning build artifacts"
	@rm -rf $(BUILD_DIR) coverage.out coverage.html
	@echo "✓ Clean complete"

## help: Show this help message
help:
	@echo "AgentNetra v$(VERSION) — Build Targets:"
	@echo ""
	@grep -E '^##' $(MAKEFILE_LIST) | sed 's/## /  /'
