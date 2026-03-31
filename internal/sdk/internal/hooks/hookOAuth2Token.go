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

// oauth2TokenHook implements OAuth2 Client Credentials authentication for
// Netskope's non-standard token endpoint.
//
// Credentials are read from environment variables:
//   - NETSKOPE_OAUTH2_CLIENT_ID
//   - NETSKOPE_OAUTH2_CLIENT_SECRET
//
// When both are set, this hook fetches a bearer token from
// POST /api/v2/platform/oauth2/token/generate and sets it on every request,
// taking priority over the API key header.
//
// The Netskope token endpoint expects a JSON body:
//
//	{"clientID": "...", "secretKey": "..."}
//
// And returns:
//
//	{"oAuth2AccessToken": "...", "expiryTime": "2024-11-04T14:24:10.000Z"}
type oauth2TokenHook struct {
	mu        sync.Mutex
	baseURL   string
	client    HTTPClient
	clientID  string
	secretKey string
	token     string
	expiresAt time.Time
}

var (
	_ sdkInitHook       = (*oauth2TokenHook)(nil)
	_ beforeRequestHook = (*oauth2TokenHook)(nil)
	_ afterErrorHook    = (*oauth2TokenHook)(nil)
)

const tokenPath = "/platform/oauth2/token/generate"

// tokenRequest is the JSON body Netskope's token endpoint expects.
type tokenRequest struct {
	ClientID  string `json:"clientID"`
	SecretKey string `json:"secretKey"`
}

// tokenResponse is the JSON body Netskope's token endpoint returns.
type tokenResponse struct {
	OAuth2AccessToken string `json:"oAuth2AccessToken"`
	ExpiryTime        string `json:"expiryTime"`
}

func (h *oauth2TokenHook) SDKInit(baseURL string, client HTTPClient) (string, HTTPClient) {
	h.baseURL = baseURL
	h.client = client
	h.clientID = os.Getenv("NETSKOPE_OAUTH2_CLIENT_ID")
	h.secretKey = os.Getenv("NETSKOPE_OAUTH2_CLIENT_SECRET")
	return baseURL, client
}

func (h *oauth2TokenHook) enabled() bool {
	return h.clientID != "" && h.secretKey != ""
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
		ClientID:  h.clientID,
		SecretKey: h.secretKey,
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

	if tokenResp.OAuth2AccessToken == "" {
		return "", time.Time{}, fmt.Errorf("token response missing oAuth2AccessToken")
	}

	// Parse the expiry time. If parsing fails, default to 1 hour from now.
	expiresAt, err := time.Parse(time.RFC3339Nano, tokenResp.ExpiryTime)
	if err != nil {
		expiresAt = time.Now().Add(1 * time.Hour)
	}

	return tokenResp.OAuth2AccessToken, expiresAt, nil
}
