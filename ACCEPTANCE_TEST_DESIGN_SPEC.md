# Acceptance Test Design Specification for terraform-provider-netskope

## Document Overview

| Field | Value |
|-------|-------|
| **Project** | terraform-provider-netskope |
| **Repository** | https://github.com/jharris-ns/terraform-provider-netskope |
| **Version** | 1.0.0 |
| **Status** | Draft |

---

## 1. Executive Summary

This document specifies the acceptance testing strategy for the Netskope Terraform Provider. Acceptance tests execute real Terraform operations against the Netskope API, validating that resources and data sources function correctly in production-like scenarios.

### 1.1 Objectives

1. Validate complete CRUD lifecycle for all resources
2. Verify data sources return accurate information
3. Test import functionality for supported resources
4. Ensure state management works correctly
5. Detect breaking changes in provider behavior
6. Validate interaction with Netskope API

### 1.2 Relationship to Other Test Types

| Test Type | Location | Purpose | Network Required |
|-----------|----------|---------|------------------|
| **Unit Tests** | `internal/sdk/internal/hooks/*_test.go` | Test custom hooks/models | No |
| **Acceptance Tests** | `internal/provider/*_test.go` | Test resources/data sources | Yes |
| **Integration Tests** | `terraform-provider-netskope-tests/crud/` | Multi-resource workflows | Yes |

### 1.3 Speakeasy Compatibility

Acceptance test files (`*_test.go`) in `internal/provider/` are **safe from Speakeasy regeneration**:

- Speakeasy does not generate `*_test.go` files in the provider directory
- Test files will persist across `speakeasy generate` runs
- No changes to `.genignore` required

---

## 2. Prerequisites

### 2.1 Dependencies

Add to `go.mod`:

```go
require (
    // Existing dependencies...
    github.com/hashicorp/terraform-plugin-testing v1.11.0
)
```

Run:
```bash
go get github.com/hashicorp/terraform-plugin-testing@latest
go mod tidy
```

### 2.2 Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `TF_ACC` | Yes | Set to `1` to enable acceptance tests |
| `NETSKOPE_API_KEY` | Yes | API key for Netskope tenant |
| `NETSKOPE_SERVER_URL` | Yes | Tenant URL (e.g., `https://tenant.goskope.com/api/v2`) |
| `TF_ACC_TERRAFORM_PATH` | No | Path to Terraform binary |
| `TF_LOG` | No | Set to `DEBUG` for verbose output |

### 2.3 Test Tenant Requirements

- Dedicated Netskope tenant for testing (not production)
- API key with full NPA permissions
- No production resources that could be affected
- Sufficient quota for test resource creation

---

## 3. Test Infrastructure

### 3.1 Provider Factory Setup

**File:** `internal/provider/provider_test.go`

```go
package provider

import (
    "os"
    "testing"

    "github.com/hashicorp/terraform-plugin-framework/providerserver"
    "github.com/hashicorp/terraform-plugin-go/tfprotov6"
    "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testAccProtoV6ProviderFactories creates provider factories for acceptance tests
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
    "netskope": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck validates required environment variables are set
func testAccPreCheck(t *testing.T) {
    if v := os.Getenv("NETSKOPE_API_KEY"); v == "" {
        t.Fatal("NETSKOPE_API_KEY must be set for acceptance tests")
    }
    if v := os.Getenv("NETSKOPE_SERVER_URL"); v == "" {
        t.Fatal("NETSKOPE_SERVER_URL must be set for acceptance tests")
    }
}

// testAccProviderConfig returns the provider configuration block
func testAccProviderConfig() string {
    return `
provider "netskope" {
    # Credentials from environment variables
}
`
}
```

### 3.2 Test Naming Convention

