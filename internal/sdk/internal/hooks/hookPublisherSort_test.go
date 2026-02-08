package hooks

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

// buildFakeAppResponseFull creates a fake HTTP response with publishers, protocols, and tags.
// This simulates what the Netskope API returns for single-app endpoints.
func buildFakeAppResponseFull(t *testing.T, publishers []map[string]interface{}, protocols []map[string]interface{}, tags []map[string]interface{}) *http.Response {
	t.Helper()
	body := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"app_name":                      "test-app",
			"host":                          "test.example.com",
			"protocols":                     protocols,
			"service_publisher_assignments": publishers,
			"tags":                          tags,
		},
	}
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal test response: %v", err)
	}
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(string(b))),
		Header:     make(http.Header),
	}
}

// buildFakeAppResponse creates a fake HTTP response with just publisher assignments.
func buildFakeAppResponse(t *testing.T, publishers []map[string]interface{}) *http.Response {
	t.Helper()
	return buildFakeAppResponseFull(t, publishers, []map[string]interface{}{}, []map[string]interface{}{})
}

// buildFakeBulkAppResponseFull creates a fake bulk HTTP response for the list endpoint.
type bulkAppInput struct {
	Publishers []map[string]interface{}
	Protocols  []map[string]interface{}
	Tags       []map[string]interface{}
}

func buildFakeBulkAppResponseFull(t *testing.T, apps []bulkAppInput) *http.Response {
	t.Helper()
	var appData []map[string]interface{}
	for _, app := range apps {
		appData = append(appData, map[string]interface{}{
			"app_name":                      "test-app",
			"host":                          "test.example.com",
			"protocols":                     app.Protocols,
			"service_publisher_assignments": app.Publishers,
			"tags":                          app.Tags,
		})
	}
	body := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"private_apps": appData,
		},
	}
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal test response: %v", err)
	}
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(string(b))),
		Header:     make(http.Header),
	}
}

// buildFakeBulkAppResponse creates a fake bulk HTTP response with just publishers per app.
func buildFakeBulkAppResponse(t *testing.T, apps [][]map[string]interface{}) *http.Response {
	t.Helper()
	var inputs []bulkAppInput
	for _, publishers := range apps {
		inputs = append(inputs, bulkAppInput{
			Publishers: publishers,
			Protocols:  []map[string]interface{}{},
			Tags:       []map[string]interface{}{},
		})
	}
	return buildFakeBulkAppResponseFull(t, inputs)
}

func hookCtxForOp(operationID string) AfterSuccessContext {
	return AfterSuccessContext{
		HookContext: HookContext{
			OperationID: operationID,
		},
	}
}

// parseDataFromResponse reads the response body and returns the "data" object.
func parseDataFromResponse(t *testing.T, res *http.Response) map[string]interface{} {
	t.Helper()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	return parsed["data"].(map[string]interface{})
}

// parsePublishersFromResponse reads the response body and extracts publisher assignments.
func parsePublishersFromResponse(t *testing.T, res *http.Response) []map[string]interface{} {
	t.Helper()
	data := parseDataFromResponse(t, res)
	raw := data["service_publisher_assignments"].([]interface{})
	var result []map[string]interface{}
	for _, r := range raw {
		result = append(result, r.(map[string]interface{}))
	}
	return result
}

// parseProtocolsFromResponse reads the response body and extracts protocols.
func parseProtocolsFromResponse(t *testing.T, res *http.Response) []map[string]interface{} {
	t.Helper()
	data := parseDataFromResponse(t, res)
	raw := data["protocols"].([]interface{})
	var result []map[string]interface{}
	for _, r := range raw {
		result = append(result, r.(map[string]interface{}))
	}
	return result
}

// parseTagsFromResponse reads the response body and extracts tags.
func parseTagsFromResponse(t *testing.T, res *http.Response) []map[string]interface{} {
	t.Helper()
	data := parseDataFromResponse(t, res)
	raw := data["tags"].([]interface{})
	var result []map[string]interface{}
	for _, r := range raw {
		result = append(result, r.(map[string]interface{}))
	}
	return result
}

