# Define the version of your provider
VERSION := $(shell \
	if [ -f speakeasy.yaml ]; then \
		grep 'version: ' speakeasy.yaml | awk '{print $$2}' | tr -d '"' | head -n 1 || echo "0.0.1"; \
	elif [ -f .speakeasy/gen.yaml ]; then \
		grep -A 10 'terraform:' .speakeasy/gen.yaml | grep '^[[:space:]]*version: ' | awk '{print $$2}' | tr -d '"' | head -n 1 || echo "0.0.1"; \
	elif [ -f VERSION ]; then \
		cat VERSION | tr -d '\n'; \
	else \
		echo "0.0.1"; \
	fi)

# Define the binary name
BINARY_NAME = terraform-provider-netskope

# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get

# Build the provider for all supported platforms
all: build-darwin build-linux build-windows

build-darwin:
	@echo "Building for Darwin (macOS) with version $(VERSION)..."
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o bin/mac/$(BINARY_NAME)_v$(VERSION) ./main.go

build-linux:
	@echo "Building for Linux with version $(VERSION)..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o bin/linux/$(BINARY_NAME)_v$(VERSION) ./main.go

build-windows:
	@echo "Building for Windows with version $(VERSION)..."
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o bin/windows/$(BINARY_NAME)_v$(VERSION).exe ./main.go

clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf bin/mac/$(BINARY_NAME)_v*
	rm -rf bin/windows/$(BINARY_NAME)_v*
	rm -rf bin/linux/$(BINARY_NAME)_v*

test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

deps:
	@echo "Fetching dependencies..."
	$(GOGET) -v ./...

# Run all acceptance tests
.PHONY: testacc
testacc:
	@echo "Running acceptance tests..."
	TF_ACC=1 $(GOTEST) -v ./internal/provider/... -timeout 120m

# Run acceptance tests with coverage
.PHONY: testacc-coverage
testacc-coverage:
	@echo "Running acceptance tests with coverage..."
	TF_ACC=1 $(GOTEST) -v -coverprofile=coverage-acc.out ./internal/provider/... -timeout 120m
	$(GOCMD) tool cover -html=coverage-acc.out -o coverage-acc.html

# Run specific resource tests
.PHONY: testacc-privateapp
testacc-privateapp:
	@echo "Running private app acceptance tests..."
	TF_ACC=1 $(GOTEST) -v ./internal/provider/... -run TestAccNPAPrivateApp -timeout 30m

.PHONY: testacc-publisher
testacc-publisher:
	@echo "Running publisher acceptance tests..."
	TF_ACC=1 $(GOTEST) -v ./internal/provider/... -run TestAccNPAPublisher -timeout 30m

.PHONY: testacc-policygroups
testacc-policygroups:
	@echo "Running policy groups acceptance tests..."
	TF_ACC=1 $(GOTEST) -v ./internal/provider/... -run TestAccNPAPolicyGroups -timeout 30m

.PHONY: testacc-rules
testacc-rules:
	@echo "Running rules acceptance tests..."
	TF_ACC=1 $(GOTEST) -v ./internal/provider/... -run TestAccNPARules -timeout 30m

# Run data source tests only
.PHONY: testacc-datasources
testacc-datasources:
	@echo "Running data source acceptance tests..."
	TF_ACC=1 $(GOTEST) -v ./internal/provider/... -run TestAcc.*DataSource -timeout 30m

# Run with debug logging
.PHONY: testacc-debug
testacc-debug:
	@echo "Running acceptance tests with debug logging..."
	TF_ACC=1 TF_LOG=DEBUG $(GOTEST) -v ./internal/provider/... -timeout 120m 2>&1 | tee test-debug.log

.PHONY: all build-darwin build-linux build-windows clean test deps testacc testacc-coverage testacc-privateapp testacc-publisher testacc-policygroups testacc-rules testacc-datasources testacc-debug
