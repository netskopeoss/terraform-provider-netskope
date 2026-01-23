# Acceptance Tests for terraform-provider-netskope

This document describes the acceptance tests for the Netskope Terraform Provider, their coverage, and how to run them.

## Overview

Acceptance tests verify that the provider works correctly against a real Netskope tenant. They create, read, update, and delete actual resources, ensuring the provider functions as expected in production environments.

## Prerequisites

### Environment Variables

```bash
export NETSKOPE_API_KEY="your-api-key"
export NETSKOPE_SERVER_URL="https://your-tenant.goskope.com/api/v2"
```

### Dependencies

The tests use the HashiCorp terraform-plugin-testing framework:

```bash
go get github.com/hashicorp/terraform-plugin-testing@latest
```

## Running Tests

### Run All Tests

```bash
make testacc
```

Or directly:

```bash
TF_ACC=1 go test -v ./internal/provider/... -timeout 120m
```

### Run Specific Resource Tests

```bash
# Publisher tests
make testacc-publisher

# Private app tests
make testacc-privateapp

# Policy groups tests
make testacc-policygroups

# Rules tests
make testacc-rules

# Data source tests only
make testacc-datasources
```

### Run with Debug Logging

```bash
make testacc-debug
```

### Run Tests Sequentially (Avoid Rate Limiting)

```bash
TF_ACC=1 go test -v ./internal/provider/... -timeout 120m -parallel 1
```

## Test Coverage

### Resource Tests

| Resource | Test | Description | Status |
|----------|------|-------------|--------|
| **netskope_npa_publisher** | | | |
| | `TestAccNPAPublisher_basic` | Create publisher with minimal config | Pass |
| | `TestAccNPAPublisher_update` | Update publisher name | Pass |
| | `TestAccNPAPublisher_import` | Import existing publisher | Pass |
| | `TestAccNPAPublisher_withUpgradeProfile` | Create with upgrade profile | Pass |
| **netskope_npa_private_app** | | | |
| | `TestAccNPAPrivateApp_basic` | Create with required attributes | Pass |
| | `TestAccNPAPrivateApp_complete` | Create with all attributes | Pass |
| | `TestAccNPAPrivateApp_update` | Update hostname and protocols | Pass |
| | `TestAccNPAPrivateApp_import` | Import existing app | Pass |
| | `TestAccNPAPrivateApp_multipleProtocols` | Create with TCP/UDP protocols | Pass |
| **netskope_npa_policy_groups** | | | |
| | `TestAccNPAPolicyGroups_basic` | Create policy group | Pass |
| | `TestAccNPAPolicyGroups_update` | Update group name | Skip* |
| | `TestAccNPAPolicyGroups_import` | Import existing group | Pass |
| **netskope_npa_rules** | | | |
| | `TestAccNPARules_basic` | Create allow rule | Pass |
| | `TestAccNPARules_update` | Update rule criteria | Skip* |
| | `TestAccNPARules_import` | Import existing rule | Pass |
| | `TestAccNPARules_denyRule` | Create block rule | Skip* |

### Data Source Tests

| Data Source | Test | Description | Status |
|-------------|------|-------------|--------|
| **netskope_npa_publisher** | `TestAccNPAPublisherDataSource_basic` | Read single publisher | Pass |
| **netskope_npa_private_app** | `TestAccNPAPrivateAppDataSource_basic` | Read single private app | Pass |
| **netskope_npa_policy_groups** | `TestAccNPAPolicyGroupsDataSource_basic` | Read single policy group | Pass |
| **netskope_npa_rules** | `TestAccNPARulesDataSource_basic` | Read single rule | Pass |
| **netskope_npa_publishers_list** | `TestAccNPAPublishersListDataSource_basic` | List all publishers | Pass |
| **netskope_npa_private_apps_list** | `TestAccNPAPrivateAppsListDataSource_basic` | List all private apps | Pass |
| **netskope_npa_policy_groups_list** | `TestAccNPAPolicyGroupsListDataSource_basic` | List all policy groups | Pass |
| **netskope_npa_rules_list** | `TestAccNPARulesListDataSource_basic` | List all rules | Pass |

## Test Summary

| Category | Total | Pass | Skip | Fail |
|----------|-------|------|------|------|
| Publisher Resource | 4 | 4 | 0 | 0 |
| Private App Resource | 5 | 5 | 0 | 0 |
| Policy Groups Resource | 3 | 2 | 1 | 0 |
| Rules Resource | 4 | 2 | 2 | 0 |
| Data Sources | 8 | 8 | 0 | 0 |
| **Total** | **24** | **21** | **3** | **0** |

## Skipped Tests

The following tests are skipped due to known provider or API limitations:

### TestAccNPAPolicyGroups_update
- **Reason**: Provider bug with policy group updates
- **Error**: SQLAlchemy bind parameter error for `new_order`
- **Impact**: Policy group names cannot be updated after creation

