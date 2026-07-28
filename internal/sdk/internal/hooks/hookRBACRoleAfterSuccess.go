package hooks

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// rbacRoleHook handles request/response normalization for RBAC role endpoints.
//
// AfterSuccess (createRBACRole, replaceRBACRoleDetails):
//   - The API returns only {"roleId": N} on create/update. This hook fetches the
//     full role via GET and replaces the response body, preserving the original
//     status code (201 for create, 200 for update). This allows the generated
//     Terraform resource to populate all computed fields after write operations.
//
// AfterSuccess (getRBACRole):
//   - The ipAllowList.ipList field in GET responses is an array of objects
//     ({ipAddress, createdAt, updatedAt}), but the OAS and SDK model expect
//     string arrays. This hook normalizes those objects to bare IP address strings.
type rbacRoleHook struct{}

var _ afterSuccessHook = (*rbacRoleHook)(nil)

func (h *rbacRoleHook) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
	switch hookCtx.OperationID {
	case "createRBACRole", "replaceRBACRoleDetails":
		return h.fetchFullRole(res)
	case "getRBACRole":
		return h.normalizeGetResponse(res)
	default:
		return res, nil
	}
}

// fetchFullRole parses the roleId from the slim {"roleId": N} create/update
// response, fetches the full role via GET, normalizes it, and replaces the
// response body while preserving the original status code.
func (h *rbacRoleHook) fetchFullRole(origRes *http.Response) (*http.Response, error) {
	body, err := io.ReadAll(origRes.Body)
	origRes.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("rbacRole hook: failed to read create/update response: %w", err)
	}

	if origRes.StatusCode < 200 || origRes.StatusCode >= 300 {
		return replaceBody(origRes, body), nil
	}

	var slim struct {
		RoleID int64 `json:"roleId"`
	}
	if err := json.Unmarshal(body, &slim); err != nil || slim.RoleID == 0 {
		// Not the expected slim response — pass through unchanged
		return replaceBody(origRes, body), nil
	}

	// Fetch the full role
	getURL := *origRes.Request.URL
	getURL.Path = fmt.Sprintf("/api/v2/rbac/roles/%d", slim.RoleID)
	getURL.RawQuery = ""

	getReq, err := http.NewRequestWithContext(origRes.Request.Context(), http.MethodGet, getURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("rbacRole hook: failed to build GET request: %w", err)
	}

	// Copy auth headers from original request
	for _, hdr := range []string{"Netskope-Api-Token", "Authorization"} {
		if v := origRes.Request.Header.Get(hdr); v != "" {
			getReq.Header.Set(hdr, v)
		}
	}

	client := &http.Client{}
	getRes, err := client.Do(getReq)
	if err != nil {
		return nil, fmt.Errorf("rbacRole hook: follow-up GET failed: %w", err)
	}

	getBody, err := io.ReadAll(getRes.Body)
	getRes.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("rbacRole hook: failed to read GET response: %w", err)
	}

	if getRes.StatusCode != 200 {
		// GET failed — return the original slim response unchanged so the SDK can
		// surface a meaningful error instead of a confusing unmarshal failure.
		return replaceBody(origRes, body), nil
	}

	normalized, err := h.normalizeRoleBody(getBody)
	if err != nil {
		return nil, err
	}

	return replaceBody(origRes, normalized), nil
}

// normalizeGetResponse normalizes the ipAllowList.ipList field in GET responses
// from an array of objects ({ipAddress, createdAt, updatedAt}) to an array of
// plain IP address strings, as expected by the OAS / SDK model.
func (h *rbacRoleHook) normalizeGetResponse(origRes *http.Response) (*http.Response, error) {
	body, err := io.ReadAll(origRes.Body)
	origRes.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("rbacRole hook: failed to read GET response: %w", err)
	}

	if origRes.StatusCode != 200 {
		return replaceBody(origRes, body), nil
	}

	normalized, err := h.normalizeRoleBody(body)
	if err != nil {
		return nil, err
	}

	return replaceBody(origRes, normalized), nil
}

// normalizeRoleBody normalizes a full role JSON body:
//   - ipAllowList.ipList: objects with {ipAddress, ...} → plain strings
func (h *rbacRoleHook) normalizeRoleBody(body []byte) ([]byte, error) {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(body, &obj); err != nil {
		return body, nil
	}

	ipAllowListRaw, ok := obj["ipAllowList"]
	if !ok {
		return body, nil
	}

	var ipAllowList map[string]json.RawMessage
	if err := json.Unmarshal(ipAllowListRaw, &ipAllowList); err != nil {
		return body, nil
	}

	ipListRaw, ok := ipAllowList["ipList"]
	if !ok {
		return body, nil
	}

	normalized, changed := h.normalizeIPList(ipListRaw)
	if !changed {
		return body, nil
	}

	ipAllowList["ipList"] = normalized
	normalizedIPAllowList, err := json.Marshal(ipAllowList)
	if err != nil {
		return body, nil
	}
	obj["ipAllowList"] = normalizedIPAllowList

	result, err := json.Marshal(obj)
	if err != nil {
		return body, nil
	}
	return result, nil
}

// normalizeIPList converts an IP list from object form
// ([{ipAddress: "1.2.3.4", createdAt: ...}]) to string form (["1.2.3.4"]).
// Returns the normalized JSON and true if any conversion was performed.
func (h *rbacRoleHook) normalizeIPList(raw json.RawMessage) (json.RawMessage, bool) {
	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil {
		return raw, false
	}

	changed := false
	result := make([]string, 0, len(items))

	for _, item := range items {
		// Try string first (already normalized)
		var s string
		if err := json.Unmarshal(item, &s); err == nil {
			result = append(result, s)
			continue
		}

		// Object form: {ipAddress: "1.2.3.4", ...}
		var obj struct {
			IPAddress string `json:"ipAddress"`
		}
		if err := json.Unmarshal(item, &obj); err == nil && obj.IPAddress != "" {
			result = append(result, obj.IPAddress)
			changed = true
			continue
		}
	}

	if !changed {
		return raw, false
	}

	normalized, err := json.Marshal(result)
	if err != nil {
		return raw, false
	}
	return normalized, true
}