| Pattern | Description |
|---------|-------------|
| `TestAcc{Resource}_basic` | Minimal resource creation |
| `TestAcc{Resource}_complete` | All attributes specified |
| `TestAcc{Resource}_update` | Attribute modification |
| `TestAcc{Resource}_import` | Import existing resource |
| `TestAcc{Resource}_disappears` | Resource deleted outside Terraform |
| `TestAcc{DataSource}_basic` | Data source read |
| `TestAcc{DataSource}_filter` | Data source with filters |

### 3.3 Test Prefix Convention

All test resources should use a consistent naming prefix to:
- Identify test-created resources
- Enable cleanup of orphaned resources
- Prevent collision with production resources

```go
const testAccResourcePrefix = "tf-acc-test"
```

---

## 4. Resource Test Specifications

### 4.1 netskope_npa_private_app

**File:** `internal/provider/npaprivateapp_resource_test.go`

#### Test Cases

| Test ID | Test Name | Description | Priority |
|---------|-----------|-------------|----------|
| PA-ACC-001 | `TestAccNPAPrivateApp_basic` | Create with required attributes | P0 |
| PA-ACC-002 | `TestAccNPAPrivateApp_complete` | Create with all attributes | P0 |
| PA-ACC-003 | `TestAccNPAPrivateApp_update` | Update app name and protocols | P0 |
| PA-ACC-004 | `TestAccNPAPrivateApp_updatePublishers` | Change publisher assignment | P1 |
| PA-ACC-005 | `TestAccNPAPrivateApp_import` | Import existing app | P0 |
| PA-ACC-006 | `TestAccNPAPrivateApp_multipleProtocols` | Multiple TCP/UDP protocols | P1 |
| PA-ACC-007 | `TestAccNPAPrivateApp_clientlessAccess` | Browser access app | P1 |
| PA-ACC-008 | `TestAccNPAPrivateApp_tags` | Create with tags | P2 |
| PA-ACC-009 | `TestAccNPAPrivateApp_disappears` | Handle external deletion | P2 |

#### Example Implementation

```go
package provider

import (
    "fmt"
    "testing"

    "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
    "github.com/hashicorp/terraform-plugin-testing/helper/resource"
    "github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccNPAPrivateApp_basic(t *testing.T) {
    rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
    resourceName := "netskope_npa_private_app.test"

    resource.ParallelTest(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        CheckDestroy:             testAccCheckNPAPrivateAppDestroy,
        Steps: []resource.TestStep{
            // Create and Read
            {
                Config: testAccNPAPrivateAppConfig_basic(rName),
                Check: resource.ComposeAggregateTestCheckFunc(
                    testAccCheckNPAPrivateAppExists(resourceName),
                    resource.TestCheckResourceAttr(resourceName, "private_app_name", rName),
                    resource.TestCheckResourceAttr(resourceName, "private_app_hostname", "192.168.1.100"),
                    resource.TestCheckResourceAttrSet(resourceName, "private_app_id"),
                    resource.TestCheckResourceAttr(resourceName, "protocols.#", "1"),
                    resource.TestCheckResourceAttr(resourceName, "protocols.0.port", "443"),
                    resource.TestCheckResourceAttr(resourceName, "protocols.0.protocol", "tcp"),
                ),
            },
            // Import
            {
                ResourceName:      resourceName,
                ImportState:       true,
                ImportStateVerify: true,
                // Attributes that may differ on import
                ImportStateVerifyIgnore: []string{},
            },
        },
    })
}

func TestAccNPAPrivateApp_update(t *testing.T) {
    rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
    rNameUpdated := fmt.Sprintf("%s-updated", rName)
    resourceName := "netskope_npa_private_app.test"

    resource.ParallelTest(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        CheckDestroy:             testAccCheckNPAPrivateAppDestroy,
        Steps: []resource.TestStep{
            // Create
            {
                Config: testAccNPAPrivateAppConfig_basic(rName),
                Check: resource.ComposeAggregateTestCheckFunc(
                    testAccCheckNPAPrivateAppExists(resourceName),
                    resource.TestCheckResourceAttr(resourceName, "private_app_name", rName),
                ),
            },
            // Update
            {
                Config: testAccNPAPrivateAppConfig_updated(rNameUpdated),
                Check: resource.ComposeAggregateTestCheckFunc(
                    testAccCheckNPAPrivateAppExists(resourceName),
                    resource.TestCheckResourceAttr(resourceName, "private_app_name", rNameUpdated),
                    resource.TestCheckResourceAttr(resourceName, "protocols.#", "2"),
                ),
            },
        },
    })
}

func TestAccNPAPrivateApp_import(t *testing.T) {
    rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
    resourceName := "netskope_npa_private_app.test"

    resource.ParallelTest(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        CheckDestroy:             testAccCheckNPAPrivateAppDestroy,
        Steps: []resource.TestStep{
            {
                Config: testAccNPAPrivateAppConfig_basic(rName),
            },
            {
                ResourceName:            resourceName,
                ImportState:             true,
                ImportStateVerify:       true,
                ImportStateIdFunc:       testAccNPAPrivateAppImportStateIdFunc(resourceName),
            },
        },
    })
}

// Configuration functions

func testAccNPAPrivateAppConfig_basic(name string) string {
    return fmt.Sprintf(`
