package hooks

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestOAuth2Hook_Disabled_WhenEnvVarsUnset(t *testing.T) {
	os.Unsetenv("NETSKOPE_OAUTH2_CLIENT_ID")
	os.Unsetenv("NETSKOPE_OAUTH2_CLIENT_SECRET")

	hook := &oauth2TokenHook{}
	hook.SDKInit("https://test.goskope.com/api/v2", http.DefaultClient)

	if hook.enabled() {
		t.Fatal("hook should be disabled when env vars are unset")
	}

	// BeforeRequest should pass through without modification
	req, _ := http.NewRequest("GET", "https://test.goskope.com/api/v2/test", nil)
	req.Header.Set("Netskope-Api-Token", "my-api-key")

	ctx := BeforeRequestContext{HookContext: HookContext{Context: context.Background()}}
	out, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Header.Get("Netskope-Api-Token") != "my-api-key" {
		t.Fatal("API key header should be preserved when hook is disabled")
	}
	if out.Header.Get("Authorization") != "" {
		t.Fatal("Authorization header should not be set when hook is disabled")
	}
}

func TestOAuth2Hook_FetchesToken(t *testing.T) {
	var tokenCalls atomic.Int32

	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCalls.Add(1)

		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected application/json content-type, got %s", ct)
		}

		var body tokenRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if body.ClientID != "test-client-id" {
			t.Errorf("expected client_id=test-client-id, got %s", body.ClientID)
		}
		if body.ClientSecret != "test-secret" {
			t.Errorf("expected client_secret=test-secret, got %s", body.ClientSecret)
		}
		if body.GrantType != "client_credentials" {
			t.Errorf("expected grant_type=client_credentials, got %s", body.GrantType)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "test-bearer-token",
			TokenType:   "Bearer",
			ExpiresIn:   3600,
		})
	}))
	defer tokenServer.Close()

	os.Setenv("NETSKOPE_OAUTH2_CLIENT_ID", "test-client-id")
	os.Setenv("NETSKOPE_OAUTH2_CLIENT_SECRET", "test-secret")
	defer os.Unsetenv("NETSKOPE_OAUTH2_CLIENT_ID")
	defer os.Unsetenv("NETSKOPE_OAUTH2_CLIENT_SECRET")

	hook := &oauth2TokenHook{}
	hook.SDKInit(tokenServer.URL, tokenServer.Client())

	// First request should fetch token
	req, _ := http.NewRequest("GET", tokenServer.URL+"/steering/apps/private", nil)
	req.Header.Set("Netskope-Api-Token", "my-api-key")

	ctx := BeforeRequestContext{HookContext: HookContext{Context: context.Background()}}
	out, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.Header.Get("Authorization") != "Bearer test-bearer-token" {
		t.Errorf("expected Bearer token, got %s", out.Header.Get("Authorization"))
	}
	if out.Header.Get("Netskope-Api-Token") != "" {
		t.Error("API key header should be removed when using OAuth2")
	}
	if tokenCalls.Load() != 1 {
		t.Errorf("expected 1 token call, got %d", tokenCalls.Load())
	}

	// Second request should use cached token
	req2, _ := http.NewRequest("GET", tokenServer.URL+"/steering/apps/private", nil)
	req2.Header.Set("Netskope-Api-Token", "my-api-key")
	out2, err := hook.BeforeRequest(ctx, req2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out2.Header.Get("Authorization") != "Bearer test-bearer-token" {
		t.Error("cached token should be used")
	}
	if tokenCalls.Load() != 1 {
		t.Errorf("expected still 1 token call (cached), got %d", tokenCalls.Load())
	}
}

