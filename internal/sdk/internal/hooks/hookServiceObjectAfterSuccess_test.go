package hooks

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// --- BeforeRequest tests ---

func TestServiceObjectHook_BeforeRequest_CapsLimitAt150(t *testing.T) {
	hook := &serviceObjectHook{}

	for _, tc := range []struct {
		name        string
		queryString string
		wantLimit   string
	}{
		{"empty limit", "", "150"},
		{"default SDK limit 500", "limit=500", "150"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			u, _ := url.Parse("https://example.com/api/v2/profiles/serviceobjects?" + tc.queryString)
			req := &http.Request{URL: u, Header: make(http.Header)}
			hookCtx := BeforeRequestContext{HookContext: HookContext{OperationID: "listServiceObjects"}}

			out, err := hook.BeforeRequest(hookCtx, req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := out.URL.Query().Get("limit"); got != tc.wantLimit {
				t.Errorf("expected limit=%s, got %s", tc.wantLimit, got)
			}
		})
	}
}

func TestServiceObjectHook_BeforeRequest_PreservesCustomLimit(t *testing.T) {
	hook := &serviceObjectHook{}

	u, _ := url.Parse("https://example.com/api/v2/profiles/serviceobjects?limit=50")
	req := &http.Request{URL: u, Header: make(http.Header)}
	hookCtx := BeforeRequestContext{HookContext: HookContext{OperationID: "listServiceObjects"}}

	out, err := hook.BeforeRequest(hookCtx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// A user-supplied limit that isn't 500 should pass through unchanged
	if got := out.URL.Query().Get("limit"); got != "50" {
		t.Errorf("expected limit=50, got %s", got)
	}
}

func TestServiceObjectHook_BeforeRequest_PassthroughForOtherOps(t *testing.T) {
	hook := &serviceObjectHook{}

	u, _ := url.Parse("https://example.com/api/v2/profiles/serviceobjects/42")
	req := &http.Request{URL: u, Header: make(http.Header)}
	hookCtx := BeforeRequestContext{HookContext: HookContext{OperationID: "getServiceObject"}}

	out, err := hook.BeforeRequest(hookCtx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != req {
		t.Error("expected same request pointer for non-list operation")
	}
}

// --- AfterSuccess tests ---

func TestServiceObjectHook_AfterSuccess_NormalizesPortIntegers(t *testing.T) {
	hook := &serviceObjectHook{}

	body := `{
		"id": "42",
		"name": "my-obj",
		"status": "applied",
		"type": "custom",
		"protocols": {
			"tcp": [80, 443],
			"udp": [53]
		}
	}`

	res := makeServiceObjectResponse(200, body)
	hookCtx := AfterSuccessContext{HookContext: HookContext{OperationID: "getServiceObject"}}

	result, err := hook.AfterSuccess(hookCtx, res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Protocols struct {
			TCP []interface{} `json:"tcp"`
			UDP []interface{} `json:"udp"`
		} `json:"protocols"`
	}
	readServiceObjectResponse(t, result, &parsed)

	if len(parsed.Protocols.TCP) != 2 {
		t.Fatalf("expected 2 TCP ports, got %d", len(parsed.Protocols.TCP))
	}
	for i, p := range parsed.Protocols.TCP {
		if _, ok := p.(string); !ok {
			t.Errorf("TCP[%d] should be string after normalization, got %T: %v", i, p, p)
		}
	}
	if parsed.Protocols.TCP[0] != "80" {
		t.Errorf("expected TCP[0]=80, got %v", parsed.Protocols.TCP[0])
	}
	if parsed.Protocols.TCP[1] != "443" {
		t.Errorf("expected TCP[1]=443, got %v", parsed.Protocols.TCP[1])
	}
	if len(parsed.Protocols.UDP) != 1 || parsed.Protocols.UDP[0] != "53" {
		t.Errorf("expected UDP=[53], got %v", parsed.Protocols.UDP)
	}
}

func TestServiceObjectHook_AfterSuccess_AlreadyStringPorts(t *testing.T) {
	hook := &serviceObjectHook{}

	body := `{
		"id": "42",
		"name": "my-obj",
		"status": "applied",
		"type": "custom",
		"protocols": {
			"tcp": ["80", "443"]
		}
	}`

	res := makeServiceObjectResponse(200, body)
	hookCtx := AfterSuccessContext{HookContext: HookContext{OperationID: "getServiceObject"}}

	result, err := hook.AfterSuccess(hookCtx, res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Protocols struct {
			TCP []string `json:"tcp"`
		} `json:"protocols"`
	}
	readServiceObjectResponse(t, result, &parsed)

	if len(parsed.Protocols.TCP) != 2 {
		t.Fatalf("expected 2 TCP ports, got %d", len(parsed.Protocols.TCP))
	}
	if parsed.Protocols.TCP[0] != "80" || parsed.Protocols.TCP[1] != "443" {
		t.Errorf("port values changed unexpectedly: %v", parsed.Protocols.TCP)
	}
}

func TestServiceObjectHook_AfterSuccess_NormalizesStatusToLowercase(t *testing.T) {
	hook := &serviceObjectHook{}

	for _, tc := range []struct {
		input string
		want  string
	}{
		{"APPLIED", "applied"},
		{"PENDING-CREATE", "pending-create"},
		{"applied", "applied"},
	} {
		t.Run(tc.input, func(t *testing.T) {
			body := `{"id":"1","name":"x","status":"` + tc.input + `","type":"custom","protocols":{"tcp":["80"]}}`
			res := makeServiceObjectResponse(200, body)
			hookCtx := AfterSuccessContext{HookContext: HookContext{OperationID: "getServiceObject"}}

			result, err := hook.AfterSuccess(hookCtx, res)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var parsed struct {
				Status string `json:"status"`
			}
			readServiceObjectResponse(t, result, &parsed)

			if parsed.Status != tc.want {
				t.Errorf("expected status=%q, got %q", tc.want, parsed.Status)
			}
		})
	}
}

func TestServiceObjectHook_AfterSuccess_NormalizesTypeField(t *testing.T) {
	hook := &serviceObjectHook{}

	for _, tc := range []struct {
		input string
		want  string
	}{
		{"CUSTOM", "custom"},
		{"custom", "custom"},
		{"PREDEFINED", "PREDEFINED"},
		{"predefined", "PREDEFINED"},
	} {
		t.Run(tc.input, func(t *testing.T) {
			body := `{"id":"1","name":"x","status":"applied","type":"` + tc.input + `","protocols":{"tcp":["80"]}}`
			res := makeServiceObjectResponse(200, body)
			hookCtx := AfterSuccessContext{HookContext: HookContext{OperationID: "getServiceObject"}}

			result, err := hook.AfterSuccess(hookCtx, res)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var parsed struct {
				Type string `json:"type"`
			}
			readServiceObjectResponse(t, result, &parsed)

			if parsed.Type != tc.want {
				t.Errorf("input=%q: expected type=%q, got %q", tc.input, tc.want, parsed.Type)
			}
		})
	}
}

func TestServiceObjectHook_AfterSuccess_NormalizesIntegerID(t *testing.T) {
	hook := &serviceObjectHook{}

	body := `{"id":42,"name":"x","status":"applied","type":"custom","protocols":{"tcp":["80"]}}`
	res := makeServiceObjectResponse(200, body)
	hookCtx := AfterSuccessContext{HookContext: HookContext{OperationID: "getServiceObject"}}

	result, err := hook.AfterSuccess(hookCtx, res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		ID interface{} `json:"id"`
	}
	readServiceObjectResponse(t, result, &parsed)

	if _, ok := parsed.ID.(string); !ok {
		t.Errorf("expected ID to be string after normalization, got %T: %v", parsed.ID, parsed.ID)
	}
	if parsed.ID != "42" {
		t.Errorf("expected ID=42, got %v", parsed.ID)
	}
}

func TestServiceObjectHook_AfterSuccess_ICMPFieldNotTouched(t *testing.T) {
	hook := &serviceObjectHook{}

	body := `{"id":"1","name":"x","status":"applied","type":"custom","protocols":{"icmp":true}}`
	res := makeServiceObjectResponse(200, body)
	hookCtx := AfterSuccessContext{HookContext: HookContext{OperationID: "getServiceObject"}}

	result, err := hook.AfterSuccess(hookCtx, res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Protocols struct {
			ICMP interface{} `json:"icmp"`
		} `json:"protocols"`
	}
	readServiceObjectResponse(t, result, &parsed)

	if parsed.Protocols.ICMP != true {
		t.Errorf("expected icmp=true, got %v", parsed.Protocols.ICMP)
	}
}

func TestServiceObjectHook_AfterSuccess_NormalizesListResponse(t *testing.T) {
	hook := &serviceObjectHook{}

	body := `{
		"services": [
			{"id": "1", "name": "svc-a", "status": "APPLIED", "type": "CUSTOM", "protocols": {"tcp": [80]}},
			{"id": "2", "name": "svc-b", "status": "applied", "type": "PREDEFINED", "protocols": {"tcp": ["443"]}}
		],
		"total": 2
	}`

	res := makeServiceObjectResponse(200, body)
	hookCtx := AfterSuccessContext{HookContext: HookContext{OperationID: "listServiceObjects"}}

	result, err := hook.AfterSuccess(hookCtx, res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Services []struct {
			Status    string        `json:"status"`
			Type      string        `json:"type"`
			Protocols struct {
				TCP []interface{} `json:"tcp"`
			} `json:"protocols"`
		} `json:"services"`
	}
	readServiceObjectResponse(t, result, &parsed)

	if len(parsed.Services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(parsed.Services))
	}

	// First service: APPLIED→applied, CUSTOM→custom, [80]→["80"]
	if parsed.Services[0].Status != "applied" {
		t.Errorf("services[0].status: expected applied, got %q", parsed.Services[0].Status)
	}
	if parsed.Services[0].Type != "custom" {
		t.Errorf("services[0].type: expected custom, got %q", parsed.Services[0].Type)
	}
	if len(parsed.Services[0].Protocols.TCP) != 1 || parsed.Services[0].Protocols.TCP[0] != "80" {
		t.Errorf("services[0].tcp: expected [80], got %v", parsed.Services[0].Protocols.TCP)
	}

	// Second service: already normalized
	if parsed.Services[1].Type != "PREDEFINED" {
		t.Errorf("services[1].type: expected PREDEFINED, got %q", parsed.Services[1].Type)
	}
}

func TestServiceObjectHook_AfterSuccess_NormalizesCreateResponse(t *testing.T) {
	hook := &serviceObjectHook{}

	// API returns integer ports in tcp_udp on create response
	body := `{
		"id": "99",
		"name": "DNS",
		"status": "applied",
		"type": "custom",
		"protocols": {
			"tcp_udp": [53]
		}
	}`

	res := makeServiceObjectResponse(200, body)
	hookCtx := AfterSuccessContext{HookContext: HookContext{OperationID: "createServiceObject"}}

	result, err := hook.AfterSuccess(hookCtx, res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Protocols struct {
			TCPUDP []interface{} `json:"tcp_udp"`
		} `json:"protocols"`
	}
	readServiceObjectResponse(t, result, &parsed)

	if len(parsed.Protocols.TCPUDP) != 1 {
		t.Fatalf("expected 1 tcp_udp port, got %d", len(parsed.Protocols.TCPUDP))
	}
	if _, ok := parsed.Protocols.TCPUDP[0].(string); !ok {
		t.Errorf("tcp_udp[0] should be string after normalization, got %T: %v", parsed.Protocols.TCPUDP[0], parsed.Protocols.TCPUDP[0])
	}
	if parsed.Protocols.TCPUDP[0] != "53" {
		t.Errorf("expected tcp_udp[0]=53, got %v", parsed.Protocols.TCPUDP[0])
	}
}

func TestServiceObjectHook_AfterSuccess_NormalizesUpdateResponse(t *testing.T) {
	hook := &serviceObjectHook{}

	body := `{
		"id": "99",
		"name": "DNS",
		"status": "applied",
		"type": "custom",
		"protocols": {
			"tcp_udp": [53]
		}
	}`

	res := makeServiceObjectResponse(200, body)
	hookCtx := AfterSuccessContext{HookContext: HookContext{OperationID: "updateServiceObject"}}

	result, err := hook.AfterSuccess(hookCtx, res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Protocols struct {
			TCPUDP []interface{} `json:"tcp_udp"`
		} `json:"protocols"`
	}
	readServiceObjectResponse(t, result, &parsed)

	if len(parsed.Protocols.TCPUDP) != 1 || parsed.Protocols.TCPUDP[0] != "53" {
		t.Errorf("expected tcp_udp=[53], got %v", parsed.Protocols.TCPUDP)
	}
}

func TestServiceObjectHook_AfterSuccess_PassthroughForOtherOps(t *testing.T) {
	hook := &serviceObjectHook{}

	body := `{"some":"data"}`
	res := &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}

	for _, opID := range []string{"createServiceObject", "updateServiceObject", "deleteServiceObject", "someOtherOp"} {
		hookCtx := AfterSuccessContext{HookContext: HookContext{OperationID: opID}}
		result, err := hook.AfterSuccess(hookCtx, res)
		if err != nil {
			t.Fatalf("op %s: unexpected error: %v", opID, err)
		}
		if result != res {
			t.Errorf("op %s: expected passthrough (same pointer), got different response", opID)
		}
	}
}