%s

resource "netskope_npa_private_app" "test" {
    private_app_name     = %q
    private_app_hostname = "192.168.1.100"

    protocols = [
        {
            port     = "443"
            protocol = "tcp"
        }
    ]

    publishers = [
        {
            publisher_id   = tostring(netskope_npa_publisher.test.publisher_id)
            publisher_name = netskope_npa_publisher.test.publisher_name
        }
    ]

    use_publisher_dns       = true
    trust_self_signed_certs = false
}

resource "netskope_npa_publisher" "test" {
    publisher_name = "%s-publisher"
}
`, testAccProviderConfig(), name, name)
}

func testAccNPAPrivateAppConfig_updated(name string) string {
    return fmt.Sprintf(`
%s

resource "netskope_npa_private_app" "test" {
    private_app_name     = %q
    private_app_hostname = "192.168.1.100,192.168.1.101"

    protocols = [
        {
            port     = "443"
            protocol = "tcp"
        },
        {
            port     = "22"
            protocol = "tcp"
        }
    ]

    publishers = [
        {
            publisher_id   = tostring(netskope_npa_publisher.test.publisher_id)
            publisher_name = netskope_npa_publisher.test.publisher_name
        }
    ]

    use_publisher_dns       = true
    trust_self_signed_certs = false
}

resource "netskope_npa_publisher" "test" {
    publisher_name = "%s-publisher"
}
`, testAccProviderConfig(), name, name)
}

// Helper functions

func testAccCheckNPAPrivateAppExists(resourceName string) resource.TestCheckFunc {
    return func(s *terraform.State) error {
        rs, ok := s.RootModule().Resources[resourceName]
        if !ok {
            return fmt.Errorf("resource not found: %s", resourceName)
        }

        if rs.Primary.ID == "" {
            return fmt.Errorf("resource ID not set")
        }

        // Optionally verify via API call
        return nil
    }
}

func testAccCheckNPAPrivateAppDestroy(s *terraform.State) error {
    for _, rs := range s.RootModule().Resources {
        if rs.Type != "netskope_npa_private_app" {
            continue
        }

        // Verify resource no longer exists via API
        // Return error if resource still exists
    }
    return nil
}

func testAccNPAPrivateAppImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
    return func(s *terraform.State) (string, error) {
        rs, ok := s.RootModule().Resources[resourceName]
        if !ok {
            return "", fmt.Errorf("resource not found: %s", resourceName)
        }
        return rs.Primary.Attributes["private_app_id"], nil
    }
}
```

---

### 4.2 netskope_npa_publisher

**File:** `internal/provider/npapublisher_resource_test.go`

#### Test Cases

