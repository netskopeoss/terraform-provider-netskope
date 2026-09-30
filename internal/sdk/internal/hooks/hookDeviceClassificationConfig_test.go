package hooks

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func dcBeforeCtx(opID string) BeforeRequestContext {
	return BeforeRequestContext{HookContext: HookContext{OperationID: opID}}
}

func dcAfterCtx(opID string) AfterSuccessContext {
	return AfterSuccessContext{HookContext: HookContext{OperationID: opID}}
}

func TestDCConfig_BeforeCreateRule_WrapsAndSetsHeaders(t *testing.T) {
	h := &dcConfigHooks{}

	body := `{"name":"test-rule","label":"Windows","os":"windows","conditions":{"$and":[]}}`
	req, _ := http.NewRequest("POST", "https://example.com/api/v2/deviceclassification/rules", bytes.NewBufferString(body))
	req.Header = make(http.Header)

	out, err := h.BeforeRequest(dcBeforeCtx("createDeviceClassificationRule"), req)
	if err != nil {
		t.Fatalf("BeforeRequest error: %v", err)
	}

	if out.Header.Get("X-DC-Rule-Name") != "test-rule" {
		t.Errorf("expected X-DC-Rule-Name=test-rule, got %q", out.Header.Get("X-DC-Rule-Name"))
	}
	if out.Header.Get("X-DC-Rule-Label") != "Windows" {
		t.Errorf("expected X-DC-Rule-Label=Windows, got %q", out.Header.Get("X-DC-Rule-Label"))
	}

	outBody, _ := io.ReadAll(out.Body)
	var arr []map[string]interface{}
	if err := json.Unmarshal(outBody, &arr); err != nil {
		t.Fatalf("expected array body, got: %s", outBody)
	}
	if len(arr) != 1 {
		t.Errorf("expected 1-element array, got %d", len(arr))
	}
	if arr[0]["name"] != "test-rule" {
		t.Errorf("expected name=test-rule in wrapped body, got %v", arr[0]["name"])
	}
}

func TestDCConfig_BeforeCreateOnPrem_WrapsInArray(t *testing.T) {
	h := &dcConfigHooks{}

	body := `{"name":"test-onprem","config":{"onpremcheck":{}}}`
	req, _ := http.NewRequest("POST", "https://example.com/api/v2/deviceclassification/client/onpremdetection", bytes.NewBufferString(body))
	req.Header = make(http.Header)

	out, err := h.BeforeRequest(dcBeforeCtx("createDeviceClassificationOnPremDetection"), req)
	if err != nil {
		t.Fatalf("BeforeRequest error: %v", err)
	}

	outBody, _ := io.ReadAll(out.Body)
	var arr []map[string]interface{}
	if err := json.Unmarshal(outBody, &arr); err != nil {
		t.Fatalf("expected array body, got: %s", outBody)
	}
	if len(arr) != 1 || arr[0]["name"] != "test-onprem" {
		t.Errorf("unexpected wrapped body: %v", arr)
	}
}

func TestDCConfig_BeforeUpdateSteering_UnwrapsToArray(t *testing.T) {
	h := &dcConfigHooks{}

	body := `{"onprem_detection_ids":[10,20,30]}`
	req, _ := http.NewRequest("PUT", "https://example.com/api/v2/deviceclassification/onpremdetection/steering/5", bytes.NewBufferString(body))
	req.Header = make(http.Header)

	out, err := h.BeforeRequest(dcBeforeCtx("updateDeviceClassificationSteeringMapping"), req)
	if err != nil {
		t.Fatalf("BeforeRequest error: %v", err)
	}

	outBody, _ := io.ReadAll(out.Body)
	var ids []int64
	if err := json.Unmarshal(outBody, &ids); err != nil {
		t.Fatalf("expected int array body, got: %s, err: %v", outBody, err)
	}
	if len(ids) != 3 || ids[0] != 10 || ids[1] != 20 || ids[2] != 30 {
		t.Errorf("unexpected ids: %v", ids)
	}
}

