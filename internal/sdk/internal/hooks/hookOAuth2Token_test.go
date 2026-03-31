package hooks

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
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
			t.Errorf("expected clientID=test-client-id, got %s", body.ClientID)
		}
		if body.SecretKey != "test-secret" {
			t.Errorf("expected secretKey=test-secret, got %s", body.SecretKey)
		}

		expiry := time.Now().Add(1 * time.Hour).Format(time.RFC3339Nano)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(tokenResponse{
			OAuth2AccessToken: "test-bearer-token",
			ExpiryTime:        expiry,
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
		expiry := time.Now().Add(1 * time.Hour).Format(time.RFC3339Nano)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(tokenResponse{
			OAuth2AccessToken: "token-v" + string(rune('0'+callCount)),
			ExpiryTime:        expiry,
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