// parseBulkPublishersFromResponse extracts publisher assignments from bulk response.
func parseBulkPublishersFromResponse(t *testing.T, res *http.Response) [][]map[string]interface{} {
	t.Helper()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	data := parsed["data"].(map[string]interface{})
	apps := data["private_apps"].([]interface{})
	var result [][]map[string]interface{}
	for _, app := range apps {
		appMap := app.(map[string]interface{})
		raw := appMap["service_publisher_assignments"].([]interface{})
		var pubs []map[string]interface{}
		for _, r := range raw {
			pubs = append(pubs, r.(map[string]interface{}))
		}
		result = append(result, pubs)
	}
	return result
}

// parseBulkFieldFromResponse extracts a named list field from each app in a bulk response.
func parseBulkFieldFromResponse(t *testing.T, res *http.Response, field string) [][]map[string]interface{} {
	t.Helper()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	data := parsed["data"].(map[string]interface{})
	apps := data["private_apps"].([]interface{})
	var result [][]map[string]interface{}
	for _, app := range apps {
		appMap := app.(map[string]interface{})
		raw := appMap[field].([]interface{})
		var items []map[string]interface{}
		for _, r := range raw {
			items = append(items, r.(map[string]interface{}))
		}
		result = append(result, items)
	}
	return result
}

// ============================================================================
// Publisher tests
// ============================================================================

// TestPublishersSortedByID verifies that the single-app hook sorts publishers
// by publisher_id, ensuring deterministic ordering regardless of API return order.
func TestPublishersSortedByID(t *testing.T) {
	hook := &myAppResponse{}

	// API returns publishers in reverse order (256 before 249)
	publishers := []map[string]interface{}{
		{"publisher_id": float64(256), "publisher_name": "Publisher-B"},
		{"publisher_id": float64(249), "publisher_name": "Publisher-A"},
		{"publisher_id": float64(300), "publisher_name": "Publisher-C"},
	}

	for _, opID := range []string{"getNPAPrivateApp", "createNPAPrivateApps", "updateNPAPrivateApp"} {
		t.Run(opID, func(t *testing.T) {
			res := buildFakeAppResponse(t, publishers)
			ctx := hookCtxForOp(opID)

			res, err := hook.AfterSuccess(ctx, res)
			if err != nil {
				t.Fatalf("AfterSuccess returned error: %v", err)
			}

			result := parsePublishersFromResponse(t, res)

			// Expect sorted by publisher_id: 249, 256, 300
			expectedOrder := []string{"249", "256", "300"}
			for i, expected := range expectedOrder {
				got := result[i]["publisher_id"].(string)
				if got != expected {
					t.Errorf("publisher[%d]: expected publisher_id=%s, got=%s", i, expected, got)
				}
			}
		})
	}
}

// TestPublisherNameWhitespaceTrimmed verifies that the single-app hook trims
// leading/trailing whitespace from publisher_name values.
func TestPublisherNameWhitespaceTrimmed(t *testing.T) {
	hook := &myAppResponse{}

	// API returns publisher names with leading spaces (the "ghost space" bug)
	publishers := []map[string]interface{}{
		{"publisher_id": float64(249), "publisher_name": "AWS-LSAC3-Prod-green"},
		{"publisher_id": float64(256), "publisher_name": " AWS-LSAC3-Prod-blue"},
		{"publisher_id": float64(260), "publisher_name": "  Publisher-with-tabs\t"},
	}

	res := buildFakeAppResponse(t, publishers)
	ctx := hookCtxForOp("getNPAPrivateApp")

	res, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess returned error: %v", err)
	}

	result := parsePublishersFromResponse(t, res)

	expectedNames := []string{"AWS-LSAC3-Prod-green", "AWS-LSAC3-Prod-blue", "Publisher-with-tabs"}
	for i, expected := range expectedNames {
		got := result[i]["publisher_name"].(string)
		if got != expected {
			t.Errorf("publisher[%d]: expected publisher_name=%q, got=%q", i, expected, got)
		}
	}
}