func TestDCConfig_AfterCreateRule_FindsRuleByName(t *testing.T) {
	rules := []map[string]interface{}{
		{"id": 1, "name": "other-rule", "label": "Windows", "os": "windows", "conditions": map[string]interface{}{}},
		{"id": 42, "name": "test-rule", "label": "Windows", "os": "windows", "conditions": map[string]interface{}{}},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(rules)
	}))
	defer server.Close()

	h := &dcConfigHooks{}

	origReq, _ := http.NewRequest("POST", server.URL+"/api/v2/deviceclassification/rules", nil)
	origReq.Header = make(http.Header)
	origReq.Header.Set("X-DC-Rule-Name", "test-rule")
	origReq.Header.Set("X-DC-Rule-Label", "Windows")
	origReq.Header.Set("Netskope-Api-Token", "test-token")

	res := &http.Response{
		StatusCode: 201,
		Body:       io.NopCloser(bytes.NewReader([]byte{})),
		Request:    origReq,
	}

	out, err := h.AfterSuccess(dcAfterCtx("createDeviceClassificationRule"), res)
	if err != nil {
		t.Fatalf("AfterSuccess error: %v", err)
	}

	outBody, _ := io.ReadAll(out.Body)
	var rule map[string]interface{}
	if err := json.Unmarshal(outBody, &rule); err != nil {
		t.Fatalf("expected rule object, got: %s", outBody)
	}
	if rule["name"] != "test-rule" {
		t.Errorf("expected name=test-rule, got %v", rule["name"])
	}
	if rule["id"] != float64(42) {
		t.Errorf("expected id=42, got %v", rule["id"])
	}
}

func TestDCConfig_AfterCreateRule_ErrorIfNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{"id": 1, "name": "other-rule", "label": "Windows"},
		})
	}))
	defer server.Close()

	// No delays so the test doesn't sleep through retries.
	h := &dcConfigHooks{ruleCreateRetryDelays: []time.Duration{}}

	origReq, _ := http.NewRequest("POST", server.URL+"/api/v2/deviceclassification/rules", nil)
	origReq.Header = make(http.Header)
	origReq.Header.Set("X-DC-Rule-Name", "missing-rule")
	origReq.Header.Set("X-DC-Rule-Label", "Windows")

	res := &http.Response{
		StatusCode: 201,
		Body:       io.NopCloser(bytes.NewReader([]byte{})),
		Request:    origReq,
	}

	_, err := h.AfterSuccess(dcAfterCtx("createDeviceClassificationRule"), res)
	if err == nil {
		t.Fatal("expected error when rule not found, got nil")
	}
}

func TestDCConfig_AfterCreateOnPrem_FetchesById(t *testing.T) {
	itemJSON := `{"id":6500,"name":"corp-network","config":{"onpremcheck":{}}}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Should receive GET /api/v2/deviceclassification/client/onpremdetection/6500
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Write([]byte(itemJSON))
	}))
	defer server.Close()

	h := &dcConfigHooks{}

	origReq, _ := http.NewRequest("POST", server.URL+"/api/v2/deviceclassification/client/onpremdetection", nil)
	origReq.Header = make(http.Header)
	origReq.Header.Set("Netskope-Api-Token", "test-token")

	createBody := `{"status":true,"data":[6500]}`
	res := &http.Response{
		StatusCode: 201,
		Body:       io.NopCloser(bytes.NewReader([]byte(createBody))),
		Request:    origReq,
	}

	out, err := h.AfterSuccess(dcAfterCtx("createDeviceClassificationOnPremDetection"), res)
	if err != nil {
		t.Fatalf("AfterSuccess error: %v", err)
	}

	outBody, _ := io.ReadAll(out.Body)
	var item map[string]interface{}
	if err := json.Unmarshal(outBody, &item); err != nil {
		t.Fatalf("expected item object, got: %s", outBody)
	}
	if item["id"] != float64(6500) {
		t.Errorf("expected id=6500, got %v", item["id"])
	}
	if item["name"] != "corp-network" {
		t.Errorf("expected name=corp-network, got %v", item["name"])
	}
}

func TestDCConfig_AfterUpdateRule_FetchesFromPath(t *testing.T) {
	ruleJSON := `{"id":99,"name":"updated-rule","label":"Mac","os":"mac","conditions":{}}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(ruleJSON))
	}))
	defer server.Close()

	h := &dcConfigHooks{}

	origReq, _ := http.NewRequest("PUT", server.URL+"/api/v2/deviceclassification/rules/99", nil)
	origReq.Header = make(http.Header)
	origReq.Header.Set("Netskope-Api-Token", "test-token")

	res := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader([]byte{})),
		Request:    origReq,
	}

	out, err := h.AfterSuccess(dcAfterCtx("updateDeviceClassificationRule"), res)
	if err != nil {
		t.Fatalf("AfterSuccess error: %v", err)
	}

	outBody, _ := io.ReadAll(out.Body)
	var rule map[string]interface{}
	if err := json.Unmarshal(outBody, &rule); err != nil {
		t.Fatalf("expected rule object, got: %s", outBody)
	}
	if rule["name"] != "updated-rule" {
		t.Errorf("expected name=updated-rule, got %v", rule["name"])
	}
}

