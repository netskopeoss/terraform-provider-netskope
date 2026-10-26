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

install:
	@echo "Building and installing provider for local development..."
	$(GOBUILD) -o $(GOPATH)/bin/$(BINARY_NAME) ./main.go
	@echo "Installed to $(GOPATH)/bin/$(BINARY_NAME)"
	@echo "dev_overrides in ~/.terraformrc points terraform to this binary."

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

# Fail if the provider version is not identical in every place it is recorded.
# Speakeasy bumps .speakeasy/gen.yaml `terraform.version` on every `speakeasy run`
# and stamps it into docs/index.md and examples/provider/provider.tf; the release
# tag comes from CHANGELOG. Keeping all four equal is what stops a regen from
# publishing docs that point at the wrong version (CLAUDE.md pitfall 6).
.PHONY: check-version
check-version:
	@gen=$$(grep -A 10 '^terraform:' .speakeasy/gen.yaml | grep '^[[:space:]]*version: ' | awk '{print $$2}' | head -n 1); \
	lock=$$(grep '^[[:space:]]*releaseVersion: ' .speakeasy/gen.lock | awk '{print $$2}'); \
	docs=$$(grep -o 'version = "[0-9.]*"' docs/index.md | head -n 1 | grep -o '[0-9.]*'); \
	ex=$$(grep -o 'version = "[0-9.]*"' examples/provider/provider.tf | head -n 1 | grep -o '[0-9.]*'); \
	cl=$$(grep -m 1 -o '^## \[[0-9.]*\]' CHANGELOG.md | grep -o '[0-9.]*'); \
	echo "gen.yaml=$$gen gen.lock=$$lock docs/index.md=$$docs examples/provider=$$ex CHANGELOG=$$cl"; \
	if [ "$$gen" = "$$lock" ] && [ "$$gen" = "$$docs" ] && [ "$$gen" = "$$ex" ] && [ "$$gen" = "$$cl" ]; then \
		echo "OK: all version references are $$gen"; \
	else \
		echo "ERROR: version references disagree (set .speakeasy/gen.yaml terraform.version and gen.lock releaseVersion to the release before 'speakeasy run')"; exit 1; \
	fi

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

.PHONY: testacc-rbaclabels
testacc-rbaclabels:
	@echo "Running RBAC labels acceptance tests..."
	TF_ACC=1 $(GOTEST) -v ./internal/provider/... -run TestAccRBACLabel -timeout 30m

.PHONY: testacc-aig
testacc-aig:
	@echo "Running AI Gateway appliance acceptance tests..."
	TF_ACC=1 $(GOTEST) -v ./internal/provider/... -run TestAccAIGAppliance -timeout 30m

.PHONY: testacc-customcategory
testacc-customcategory:
	@echo "Running custom category acceptance tests..."
	TF_ACC=1 $(GOTEST) -v ./internal/provider/... -run TestAccCustomCategory -timeout 30m

.PHONY: testacc-serviceobject
testacc-serviceobject:
	@echo "Running service object acceptance tests..."
	TF_ACC=1 $(GOTEST) -v ./internal/provider/... -run TestAccServiceObject -timeout 30m

.PHONY: testacc-rbacrole
testacc-rbacrole:
	@echo "Running RBAC role acceptance tests..."
	TF_ACC=1 $(GOTEST) -v ./internal/provider/... -run TestAccRBACRole -timeout 30m

.PHONY: testacc-oauth2
testacc-oauth2:
	@echo "Running OAuth2 token data source acceptance tests..."
	@test -n "$(NETSKOPE_OAUTH2_CLIENT_ID)" || (echo "Error: NETSKOPE_OAUTH2_CLIENT_ID is not set" && exit 1)
	@test -n "$(NETSKOPE_OAUTH2_CLIENT_SECRET)" || (echo "Error: NETSKOPE_OAUTH2_CLIENT_SECRET is not set" && exit 1)
	TF_ACC=1 $(GOTEST) -v ./internal/provider/... -run TestAccPlatformOAuth2Token -timeout 10m

.PHONY: testacc-oauth2-provider
testacc-oauth2-provider:
	@echo "Running provider-level OAuth2 auth acceptance tests (no API key)..."
	@test -n "$(NETSKOPE_OAUTH2_CLIENT_ID)" || (echo "Error: NETSKOPE_OAUTH2_CLIENT_ID is not set" && exit 1)
	@test -n "$(NETSKOPE_OAUTH2_CLIENT_SECRET)" || (echo "Error: NETSKOPE_OAUTH2_CLIENT_SECRET is not set" && exit 1)
	NETSKOPE_API_KEY="" TF_ACC=1 $(GOTEST) -v ./internal/provider/... -run TestAccOAuth2ProviderAuth -timeout 10m

.PHONY: testacc-deviceclassification
testacc-deviceclassification:
	@echo "Running device classification acceptance tests..."
	TF_ACC=1 $(GOTEST) -v ./internal/provider/... -run TestAccDeviceClassification -timeout 30m

.PHONY: all install build-darwin build-linux build-windows clean test deps testacc testacc-coverage testacc-privateapp testacc-publisher testacc-policygroups testacc-rules testacc-datasources testacc-debug testacc-rbaclabels testacc-aig testacc-customcategory testacc-serviceobject testacc-rbacrole testacc-oauth2 testacc-oauth2-provider testacc-deviceclassification

# -----------------------------------------------------------------------------
# Pre-PR checks (see .claude/skills/testacc/SKILL.md)
# -----------------------------------------------------------------------------

# Layer 1: build, vet, unit + hook tests, generated-file diff, secret scan.
# No tenant required.
.PHONY: preflight
preflight:
	@echo "==> go build"
	$(GOBUILD) ./...
	@echo "==> go vet"
	$(GOCMD) vet ./...
	@echo "==> unit + hook tests (TF_ACC unset, TestAcc* skip)"
	$(GOTEST) ./internal/sdk/internal/hooks/ ./internal/provider/ -run 'Test[^A]' -timeout 10m
	@echo "==> files changed outside hooks/tests/docs since main (must be justified — see CLAUDE.md rule 1)"
	@git diff --name-only main...HEAD -- '*.go' | grep -v -E '^internal/sdk/internal/hooks/|_test\.go$$|_planmodify\.go$$' || echo "   (none)"
	@echo "==> secret scan"
	@if command -v gitleaks >/dev/null 2>&1; then gitleaks detect --source . --log-opts "origin/main..HEAD" --no-banner; else echo "   gitleaks not installed — skipped (CI runs it)"; fi

# Layer 2: NPA rules template/ordering subset — the paths behind #79/#105/#118/#124/#129.
# Run against the current tenant env (source .env or the policydemo env first).
# Set NETSKOPE_TEST_BLOCK_TEMPLATE(_2) and NETSKOPE_TEST_PERIODIC_REAUTH_TEMPLATE(_2)
# (display names) for the template-change tests.
NPA_TEMPLATE_TESTS = 'TestAccNPARules_(denyRule|blockRuleUpdate|blockRuleTemplateChange|updateWithIgnoredRuleOrder|importBlockRule|periodicReauth|withPeriodicReauth)|TestAccNPARulesOrder_|TestAccNPAPrivateApp_sharedTag|TestAccDrift_NPARules'
.PHONY: testacc-npa-templates
testacc-npa-templates:
	@echo "Running NPA template/ordering subset against $$NETSKOPE_SERVER_URL ..."
	TF_ACC=1 $(GOTEST) -v ./internal/provider/... -run $(NPA_TEMPLATE_TESTS) -timeout 60m -parallel $${TESTACC_PARALLEL:-4}