| Test ID | Test Name | Description | Priority |
|---------|-----------|-------------|----------|
| PB-ACC-001 | `TestAccNPAPublisher_basic` | Create basic publisher | P0 |
| PB-ACC-002 | `TestAccNPAPublisher_withUpgradeProfile` | Create with upgrade profile | P1 |
| PB-ACC-003 | `TestAccNPAPublisher_update` | Update publisher name | P0 |
| PB-ACC-004 | `TestAccNPAPublisher_import` | Import existing publisher | P0 |
| PB-ACC-005 | `TestAccNPAPublisher_disappears` | Handle external deletion | P2 |

#### Example Implementation

```go
func TestAccNPAPublisher_basic(t *testing.T) {
    rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
    resourceName := "netskope_npa_publisher.test"

    resource.ParallelTest(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        CheckDestroy:             testAccCheckNPAPublisherDestroy,
        Steps: []resource.TestStep{
            {
                Config: testAccNPAPublisherConfig_basic(rName),
                Check: resource.ComposeAggregateTestCheckFunc(
                    testAccCheckNPAPublisherExists(resourceName),
                    resource.TestCheckResourceAttr(resourceName, "publisher_name", rName),
                    resource.TestCheckResourceAttrSet(resourceName, "publisher_id"),
                ),
            },
            {
                ResourceName:      resourceName,
                ImportState:       true,
                ImportStateVerify: true,
            },
        },
    })
}

func testAccNPAPublisherConfig_basic(name string) string {
    return fmt.Sprintf(`
%s

resource "netskope_npa_publisher" "test" {
    publisher_name = %q
}
`, testAccProviderConfig(), name)
}
```

---

### 4.3 netskope_npa_policy_groups

**File:** `internal/provider/npapolicygroups_resource_test.go`

#### Test Cases

| Test ID | Test Name | Description | Priority |
|---------|-----------|-------------|----------|
| PG-ACC-001 | `TestAccNPAPolicyGroups_basic` | Create policy group | P0 |
| PG-ACC-002 | `TestAccNPAPolicyGroups_update` | Update group name | P0 |
| PG-ACC-003 | `TestAccNPAPolicyGroups_import` | Import existing group | P0 |

---

### 4.4 netskope_npa_rules

**File:** `internal/provider/nparules_resource_test.go`

#### Test Cases

| Test ID | Test Name | Description | Priority |
|---------|-----------|-------------|----------|
| RU-ACC-001 | `TestAccNPARules_basic` | Create allow rule | P0 |
| RU-ACC-002 | `TestAccNPARules_denyRule` | Create deny rule | P1 |
| RU-ACC-003 | `TestAccNPARules_withUserCriteria` | Rule with user matching | P1 |
| RU-ACC-004 | `TestAccNPARules_update` | Update rule criteria | P0 |
| RU-ACC-005 | `TestAccNPARules_import` | Import existing rule | P0 |
| RU-ACC-006 | `TestAccNPARules_withDLP` | Rule with DLP actions | P2 |

#### Example Implementation

```go
func TestAccNPARules_basic(t *testing.T) {
    rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
    resourceName := "netskope_npa_rules.test"

    resource.ParallelTest(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        CheckDestroy:             testAccCheckNPARulesDestroy,
        Steps: []resource.TestStep{
            {
                Config: testAccNPARulesConfig_basic(rName),
                Check: resource.ComposeAggregateTestCheckFunc(
                    testAccCheckNPARulesExists(resourceName),
                    resource.TestCheckResourceAttr(resourceName, "rule_name", rName),
                    resource.TestCheckResourceAttr(resourceName, "enabled", "1"),
                    resource.TestCheckResourceAttr(resourceName, "rule_data.policy_type", "private-app"),
                    resource.TestCheckResourceAttr(resourceName, "rule_data.match_criteria_action.action_name", "allow"),
                ),
            },
            {
                ResourceName:      resourceName,
                ImportState:       true,
                ImportStateVerify: true,
            },
        },
    })
}

func testAccNPARulesConfig_basic(name string) string {
    return fmt.Sprintf(`
