# Unit Testing Speakeasy Hooks

This guide explains how to write unit tests for SDK hooks in a Speakeasy-generated Terraform provider. It covers Go testing fundamentals, the hook architecture, and patterns used in this codebase.

## Table of Contents

1. [Go Unit Test Fundamentals](#go-unit-test-fundamentals)
2. [Hook Architecture](#hook-architecture)
3. [Test File Structure](#test-file-structure)
4. [Writing a Hook Test](#writing-a-hook-test)
5. [Helper Functions](#helper-functions)
6. [Common Patterns](#common-patterns)
7. [Running Tests](#running-tests)

---

## Go Unit Test Fundamentals

### File Naming

Test files must end with `_test.go`:

```
hookMyAppAfterSuccess.go       # Production code
hookPublisherSort_test.go      # Test code
```

Go compiles `_test.go` files only when running tests, not in production builds.

### Test Function Signature

Every test function must:
- Start with `Test` (capital T)
- Accept exactly one parameter: `t *testing.T`

```go
func TestPublishersSortedByID(t *testing.T) {
    // test code
}
```

### The `*testing.T` Parameter

The `t` parameter provides methods for test control:

| Method | Purpose |
|--------|---------|
| `t.Errorf("msg")` | Report failure, continue running |
| `t.Fatalf("msg")` | Report failure, stop this test |
| `t.Run("name", func(t *testing.T){...})` | Run a subtest |
| `t.Helper()` | Mark function as helper (cleaner stack traces) |
| `t.Skip("reason")` | Skip this test |

### Test Structure: Arrange-Act-Assert

```go
func TestSomething(t *testing.T) {
    // ARRANGE: Set up test data
    input := createTestData()

    // ACT: Call the code under test
    result, err := FunctionUnderTest(input)

    // ASSERT: Verify the results
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result != expected {
        t.Errorf("expected %v, got %v", expected, result)
    }
}
```

---

## Hook Architecture

### Why Structs Instead of Functions?

Hooks are implemented as structs with methods rather than plain functions because Speakeasy uses an **interface-based hook system**.

```go
// The SDK defines this interface
type afterSuccessHook interface {
    AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error)
}

// Our hook implements the interface
type myAppResponse struct {
    Data   models.AppData `json:"data"`
    Status string         `json:"status"`
}

func (i *myAppResponse) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
    // Hook logic here
}
```

### Interface Verification

The codebase uses compile-time interface verification:

```go
var (
    _ afterSuccessHook = (*myAppResponse)(nil)
)
```

This line ensures `myAppResponse` implements `afterSuccessHook` at compile time. If the method signature is wrong, the build fails.

### Hook Registration

Hooks are registered in `registration.go` and called by the SDK:

```go
// Simplified example of how the SDK calls hooks
for _, hook := range registeredAfterSuccessHooks {
    res, err = hook.AfterSuccess(ctx, res)
}
```

Each hook checks the `OperationID` to decide if it should process the response:

```go
func (i *myAppResponse) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
    if hookCtx.OperationID == "getNPAPrivateApp" ||
       hookCtx.OperationID == "createNPAPrivateApps" ||
       hookCtx.OperationID == "updateNPAPrivateApp" {
        // Process this response
    }
    return res, nil  // Pass through unchanged for other operations
}
```

### Hook Types

| Interface | When Called | Use Case |
|-----------|-------------|----------|
| `beforeRequestHook` | Before HTTP request sent | Modify request body, add headers |
| `afterSuccessHook` | After successful HTTP response | Transform response data |
| `afterErrorHook` | After HTTP error | Custom error handling |

---

## Test File Structure

A typical hook test file has these sections:

```go
package hooks

import (
    "encoding/json"
    "io"
    "net/http"
    "strings"
    "testing"
)

// ============================================================================
// Helper Functions
// ============================================================================

func buildFakeAppResponse(t *testing.T, publishers []map[string]interface{}) *http.Response {
    // Creates mock HTTP response
}

func parsePublishersFromResponse(t *testing.T, res *http.Response) []map[string]interface{} {
    // Extracts data from response for assertions
}

// ============================================================================
// Tests
// ============================================================================

func TestPublishersSortedByID(t *testing.T) {
    // Test implementation
}
```

---

## Writing a Hook Test

### Step 1: Create the Hook Instance

```go
func TestPublishersSortedByID(t *testing.T) {
    hook := &myAppResponse{}  // Create instance of hook to test
```

### Step 2: Prepare Test Data

Create data that simulates what the API returns:

```go
    // Simulate API returning publishers in wrong order
    publishers := []map[string]interface{}{
        {"publisher_id": float64(256), "publisher_name": "Publisher-B"},
        {"publisher_id": float64(249), "publisher_name": "Publisher-A"},
        {"publisher_id": float64(300), "publisher_name": "Publisher-C"},
    }
```

**Note:** JSON numbers unmarshal to `float64` in Go, not `int`.

### Step 3: Build Mock HTTP Response

```go
    res := buildFakeAppResponse(t, publishers)
```

The helper creates an `*http.Response` with a JSON body matching the API structure.

### Step 4: Create Hook Context

```go
    ctx := hookCtxForOp("getNPAPrivateApp")
```

The context tells the hook which API operation this is.

### Step 5: Call the Hook

```go
    res, err := hook.AfterSuccess(ctx, res)
    if err != nil {
        t.Fatalf("AfterSuccess returned error: %v", err)
    }
```

### Step 6: Parse and Assert Results

```go
    result := parsePublishersFromResponse(t, res)

    expectedOrder := []string{"249", "256", "300"}
    for i, expected := range expectedOrder {
        got := result[i]["publisher_id"].(string)
        if got != expected {
            t.Errorf("publisher[%d]: expected %s, got %s", i, expected, got)
        }
    }
}
```

---

## Helper Functions

### Building Mock HTTP Responses

```go
func buildFakeAppResponse(t *testing.T, publishers []map[string]interface{}) *http.Response {
    t.Helper()  // Marks this as a helper for better error reporting

    // Build JSON structure matching API response
    body := map[string]interface{}{
        "status": "success",
        "data": map[string]interface{}{
            "app_name":                      "test-app",
            "host":                          "test.example.com",
            "protocols":                     []map[string]interface{}{},
            "service_publisher_assignments": publishers,
            "tags":                          []map[string]interface{}{},
        },
    }

    // Convert to JSON bytes
    b, err := json.Marshal(body)
    if err != nil {
        t.Fatalf("failed to marshal test response: %v", err)
    }

    // Create fake HTTP response
    return &http.Response{
        StatusCode: 200,
        Body:       io.NopCloser(strings.NewReader(string(b))),
        Header:     make(http.Header),
    }
}
```

**Key points:**
- `t.Helper()` makes error messages point to the test, not this function
- `io.NopCloser` wraps a string reader to satisfy `io.ReadCloser` interface
- The JSON structure must match what the real API returns

### Parsing Response Data

```go
func parsePublishersFromResponse(t *testing.T, res *http.Response) []map[string]interface{} {
    t.Helper()

    // Read response body
    body, err := io.ReadAll(res.Body)
    if err != nil {
        t.Fatalf("failed to read response body: %v", err)
    }

    // Parse JSON
    var parsed map[string]interface{}
    if err := json.Unmarshal(body, &parsed); err != nil {
        t.Fatalf("failed to unmarshal response: %v", err)
    }

    // Navigate to the publishers array
    data := parsed["data"].(map[string]interface{})
    raw := data["service_publisher_assignments"].([]interface{})

    // Convert to typed slice
    var result []map[string]interface{}
    for _, r := range raw {
        result = append(result, r.(map[string]interface{}))
    }

    return result
}
```

### Creating Hook Context

```go
func hookCtxForOp(operationID string) AfterSuccessContext {
    return AfterSuccessContext{
        HookContext: HookContext{
            OperationID: operationID,
        },
    }
}
```

---

## Common Patterns

### Subtests for Multiple Scenarios

Use `t.Run()` to test multiple cases with shared logic:

```go
func TestPublishersSortedByID(t *testing.T) {
    hook := &myAppResponse{}
    publishers := []map[string]interface{}{...}

    // Test all three operation IDs
    for _, opID := range []string{"getNPAPrivateApp", "createNPAPrivateApps", "updateNPAPrivateApp"} {
        t.Run(opID, func(t *testing.T) {
            res := buildFakeAppResponse(t, publishers)
            ctx := hookCtxForOp(opID)

            res, err := hook.AfterSuccess(ctx, res)
            // assertions...
        })
    }
}
```

Output:
```
=== RUN   TestPublishersSortedByID
=== RUN   TestPublishersSortedByID/getNPAPrivateApp
=== RUN   TestPublishersSortedByID/createNPAPrivateApps
=== RUN   TestPublishersSortedByID/updateNPAPrivateApp
```

### Table-Driven Tests

For testing many input/output combinations:

```go
func TestIsNotFoundError(t *testing.T) {
    tests := []struct {
        name     string
        message  string
        expected bool
    }{
        {"exact match", "not found", true},
        {"case insensitive", "NOT FOUND", true},
        {"partial match", "No private app with id 123", true},
        {"no match", "some other error", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := isNotFoundError(tt.message)
            if got != tt.expected {
                t.Errorf("isNotFoundError(%q) = %v, want %v", tt.message, got, tt.expected)
            }
        })
    }
}
```

### Idempotency Tests

Verify that running the hook twice produces identical results:

```go
func TestPublishersIdempotent(t *testing.T) {
    hook := &myAppResponse{}
    publishers := []map[string]interface{}{...}
    ctx := hookCtxForOp("getNPAPrivateApp")

    // First pass
    res1 := buildFakeAppResponse(t, publishers)
    res1, _ = hook.AfterSuccess(ctx, res1)
    result1 := parsePublishersFromResponse(t, res1)

    // Second pass with same input
    res2 := buildFakeAppResponse(t, publishers)
    res2, _ = hook.AfterSuccess(ctx, res2)
    result2 := parsePublishersFromResponse(t, res2)

    // Compare results
    for i := range result1 {
        if result1[i]["publisher_id"] != result2[i]["publisher_id"] {
            t.Errorf("not idempotent at index %d", i)
        }
    }
}
```

### Edge Case Tests

Always test edge cases:

```go
func TestEmptyListsDoNotPanic(t *testing.T) {
    hook := &myAppResponse{}

    // Empty lists
    res := buildFakeAppResponseFull(t,
        []map[string]interface{}{},  // empty publishers
        []map[string]interface{}{},  // empty protocols
        []map[string]interface{}{},  // empty tags
    )

    res, err := hook.AfterSuccess(hookCtxForOp("getNPAPrivateApp"), res)
    if err != nil {
        t.Fatalf("hook failed on empty lists: %v", err)
    }
}

func TestNilPublisherIdHandling(t *testing.T) {
    hook := &myAppResponse{}

    publishers := []map[string]interface{}{
        {"publisher_id": nil, "publisher_name": "Nil-ID"},
        {"publisher_id": float64(100), "publisher_name": "Valid"},
    }

    res := buildFakeAppResponse(t, publishers)
    res, err := hook.AfterSuccess(hookCtxForOp("getNPAPrivateApp"), res)
    if err != nil {
        t.Fatalf("hook failed with nil publisher_id: %v", err)
    }
}
```

---

## Running Tests

### Run All Hook Tests

```bash
go test -v ./internal/sdk/internal/hooks/
```

### Run Specific Test

```bash
go test -v ./internal/sdk/internal/hooks/ -run TestPublishersSortedByID
```

### Run Tests Matching Pattern

```bash
go test -v ./internal/sdk/internal/hooks/ -run "TestEmpty|TestNil"
```

### Run Subtests

```bash
# Run only the getNPAPrivateApp subtest
go test -v ./internal/sdk/internal/hooks/ -run TestPublishersSortedByID/getNPAPrivateApp
```

### Show Coverage

```bash
go test -v ./internal/sdk/internal/hooks/ -cover
```

### Generate Coverage Report

```bash
go test ./internal/sdk/internal/hooks/ -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## Checklist for New Hook Tests

When adding tests for a new hook:

- [ ] Create test file named `hook<Name>_test.go`
- [ ] Add helper functions to build mock requests/responses
- [ ] Test all operation IDs the hook handles
- [ ] Test the passthrough case (non-matching operation ID)
- [ ] Test edge cases (empty lists, nil values, missing fields)
- [ ] Test idempotency (running twice produces same result)
- [ ] Test error handling paths
- [ ] Run tests and verify all pass

---

## Reference Files

| File | Purpose |
|------|---------|
| `hookPublisherSort_test.go` | Example tests for AfterSuccess hooks |
| `hookMyAppAfterSuccess.go` | Single-app response hook |
| `hookMyBulkAppAfterSuccess.go` | Bulk-app response hook |
| `hookErrorStatusResponse.go` | Error handling hook (needs tests) |
| `hookPrivateAppRequest.go` | BeforeRequest hook (needs tests) |
