package hooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// dcConfigHooks handles request/response transformations for the DC rules and
// on-prem detection endpoints, which differ from the standard REST pattern:
//
//   - POST /rules accepts an array body and returns 201 with empty body; a
//     follow-up GET /rules?label=… is used to resolve the created rule ID.
//   - PUT /rules/{id} returns 200 with empty body; a follow-up GET resolves state.
//   - POST /client/onpremdetection accepts an array body and returns 201 with
//     {"status":true,"data":[id]}; a follow-up GET /client/onpremdetection/{id}
//     is required to resolve the full item.
//   - PUT /client/onpremdetection/{id} returns 200; a follow-up GET resolves state.
//   - GET /rules and GET /client/onpremdetection return flat arrays that must be
//     wrapped into {"rules":[…]} and {"onpremdetection":[…]} respectively.
//   - PUT /onpremdetection/steering/{id} accepts a bare int array, but the SDK
//     sends {"onprem_detection_ids":[…]} which must be unwrapped.
type dcConfigHooks struct {
	// ruleCreateRetryDelays controls the backoff between GET retries when a
	// newly created rule is not immediately visible in the list. Configurable
	// to allow fast tests without real sleeps.
	ruleCreateRetryDelays []time.Duration
}

// defaultRuleCreateRetryDelays are the production retry delays for eventual
// consistency after POST /rules.
var defaultRuleCreateRetryDelays = []time.Duration{
	1 * time.Second,
	2 * time.Second,
	4 * time.Second,
	8 * time.Second,
	16 * time.Second,
}

func (h *dcConfigHooks) retryDelays() []time.Duration {
	if h.ruleCreateRetryDelays != nil {
		return h.ruleCreateRetryDelays
	}
	return defaultRuleCreateRetryDelays
}

var _ beforeRequestHook = (*dcConfigHooks)(nil)
var _ afterSuccessHook = (*dcConfigHooks)(nil)

func (h *dcConfigHooks) BeforeRequest(hookCtx BeforeRequestContext, req *http.Request) (*http.Request, error) {
	switch hookCtx.OperationID {
	case "createDeviceClassificationRule":
		return h.beforeCreateRule(req)
	case "createDeviceClassificationOnPremDetection":
		return h.wrapInArray(req, "createDeviceClassificationOnPremDetection")
	case "updateDeviceClassificationSteeringMapping":
		return h.beforeUpdateSteering(req)
	case "listDeviceClassificationRules", "listDeviceClassificationOnPremDetection":
		return h.beforeList(req)
	}
	return req, nil
}

// beforeList strips the offset query parameter when limit is not provided.
// The list endpoints require limit when offset is set, but the SDK sends
// offset=0 by default (from the OAS default value). Sending offset without
// limit causes a 400 Bad Request.
func (h *dcConfigHooks) beforeList(req *http.Request) (*http.Request, error) {
	q := req.URL.Query()
	if q.Get("limit") == "" {
		q.Del("offset")
		req.URL.RawQuery = q.Encode()
	}
	return req, nil
}

func (h *dcConfigHooks) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
	switch hookCtx.OperationID {
	case "createDeviceClassificationRule":
		return h.afterCreateRule(res)
	case "updateDeviceClassificationRule":
		return h.afterUpdateByPathID(res, "rules")
	case "listDeviceClassificationRules":
		return h.wrapFlatArray(res, "rules")
	case "createDeviceClassificationOnPremDetection":
		return h.afterCreateOnPrem(res)
	case "updateDeviceClassificationOnPremDetection":
		return h.afterUpdateByPathID(res, "client/onpremdetection")
	case "listDeviceClassificationOnPremDetection":
		return h.wrapFlatArray(res, "onpremdetection")
	}
	return res, nil
}

// beforeCreateRule parses the single-rule body, records name and label as
// X-DC-Rule-Name / X-DC-Rule-Label request headers for the AfterSuccess hook,
// then wraps the body in an array as required by the API.
func (h *dcConfigHooks) beforeCreateRule(req *http.Request) (*http.Request, error) {
	body, err := io.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("dcConfig createRule before: read body: %w", err)
	}

	var item struct {
		Name  string `json:"name"`
		Label string `json:"label"`
	}
	if err := json.Unmarshal(body, &item); err != nil {
		return nil, fmt.Errorf("dcConfig createRule before: unmarshal: %w", err)
	}

	req.Header.Set("X-DC-Rule-Name", item.Name)
	req.Header.Set("X-DC-Rule-Label", item.Label)

	wrapped, err := json.Marshal([]json.RawMessage{body})
	if err != nil {
		return nil, fmt.Errorf("dcConfig createRule before: wrap: %w", err)
	}

	req.Body = io.NopCloser(bytes.NewReader(wrapped))
	req.ContentLength = int64(len(wrapped))
	req.Header.Set("Content-Length", strconv.Itoa(len(wrapped)))
	return req, nil
}