%s

resource "netskope_npa_policy_groups" "test" {
    name        = "%s-group"
    description = "Test policy group"
}

resource "netskope_npa_publisher" "test" {
    publisher_name = "%s-publisher"
}

resource "netskope_npa_private_app" "test" {
    private_app_name     = "%s-app"
    private_app_hostname = "192.168.1.100"

    protocols = [
        {
            port     = "443"
            protocol = "tcp"
        }
    ]

    publishers = [
        {
            publisher_id   = tostring(netskope_npa_publisher.test.publisher_id)
            publisher_name = netskope_npa_publisher.test.publisher_name
        }
    ]
}

resource "netskope_npa_rules" "test" {
    rule_name   = %q
    description = "Test rule"
    enabled     = "1"
    group_id    = netskope_npa_policy_groups.test.id

    rule_data = {
        policy_type = "private-app"

        match_criteria_action = {
            action_name = "allow"
        }

        user_id      = ["*"]
        private_apps = [netskope_npa_private_app.test.private_app_name]
        access_method = ["Client"]
    }
}
`, testAccProviderConfig(), name, name, name, name)
}
```

---

### 4.5 Additional Resources

#### netskope_npa_publisher_upgrade_profile

| Test ID | Test Name | Priority |
|---------|-----------|----------|
| UP-ACC-001 | `TestAccNPAPublisherUpgradeProfile_basic` | P1 |
| UP-ACC-002 | `TestAccNPAPublisherUpgradeProfile_update` | P1 |
| UP-ACC-003 | `TestAccNPAPublisherUpgradeProfile_import` | P1 |

#### netskope_npa_publisher_token

| Test ID | Test Name | Priority |
|---------|-----------|----------|
| PT-ACC-001 | `TestAccNPAPublisherToken_basic` | P1 |

#### netskope_npa_private_app_public_host

| Test ID | Test Name | Priority |
|---------|-----------|----------|
| PH-ACC-001 | `TestAccNPAPrivateAppPublicHost_basic` | P2 |
| PH-ACC-002 | `TestAccNPAPrivateAppPublicHost_update` | P2 |

#### netskope_npa_publishers_alerts_configuration

| Test ID | Test Name | Priority |
|---------|-----------|----------|
| AC-ACC-001 | `TestAccNPAPublishersAlertsConfiguration_basic` | P2 |

---

## 5. Data Source Test Specifications

### 5.1 netskope_npa_private_app (Data Source)

**File:** `internal/provider/npaprivateapp_data_source_test.go`

| Test ID | Test Name | Description | Priority |
|---------|-----------|-------------|----------|
| PA-DS-001 | `TestAccNPAPrivateAppDataSource_basic` | Fetch by ID | P0 |
| PA-DS-002 | `TestAccNPAPrivateAppDataSource_notFound` | Handle missing app | P1 |

```go
func TestAccNPAPrivateAppDataSource_basic(t *testing.T) {
    rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
    resourceName := "netskope_npa_private_app.test"
    dataSourceName := "data.netskope_npa_private_app.test"

    resource.ParallelTest(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: testAccNPAPrivateAppDataSourceConfig_basic(rName),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttrPair(
                        dataSourceName, "private_app_id",
                        resourceName, "private_app_id",
                    ),
                    resource.TestCheckResourceAttrPair(
                        dataSourceName, "private_app_name",
                        resourceName, "private_app_name",
                    ),
                ),
            },
        },
    })
}

func testAccNPAPrivateAppDataSourceConfig_basic(name string) string {
    return fmt.Sprintf(`
%s

resource "netskope_npa_publisher" "test" {
    publisher_name = "%s-publisher"
}

resource "netskope_npa_private_app" "test" {
    private_app_name     = %q
    private_app_hostname = "192.168.1.100"

    protocols = [
        {
            port     = "443"
            protocol = "tcp"
        }
    ]

    publishers = [
        {
            publisher_id   = tostring(netskope_npa_publisher.test.publisher_id)
            publisher_name = netskope_npa_publisher.test.publisher_name
        }
    ]
}

