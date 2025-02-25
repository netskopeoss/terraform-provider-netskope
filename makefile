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

.PHONY: all build-darwin build-linux build-windows clean test deps