func TestServiceObjectHook_NormalizesErrCodeIntegerToString(t *testing.T) {
	hook := &serviceObjectHook{}

	body := `{"err_code": 400, "err_msg": "bad request"}`
	res := makeServiceObjectResponse(400, body)
	hookCtx := AfterSuccessContext{HookContext: HookContext{OperationID: "getServiceObject"}}

	result, err := hook.AfterSuccess(hookCtx, res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resultBody, _ := io.ReadAll(result.Body)

	var parsed struct {
		ErrCode interface{} `json:"err_code"`
	}
	if err := json.Unmarshal(resultBody, &parsed); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	if _, ok := parsed.ErrCode.(string); !ok {
		t.Errorf("expected err_code to be string, got %T: %v", parsed.ErrCode, parsed.ErrCode)
	}
	if parsed.ErrCode != "400" {
		t.Errorf("expected err_code=400, got %v", parsed.ErrCode)
	}
}

// --- helpers ---

func makeServiceObjectResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    &http.Request{URL: &url.URL{}},
	}
}

func readServiceObjectResponse(t *testing.T, res *http.Response, dst interface{}) {
	t.Helper()
	b, _ := io.ReadAll(res.Body)
	if err := json.Unmarshal(b, dst); err != nil {
		t.Fatalf("failed to parse response body: %v\nbody: %s", err, b)
	}
}
