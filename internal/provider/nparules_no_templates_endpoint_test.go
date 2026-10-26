package provider

// TestNPARules_NoTemplatesEndpointCall guards against re-introducing calls to
// /api/v2/templates/usernotifications (#118, #124, #129). That endpoint is
// unavailable on 403 tenants and was the root cause of every NPA template drift
// issue in v0.4.11–v0.4.12.
//
// The old hook code used its own http.Client{Timeout: 10*time.Second} (not the
// SDK's configured client) to call the templates endpoint. To catch that pattern
// if it were reintroduced, this test replaces BOTH the SDK client AND
// http.DefaultTransport for the duration of the test. Any HTTP request — whether
// it goes through the SDK's configured client or a freshly constructed http.Client{}
// — is intercepted and fails the test if the path contains "/templates/".
//
// Negative control (manual): temporarily add a call to ensureNPATemplatesAPILoaded
// (or equivalent) inside hookMyPolicyBeforeRequest.go's BeforeRequest switch for
// "createNPARules". Re-run this test — it must FAIL. Restore the hook and confirm
// `git diff internal/sdk/internal/hooks/` is clean apart from Task 3.2's debug-flag
// changes.

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/netskopeoss/terraform-provider-netskope/internal/sdk"
	"github.com/netskopeoss/terraform-provider-netskope/internal/sdk/models/operations"
	"github.com/netskopeoss/terraform-provider-netskope/internal/sdk/models/shared"
)

const cannedBlockRuleResponse = `{"data":{"rule_id":"1","rule_name":"test-block-rule","enabled":"1","rule_data":{"policy_type":"private-app","match_criteria_action":{"action_name":"block","emit_alert":true,"template":"block_page.html"},"access_method":["Client"]}},"status":"success"}`

// templatesGuardClient implements sdk.HTTPClient. It records every URL path and
// fails the test immediately if any path contains "/templates/".
type templatesGuardClient struct {
	t     *testing.T
	paths *[]string
}

func (c *templatesGuardClient) Do(req *http.Request) (*http.Response, error) {
	c.recordAndGuard(req.URL.Path)
	return &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(cannedBlockRuleResponse)),
	}, nil
}

func (c *templatesGuardClient) recordAndGuard(path string) {
	*c.paths = append(*c.paths, path)
	if strings.Contains(path, "/templates/") {
		c.t.Errorf("forbidden endpoint called: %s — provider must not call /templates/ (see #118, #124, #129)", path)
	}
}

// templatesGuardTransport implements http.RoundTripper so that hooks creating
// their own http.Client{} are also intercepted. The old ensureNPATemplatesAPILoaded
// used http.Client{Timeout: 10*time.Second}, which goes through http.DefaultTransport.
type templatesGuardTransport struct {
	t     *testing.T
	paths *[]string
}

func (tr *templatesGuardTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	path := req.URL.Path
	*tr.paths = append(*tr.paths, "(default-transport) "+path)
	if strings.Contains(path, "/templates/") {
		tr.t.Errorf("forbidden endpoint called via default transport: %s — provider must not call /templates/ (see #118, #124, #129)", path)
	}
	return &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(cannedBlockRuleResponse)),
	}, nil
}

func TestNPARules_NoTemplatesEndpointCall(t *testing.T) {
	var recordedPaths []string

	fake := &templatesGuardClient{t: t, paths: &recordedPaths}

	// Also guard http.DefaultTransport so hook-level http.Client{} calls are caught.
	origTransport := http.DefaultTransport
	http.DefaultTransport = &templatesGuardTransport{t: t, paths: &recordedPaths}
	t.Cleanup(func() { http.DefaultTransport = origTransport })

	client := sdk.New(
		sdk.WithServerURL("http://fake/api/v2"),
		sdk.WithClient(fake),
	)

	actionName := shared.ActionNameBlock
	emitAlert := true
	tmpl := "Default Template"
	enabled := "1"
	groupID := "1"
	ruleName := "test-block-rule"
	policyType := shared.PolicyTypePrivateApp

	ruleBody := shared.NpaPolicyRequest{
		RuleName: &ruleName,
		Enabled:  &enabled,
		GroupID:  &groupID,
		RuleData: &shared.NpaPolicyRuleData{
			PolicyType: &policyType,
			MatchCriteriaAction: &shared.MatchCriteriaAction{
				ActionName: &actionName,
				EmitAlert:  &emitAlert,
				Template:   &tmpl,
			},
			AccessMethod: []shared.AccessMethod{shared.AccessMethodClient},
		},
	}

	ctx := context.Background()

	// CreateNPARules — must not trigger /templates/
	_, _ = client.CreateNPARules(ctx, ruleBody)

	// GetNPARules — must not trigger /templates/
	_, _ = client.GetNPARules(ctx, operations.GetNPARulesRequest{ID: "1"})

	// UpdateNPARules — must not trigger /templates/
	_, _ = client.UpdateNPARules(ctx, operations.UpdateNPARulesRequest{
		ID:               "1",
		NpaPolicyRequest: ruleBody,
	})

	if len(recordedPaths) == 0 {
		t.Error("no HTTP requests were made; expected at least 3 (create, get, update)")
	}

	// Secondary assertion on recorded paths (Do and RoundTrip already fail
	// immediately above, but this gives a clear summary on test failure).
	for _, p := range recordedPaths {
		if strings.Contains(p, "/templates/") {
			t.Errorf("recorded path contains /templates/: %s", p)
		}
	}

	t.Logf("recorded %d request paths: %v", len(recordedPaths), recordedPaths)
}