// TestBulkPublishersSortedByID verifies the bulk-app hook sorts publishers
// by publisher_id for every app in the response.
func TestBulkPublishersSortedByID(t *testing.T) {
	hook := &myBulkAppResponse{}

	apps := [][]map[string]interface{}{
		{
			{"publisher_id": float64(300), "publisher_name": "Publisher-C"},
			{"publisher_id": float64(100), "publisher_name": "Publisher-A"},
			{"publisher_id": float64(200), "publisher_name": "Publisher-B"},
		},
		{
			{"publisher_id": float64(50), "publisher_name": "Pub-2"},
			{"publisher_id": float64(10), "publisher_name": "Pub-1"},
		},
	}

	res := buildFakeBulkAppResponse(t, apps)
	ctx := hookCtxForOp("listNPAPrivateApps")

	res, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess returned error: %v", err)
	}

	result := parseBulkPublishersFromResponse(t, res)

	// App 0: expect 100, 200, 300
	expectedApp0 := []string{"100", "200", "300"}
	for i, expected := range expectedApp0 {
		got := result[0][i]["publisher_id"].(string)
		if got != expected {
			t.Errorf("app[0] publisher[%d]: expected publisher_id=%s, got=%s", i, expected, got)
		}
	}

	// App 1: expect 10, 50
	expectedApp1 := []string{"10", "50"}
	for i, expected := range expectedApp1 {
		got := result[1][i]["publisher_id"].(string)
		if got != expected {
			t.Errorf("app[1] publisher[%d]: expected publisher_id=%s, got=%s", i, expected, got)
		}
	}
}

// TestBulkPublisherNameWhitespaceTrimmed verifies the bulk-app hook trims
// leading/trailing whitespace from publisher_name values.
func TestBulkPublisherNameWhitespaceTrimmed(t *testing.T) {
	hook := &myBulkAppResponse{}

	apps := [][]map[string]interface{}{
		{
			{"publisher_id": float64(1), "publisher_name": " Leading-Space"},
			{"publisher_id": float64(2), "publisher_name": "Trailing-Space "},
		},
	}

	res := buildFakeBulkAppResponse(t, apps)
	ctx := hookCtxForOp("listNPAPrivateApps")

	res, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess returned error: %v", err)
	}

	result := parseBulkPublishersFromResponse(t, res)

	expectedNames := []string{"Leading-Space", "Trailing-Space"}
	for i, expected := range expectedNames {
		got := result[0][i]["publisher_name"].(string)
		if got != expected {
			t.Errorf("publisher[%d]: expected publisher_name=%q, got=%q", i, expected, got)
		}
	}
}

// TestPublishersIdempotent verifies that running the hook twice on the same
// data produces identical results (important for plan stability).
func TestPublishersIdempotent(t *testing.T) {
	hook := &myAppResponse{}

	publishers := []map[string]interface{}{
		{"publisher_id": float64(256), "publisher_name": " Publisher-B"},
		{"publisher_id": float64(249), "publisher_name": "Publisher-A"},
	}

	ctx := hookCtxForOp("getNPAPrivateApp")

	// First pass
	res1 := buildFakeAppResponse(t, publishers)
	res1, err := hook.AfterSuccess(ctx, res1)
	if err != nil {
		t.Fatalf("first AfterSuccess returned error: %v", err)
	}
	result1 := parsePublishersFromResponse(t, res1)

	// Second pass with same input
	res2 := buildFakeAppResponse(t, publishers)
	res2, err = hook.AfterSuccess(ctx, res2)
	if err != nil {
		t.Fatalf("second AfterSuccess returned error: %v", err)
	}
	result2 := parsePublishersFromResponse(t, res2)

	// Both passes should produce identical output
	for i := range result1 {
		id1 := result1[i]["publisher_id"].(string)
		id2 := result2[i]["publisher_id"].(string)
		name1 := result1[i]["publisher_name"].(string)
		name2 := result2[i]["publisher_name"].(string)
		if id1 != id2 || name1 != name2 {
			t.Errorf("publisher[%d]: pass1=(%s,%s) != pass2=(%s,%s)", i, id1, name1, id2, name2)
		}
	}
}

// ============================================================================
// Protocol tests
// ============================================================================