func TestDCConfig_AfterListRules_WrapsInRulesKey(t *testing.T) {
	h := &dcConfigHooks{}

	rawArray := `[{"id":1,"name":"rule1"},{"id":2,"name":"rule2"}]`
	origReq, _ := http.NewRequest("GET", "https://example.com/api/v2/deviceclassification/rules", nil)
	res := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader([]byte(rawArray))),
		Request:    origReq,
	}

	out, err := h.AfterSuccess(dcAfterCtx("listDeviceClassificationRules"), res)
	if err != nil {
		t.Fatalf("AfterSuccess error: %v", err)
	}

	outBody, _ := io.ReadAll(out.Body)
	var wrapped map[string]interface{}
	if err := json.Unmarshal(outBody, &wrapped); err != nil {
		t.Fatalf("expected object, got: %s", outBody)
	}
	rules, ok := wrapped["rules"]
	if !ok {
		t.Errorf("expected 'rules' key in response, got: %v", wrapped)
	}
	if len(rules.([]interface{})) != 2 {
		t.Errorf("expected 2 rules, got %v", rules)
	}
}

func TestDCConfig_AfterListOnPrem_WrapsInOnpremdetectionKey(t *testing.T) {
	h := &dcConfigHooks{}

	rawArray := `[{"id":1,"name":"label1"},{"id":2,"name":"label2"}]`
	origReq, _ := http.NewRequest("GET", "https://example.com/api/v2/deviceclassification/client/onpremdetection", nil)
	res := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader([]byte(rawArray))),
		Request:    origReq,
	}

	out, err := h.AfterSuccess(dcAfterCtx("listDeviceClassificationOnPremDetection"), res)
	if err != nil {
		t.Fatalf("AfterSuccess error: %v", err)
	}

	outBody, _ := io.ReadAll(out.Body)
	var wrapped map[string]interface{}
	if err := json.Unmarshal(outBody, &wrapped); err != nil {
		t.Fatalf("expected object, got: %s", outBody)
	}
	if _, ok := wrapped["onpremdetection"]; !ok {
		t.Errorf("expected 'onpremdetection' key, got: %v", wrapped)
	}
}

func TestDCConfig_BeforeListRules_StripsOffsetWithoutLimit(t *testing.T) {
	h := &dcConfigHooks{}

	req, _ := http.NewRequest("GET", "https://example.com/api/v2/deviceclassification/rules?offset=0", nil)
	req.Header = make(http.Header)

	out, err := h.BeforeRequest(dcBeforeCtx("listDeviceClassificationRules"), req)
	if err != nil {
		t.Fatalf("BeforeRequest error: %v", err)
	}

	q := out.URL.Query()
	if q.Get("offset") != "" {
		t.Errorf("expected offset to be removed, got %q", q.Get("offset"))
	}
}

func TestDCConfig_BeforeListRules_KeepsOffsetWhenLimitSet(t *testing.T) {
	h := &dcConfigHooks{}

	req, _ := http.NewRequest("GET", "https://example.com/api/v2/deviceclassification/rules?offset=10&limit=50", nil)
	req.Header = make(http.Header)

	out, err := h.BeforeRequest(dcBeforeCtx("listDeviceClassificationRules"), req)
	if err != nil {
		t.Fatalf("BeforeRequest error: %v", err)
	}

	q := out.URL.Query()
	if q.Get("offset") != "10" {
		t.Errorf("expected offset=10, got %q", q.Get("offset"))
	}
	if q.Get("limit") != "50" {
		t.Errorf("expected limit=50, got %q", q.Get("limit"))
	}
}

func TestDCConfig_PassthroughForUnrelatedOps(t *testing.T) {
	h := &dcConfigHooks{}

	body := `{"foo":"bar"}`
	req, _ := http.NewRequest("GET", "https://example.com/", bytes.NewBufferString(body))
	req.Header = make(http.Header)

	out, err := h.BeforeRequest(dcBeforeCtx("someOtherOperation"), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != req {
		t.Error("expected passthrough (same request returned)")
	}

	res := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader([]byte(body))),
		Request:    req,
	}
	out2, err := h.AfterSuccess(dcAfterCtx("someOtherOperation"), res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out2 != res {
		t.Error("expected passthrough (same response returned)")
	}
}