data "netskope_npa_private_app" "test" {
    private_app_id = netskope_npa_private_app.test.private_app_id
}
`, testAccProviderConfig(), name, name)
}
```

### 5.2 List Data Sources

| Data Source | Test Name | Priority |
|-------------|-----------|----------|
| `netskope_npa_private_apps_list` | `TestAccNPAPrivateAppsListDataSource_basic` | P1 |
| `netskope_npa_publishers_list` | `TestAccNPAPublishersListDataSource_basic` | P1 |
| `netskope_npa_policy_groups_list` | `TestAccNPAPolicyGroupsListDataSource_basic` | P1 |
| `netskope_npa_rules_list` | `TestAccNPARulesListDataSource_basic` | P1 |
| `netskope_npa_publisher_upgrade_profiles_list` | `TestAccNPAPublisherUpgradeProfilesListDataSource_basic` | P2 |

---

## 6. Test Execution

### 6.1 Makefile Targets

Add to `Makefile`:

```makefile
# Run all acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test -v ./internal/provider/... -timeout 120m

# Run acceptance tests with coverage
.PHONY: testacc-coverage
testacc-coverage:
	TF_ACC=1 go test -v -coverprofile=coverage-acc.out ./internal/provider/... -timeout 120m
	go tool cover -html=coverage-acc.out -o coverage-acc.html

# Run specific resource tests
.PHONY: testacc-privateapp
testacc-privateapp:
	TF_ACC=1 go test -v ./internal/provider/... -run TestAccNPAPrivateApp -timeout 30m

.PHONY: testacc-publisher
testacc-publisher:
	TF_ACC=1 go test -v ./internal/provider/... -run TestAccNPAPublisher -timeout 30m

.PHONY: testacc-rules
testacc-rules:
	TF_ACC=1 go test -v ./internal/provider/... -run TestAccNPARules -timeout 30m

# Run data source tests only
.PHONY: testacc-datasources
testacc-datasources:
	TF_ACC=1 go test -v ./internal/provider/... -run TestAcc.*DataSource -timeout 30m

# Run with debug logging
.PHONY: testacc-debug
testacc-debug:
	TF_ACC=1 TF_LOG=DEBUG go test -v ./internal/provider/... -timeout 120m 2>&1 | tee test-debug.log
```

### 6.2 Running Tests Locally

```bash
# Set environment variables
export NETSKOPE_API_KEY="your-api-key"
export NETSKOPE_SERVER_URL="https://your-tenant.goskope.com/api/v2"

# Run all acceptance tests
make testacc

# Run specific test
TF_ACC=1 go test -v ./internal/provider/... -run TestAccNPAPrivateApp_basic -timeout 10m

# Run with verbose Terraform logging
TF_ACC=1 TF_LOG=DEBUG go test -v ./internal/provider/... -run TestAccNPAPrivateApp_basic
```

### 6.3 Parallel Test Execution

Tests use `resource.ParallelTest()` to run concurrently. To limit parallelism:

```bash
TF_ACC=1 go test -v -parallel 4 ./internal/provider/... -timeout 120m
```

---

## 7. CI/CD Integration

### 7.1 GitHub Actions Workflow

**File:** `.github/workflows/acceptance-tests.yaml`