func TestOAuth2Hook_ClearsTokenOn401(t *testing.T) {
	callCount := 0
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "token-v" + string(rune('0'+callCount)),
			TokenType:   "Bearer",
			ExpiresIn:   3600,
		})
	}))
	defer tokenServer.Close()

	os.Setenv("NETSKOPE_OAUTH2_CLIENT_ID", "client")
	os.Setenv("NETSKOPE_OAUTH2_CLIENT_SECRET", "secret")
	defer os.Unsetenv("NETSKOPE_OAUTH2_CLIENT_ID")
	defer os.Unsetenv("NETSKOPE_OAUTH2_CLIENT_SECRET")

	hook := &oauth2TokenHook{}
	hook.SDKInit(tokenServer.URL, tokenServer.Client())

	ctx := BeforeRequestContext{HookContext: HookContext{Context: context.Background()}}

	// Fetch initial token
	req, _ := http.NewRequest("GET", tokenServer.URL+"/test", nil)
	hook.BeforeRequest(ctx, req)

	if callCount != 1 {
		t.Fatalf("expected 1 token fetch, got %d", callCount)
	}

	// Simulate 401 response
	resp401 := &http.Response{StatusCode: http.StatusUnauthorized}
	errCtx := AfterErrorContext{HookContext: HookContext{Context: context.Background()}}
	hook.AfterError(errCtx, resp401, nil)

	// Next request should fetch a new token
	req2, _ := http.NewRequest("GET", tokenServer.URL+"/test", nil)
	hook.BeforeRequest(ctx, req2)

	if callCount != 2 {
		t.Errorf("expected 2 token fetches after 401 clear, got %d", callCount)
	}
}

