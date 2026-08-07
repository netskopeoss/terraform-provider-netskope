# Running Tests from the CLI

This guide covers how to run unit and acceptance tests for the Netskope Terraform Provider.

## Prerequisites

- Go 1.21 or later
- Access to a Netskope tenant (for acceptance tests)

## Unit Tests

Unit tests run without API credentials and do not make network calls.

```bash
# Run all unit tests
go test -v ./...

# Run with timeout
go test -v ./... -timeout 10m

# Run specific package
go test -v ./internal/provider/...
```

## Acceptance Tests

Acceptance tests create real resources in your Netskope tenant. They require environment variables to be set.

### Required Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `TF_ACC` | Enable acceptance tests | `1` |
| `NETSKOPE_API_KEY` | Your Netskope API key (base64 encoded) | `<your-api-key>` |
| `NETSKOPE_SERVER_URL` | Your tenant API URL | `https://tenant.goskope.com/api/v2` |

### Running Acceptance Tests

#### Option 1: Inline Environment Variables

```bash
# Note: -parallel 1 runs tests sequentially to avoid API rate limits
NETSKOPE_API_KEY="your-api-key" \
NETSKOPE_SERVER_URL="https://your-tenant.goskope.com/api/v2" \
TF_ACC=1 go test -v ./internal/provider/... -timeout 120m -parallel 1
```

#### Option 2: Export Variables First

```bash
export NETSKOPE_API_KEY="your-api-key"
export NETSKOPE_SERVER_URL="https://your-tenant.goskope.com/api/v2"
export TF_ACC=1

go test -v ./internal/provider/... -timeout 120m -parallel 1
```

#### Option 3: Using a .env File

Create a `.env` file (add to `.gitignore`):

```bash
# .env
export NETSKOPE_API_KEY="your-api-key"
export NETSKOPE_SERVER_URL="https://your-tenant.goskope.com/api/v2"
export TF_ACC=1
```

Then source it before running tests:

```bash
source .env
go test -v ./internal/provider/... -timeout 120m -parallel 1
```

### Running Specific Tests

```bash
# Run a single test
TF_ACC=1 go test -v -run TestAccNPAPublisher_basic ./internal/provider/...

# Run tests matching a pattern
TF_ACC=1 go test -v -run "TestAccNPAPrivateApp" ./internal/provider/...

# Run all publisher tests
TF_ACC=1 go test -v -run "TestAccNPAPublisher" ./internal/provider/...

# Run all data source tests
TF_ACC=1 go test -v -run "DataSource" ./internal/provider/...
```

### Test Categories

| Pattern | Description |
|---------|-------------|
| `TestAccNPAPublisher` | Publisher resource tests |
| `TestAccNPAPrivateApp` | Private application tests |
| `TestAccNPAPolicyGroups` | Policy group tests |
| `TestAccNPARules` | NPA rules tests |
| `TestAccNPAPublisherUpgradeProfile` | Upgrade profile tests |
| `*DataSource*` | All data source tests |
| `*List*` | All list data source tests |

## Common Options

| Flag | Description |
|------|-------------|
| `-v` | Verbose output |
| `-timeout 120m` | Set test timeout (default 10m) |
| `-parallel 1` | Run tests sequentially (recommended - avoids API rate limits and race conditions) |
| `-run <pattern>` | Run only tests matching pattern |
| `-count 1` | Disable test caching |

## Using Make Targets

If available, use the Makefile targets:

```bash
# Run all acceptance tests
make testacc

# Run specific resource tests
make testacc-publisher
make testacc-privateapp
make testacc-policygroups
make testacc-rules

# Run with debug logging
make testacc-debug
```

## Troubleshooting

### Tests Skip with "TF_ACC must be set"

Ensure `TF_ACC=1` is set:

```bash
TF_ACC=1 go test -v ./internal/provider/...
```

### Tests Fail with "NETSKOPE_API_KEY must be set"

Set the required environment variables:

```bash
export NETSKOPE_API_KEY="your-api-key"
export NETSKOPE_SERVER_URL="https://your-tenant.goskope.com/api/v2"
```

### API Rate Limiting

Run tests sequentially with `-parallel 1`:

```bash
TF_ACC=1 go test -v ./internal/provider/... -parallel 1
```

### Test Timeout

Increase the timeout for full test suite:

```bash
TF_ACC=1 go test -v ./internal/provider/... -timeout 120m
```

### Cached Results

Force tests to run without cache:

```bash
TF_ACC=1 go test -v ./internal/provider/... -count 1
```

## Example: Full Test Run

```bash
# Complete acceptance test run
NETSKOPE_API_KEY="your-api-key" \
NETSKOPE_SERVER_URL="https://<your-tenant>.goskope.com/api/v2" \
TF_ACC=1 go test -v ./internal/provider/... -timeout 120m -parallel 1
```

## CI/CD Integration

For GitHub Actions, set these as repository secrets:

- `NETSKOPE_API_KEY`
- `NETSKOPE_SERVER_URL`

See `.github/workflows/test.yml` for the CI configuration.