// TestProtocolsSortedByTypeThenPort verifies that the single-app hook sorts
// protocols by type (alphabetically) then by port (numerically).
// The API returns protocols in non-deterministic order, causing perpetual diffs
// (KNOWN_API_ISSUES #11).
func TestProtocolsSortedByTypeThenPort(t *testing.T) {
	hook := &myAppResponse{}

	// API returns protocols in wrong order: UDP before TCP, high ports before low
	protocols := []map[string]interface{}{
		{"transport": "udp", "port": "123", "id": float64(4)},
		{"transport": "tcp", "port": "443", "id": float64(2)},
		{"transport": "udp", "port": "53", "id": float64(3)},
		{"transport": "tcp", "port": "22", "id": float64(1)},
	}

	res := buildFakeAppResponseFull(t, []map[string]interface{}{}, protocols, []map[string]interface{}{})
	ctx := hookCtxForOp("getNPAPrivateApp")

	res, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess returned error: %v", err)
	}

	result := parseProtocolsFromResponse(t, res)

	// Expected: tcp:22, tcp:443, udp:53, udp:123
	type expected struct {
		typ  string
		port string
	}
	expectedOrder := []expected{
		{"tcp", "22"},
		{"tcp", "443"},
		{"udp", "53"},
		{"udp", "123"},
	}

	if len(result) != len(expectedOrder) {
		t.Fatalf("expected %d protocols, got %d", len(expectedOrder), len(result))
	}

	for i, exp := range expectedOrder {
		// After the hook, transport is copied to type
		gotType := result[i]["type"].(string)
		gotPort := result[i]["port"].(string)
		if gotType != exp.typ || gotPort != exp.port {
			t.Errorf("protocol[%d]: expected %s:%s, got %s:%s", i, exp.typ, exp.port, gotType, gotPort)
		}
	}
}

// TestProtocolsSortedSameType verifies sorting when all protocols share the
// same type — should sort by port ascending.
func TestProtocolsSortedSameType(t *testing.T) {
	hook := &myAppResponse{}

	protocols := []map[string]interface{}{
		{"transport": "tcp", "port": "8080", "id": float64(3)},
		{"transport": "tcp", "port": "22", "id": float64(1)},
		{"transport": "tcp", "port": "443", "id": float64(2)},
	}

	res := buildFakeAppResponseFull(t, []map[string]interface{}{}, protocols, []map[string]interface{}{})
	ctx := hookCtxForOp("getNPAPrivateApp")

	res, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess returned error: %v", err)
	}

	result := parseProtocolsFromResponse(t, res)

	expectedPorts := []string{"22", "443", "8080"}
	for i, expected := range expectedPorts {
		got := result[i]["port"].(string)
		if got != expected {
			t.Errorf("protocol[%d]: expected port=%s, got=%s", i, expected, got)
		}
	}
}

// TestBulkProtocolsSorted verifies the bulk-app hook sorts protocols for each app.
func TestBulkProtocolsSorted(t *testing.T) {
	hook := &myBulkAppResponse{}

	apps := []bulkAppInput{
		{
			Publishers: []map[string]interface{}{},
			Protocols: []map[string]interface{}{
				{"transport": "tcp", "port": "443", "id": float64(2)},
				{"transport": "tcp", "port": "22", "id": float64(1)},
			},
			Tags: []map[string]interface{}{},
		},
		{
			Publishers: []map[string]interface{}{},
			Protocols: []map[string]interface{}{
				{"transport": "udp", "port": "53", "id": float64(2)},
				{"transport": "tcp", "port": "80", "id": float64(1)},
			},
			Tags: []map[string]interface{}{},
		},
	}

	res := buildFakeBulkAppResponseFull(t, apps)
	ctx := hookCtxForOp("listNPAPrivateApps")

	res, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess returned error: %v", err)
	}

	result := parseBulkFieldFromResponse(t, res, "protocols")

	// App 0: tcp:22, tcp:443
	if result[0][0]["port"].(string) != "22" || result[0][1]["port"].(string) != "443" {
		t.Errorf("app[0]: expected tcp:22, tcp:443, got %s:%s, %s:%s",
			result[0][0]["type"], result[0][0]["port"],
			result[0][1]["type"], result[0][1]["port"])
	}

	// App 1: tcp:80, udp:53
	if result[1][0]["type"].(string) != "tcp" || result[1][1]["type"].(string) != "udp" {
		t.Errorf("app[1]: expected tcp:80 then udp:53, got %s:%s, %s:%s",
			result[1][0]["type"], result[1][0]["port"],
			result[1][1]["type"], result[1][1]["port"])
	}
}

// ============================================================================
// Tag tests
// ============================================================================