```yaml
name: Acceptance Tests

on:
  push:
    branches: [main]
    paths:
      - 'internal/provider/**'
      - '.github/workflows/acceptance-tests.yaml'
  pull_request:
    branches: [main]
    paths:
      - 'internal/provider/**'
  schedule:
    # Run nightly at 2 AM UTC
    - cron: '0 2 * * *'
  workflow_dispatch:
    inputs:
      test_pattern:
        description: 'Test pattern to run (e.g., TestAccNPAPrivateApp)'
        required: false
        default: ''

env:
  GO_VERSION: '1.23'
  TERRAFORM_VERSION: '1.6.0'

jobs:
  acceptance-tests:
    runs-on: ubuntu-latest
    timeout-minutes: 120

    # Use GitHub environments for secrets management
    environment: acceptance-tests

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: Set up Terraform
        uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: ${{ env.TERRAFORM_VERSION }}
          terraform_wrapper: false

      - name: Cache Go modules
        uses: actions/cache@v4
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
          restore-keys: |
            ${{ runner.os }}-go-

      - name: Download dependencies
        run: go mod download

      - name: Run Acceptance Tests
        env:
          TF_ACC: '1'
          NETSKOPE_API_KEY: ${{ secrets.NETSKOPE_API_KEY }}
          NETSKOPE_SERVER_URL: ${{ secrets.NETSKOPE_SERVER_URL }}
        run: |
          if [ -n "${{ github.event.inputs.test_pattern }}" ]; then
            go test -v ./internal/provider/... -run "${{ github.event.inputs.test_pattern }}" -timeout 60m
          else
            go test -v ./internal/provider/... -timeout 120m
          fi

      - name: Upload Test Results
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: test-results
          path: |
            test-debug.log
            coverage-acc.out
```

### 7.2 GitHub Secrets Configuration

Configure these secrets in the repository settings:

| Secret | Description |
|--------|-------------|
| `NETSKOPE_API_KEY` | API key for test tenant |
| `NETSKOPE_SERVER_URL` | Test tenant URL |

### 7.3 Branch Protection

Recommended branch protection rules for `main`:

- Require acceptance tests to pass before merge
- Require up-to-date branches
- Allow bypass for emergency fixes only

---

## 8. Test Cleanup and Resource Management

### 8.1 Automatic Cleanup

The testing framework automatically destroys resources after each test. Implement `CheckDestroy` functions to verify:

```go
func testAccCheckNPAPrivateAppDestroy(s *terraform.State) error {
    // Get SDK client
    // client := ...

    for _, rs := range s.RootModule().Resources {
        if rs.Type != "netskope_npa_private_app" {
            continue
        }

        id := rs.Primary.Attributes["private_app_id"]

        // Attempt to fetch the resource
        // If found, return error
        // If not found (404), return nil
    }
    return nil
}
```

### 8.2 Manual Cleanup Script

For orphaned test resources:

**File:** `scripts/cleanup-test-resources.sh`

```bash
#!/bin/bash
# Cleanup orphaned acceptance test resources

set -e

PREFIX="tf-acc-test"

echo "Fetching test resources with prefix: $PREFIX"

# List and delete private apps
echo "Cleaning up private apps..."
# Use API calls to list and delete

# List and delete publishers
echo "Cleaning up publishers..."
# Use API calls to list and delete

# List and delete policy groups
echo "Cleaning up policy groups..."
# Use API calls to list and delete

echo "Cleanup complete"
```

### 8.3 Sweep Functions

Implement sweep functions for CI cleanup:

```go
func init() {
    resource.AddTestSweepers("netskope_npa_private_app", &resource.Sweeper{
        Name: "netskope_npa_private_app",
        F:    sweepNPAPrivateApps,
    })
}

func sweepNPAPrivateApps(region string) error {
    // List all private apps with test prefix
    // Delete each one
    return nil
}
```

Run sweepers:
```bash
go test ./internal/provider/... -sweep=all -v
```

---

## 9. Coverage Goals

### 9.1 Resource Coverage Matrix

