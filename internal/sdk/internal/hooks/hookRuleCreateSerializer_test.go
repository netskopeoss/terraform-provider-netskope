package hooks

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestRuleCreateSerializer_SerializesConcurrentCreates verifies that
// concurrent createNPARules requests are serialized (no two run in parallel).
func TestRuleCreateSerializer_SerializesConcurrentCreates(t *testing.T) {
	hook := &ruleCreateSerializer{}

	const numGoroutines = 5
	var mu sync.Mutex
	var active int
	var maxActive int
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			ctx := BeforeRequestContext{
				HookContext: HookContext{OperationID: "createNPARules"},
			}
			req, _ := http.NewRequest("POST", "https://example.com/policy/npa/rules",
				strings.NewReader(fmt.Sprintf(`{"rule_name":"rule-%d"}`, id)))

			// Acquire the lock (BeforeRequest)
			_, err := hook.BeforeRequest(ctx, req)
			if err != nil {
				t.Errorf("BeforeRequest failed: %v", err)
				return
			}

			// Track concurrency — only one goroutine should be active at a time
			mu.Lock()
			active++
			if active > maxActive {
				maxActive = active
			}
			if active > 1 {
				t.Errorf("expected max 1 concurrent create, got %d active", active)
			}
			mu.Unlock()

			// Simulate API call duration
			// (no sleep needed — goroutine scheduling is enough to detect races)

			mu.Lock()
			active--
			mu.Unlock()

			// Release the lock (AfterSuccess)
			successCtx := AfterSuccessContext{
				HookContext: HookContext{OperationID: "createNPARules"},
			}
			res := &http.Response{StatusCode: 200}
			_, err = hook.AfterSuccess(successCtx, res)
			if err != nil {
				t.Errorf("AfterSuccess failed: %v", err)
			}
		}(i)
	}

	wg.Wait()

	if maxActive != 1 {
		t.Errorf("expected max 1 concurrent create, got %d", maxActive)
	}
}

// TestRuleCreateSerializer_AfterErrorUnlocks verifies that the mutex is
// released when AfterError is called (e.g., on HTTP failure).
func TestRuleCreateSerializer_AfterErrorUnlocks(t *testing.T) {
	hook := &ruleCreateSerializer{}

	ctx := BeforeRequestContext{
		HookContext: HookContext{OperationID: "createNPARules"},
	}
	req, _ := http.NewRequest("POST", "https://example.com/policy/npa/rules",
		strings.NewReader(`{"rule_name":"rule-1"}`))

	// Acquire
	_, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}

	// Release via error path
	errorCtx := AfterErrorContext{
		HookContext: HookContext{OperationID: "createNPARules"},
	}
	_, err = hook.AfterError(errorCtx, nil, fmt.Errorf("connection refused"))
	if err == nil {
		t.Fatal("expected error to be passed through")
	}

	// Verify we can acquire again (not deadlocked)
	done := make(chan bool, 1)
	go func() {
		_, _ = hook.BeforeRequest(ctx, req)
		done <- true
		// Clean up
		successCtx := AfterSuccessContext{
			HookContext: HookContext{OperationID: "createNPARules"},
		}
		hook.AfterSuccess(successCtx, &http.Response{StatusCode: 200})
	}()

	select {
	case <-done:
		// OK — mutex was released by AfterError
	case <-time.After(2 * time.Second):
		t.Fatal("mutex was not released by AfterError — deadlocked")
	}
}

// TestRuleCreateSerializer_NonCreatePassthrough verifies that non-create
// operations are not serialized (mutex is not acquired).
func TestRuleCreateSerializer_NonCreatePassthrough(t *testing.T) {
	hook := &ruleCreateSerializer{}

	operations := []string{"updateNPARules", "deleteNPARules", "readNPARules", "createNPAPrivateApp"}

	for _, op := range operations {
		ctx := BeforeRequestContext{
			HookContext: HookContext{OperationID: op},
		}
		req, _ := http.NewRequest("POST", "https://example.com/test",
			strings.NewReader(`{}`))

		_, err := hook.BeforeRequest(ctx, req)
		if err != nil {
			t.Errorf("BeforeRequest failed for %s: %v", op, err)
		}

		// These should NOT call unlock since lock was never acquired
		// Verify no panic on AfterSuccess for non-create operations
		successCtx := AfterSuccessContext{
			HookContext: HookContext{OperationID: op},
		}
		_, err = hook.AfterSuccess(successCtx, &http.Response{StatusCode: 200})
		if err != nil {
			t.Errorf("AfterSuccess failed for %s: %v", op, err)
		}
	}
}