// TestTagsSortedByID verifies that the single-app hook sorts tags by tag_id,
// ensuring deterministic ordering regardless of API return order.
func TestTagsSortedByID(t *testing.T) {
	hook := &myAppResponse{}

	tags := []map[string]interface{}{
		{"tag_id": float64(26), "tag_name": "npasp-sre"},
		{"tag_id": float64(10), "tag_name": "npasp-dev"},
		{"tag_id": float64(22), "tag_name": "npasp-terraform-managed"},
	}

	res := buildFakeAppResponseFull(t, []map[string]interface{}{}, []map[string]interface{}{}, tags)
	ctx := hookCtxForOp("getNPAPrivateApp")

	res, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess returned error: %v", err)
	}

	result := parseTagsFromResponse(t, res)

	// Expected: sorted by tag_id: 10, 22, 26
	expectedIDs := []float64{10, 22, 26}
	expectedNames := []string{"npasp-dev", "npasp-terraform-managed", "npasp-sre"}
	for i, expectedID := range expectedIDs {
		gotID := result[i]["tag_id"].(float64)
		gotName := result[i]["tag_name"].(string)
		if gotID != expectedID {
			t.Errorf("tag[%d]: expected tag_id=%.0f, got=%.0f", i, expectedID, gotID)
		}
		if gotName != expectedNames[i] {
			t.Errorf("tag[%d]: expected tag_name=%q, got=%q", i, expectedNames[i], gotName)
		}
	}
}

// TestBulkTagsSorted verifies the bulk-app hook sorts tags for each app.
func TestBulkTagsSorted(t *testing.T) {
	hook := &myBulkAppResponse{}

	apps := []bulkAppInput{
		{
			Publishers: []map[string]interface{}{},
			Protocols:  []map[string]interface{}{},
			Tags: []map[string]interface{}{
				{"tag_id": float64(30), "tag_name": "tag-c"},
				{"tag_id": float64(10), "tag_name": "tag-a"},
				{"tag_id": float64(20), "tag_name": "tag-b"},
			},
		},
	}

	res := buildFakeBulkAppResponseFull(t, apps)
	ctx := hookCtxForOp("listNPAPrivateApps")

	res, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess returned error: %v", err)
	}

	result := parseBulkFieldFromResponse(t, res, "tags")

	// Expect: 10, 20, 30
	expectedIDs := []float64{10, 20, 30}
	for i, expectedID := range expectedIDs {
		gotID := result[0][i]["tag_id"].(float64)
		if gotID != expectedID {
			t.Errorf("tag[%d]: expected tag_id=%.0f, got=%.0f", i, expectedID, gotID)
		}
	}
}

// ============================================================================
// Combined / idempotency tests
// ============================================================================

// TestAllFieldsIdempotent verifies that running the hook twice on a response
// with publishers, protocols, and tags produces identical results both times.
func TestAllFieldsIdempotent(t *testing.T) {
	hook := &myAppResponse{}

	publishers := []map[string]interface{}{
		{"publisher_id": float64(256), "publisher_name": " Publisher-B"},
		{"publisher_id": float64(249), "publisher_name": "Publisher-A"},
	}
	protocols := []map[string]interface{}{
		{"transport": "udp", "port": "53", "id": float64(2)},
		{"transport": "tcp", "port": "443", "id": float64(1)},
	}
	tags := []map[string]interface{}{
		{"tag_id": float64(22), "tag_name": "tag-b"},
		{"tag_id": float64(10), "tag_name": "tag-a"},
	}

	ctx := hookCtxForOp("getNPAPrivateApp")

	// First pass
	res1 := buildFakeAppResponseFull(t, publishers, protocols, tags)
	res1, err := hook.AfterSuccess(ctx, res1)
	if err != nil {
		t.Fatalf("first pass error: %v", err)
	}
	pubs1 := parsePublishersFromResponse(t, res1)

	res1b := buildFakeAppResponseFull(t, publishers, protocols, tags)
	res1b, err = hook.AfterSuccess(ctx, res1b)
	if err != nil {
		t.Fatalf("second pass error: %v", err)
	}
	pubs2 := parsePublishersFromResponse(t, res1b)

	for i := range pubs1 {
		if pubs1[i]["publisher_id"] != pubs2[i]["publisher_id"] || pubs1[i]["publisher_name"] != pubs2[i]["publisher_name"] {
			t.Errorf("publisher[%d] not idempotent: %v vs %v", i, pubs1[i], pubs2[i])
		}
	}

	// Check protocols idempotency
	res2a := buildFakeAppResponseFull(t, publishers, protocols, tags)
	res2a, _ = hook.AfterSuccess(ctx, res2a)
	protos1 := parseProtocolsFromResponse(t, res2a)

	res2b := buildFakeAppResponseFull(t, publishers, protocols, tags)
	res2b, _ = hook.AfterSuccess(ctx, res2b)
	protos2 := parseProtocolsFromResponse(t, res2b)

	for i := range protos1 {
		if protos1[i]["type"] != protos2[i]["type"] || protos1[i]["port"] != protos2[i]["port"] {
			t.Errorf("protocol[%d] not idempotent: %v vs %v", i, protos1[i], protos2[i])
		}
	}

	// Check tags idempotency
	res3a := buildFakeAppResponseFull(t, publishers, protocols, tags)
	res3a, _ = hook.AfterSuccess(ctx, res3a)
	tags1 := parseTagsFromResponse(t, res3a)

	res3b := buildFakeAppResponseFull(t, publishers, protocols, tags)
	res3b, _ = hook.AfterSuccess(ctx, res3b)
	tags2 := parseTagsFromResponse(t, res3b)

	for i := range tags1 {
		if tags1[i]["tag_id"] != tags2[i]["tag_id"] || tags1[i]["tag_name"] != tags2[i]["tag_name"] {
			t.Errorf("tag[%d] not idempotent: %v vs %v", i, tags1[i], tags2[i])
		}
	}
}