// wrapInArray wraps a single JSON object body into an array, as required by the
// batch-upsert endpoints.
func (h *dcConfigHooks) wrapInArray(req *http.Request, op string) (*http.Request, error) {
	body, err := io.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("dcConfig %s before: read body: %w", op, err)
	}

	var item json.RawMessage
	if err := json.Unmarshal(body, &item); err != nil {
		return nil, fmt.Errorf("dcConfig %s before: unmarshal: %w", op, err)
	}

	wrapped, err := json.Marshal([]json.RawMessage{item})
	if err != nil {
		return nil, fmt.Errorf("dcConfig %s before: wrap: %w", op, err)
	}

	req.Body = io.NopCloser(bytes.NewReader(wrapped))
	req.ContentLength = int64(len(wrapped))
	req.Header.Set("Content-Length", strconv.Itoa(len(wrapped)))
	return req, nil
}

// beforeUpdateSteering unwraps {"onprem_detection_ids":[…]} to a bare int array
// as required by PUT /onpremdetection/steering/{id}.
func (h *dcConfigHooks) beforeUpdateSteering(req *http.Request) (*http.Request, error) {
	body, err := io.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("dcConfig steering before: read body: %w", err)
	}

	var wrapper struct {
		OnpremDetectionIds []int64 `json:"onprem_detection_ids"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, fmt.Errorf("dcConfig steering before: unmarshal: %w", err)
	}

	bare, err := json.Marshal(wrapper.OnpremDetectionIds)
	if err != nil {
		return nil, fmt.Errorf("dcConfig steering before: marshal: %w", err)
	}

	req.Body = io.NopCloser(bytes.NewReader(bare))
	req.ContentLength = int64(len(bare))
	req.Header.Set("Content-Length", strconv.Itoa(len(bare)))
	return req, nil
}

// afterCreateRule handles the 201 empty body from POST /rules by fetching the
// rule via GET /rules?label={label} and locating it by name. Retries with
// backoff to handle eventual consistency (the list may not reflect the new
// rule immediately after creation).
func (h *dcConfigHooks) afterCreateRule(res *http.Response) (*http.Response, error) {
	io.ReadAll(res.Body) //nolint:errcheck
	res.Body.Close()

	name := res.Request.Header.Get("X-DC-Rule-Name")
	label := res.Request.Header.Get("X-DC-Rule-Label")
	if name == "" || label == "" {
		return nil, fmt.Errorf("dcConfig createRule after: missing X-DC-Rule-Name or X-DC-Rule-Label headers")
	}

	getURL := *res.Request.URL
	getURL.Path = "/api/v2/deviceclassification/rules"
	getURL.RawQuery = "label=" + url.QueryEscape(label)

	client := &http.Client{}

	delays := h.retryDelays()
	for attempt := 0; attempt <= len(delays); attempt++ {
		getReq, err := http.NewRequestWithContext(res.Request.Context(), "GET", getURL.String(), nil)
		if err != nil {
			return nil, fmt.Errorf("dcConfig createRule after: build GET: %w", err)
		}
		dcCopyAuthHeaders(res.Request, getReq)

		getRes, err := client.Do(getReq)
		if err != nil {
			return nil, fmt.Errorf("dcConfig createRule after: GET /rules: %w", err)
		}

		rawList, err := io.ReadAll(getRes.Body)
		getRes.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("dcConfig createRule after: read list: %w", err)
		}

		var rules []json.RawMessage
		if err := json.Unmarshal(rawList, &rules); err != nil {
			return nil, fmt.Errorf("dcConfig createRule after: parse list: %w", err)
		}

		for _, raw := range rules {
			var r struct {
				Name string `json:"name"`
			}
			if err := json.Unmarshal(raw, &r); err != nil {
				continue
			}
			if r.Name == name {
				return replaceBody(res, []byte(raw)), nil
			}
		}

		if attempt < len(delays) {
			time.Sleep(delays[attempt])
		}
	}

	return nil, fmt.Errorf("dcConfig createRule after: rule %q not found in label %q after retries", name, label)
}

// afterCreateOnPrem handles the 201 response from POST /client/onpremdetection,
// which returns {"status":true,"data":[id]} instead of the full item. It extracts
// the ID from data[0] and fetches the full item via GET /client/onpremdetection/{id}.
func (h *dcConfigHooks) afterCreateOnPrem(res *http.Response) (*http.Response, error) {
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("dcConfig createOnPrem after: read body: %w", err)
	}

	var created struct {
		Status bool    `json:"status"`
		Data   []int64 `json:"data"`
	}
	if err := json.Unmarshal(body, &created); err != nil {
		return nil, fmt.Errorf("dcConfig createOnPrem after: parse body %q: %w", string(body), err)
	}
	if len(created.Data) == 0 {
		return nil, fmt.Errorf("dcConfig createOnPrem after: empty data array in response")
	}
	id := created.Data[0]

	getURL := *res.Request.URL
	getURL.Path = fmt.Sprintf("/api/v2/deviceclassification/client/onpremdetection/%d", id)
	getURL.RawQuery = ""

	getReq, err := http.NewRequestWithContext(res.Request.Context(), "GET", getURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("dcConfig createOnPrem after: build GET: %w", err)
	}
	dcCopyAuthHeaders(res.Request, getReq)

	client := &http.Client{}
	getRes, err := client.Do(getReq)
	if err != nil {
		return nil, fmt.Errorf("dcConfig createOnPrem after: GET: %w", err)
	}

	itemBody, err := io.ReadAll(getRes.Body)
	getRes.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("dcConfig createOnPrem after: read item: %w", err)
	}

	return replaceBody(res, itemBody), nil
}

// afterUpdateByPathID handles 200 empty-body responses from PUT endpoints by
// doing a follow-up GET /{pathSuffix}/{id} using the ID extracted from the URL.
// pathSuffix is relative to /api/v2/deviceclassification/ (e.g. "rules" or
// "client/onpremdetection").
func (h *dcConfigHooks) afterUpdateByPathID(res *http.Response, pathSuffix string) (*http.Response, error) {
	io.ReadAll(res.Body) //nolint:errcheck
	res.Body.Close()

	parts := splitPath(res.Request.URL.Path)
	if len(parts) == 0 {
		return nil, fmt.Errorf("dcConfig afterUpdate(%s): cannot extract ID from URL %s", pathSuffix, res.Request.URL.Path)
	}
	idStr := parts[len(parts)-1]

	getURL := *res.Request.URL
	getURL.Path = fmt.Sprintf("/api/v2/deviceclassification/%s/%s", pathSuffix, idStr)
	getURL.RawQuery = ""

	getReq, err := http.NewRequestWithContext(res.Request.Context(), "GET", getURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("dcConfig afterUpdate(%s): build GET: %w", pathSuffix, err)
	}
	dcCopyAuthHeaders(res.Request, getReq)

	client := &http.Client{}
	getRes, err := client.Do(getReq)
	if err != nil {
		return nil, fmt.Errorf("dcConfig afterUpdate(%s): GET: %w", pathSuffix, err)
	}

	body, err := io.ReadAll(getRes.Body)
	getRes.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("dcConfig afterUpdate(%s): read: %w", pathSuffix, err)
	}

	return replaceBody(res, body), nil
}

// wrapFlatArray wraps a flat JSON array response into {"key":[…]}.
func (h *dcConfigHooks) wrapFlatArray(res *http.Response, key string) (*http.Response, error) {
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("dcConfig wrapFlatArray(%s): read: %w", key, err)
	}

	var items []json.RawMessage
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, fmt.Errorf("dcConfig wrapFlatArray(%s): unmarshal: %w", key, err)
	}

	wrapped, err := json.Marshal(map[string]interface{}{key: items})
	if err != nil {
		return nil, fmt.Errorf("dcConfig wrapFlatArray(%s): marshal: %w", key, err)
	}

	return replaceBody(res, wrapped), nil
}

// dcCopyAuthHeaders copies Netskope authentication headers to a new request.
func dcCopyAuthHeaders(src, dst *http.Request) {
	for _, h := range []string{"Netskope-Api-Token", "Authorization"} {
		if v := src.Header.Get(h); v != "" {
			dst.Header.Set(h, v)
		}
	}
}
