package hooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// oauth2TokenHook implements OAuth2 Client Credentials authentication (RFC 6749)
// for the Netskope API.
//
// Credentials are read from environment variables:
//   - NETSKOPE_OAUTH2_CLIENT_ID
//   - NETSKOPE_OAUTH2_CLIENT_SECRET
//
// When both are set, this hook fetches a bearer token from
// POST /api/v2/platform/oauth2/token and sets it on every request,
// taking priority over the API key header.
//
// The endpoint expects a JSON body:
//
//	{"client_id": "...", "client_secret": "...", "grant_type": "client_credentials"}
//
// And returns:
//
//	{"access_token": "...", "token_type": "Bearer", "expires_in": 3600}
type oauth2TokenHook struct {
	mu           sync.Mutex
	baseURL      string
	client       HTTPClient
	clientID     string
	clientSecret string
	token        string
	expiresAt    time.Time
}

var (
	_ sdkInitHook       = (*oauth2TokenHook)(nil)
	_ beforeRequestHook = (*oauth2TokenHook)(nil)
	_ afterErrorHook    = (*oauth2TokenHook)(nil)
)

const tokenPath = "/platform/oauth2/token"

// tokenRequest is the JSON body the RFC 6749 token endpoint expects.
type tokenRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	GrantType    string `json:"grant_type"`
}

// tokenResponse is the JSON body the RFC 6749 token endpoint returns.
type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func (h *oauth2TokenHook) SDKInit(baseURL string, client HTTPClient) (string, HTTPClient) {
	h.baseURL = baseURL
	h.client = client
	// Only activate OAuth2 when no API key is configured. API key takes precedence
	// so that setting both in the environment does not silently break API-key-based
	// workflows (e.g. acceptance tests that source .env for both).
	if os.Getenv("NETSKOPE_API_KEY") == "" {
		h.clientID = os.Getenv("NETSKOPE_OAUTH2_CLIENT_ID")
		h.clientSecret = os.Getenv("NETSKOPE_OAUTH2_CLIENT_SECRET")
	}
	return baseURL, client
}

func (h *oauth2TokenHook) enabled() bool {
	return h.clientID != "" && h.clientSecret != ""
}

func (h *oauth2TokenHook) BeforeRequest(hookCtx BeforeRequestContext, req *http.Request) (*http.Request, error) {
	if !h.enabled() {
		return req, nil
	}

	token, err := h.getToken(hookCtx)
	if err != nil {
		return nil, &FailEarly{Cause: fmt.Errorf("oauth2 token fetch failed: %w", err)}
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Del("Netskope-Api-Token")
	return req, nil
}

func (h *oauth2TokenHook) AfterError(hookCtx AfterErrorContext, res *http.Response, err error) (*http.Response, error) {
	if !h.enabled() {
		return res, err
	}

	// On 401, clear the cached token so the next request fetches a new one.
	if res != nil && res.StatusCode == http.StatusUnauthorized {
		h.mu.Lock()
		h.token = ""
		h.expiresAt = time.Time{}
		h.mu.Unlock()
	}

	return res, err
}

func (h *oauth2TokenHook) getToken(hookCtx BeforeRequestContext) (string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Return cached token if still valid (with 60s buffer before expiry).
	if h.token != "" && time.Now().Add(60*time.Second).Before(h.expiresAt) {
		return h.token, nil
	}

	token, expiresAt, err := h.fetchToken(hookCtx)
	if err != nil {
		return "", err
	}

	h.token = token
	h.expiresAt = expiresAt
	return token, nil
}

func (h *oauth2TokenHook) fetchToken(hookCtx BeforeRequestContext) (string, time.Time, error) {
	tokenURL := strings.TrimSuffix(h.baseURL, "/") + tokenPath

	body, err := json.Marshal(tokenRequest{
		ClientID:     h.clientID,
		ClientSecret: h.clientSecret,
		GrantType:    "client_credentials",
	})
	if err != nil {
		return "", time.Time{}, fmt.Errorf("marshal token request: %w", err)
	}

	req, err := http.NewRequestWithContext(hookCtx.Context, http.MethodPost, tokenURL, bytes.NewReader(body))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("send token request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("read token response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", time.Time{}, fmt.Errorf("token endpoint returned %d: %s", resp.StatusCode, respBody)
	}

	var tokenResp tokenResponse
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return "", time.Time{}, fmt.Errorf("decode token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return "", time.Time{}, fmt.Errorf("token response missing access_token")
	}

	// Calculate expiry from expires_in seconds. Default to 1 hour if zero.
	expiresIn := tokenResp.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 3600
	}
	expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Second)

	return tokenResp.AccessToken, expiresAt, nil
}