// ============================================================================
// Edge case tests
// ============================================================================

// TestEmptyListsDoNotPanic verifies that the hook handles empty lists gracefully.
// This is critical because sort.Slice on nil/empty slices should not panic.
func TestEmptyListsDoNotPanic(t *testing.T) {
	hook := &myAppResponse{}

	// All lists are empty
	res := buildFakeAppResponseFull(t, []map[string]interface{}{}, []map[string]interface{}{}, []map[string]interface{}{})
	ctx := hookCtxForOp("getNPAPrivateApp")

	res, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess returned error on empty lists: %v", err)
	}

	// Verify the response is still valid and lists are empty
	data := parseDataFromResponse(t, res)
	if pubs, ok := data["service_publisher_assignments"].([]interface{}); !ok || len(pubs) != 0 {
		t.Errorf("expected empty publishers, got %v", data["service_publisher_assignments"])
	}
	if protos, ok := data["protocols"].([]interface{}); !ok || len(protos) != 0 {
		t.Errorf("expected empty protocols, got %v", data["protocols"])
	}
	if tags, ok := data["tags"].([]interface{}); !ok || len(tags) != 0 {
		t.Errorf("expected empty tags, got %v", data["tags"])
	}
}

// TestBulkEmptyListsDoNotPanic verifies that the bulk hook handles empty lists.
func TestBulkEmptyListsDoNotPanic(t *testing.T) {
	hook := &myBulkAppResponse{}

	apps := []bulkAppInput{
		{
			Publishers: []map[string]interface{}{},
			Protocols:  []map[string]interface{}{},
			Tags:       []map[string]interface{}{},
		},
	}

	res := buildFakeBulkAppResponseFull(t, apps)
	ctx := hookCtxForOp("listNPAPrivateApps")

	res, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess returned error on empty lists: %v", err)
	}

	// Just verify it didn't panic and returned valid response
	if res == nil {
		t.Fatal("expected non-nil response")
	}
}

// TestNilPublisherIdHandling verifies behavior when publisher_id is nil.
// The hook should handle this gracefully without panicking during sort.
func TestNilPublisherIdHandling(t *testing.T) {
	hook := &myAppResponse{}

	// One publisher with nil ID, one with valid ID
	publishers := []map[string]interface{}{
		{"publisher_id": float64(100), "publisher_name": "Valid-Publisher"},
		{"publisher_id": nil, "publisher_name": "Nil-ID-Publisher"},
	}

	res := buildFakeAppResponse(t, publishers)
	ctx := hookCtxForOp("getNPAPrivateApp")

	res, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess returned error with nil publisher_id: %v", err)
	}

	result := parsePublishersFromResponse(t, res)
	if len(result) != 2 {
		t.Fatalf("expected 2 publishers, got %d", len(result))
	}

	// Verify whitespace was still trimmed on the valid publisher
	for _, pub := range result {
		name := pub["publisher_name"].(string)
		if name != strings.TrimSpace(name) {
			t.Errorf("publisher_name not trimmed: %q", name)
		}
	}
}

