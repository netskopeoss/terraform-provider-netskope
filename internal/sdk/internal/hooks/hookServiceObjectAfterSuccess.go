package hooks

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// serviceObjectHook handles request/response normalization for service object endpoints.
//
// BeforeRequest:
//   - Caps the `limit` query parameter at 150 (the API maximum). The generated SDK
//     sets a default of 500 which causes 400 errors.
//
// AfterSuccess:
//   - Protocol port arrays: predefined objects return ports as integers (e.g.
//     [80, 443]) but the SDK model expects []string. Converts integers to strings.
//   - Status field: predefined objects return status as uppercase ("APPLIED")
//     but the generated enum only handles lowercase values. Normalizes to lowercase.
//   - err_code field: error responses return an integer but the SDK model expects
//     a string. Converts to string as a safety net.
type serviceObjectHook struct{}

var _ beforeRequestHook = (*serviceObjectHook)(nil)
var _ afterSuccessHook = (*serviceObjectHook)(nil)

func (h *serviceObjectHook) BeforeRequest(hookCtx BeforeRequestContext, req *http.Request) (*http.Request, error) {
	if hookCtx.OperationID != "listServiceObjects" {
		return req, nil
	}

	q := req.URL.Query()
	if limit := q.Get("limit"); limit == "" || limit == "500" {
		q.Set("limit", "150")
		req.URL.RawQuery = q.Encode()
	}

	return req, nil
}

func (h *serviceObjectHook) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
	switch hookCtx.OperationID {
	case "listServiceObjects":
		return h.normalizeListResponse(res)
	case "getServiceObject":
		return h.normalizeItemResponse(res)
	default:
		return res, nil
	}
}

func (h *serviceObjectHook) normalizeListResponse(res *http.Response) (*http.Response, error) {
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("serviceObject list hook: failed to read response: %w", err)
	}

	if res.StatusCode != 200 {
		// Normalize err_code from integer to string in error responses
		return h.normalizeErrorResponse(res, body), nil
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return replaceBody(res, body), nil
	}

	servicesRaw, ok := raw["services"]
	if !ok {
		return replaceBody(res, body), nil
	}

	var items []map[string]json.RawMessage
	if err := json.Unmarshal(servicesRaw, &items); err != nil {
		return replaceBody(res, body), nil
	}

	for i := range items {
		h.normalizeServiceObject(items[i])
	}

	normalizedServices, err := json.Marshal(items)
	if err != nil {
		return replaceBody(res, body), nil
	}
	raw["services"] = normalizedServices

	normalized, err := json.Marshal(raw)
	if err != nil {
		return replaceBody(res, body), nil
	}

	return replaceBody(res, normalized), nil
}

func (h *serviceObjectHook) normalizeItemResponse(res *http.Response) (*http.Response, error) {
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("serviceObject item hook: failed to read response: %w", err)
	}

	if res.StatusCode != 200 {
		return h.normalizeErrorResponse(res, body), nil
	}

	var obj map[string]json.RawMessage
	if err := json.Unmarshal(body, &obj); err != nil {
		return replaceBody(res, body), nil
	}

	h.normalizeServiceObject(obj)

	normalized, err := json.Marshal(obj)
	if err != nil {
		return replaceBody(res, body), nil
	}

	return replaceBody(res, normalized), nil
}

// normalizeErrorResponse converts the err_code field from integer to string.
// The API returns {"err_code": 400, ...} but the SDK model expects *string.
func (h *serviceObjectHook) normalizeErrorResponse(res *http.Response, body []byte) *http.Response {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(body, &obj); err != nil {
		return replaceBody(res, body)
	}

	if raw, ok := obj["err_code"]; ok {
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			// It's not a string; try integer
			var n int64
			if err := json.Unmarshal(raw, &n); err == nil {
				if normalized, err := json.Marshal(strconv.FormatInt(n, 10)); err == nil {
					obj["err_code"] = normalized
				}
			}
		}
	}

	normalized, err := json.Marshal(obj)
	if err != nil {
		return replaceBody(res, body)
	}
	return replaceBody(res, normalized)
}

// normalizeServiceObject normalizes the status, type fields and protocol port arrays.
func (h *serviceObjectHook) normalizeServiceObject(obj map[string]json.RawMessage) {
	// Normalize status to lowercase (predefined objects return "APPLIED", "PENDING-CREATE", etc.)
	if statusRaw, ok := obj["status"]; ok {
		var status string
		if err := json.Unmarshal(statusRaw, &status); err == nil {
			lower := strings.ToLower(status)
			if lower != status {
				if normalized, err := json.Marshal(lower); err == nil {
					obj["status"] = normalized
				}
			}
		}
	}

	// Normalize type field: API returns "CUSTOM" but enum expects "custom";
	// "PREDEFINED" is correct and stays as-is.
	if typeRaw, ok := obj["type"]; ok {
		var objType string
		if err := json.Unmarshal(typeRaw, &objType); err == nil {
			var canonical string
			switch strings.ToLower(objType) {
			case "custom":
				canonical = "custom"
			case "predefined":
				canonical = "PREDEFINED"
			}
			if canonical != "" && canonical != objType {
				if normalized, err := json.Marshal(canonical); err == nil {
					obj["type"] = normalized
				}
			}
		}
	}

	// Defensively convert numeric IDs to strings
	if idRaw, ok := obj["id"]; ok {
		var idStr string
		if err := json.Unmarshal(idRaw, &idStr); err != nil {
			var idNum int64
			if err := json.Unmarshal(idRaw, &idNum); err == nil {
				if normalized, err := json.Marshal(strconv.FormatInt(idNum, 10)); err == nil {
					obj["id"] = normalized
				}
			}
		}
	}

	// Normalize protocol port arrays (convert integers to strings)
	protocolsRaw, ok := obj["protocols"]
	if !ok {
		return
	}

	var protocols map[string]json.RawMessage
	if err := json.Unmarshal(protocolsRaw, &protocols); err != nil {
		return
	}

	changed := false
	for key := range protocols {
		if key == "icmp" {
			continue // bool field, not a port array
		}
		if normalized, ok := h.normalizePortArray(protocols[key]); ok {
			protocols[key] = normalized
			changed = true
		}
	}

	if changed {
		if normalizedProtocols, err := json.Marshal(protocols); err == nil {
			obj["protocols"] = normalizedProtocols
		}
	}
}

// normalizePortArray converts integer port values to strings. Returns the
// normalized JSON and true if any conversion was performed.
func (h *serviceObjectHook) normalizePortArray(val json.RawMessage) (json.RawMessage, bool) {
	var ports []json.RawMessage
	if err := json.Unmarshal(val, &ports); err != nil {
		return val, false
	}

	changed := false
	normalized := make([]string, 0, len(ports))
	for _, portRaw := range ports {
		var s string
		if err := json.Unmarshal(portRaw, &s); err == nil {
			normalized = append(normalized, s)
			continue
		}
		var n int64
		if err := json.Unmarshal(portRaw, &n); err == nil {
			normalized = append(normalized, strconv.FormatInt(n, 10))
			changed = true
			continue
		}
	}

	if !changed {
		return val, false
	}

	normalizedJSON, err := json.Marshal(normalized)
	if err != nil {
		return val, false
	}
	return normalizedJSON, true
}
