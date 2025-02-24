# Define the version of your provider
VERSION = 0.3.0

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
	@echo "Building for Darwin (macOS)..."
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o bin/mac/$(BINARY_NAME)_v$(VERSION) ./main.go

build-linux:
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o bin/linux/$(BINARY_NAME)_v$(VERSION) ./main.go

build-windows:
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o bin/windows/$(BINARY_NAME)_v$(VERSION).exe ./main.go

clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f bin/mac/$(BINARY_NAME)_v$(VERSION)*
	rm -f bin/windows/$(BINARY_NAME)_v$(VERSION)*
	rm -f bin/linux/$(BINARY_NAME)_v$(VERSION)*

test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

deps:
	@echo "Fetching dependencies..."
	$(GOGET) -v ./...

.PHONY: all build-darwin build-linux build-windows clean test deps