// TestIntTypedPublisherId verifies that the int type branch is exercised.
// JSON unmarshaling typically produces float64, but the code handles int too.
func TestIntTypedPublisherId(t *testing.T) {
	hook := &myAppResponse{}

	// Use int type instead of float64
	publishers := []map[string]interface{}{
		{"publisher_id": int(200), "publisher_name": "Publisher-B"},
		{"publisher_id": int(100), "publisher_name": "Publisher-A"},
	}

	res := buildFakeAppResponse(t, publishers)
	ctx := hookCtxForOp("getNPAPrivateApp")

	res, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess returned error: %v", err)
	}

	result := parsePublishersFromResponse(t, res)

	// Should be sorted: 100, 200
	expectedOrder := []string{"100", "200"}
	for i, expected := range expectedOrder {
		got := result[i]["publisher_id"].(string)
		if got != expected {
			t.Errorf("publisher[%d]: expected publisher_id=%s, got=%s", i, expected, got)
		}
	}
}

// TestStringPublisherIdPassthrough verifies behavior when publisher_id is already a string.
// The type switch doesn't handle string, so it should pass through unchanged.
func TestStringPublisherIdPassthrough(t *testing.T) {
	hook := &myAppResponse{}

	// Publisher ID already a string (unusual but possible)
	publishers := []map[string]interface{}{
		{"publisher_id": "200", "publisher_name": "Publisher-B"},
		{"publisher_id": "100", "publisher_name": "Publisher-A"},
	}

	res := buildFakeAppResponse(t, publishers)
	ctx := hookCtxForOp("getNPAPrivateApp")

	res, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess returned error: %v", err)
	}

	result := parsePublishersFromResponse(t, res)

	// Should still be sorted numerically via strconv.Atoi fallback: 100, 200
	expectedOrder := []string{"100", "200"}
	for i, expected := range expectedOrder {
		got := result[i]["publisher_id"].(string)
		if got != expected {
			t.Errorf("publisher[%d]: expected publisher_id=%s, got=%s", i, expected, got)
		}
	}
}

// TestNonNumericPortFallback verifies the string comparison fallback for non-numeric ports.
func TestNonNumericPortFallback(t *testing.T) {
	hook := &myAppResponse{}

	// Mix of numeric and non-numeric ports
	protocols := []map[string]interface{}{
		{"transport": "tcp", "port": "8080", "id": float64(3)},
		{"transport": "tcp", "port": "1-1024", "id": float64(1)}, // Port range
		{"transport": "tcp", "port": "443", "id": float64(2)},
	}

	res := buildFakeAppResponseFull(t, []map[string]interface{}{}, protocols, []map[string]interface{}{})
	ctx := hookCtxForOp("getNPAPrivateApp")

	res, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess returned error: %v", err)
	}

	result := parseProtocolsFromResponse(t, res)

	// String comparison: "1-1024" < "443" < "8080"
	expectedPorts := []string{"1-1024", "443", "8080"}
	for i, expected := range expectedPorts {
		got := result[i]["port"].(string)
		if got != expected {
			t.Errorf("protocol[%d]: expected port=%s, got=%s", i, expected, got)
		}
	}
}

// TestAppNameBracketTrimming verifies that app_name has brackets trimmed.
func TestAppNameBracketTrimming(t *testing.T) {
	hook := &myAppResponse{}

	// Build response with bracketed app name
	body := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"app_name":                      "[my-test-app]",
			"host":                          "test.example.com",
			"protocols":                     []map[string]interface{}{},
			"service_publisher_assignments": []map[string]interface{}{},
			"tags":                          []map[string]interface{}{},
		},
	}
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal test response: %v", err)
	}
	res := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(string(b))),
		Header:     make(http.Header),
	}

	ctx := hookCtxForOp("getNPAPrivateApp")
	res, err = hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess returned error: %v", err)
	}

	data := parseDataFromResponse(t, res)
	appName := data["app_name"].(string)
	if appName != "my-test-app" {
		t.Errorf("expected app_name='my-test-app', got=%q", appName)
	}
}

// TestNonMatchingOperationIdPassthrough verifies that unrecognized operation IDs
// pass through without modification.
func TestNonMatchingOperationIdPassthrough(t *testing.T) {
	hook := &myAppResponse{}

	originalBody := `{"status":"success","data":{"app_name":"[unchanged]","protocols":[]}}`
	res := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(originalBody)),
		Header:     make(http.Header),
	}

	// Use an operation ID that doesn't match any handled cases
	ctx := hookCtxForOp("someOtherOperation")
	res, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess returned error: %v", err)
	}

	// Body should be unchanged (brackets still present)
	resultBody, _ := io.ReadAll(res.Body)
	if string(resultBody) != originalBody {
		t.Errorf("expected body unchanged, got=%s", string(resultBody))
	}
}
