//go:build retrytest

// Retry behaviour tests. They start a local HTTP server and rely on real
// wall-clock time, so they are excluded from the default test run.
//
// Run explicitly:
//
//	go test -tags retrytest -v ./internal/sdk/internal/utils/ -timeout 30s
//	make test-retry

package utils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/netskopeoss/terraform-provider-netskope/internal/sdk/retry"
)

// retryOnce returns an httptest.Server whose handler returns statusCode on the
// first request and 200 on every subsequent request. The caller must close it.
func retryOnce(statusCode int, headers map[string]string) *httptest.Server {
	calls := 0
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			for k, v := range headers {
				w.Header().Set(k, v)
			}
			w.WriteHeader(statusCode)
			return
		}
		w.WriteHeader(200)
	}))
}

// alwaysReturn returns an httptest.Server that always responds with statusCode.
func alwaysReturn(statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
	}))
}

func doGet(url string) (*http.Response, error) {
	return http.Get(url) //nolint:noctx
}

// TestRetryAfterCappedToMaxInterval verifies that a Retry-After header larger
// than MaxInterval is capped to MaxInterval. Without the cap a single 429 with
// Retry-After: 60 would stall the provider for a full minute regardless of the
// configured retry budget.
func TestRetryAfterCappedToMaxInterval(t *testing.T) {
	srv := retryOnce(429, map[string]string{"Retry-After": "60"})
	defer srv.Close()

	start := time.Now()
	resp, err := Retry(context.Background(), Retries{
		Config: &retry.Config{
			Strategy: "backoff",
			Backoff: &retry.BackoffStrategy{
				InitialInterval: 10,   // 10 ms
				MaxInterval:     50,   // 50 ms cap — well under 60 s
				Exponent:        1.5,
				MaxElapsedTime:  5000, // 5 s budget
			},
			RetryConnectionErrors: false,
		},
		StatusCodes: []string{"429"},
	}, func() (*http.Response, error) { return doGet(srv.URL) })
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	// If Retry-After were not capped the elapsed time would be ~60 s.
	// With the MaxInterval cap the retry sleeps at most 50 ms.
	const limit = 500 * time.Millisecond
	if elapsed > limit {
		t.Errorf("Retry-After was not capped to MaxInterval: elapsed %v (want <%v)", elapsed, limit)
	}
}

// TestBudgetCheckBeforeSleep verifies that when the next sleep interval would
// exceed the remaining budget the function returns immediately rather than
// sleeping into (and past) the deadline. Without this check the caller has no
// effective control over how long the provider blocks.
//
// Note: Retry returns the last HTTP response (not an error) when the budget is
// exhausted by a TemporaryError — the caller receives the 429 and can inspect
// the status code. The key assertion here is elapsed time.
func TestBudgetCheckBeforeSleep(t *testing.T) {
	srv := alwaysReturn(429)
	defer srv.Close()

	// Budget is 80 ms. First sleep would be 200 ms — exceeds budget, so the
	// function should return without sleeping at all.
	start := time.Now()
	resp, err := Retry(context.Background(), Retries{
		Config: &retry.Config{
			Strategy: "backoff",
			Backoff: &retry.BackoffStrategy{
				InitialInterval: 200, // 200 ms — larger than budget
				MaxInterval:     200,
				Exponent:        1.0,
				MaxElapsedTime:  80, // 80 ms budget
			},
			RetryConnectionErrors: false,
		},
		StatusCodes: []string{"429"},
	}, func() (*http.Response, error) { return doGet(srv.URL) })
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected transport error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected a response, got nil")
	}
	resp.Body.Close()
	if resp.StatusCode != 429 {
		t.Errorf("expected 429 (budget exhausted on rate limit), got %d", resp.StatusCode)
	}

	// Without the pre-sleep check the function would sleep 200 ms before
	// returning. With the fix it returns as soon as it sees the sleep would
	// overshoot the remaining budget.
	const limit = 150 * time.Millisecond
	if elapsed > limit {
		t.Errorf("budget check did not fire before sleep: elapsed %v (want <%v — function slept when it should have returned immediately)", elapsed, limit)
	}
}

// TestStrategyNoneSkipsRetryLoop verifies that Strategy:"none" passes the
// operation result through without entering the retry loop. This is the
// behaviour when retry_disabled=true in the provider config.
func TestStrategyNoneSkipsRetryLoop(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(429)
	}))
	defer srv.Close()

	resp, _ := Retry(context.Background(), Retries{
		Config:      &retry.Config{Strategy: "none"},
		StatusCodes: []string{"429"},
	}, func() (*http.Response, error) { return doGet(srv.URL) })
	if resp != nil {
		resp.Body.Close()
	}

	if calls != 1 {
		t.Errorf("strategy=none: expected exactly 1 call, got %d", calls)
	}
}

// TestRetrySucceedsBeforeBudgetExhausted verifies the happy-retry path: a
// transient 429 followed by 200 succeeds and returns the successful response.
func TestRetrySucceedsBeforeBudgetExhausted(t *testing.T) {
	srv := retryOnce(429, nil)
	defer srv.Close()

	resp, err := Retry(context.Background(), Retries{
		Config: &retry.Config{
			Strategy: "backoff",
			Backoff: &retry.BackoffStrategy{
				InitialInterval: 10,
				MaxInterval:     50,
				Exponent:        1.5,
				MaxElapsedTime:  5000,
			},
			RetryConnectionErrors: false,
		},
		StatusCodes: []string{"429"},
	}, func() (*http.Response, error) { return doGet(srv.URL) })

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200 after retry, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

// TestContextCancellationStopsRetry verifies that cancelling the context
// causes the retry loop to exit promptly rather than continuing to retry.
func TestContextCancellationStopsRetry(t *testing.T) {
	srv := alwaysReturn(429)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := Retry(ctx, Retries{
		Config: &retry.Config{
			Strategy: "backoff",
			Backoff: &retry.BackoffStrategy{
				InitialInterval: 200,
				MaxInterval:     200,
				Exponent:        1.0,
				MaxElapsedTime:  30000, // 30 s — much larger than context timeout
			},
			RetryConnectionErrors: false,
		},
		StatusCodes: []string{"429"},
	}, func() (*http.Response, error) { return doGet(srv.URL) })
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected error on context cancellation, got nil")
	}

	// Should exit within ~context timeout (80 ms), not the retry budget (30 s).
	const limit = 300 * time.Millisecond
	if elapsed > limit {
		t.Errorf("context cancellation did not stop retry: elapsed %v (want <%v)", elapsed, limit)
	}
}
