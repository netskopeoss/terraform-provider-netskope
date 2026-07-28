package hooks

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestRBACRoleHook_GetResponse_NormalizesIPList(t *testing.T) {
	hook := &rbacRoleHook{}

	body := `{
		"roleId": 5,
		"roleName": "test-role",
		"roleDescription": "desc",
		"apiGroups": [{"apiGroupId": 107, "permission": "rw"}],
		"ipAllowList": {
			"enableIpAllowList": true,
			"ipList": [
				{"ipAddress": "10.0.0.1", "createdAt": "2026-07-27T11:13:22.000Z", "updatedAt": "2026-07-27T11:13:22.000Z"},
				{"ipAddress": "192.168.1.1", "createdAt": "2026-07-27T11:13:22.000Z", "updatedAt": "2026-07-27T11:13:22.000Z"}
			]
		}
	}`

	res := &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    &http.Request{URL: &url.URL{}},
	}

	hookCtx := AfterSuccessContext{HookContext: HookContext{OperationID: "getRBACRole"}}
	result, err := hook.AfterSuccess(hookCtx, res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resultBody, _ := io.ReadAll(result.Body)

	var parsed struct {
		IPAllowList struct {
			EnableIPAllowList bool            `json:"enableIpAllowList"`
			IPList            []interface{}   `json:"ipList"`
		} `json:"ipAllowList"`
	}
	if err := json.Unmarshal(resultBody, &parsed); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	if len(parsed.IPAllowList.IPList) != 2 {
		t.Fatalf("expected 2 IPs, got %d", len(parsed.IPAllowList.IPList))
	}

	// After normalization, ipList items should be strings
	for i, item := range parsed.IPAllowList.IPList {
		if _, ok := item.(string); !ok {
			t.Errorf("ipList[%d] should be string, got %T: %v", i, item, item)
		}
	}

	if parsed.IPAllowList.IPList[0] != "10.0.0.1" {
		t.Errorf("expected ipList[0] = 10.0.0.1, got %v", parsed.IPAllowList.IPList[0])
	}
	if parsed.IPAllowList.IPList[1] != "192.168.1.1" {
		t.Errorf("expected ipList[1] = 192.168.1.1, got %v", parsed.IPAllowList.IPList[1])
	}
}

func TestRBACRoleHook_GetResponse_EmptyIPList(t *testing.T) {
	hook := &rbacRoleHook{}

	body := `{
		"roleId": 5,
		"roleName": "test-role",
		"roleDescription": "desc",
		"apiGroups": [],
		"ipAllowList": {
			"enableIpAllowList": false,
			"ipList": []
		}
	}`

	res := &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    &http.Request{URL: &url.URL{}},
	}

	hookCtx := AfterSuccessContext{HookContext: HookContext{OperationID: "getRBACRole"}}
	result, err := hook.AfterSuccess(hookCtx, res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resultBody, _ := io.ReadAll(result.Body)

	var parsed struct {
		IPAllowList struct {
			IPList []interface{} `json:"ipList"`
		} `json:"ipAllowList"`
	}
	if err := json.Unmarshal(resultBody, &parsed); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	if len(parsed.IPAllowList.IPList) != 0 {
		t.Errorf("expected empty ipList, got %v", parsed.IPAllowList.IPList)
	}
}

func TestRBACRoleHook_GetResponse_AlreadyNormalizedStrings(t *testing.T) {
	hook := &rbacRoleHook{}

	// ipList already contains strings (should pass through unchanged)
	body := `{
		"roleId": 5,
		"roleName": "test-role",
		"roleDescription": "desc",
		"apiGroups": [],
		"ipAllowList": {
			"enableIpAllowList": true,
			"ipList": ["10.0.0.1", "10.0.0.2"]
		}
	}`

	res := &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    &http.Request{URL: &url.URL{}},
	}

	hookCtx := AfterSuccessContext{HookContext: HookContext{OperationID: "getRBACRole"}}
	result, err := hook.AfterSuccess(hookCtx, res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resultBody, _ := io.ReadAll(result.Body)

	var parsed struct {
		IPAllowList struct {
			IPList []string `json:"ipList"`
		} `json:"ipAllowList"`
	}
	if err := json.Unmarshal(resultBody, &parsed); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	if len(parsed.IPAllowList.IPList) != 2 {
		t.Fatalf("expected 2 IPs, got %d", len(parsed.IPAllowList.IPList))
	}
}

func TestRBACRoleHook_UnrelatedOperationsPassthrough(t *testing.T) {
	hook := &rbacRoleHook{}

	body := `{"some": "data"}`
	res := &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}

	for _, opID := range []string{"listRBACRoles", "deleteRBACRole", "someOtherOp"} {
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

func TestRBACRoleHook_NormalizeRoleBody_NoIPAllowList(t *testing.T) {
	hook := &rbacRoleHook{}

	// Body without ipAllowList should pass through unchanged
	body := []byte(`{"roleId": 5, "roleName": "test"}`)
	result, err := hook.normalizeRoleBody(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var orig, norm map[string]interface{}
	json.Unmarshal(body, &orig)
	json.Unmarshal(result, &norm)

	if orig["roleId"] != norm["roleId"] {
		t.Errorf("roleId mismatch: %v != %v", orig["roleId"], norm["roleId"])
	}
}