| Resource | Create | Read | Update | Delete | Import | Target |
|----------|--------|------|--------|--------|--------|--------|
| `netskope_npa_private_app` | ✓ | ✓ | ✓ | ✓ | ✓ | 100% |
| `netskope_npa_publisher` | ✓ | ✓ | ✓ | ✓ | ✓ | 100% |
| `netskope_npa_policy_groups` | ✓ | ✓ | ✓ | ✓ | ✓ | 100% |
| `netskope_npa_rules` | ✓ | ✓ | ✓ | ✓ | ✓ | 100% |
| `netskope_npa_publisher_upgrade_profile` | ✓ | ✓ | ✓ | ✓ | ✓ | 80% |
| `netskope_npa_publisher_token` | ✓ | ✓ | - | ✓ | - | 80% |
| `netskope_npa_private_app_public_host` | ✓ | ✓ | ✓ | ✓ | ✓ | 60% |
| `netskope_npa_publishers_alerts_configuration` | ✓ | ✓ | ✓ | ✓ | ✓ | 60% |
| `netskope_npa_publishers_bulk_*` | ✓ | - | - | - | - | 40% |

### 9.2 Data Source Coverage Matrix

| Data Source | Read | Target |
|-------------|------|--------|
| `netskope_npa_private_app` | ✓ | 100% |
| `netskope_npa_private_apps_list` | ✓ | 80% |
| `netskope_npa_publisher` | ✓ | 100% |
| `netskope_npa_publishers_list` | ✓ | 80% |
| `netskope_npa_policy_groups` | ✓ | 100% |
| `netskope_npa_policy_groups_list` | ✓ | 80% |
| `netskope_npa_rules` | ✓ | 100% |
| `netskope_npa_rules_list` | ✓ | 80% |
| Other data sources | ✓ | 60% |

---

## 10. Implementation Roadmap

### Phase 1: Foundation (Week 1)

- [ ] Add `terraform-plugin-testing` dependency
- [ ] Create `provider_test.go` with factories and helpers
- [ ] Implement `TestAccNPAPublisher_basic` (simplest resource)
- [ ] Implement `TestAccNPAPublisher_import`
- [ ] Add Makefile targets

### Phase 2: Core Resources (Week 2-3)

- [ ] Implement all `TestAccNPAPrivateApp_*` tests
- [ ] Implement all `TestAccNPAPolicyGroups_*` tests
- [ ] Implement all `TestAccNPARules_*` tests
- [ ] Implement data source tests for core resources

### Phase 3: Secondary Resources (Week 4)

- [ ] Implement `TestAccNPAPublisherUpgradeProfile_*` tests
- [ ] Implement `TestAccNPAPublisherToken_*` tests
- [ ] Implement remaining data source tests

### Phase 4: CI/CD Integration (Week 5)

- [ ] Create GitHub Actions workflow
- [ ] Configure secrets and environments
- [ ] Implement sweep functions
- [ ] Create cleanup scripts
- [ ] Document test procedures

---

## 11. Appendix

### 11.1 Test File Structure

```
terraform-provider-netskope/
└── internal/
    └── provider/
        ├── provider.go                          # Generated
        ├── provider_test.go                     # NEW - Test infrastructure
        │
        ├── npaprivateapp_resource.go            # Generated
        ├── npaprivateapp_resource_test.go       # NEW - Acceptance tests
        │
        ├── npaprivateapp_data_source.go         # Generated
        ├── npaprivateapp_data_source_test.go    # NEW - Data source tests
        │
        ├── npapublisher_resource.go             # Generated
        ├── npapublisher_resource_test.go        # NEW - Acceptance tests
        │
        ├── npapolicygroups_resource.go          # Generated
        ├── npapolicygroups_resource_test.go     # NEW - Acceptance tests
        │
        ├── nparules_resource.go                 # Generated
        ├── nparules_resource_test.go            # NEW - Acceptance tests
        │
        └── ... (other resources and tests)
```

### 11.2 References

- [HashiCorp Acceptance Testing Guide](https://developer.hashicorp.com/terraform/plugin/testing/acceptance-tests)
- [terraform-plugin-testing Module](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-testing)
- [AWS Provider Test Examples](https://github.com/hashicorp/terraform-provider-aws/tree/main/internal/service)

### 11.3 Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0.0 | 2026-01-21 | - | Initial specification |