### TestAccNPARules_update
- **Reason**: Provider issue with stale private app references
- **Error**: "Private app doesn't exist" when updating rules
- **Impact**: Rules cannot be updated if dependent resources are recreated

### TestAccNPARules_denyRule
- **Reason**: API requires undocumented `template` field for block actions
- **Error**: "template field is required for block action"
- **Impact**: Block/deny rules require additional configuration not documented

## Known Issues

### Plan Drift After Apply

Some resources show plan drift after apply due to computed fields returned by the API that weren't specified in the configuration. Tests use `ExpectNonEmptyPlan: true` to acknowledge this behavior.

Affected resources:
- `netskope_npa_private_app` - publishers block, real_host, protocols computed fields
- Resources dependent on private apps inherit this behavior

### Import State Verification

Some fields are excluded from import state verification because:
1. They are only used during creation (e.g., `group_order`, `rule_order`)
2. They have computed values not stored in state
3. They differ between config and API response format

## Test Infrastructure

### Test Helpers

| Function | Description |
|----------|-------------|
| `testAccPreCheck(t)` | Validates required environment variables |
| `testAccProviderConfig()` | Returns base provider configuration |
| `testAccProtoV6ProviderFactories` | Provider factory for Protocol v6 |
| `testAccResourcePrefix` | Prefix for test resources (`tf-acc-test`) |

### Resource Naming

All test resources use the prefix `tf-acc-test-` followed by a random 8-character string to:
- Identify resources created by acceptance tests
- Avoid naming conflicts with existing resources
- Enable easy cleanup if tests fail

## Cleanup

The acceptance test framework automatically destroys resources after each test. If tests fail mid-execution, resources may remain in the tenant. To identify and clean up test resources:

1. Look for resources with names starting with `tf-acc-test-`
2. Delete them via the Netskope admin console or API

## Contributing

When adding new tests:

1. Follow the existing naming convention: `TestAcc<Resource>_<scenario>`
2. Use `testAccResourcePrefix` for resource names
3. Include both create and import steps where applicable
4. Document any skipped tests with clear reasons
5. Add `ExpectNonEmptyPlan: true` if the provider has known drift issues
6. Update this document with new test coverage

## Makefile Targets

| Target | Description |
|--------|-------------|
| `testacc` | Run all acceptance tests |
| `testacc-coverage` | Run with coverage report |
| `testacc-privateapp` | Run private app tests only |
| `testacc-publisher` | Run publisher tests only |
| `testacc-policygroups` | Run policy groups tests only |
| `testacc-rules` | Run rules tests only |
| `testacc-datasources` | Run data source tests only |
| `testacc-debug` | Run with debug logging |

## CI/CD Integration

### GitHub Actions

A GitHub Actions workflow is configured at `.github/workflows/test.yml` to automate testing:

| Trigger | Job | Description |
|---------|-----|-------------|
| Push to `main` | Unit Tests | Runs `go test` on all packages |
| Pull Request to `main` | Unit Tests | Runs `go test` on all packages |
| Push to `main` | Acceptance Tests | Runs full acceptance test suite |

**Note:** Acceptance tests only run on pushes to `main`, not on PRs, to protect secrets.

### Setting Up GitHub Secrets

1. Go to your repository on GitHub
2. Navigate to **Settings** > **Secrets and variables** > **Actions**
3. Click **New repository secret**
4. Add the following secrets:

| Secret Name | Value |
|-------------|-------|
| `NETSKOPE_API_KEY` | Your Netskope API key |
| `NETSKOPE_SERVER_URL` | `https://your-tenant.goskope.com/api/v2` |

### Manual Workflow Dispatch (Optional)

To enable manual triggering of acceptance tests, add this to the workflow:

```yaml
on:
  workflow_dispatch:
    inputs:
      test_filter:
        description: 'Test filter (e.g., TestAccNPAPublisher)'
        required: false
        default: ''
```

### Running Acceptance Tests on PRs

If you want to run acceptance tests on PRs from trusted contributors:

1. Create a separate workflow that requires approval
2. Use `pull_request_target` event with environment protection rules
3. Configure an environment with required reviewers

Example environment setup:
1. Go to **Settings** > **Environments**
2. Create environment named `acceptance-tests`
3. Add required reviewers
4. Update workflow to use `environment: acceptance-tests`

### Git Hooks (Local)

For local pre-commit or pre-push hooks, you can use tools like:

**Using pre-commit framework:**

```bash
pip install pre-commit
```

Create `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: local
    hooks:
      - id: go-test
        name: Go Unit Tests
        entry: go test ./... -short
        language: system
        pass_filenames: false
        types: [go]
```

Install hooks:

```bash
pre-commit install
```

**Using Git hooks directly:**

Create `.git/hooks/pre-push`:

```bash
#!/bin/bash
echo "Running unit tests..."
go test ./... -short
if [ $? -ne 0 ]; then
    echo "Tests failed. Push aborted."
    exit 1
fi
```

Make it executable:

```bash
chmod +x .git/hooks/pre-push
```