func TestOAuth2Hook_TokenEndpointError(t *testing.T) {
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "invalid credentials"}`))
	}))
	defer tokenServer.Close()

	os.Setenv("NETSKOPE_OAUTH2_CLIENT_ID", "bad-client")
	os.Setenv("NETSKOPE_OAUTH2_CLIENT_SECRET", "bad-secret")
	defer os.Unsetenv("NETSKOPE_OAUTH2_CLIENT_ID")
	defer os.Unsetenv("NETSKOPE_OAUTH2_CLIENT_SECRET")

	hook := &oauth2TokenHook{}
	hook.SDKInit(tokenServer.URL, tokenServer.Client())

	req, _ := http.NewRequest("GET", tokenServer.URL+"/test", nil)
	ctx := BeforeRequestContext{HookContext: HookContext{Context: context.Background()}}

	_, err := hook.BeforeRequest(ctx, req)
	if err == nil {
		t.Fatal("expected error from failed token fetch")
	}
}

func TestOAuth2Hook_RefreshesNearExpiry(t *testing.T) {
	var tokenCalls atomic.Int32

	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCalls.Add(1)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "new-token",
			TokenType:   "Bearer",
			ExpiresIn:   3600,
		})
	}))
	defer tokenServer.Close()

	os.Setenv("NETSKOPE_OAUTH2_CLIENT_ID", "client")
	os.Setenv("NETSKOPE_OAUTH2_CLIENT_SECRET", "secret")
	defer os.Unsetenv("NETSKOPE_OAUTH2_CLIENT_ID")
	defer os.Unsetenv("NETSKOPE_OAUTH2_CLIENT_SECRET")

	hook := &oauth2TokenHook{}
	hook.SDKInit(tokenServer.URL, tokenServer.Client())

	// Inject a cached token that expires in 30 seconds (within the 60s refresh buffer).
	hook.token = "expiring-soon-token"
	hook.expiresAt = time.Now().Add(30 * time.Second)

	ctx := BeforeRequestContext{HookContext: HookContext{Context: context.Background()}}
	req, _ := http.NewRequest("GET", tokenServer.URL+"/test", nil)
	out, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tokenCalls.Load() != 1 {
		t.Errorf("expected token refresh for near-expiry token, got %d calls", tokenCalls.Load())
	}
	if out.Header.Get("Authorization") == "Bearer expiring-soon-token" {
		t.Error("near-expiry token should have been replaced")
	}
	if out.Header.Get("Authorization") != "Bearer new-token" {
		t.Errorf("expected Bearer new-token, got %s", out.Header.Get("Authorization"))
	}
}

func TestOAuth2Hook_ZeroExpiresIn_DefaultsTo3600(t *testing.T) {
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// expires_in: 0 — hook should default to 3600
		json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "token",
			TokenType:   "Bearer",
			ExpiresIn:   0,
		})
	}))
	defer tokenServer.Close()

	os.Setenv("NETSKOPE_OAUTH2_CLIENT_ID", "client")
	os.Setenv("NETSKOPE_OAUTH2_CLIENT_SECRET", "secret")
	defer os.Unsetenv("NETSKOPE_OAUTH2_CLIENT_ID")
	defer os.Unsetenv("NETSKOPE_OAUTH2_CLIENT_SECRET")

	hook := &oauth2TokenHook{}
	hook.SDKInit(tokenServer.URL, tokenServer.Client())

	ctx := BeforeRequestContext{HookContext: HookContext{Context: context.Background()}}
	req, _ := http.NewRequest("GET", tokenServer.URL+"/test", nil)
	if _, err := hook.BeforeRequest(ctx, req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// expiresAt should be approximately 3600 seconds from now.
	minExpected := time.Now().Add(3595 * time.Second)
	if hook.expiresAt.Before(minExpected) {
		t.Errorf("zero expires_in should default to 3600s; expiresAt=%v, want >= %v", hook.expiresAt, minExpected)
	}
}

func TestOAuth2Hook_ConcurrentFetch_OnlyOneTokenCall(t *testing.T) {
	var tokenCalls atomic.Int32

	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCalls.Add(1)
		// Small delay to widen the concurrency window.
		time.Sleep(20 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "shared-token",
			TokenType:   "Bearer",
			ExpiresIn:   3600,
		})
	}))
	defer tokenServer.Close()

	os.Setenv("NETSKOPE_OAUTH2_CLIENT_ID", "client")
	os.Setenv("NETSKOPE_OAUTH2_CLIENT_SECRET", "secret")
	defer os.Unsetenv("NETSKOPE_OAUTH2_CLIENT_ID")
	defer os.Unsetenv("NETSKOPE_OAUTH2_CLIENT_SECRET")

	hook := &oauth2TokenHook{}
	hook.SDKInit(tokenServer.URL, tokenServer.Client())

	const goroutines = 10
	errs := make([]error, goroutines)
	var wg sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			ctx := BeforeRequestContext{HookContext: HookContext{Context: context.Background()}}
			req, _ := http.NewRequest("GET", tokenServer.URL+"/test", nil)
			_, errs[idx] = hook.BeforeRequest(ctx, req)
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("goroutine %d: unexpected error: %v", i, err)
		}
	}
	if tokenCalls.Load() != 1 {
		t.Errorf("expected exactly 1 token fetch for %d concurrent requests, got %d", goroutines, tokenCalls.Load())
	}
}

func TestOAuth2Hook_TokenURL_TrailingSlash(t *testing.T) {
	var requestedPath string

	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "token",
			TokenType:   "Bearer",
			ExpiresIn:   3600,
		})
	}))
	defer tokenServer.Close()

	os.Setenv("NETSKOPE_OAUTH2_CLIENT_ID", "client")
	os.Setenv("NETSKOPE_OAUTH2_CLIENT_SECRET", "secret")
	defer os.Unsetenv("NETSKOPE_OAUTH2_CLIENT_ID")
	defer os.Unsetenv("NETSKOPE_OAUTH2_CLIENT_SECRET")

	hook := &oauth2TokenHook{}
	// baseURL has a trailing slash — should not produce double slash in token URL.
	hook.SDKInit(tokenServer.URL+"/", tokenServer.Client())

	ctx := BeforeRequestContext{HookContext: HookContext{Context: context.Background()}}
	req, _ := http.NewRequest("GET", tokenServer.URL+"/test", nil)
	if _, err := hook.BeforeRequest(ctx, req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if requestedPath != "/platform/oauth2/token" {
		t.Errorf("expected token path /platform/oauth2/token, got %s", requestedPath)
	}
}
