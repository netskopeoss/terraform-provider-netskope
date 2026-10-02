package provider

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/netskopeoss/terraform-provider-netskope/internal/sdk"
	sdkerrors "github.com/netskopeoss/terraform-provider-netskope/internal/sdk/models/errors"
	"github.com/netskopeoss/terraform-provider-netskope/internal/sdk/models/shared"
	"github.com/netskopeoss/terraform-provider-netskope/internal/sdk/retry"
)

type unreadableFailureBody struct{ calls int }

func (b *unreadableFailureBody) Read([]byte) (int, error) {
	b.calls++
	return 0, errors.New("secret response body must not be read")
}

// Exercise a real SDK error (including its wrapping and response handling)
// without credentials, Terraform logging, or a live Netskope tenant.
func TestAPIFailurePublisherHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/infrastructure/publishers" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("X-Netskope-Request-Id", "publisher-failure-123")
		w.Header().Set("Set-Cookie", "cookie-secret")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = io.WriteString(w, `{"token":"registration-token-secret","error":"body-secret"}`)
	}))
	defer server.Close()
	client := sdk.New(sdk.WithServerURL(server.URL), sdk.WithSecurity(shared.Security{APIKey: "api-token-secret"}), sdk.WithRetryConfig(retry.Config{Strategy: "none"}))
	res, err := client.NPAPublishers.ListObjects(context.Background())
	if err == nil {
		t.Fatal("expected failed publisher request")
	}
	got := apiErrorDetails(err)
	if res != nil && res.RawResponse != nil {
		got = debugResponse(res.RawResponse)
	}
	if want := `method="GET" status=503 request_id="publisher-failure-123"`; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestAPIFailureOAuthHTTP(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Netskope-Request-Id", "oauth-failure-123")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"error":"invalid_client","error_description":"client-secret"}`)
	}))
	defer server.Close()
	d := &PlatformOAuth2TokenDataSource{client: sdk.New(sdk.WithServerURL(server.URL))}
	_, err := d.fetchToken(context.Background(), "client-id", "client-secret", "client_credentials")
	if err == nil {
		t.Fatal("expected failed token request")
	}
	if got, want := apiErrorDetails(err), `method="POST" status=401 request_id="oauth-failure-123"`; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func (b *unreadableFailureBody) Close() error { b.calls++; return nil }

func TestAPIFailureResponse(t *testing.T) {
	body := &unreadableFailureBody{}
	req, err := http.NewRequest(http.MethodPost, "https://user:password@example.com/api?token=secret", strings.NewReader("request-body-secret"))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Netskope-Api-Token", "api-token-secret")
	req.Header.Set("Authorization", "Bearer authorization-secret")
	response := &http.Response{
		StatusCode: http.StatusServiceUnavailable,
		Status:     "503 secret status reason",
		Request:    req,
		Header: http.Header{
			"X-Netskope-Request-Id": {"request-123"},
			"Set-Cookie":            {"cookie-secret"},
			"Other-Header":          {"header-secret"},
		},
		Body: body,
	}
	got := debugResponse(response)
	if want := `method="POST" status=503 request_id="request-123"`; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if body.calls != 0 {
		t.Fatalf("response body was accessed %d times", body.calls)
	}
	if req.Header.Get("Netskope-Api-Token") != "api-token-secret" {
		t.Fatal("request header was mutated")
	}
	content, err := io.ReadAll(req.Body)
	if err != nil || string(content) != "request-body-secret" {
		t.Fatal("request body was changed")
	}
}

func TestAPIFailureMissingMetadata(t *testing.T) {
	for _, tc := range []struct {
		name     string
		response *http.Response
		want     string
	}{
		{"no response", nil, `method="unavailable" status=unavailable request_id="unavailable"`},
		{"no request or headers", &http.Response{StatusCode: 503}, `method="unavailable" status=503 request_id="unavailable"`},
		{"empty method", &http.Response{StatusCode: 500, Request: &http.Request{}}, `method="unavailable" status=500 request_id="unavailable"`},
		{"escaped request ID", &http.Response{StatusCode: 503, Header: http.Header{"X-Netskope-Request-Id": {"id\nforged=entry"}}}, `method="unavailable" status=503 request_id="id\nforged=entry"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := debugResponse(tc.response); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestAPIFailureErrors(t *testing.T) {
	response := &http.Response{StatusCode: 503, Request: &http.Request{Method: "GET"}, Header: http.Header{"X-Netskope-Request-Id": {"request-123"}}}
	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"wrapped SDK response", fmt.Errorf("wrapper-secret: %w", sdkerrors.NewSDKError("message-secret", 503, "body-secret", response)), `method="GET" status=503 request_id="request-123"`},
		{"SDK without response", sdkerrors.NewSDKError("message-secret", 502, "body-secret", nil), `method="unavailable" status=502 request_id="unavailable"`},
		{"hand-written HTTP operation", &apiFailureError{response: response}, `method="GET" status=503 request_id="request-123"`},
		{"wrapped transport error", fmt.Errorf("wrapper-secret: %w", &url.Error{Op: "Get", URL: "https://user:password@example.com?token=secret", Err: errors.New("transport-secret")}), `method="GET" status=unavailable request_id="unavailable"`},
		{"unknown error", errors.New("arbitrary-secret"), `method="unavailable" status=unavailable request_id="unavailable"`},
		{"nil error", nil, `method="unavailable" status=unavailable request_id="unavailable"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := apiErrorDetails(tc.err); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
