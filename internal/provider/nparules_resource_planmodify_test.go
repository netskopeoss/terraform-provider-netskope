package provider

import (
	"context"
	"encoding/json"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	tfTypes "github.com/netskopeoss/terraform-provider-netskope/internal/provider/types"
)

// =============================================================================
// fakeTemplateStore — map-backed templateStore for unit tests.
// Records every SetKey call so tests can verify the key was written.
// =============================================================================

type fakeTemplateStore struct {
	store    map[string][]byte
	setCalls []string // keys passed to SetKey, in order
}

func (f *fakeTemplateStore) GetKey(_ context.Context, key string) ([]byte, diag.Diagnostics) {
	if f.store == nil {
		return nil, nil
	}
	return f.store[key], nil
}

func (f *fakeTemplateStore) SetKey(_ context.Context, key string, value []byte) diag.Diagnostics {
	if f.store == nil {
		f.store = make(map[string][]byte)
	}
	f.store[key] = value
	f.setCalls = append(f.setCalls, key)
	return nil
}

// seed stores a JSON-encoded string under key, matching what suppressTemplateDrift
// writes via json.Marshal. Tests that seed the store and then run suppressTemplateDrift
// must use this so GetKey returns bytes that json.Unmarshal can decode.
func (f *fakeTemplateStore) seed(key, value string) {
	if f.store == nil {
		f.store = make(map[string][]byte)
	}
	encoded, _ := json.Marshal(value)
	f.store[key] = encoded
}

// decodeStoreStr JSON-decodes a byte slice from the fake store back to a string.
// Returns "" on nil input or decode error.
func decodeStoreStr(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return string(b) // return raw for debugging
	}
	return s
}

// =============================================================================
// Framework object helpers
// =============================================================================

// npaRulesSchema returns the schema for NPARulesResource.
func npaRulesSchema(ctx context.Context) resourceschema.Schema {
	var resp resource.SchemaResponse
	r := &NPARulesResource{}
	r.Schema(ctx, resource.SchemaRequest{}, &resp)
	return resp.Schema
}

// npaMinModel returns a minimal NPARulesResourceModel with only the template
// field set. All other fields are zero (null) so tfsdk.State.Set succeeds.
func npaMinModel(template fwtypes.String) *NPARulesResourceModel {
	return &NPARulesResourceModel{
		RuleData: &tfTypes.NpaPolicyRuleData{
			MatchCriteriaAction: &tfTypes.MatchCriteriaAction{
				Template: template,
			},
		},
	}
}

// npaMinModelDesc returns a minimal model with template and description set.
func npaMinModelDesc(template fwtypes.String, desc fwtypes.String) *NPARulesResourceModel {
	return &NPARulesResourceModel{
		Description: desc,
		RuleData: &tfTypes.NpaPolicyRuleData{
			MatchCriteriaAction: &tfTypes.MatchCriteriaAction{
				Template: template,
			},
		},
	}
}

// buildFWState marshals model into a tfsdk.State.
func buildFWState(ctx context.Context, t *testing.T, sch resourceschema.Schema, model interface{}) tfsdk.State {
	t.Helper()
	st := tfsdk.State{Schema: sch}
	if d := st.Set(ctx, model); d.HasError() {
		t.Fatalf("buildFWState: %v", d)
	}
	return st
}

// buildFWPlan marshals model into a tfsdk.Plan.
func buildFWPlan(ctx context.Context, t *testing.T, sch resourceschema.Schema, model interface{}) tfsdk.Plan {
	t.Helper()
	p := tfsdk.Plan{Schema: sch}
	if d := p.Set(ctx, model); d.HasError() {
		t.Fatalf("buildFWPlan: %v", d)
	}
	return p
}

// buildFWConfig builds a tfsdk.Config from model by sharing the Raw from a State.
// tfsdk.Config has no Set method; this is the standard workaround.
func buildFWConfig(ctx context.Context, t *testing.T, sch resourceschema.Schema, model interface{}) tfsdk.Config {
	t.Helper()
	return tfsdk.Config{Schema: sch, Raw: buildFWState(ctx, t, sch, model).Raw}
}

// nullFWState returns a null (create-plan) tfsdk.State.
func nullFWState(ctx context.Context, sch resourceschema.Schema) tfsdk.State {
	return tfsdk.State{Schema: sch, Raw: tftypes.NewValue(sch.Type().TerraformType(ctx), nil)}
}

// nullFWPlan returns a null (destroy-plan) tfsdk.Plan.
func nullFWPlan(ctx context.Context, sch resourceschema.Schema) tfsdk.Plan {
	return tfsdk.Plan{Schema: sch, Raw: tftypes.NewValue(sch.Type().TerraformType(ctx), nil)}
}

// tplPath is the attribute path to rule_data.match_criteria_action.template.
var tplPath = path.Root("rule_data").AtName("match_criteria_action").AtName("template")

// privateAppsAttrPath is the attribute path to rule_data.private_apps.
var privateAppsAttrPath = path.Root("rule_data").AtName("private_apps")

// readPlanTemplate reads rule_data.match_criteria_action.template from resp.Plan.
func readPlanTemplate(ctx context.Context, t *testing.T, resp *resource.ModifyPlanResponse) string {
	t.Helper()
	var v fwtypes.String
	if d := resp.Plan.GetAttribute(ctx, tplPath, &v); d.HasError() {
		t.Fatalf("readPlanTemplate: %v", d)
	}
	if v.IsNull() || v.IsUnknown() {
		return ""
	}
	return v.ValueString()
}

// callSuppressTemplateDrift is a convenience wrapper: builds framework objects from
// models and calls suppressTemplateDrift, returning the ModifyPlanResponse.
func callSuppressTemplateDrift(
	t *testing.T,
	ctx context.Context,
	sch resourceschema.Schema,
	stateModel, planModel, cfgModel *NPARulesResourceModel,
	rs, ws templateStore,
) *resource.ModifyPlanResponse {
	t.Helper()
	state := buildFWState(ctx, t, sch, stateModel)
	plan := buildFWPlan(ctx, t, sch, planModel)
	cfg := buildFWConfig(ctx, t, sch, cfgModel)
	req := resource.ModifyPlanRequest{State: state, Plan: plan, Config: cfg}
	resp := &resource.ModifyPlanResponse{Plan: plan}
	suppressTemplateDrift(ctx, req, resp, rs, ws)
	return resp
}

// =============================================================================
// String list normalization tests
//
// These test the core logic used by normalizeStringListAttr:
// sort both value lists and compare.
// =============================================================================

func TestStringList_SameSetDifferentOrder(t *testing.T) {
	plan := []string{"app-b", "app-a", "app-c"}
	state := []string{"app-a", "app-b", "app-c"}

	sort.Strings(plan)
	sort.Strings(state)

	for i := range plan {
		if plan[i] != state[i] {
			t.Fatalf("keys should match after sort, but position %d differs: %s vs %s", i, plan[i], state[i])
		}
	}
}

func TestStringList_DifferentElements(t *testing.T) {
	plan := []string{"app-b", "app-c"}
	state := []string{"app-a", "app-b"}

	sort.Strings(plan)
	sort.Strings(state)

	match := true
	for i := range plan {
		if plan[i] != state[i] {
			match = false
			break
		}
	}
	if match {
		t.Fatal("keys should NOT match — different elements")
	}
}

func TestStringList_DifferentLength(t *testing.T) {
	plan := []string{"app-a", "app-b", "app-c"}
	state := []string{"app-a", "app-b"}

	// Different lengths means a real change — normalization should skip.
	if len(plan) == len(state) {
		t.Fatal("lists should have different lengths")
	}
}

func TestStringList_EmptyLists(t *testing.T) {
	plan := []string{}
	state := []string{}

	sort.Strings(plan)
	sort.Strings(state)

	if len(plan) != len(state) {
		t.Fatal("empty lists should have same length")
	}
}

func TestStringList_DuplicateElements(t *testing.T) {
	plan := []string{"app-a", "app-a", "app-b"}
	state := []string{"app-b", "app-a", "app-a"}

	sort.Strings(plan)
	sort.Strings(state)

	for i := range plan {
		if plan[i] != state[i] {
			t.Fatalf("keys should match (duplicates handled), but position %d differs", i)
		}
	}
}

func TestStringList_DuplicateCountMismatch(t *testing.T) {
	plan := []string{"app-a", "app-a", "app-b"}
	state := []string{"app-a", "app-b", "app-b"}

	sort.Strings(plan)
	sort.Strings(state)

	match := true
	for i := range plan {
		if plan[i] != state[i] {
			match = false
			break
		}
	}
	if match {
		t.Fatal("keys should NOT match — duplicate counts differ")
	}
}

// =============================================================================
// Int64 list normalization tests
// =============================================================================

func TestInt64List_SameSetDifferentOrder(t *testing.T) {
	plan := []int64{3, 1, 2}
	state := []int64{1, 2, 3}

	sort.Slice(plan, func(i, j int) bool { return plan[i] < plan[j] })
	sort.Slice(state, func(i, j int) bool { return state[i] < state[j] })

	for i := range plan {
		if plan[i] != state[i] {
			t.Fatalf("values should match after sort, but position %d differs: %d vs %d", i, plan[i], state[i])
		}
	}
}

func TestInt64List_DifferentElements(t *testing.T) {
	plan := []int64{1, 3}
	state := []int64{1, 2}

	sort.Slice(plan, func(i, j int) bool { return plan[i] < plan[j] })
	sort.Slice(state, func(i, j int) bool { return state[i] < state[j] })

	match := true
	for i := range plan {
		if plan[i] != state[i] {
			match = false
			break
		}
	}
	if match {
		t.Fatal("values should NOT match — different elements")
	}
}

func TestInt64List_EmptyLists(t *testing.T) {
	plan := []int64{}
	state := []int64{}

	if len(plan) != len(state) {
		t.Fatal("empty lists should have same length")
	}
}

// =============================================================================
// TestTemplatePlanDecision — table test for the pure decision core.
//
// otherChanges means "after tentatively setting plan.template = st, other
// attributes still differ between plan and state". This is what the caller
// (suppressTemplateDrift) computes via Raw.Equal after tentative suppression.
// =============================================================================

func TestTemplatePlanDecision(t *testing.T) {
	const display = "Private Apps Block Template"
	const filename = "5.html"

	tests := []struct {
		name         string
		cfg          string
		st           string
		last         string
		otherChanges bool
		wantVal      string
		wantRecord   bool
	}{
		// Branch 1: nothing to reconcile
		{
			name: "B1_same_display_names",
			cfg:  display, st: display,
			wantVal: display,
		},
		{
			name: "B1_config_is_filename",
			cfg:  filename, st: filename,
			wantVal: filename,
		},
		{
			name: "B1_cfg_is_filename_different_state",
			cfg:  "10.html", st: filename, last: display,
			wantVal: "10.html",
		},
		{
			name: "B1_state_not_filename",
			cfg:  display, st: display,
			wantVal: display,
		},
		{
			name: "B1_empty_cfg",
			cfg:  "", st: filename, last: display,
			wantVal: "",
		},
		{
			name: "B1_empty_st",
			cfg:  display, st: "", last: display,
			wantVal: display,
		},
		// Branch 2: suppress (lastSent == cfg, no other changes)
		{
			name: "B2_suppress_no_other_changes",
			cfg:  display, st: filename, last: display,
			otherChanges: false,
			wantVal:      filename, wantRecord: false,
		},
		// Branch 2: restore cfg when other changes present (PUT happening anyway)
		{
			name: "B2_restore_cfg_other_changes",
			cfg:  display, st: filename, last: display,
			otherChanges: true,
			wantVal:      display, wantRecord: false,
		},
		// Branch 3: no private state (first apply / import / upgrade from ≤0.4.12)
		{
			name: "B3_no_private_state",
			cfg:  display, st: filename,
			otherChanges: false,
			wantVal:      display, wantRecord: true,
		},
		// Branch 3: private state has a different display name (user changed template)
		{
			name: "B3_template_changed",
			cfg:  "New Block Template", st: filename, last: display,
			otherChanges: false,
			wantVal:      "New Block Template", wantRecord: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotVal, gotRecord := templatePlanDecision(tc.cfg, tc.st, tc.last, tc.otherChanges)
			if gotVal != tc.wantVal {
				t.Errorf("planValue: got %q, want %q", gotVal, tc.wantVal)
			}
			if gotRecord != tc.wantRecord {
				t.Errorf("record: got %v, want %v", gotRecord, tc.wantRecord)
			}
		})
	}
}

// =============================================================================
// suppressTemplateDrift unit tests (U1–U11)
//
// Each test calls suppressTemplateDrift directly, injecting a fakeTemplateStore
// so private state can be seeded and inspected without importing the internal
// privatestate package.
// =============================================================================

// U1: Branch 1 — config and state hold identical display names (no diff).
func TestSuppressTemplateDrift_U1_SameDisplayNames(t *testing.T) {
	ctx := context.Background()
	sch := npaRulesSchema(ctx)
	const display = "Private Apps Block Template"

	m := npaMinModel(fwtypes.StringValue(display))
	var rs, ws fakeTemplateStore
	resp := callSuppressTemplateDrift(t, ctx, sch, m, m, m, &rs, &ws)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics)
	}
	if got := readPlanTemplate(ctx, t, resp); got != display {
		t.Errorf("plan.template = %q, want %q (branch 1: values match)", got, display)
	}
	// Step 1 should have recorded the display name.
	if written, _ := ws.GetKey(ctx, npaTemplatePrivateKey); decodeStoreStr(written) != display {
		t.Errorf("writeStore[%q] = %q, want %q", npaTemplatePrivateKey, decodeStoreStr(written), display)
	}
}

// U2: Branch 1 — config holds a .html filename (user explicitly uses filename).
func TestSuppressTemplateDrift_U2_ConfigIsFilename(t *testing.T) {
	ctx := context.Background()
	sch := npaRulesSchema(ctx)
	const filename = "5.html"

	m := npaMinModel(fwtypes.StringValue(filename))
	var rs, ws fakeTemplateStore
	resp := callSuppressTemplateDrift(t, ctx, sch, m, m, m, &rs, &ws)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics)
	}
	if got := readPlanTemplate(ctx, t, resp); got != filename {
		t.Errorf("plan.template = %q, want %q (branch 1: config is filename)", got, filename)
	}
}

// U3: Branch 1 — state holds a display name (not .html), no reconciliation needed.
func TestSuppressTemplateDrift_U3_StateNotFilename(t *testing.T) {
	ctx := context.Background()
	sch := npaRulesSchema(ctx)
	const display = "Private Apps Block Template"

	m := npaMinModel(fwtypes.StringValue(display))
	var rs, ws fakeTemplateStore
	resp := callSuppressTemplateDrift(t, ctx, sch, m, m, m, &rs, &ws)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics)
	}
	if got := readPlanTemplate(ctx, t, resp); got != display {
		t.Errorf("plan.template = %q, want %q (branch 1: state is not filename)", got, display)
	}
}

// U4: Branch 2 — suppress (lastSent matches config, no other changes).
// Expected: plan.template = state.template ("5.html").
func TestSuppressTemplateDrift_U4_Branch2_SuppressNoDiff(t *testing.T) {
	ctx := context.Background()
	sch := npaRulesSchema(ctx)
	const (
		display  = "Private Apps Block Template"
		filename = "5.html"
	)

	stateModel := npaMinModel(fwtypes.StringValue(filename))
	planModel := npaMinModel(fwtypes.StringValue(display))

	var rs fakeTemplateStore
	rs.seed(npaTemplatePrivateKey, display) // last sent = display name
	var ws fakeTemplateStore

	resp := callSuppressTemplateDrift(t, ctx, sch, stateModel, planModel, planModel, &rs, &ws)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics)
	}
	if got := readPlanTemplate(ctx, t, resp); got != filename {
		t.Errorf("plan.template = %q, want %q (branch 2: suppressed to state filename)", got, filename)
	}
}

// U5: Branch 2 with other changes — plan.template restored to display name.
// When another attribute also changed, a PUT happens anyway; send the display
// name so the API accepts it (filenames rejected on some tenants).
func TestSuppressTemplateDrift_U5_Branch2_OtherChanges(t *testing.T) {
	ctx := context.Background()
	sch := npaRulesSchema(ctx)
	const (
		display  = "Private Apps Block Template"
		filename = "5.html"
	)

	stateModel := npaMinModelDesc(fwtypes.StringValue(filename), fwtypes.StringValue("old"))
	planModel := npaMinModelDesc(fwtypes.StringValue(display), fwtypes.StringValue("new"))

	var rs fakeTemplateStore
	rs.seed(npaTemplatePrivateKey, display)
	var ws fakeTemplateStore

	resp := callSuppressTemplateDrift(t, ctx, sch, stateModel, planModel, planModel, &rs, &ws)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics)
	}
	// Other changes (description differs) → plan.template should be the display name.
	if got := readPlanTemplate(ctx, t, resp); got != display {
		t.Errorf("plan.template = %q, want %q (branch 2: other changes, display name restored)", got, display)
	}
}

// U6: Branch 3 — no private state (first apply / import / upgrade from ≤0.4.12).
// Expected: plan.template = config.template (honest diff, one-time update).
func TestSuppressTemplateDrift_U6_Branch3_NoPrivateState(t *testing.T) {
	ctx := context.Background()
	sch := npaRulesSchema(ctx)
	const (
		display  = "Private Apps Block Template"
		filename = "5.html"
	)

	stateModel := npaMinModel(fwtypes.StringValue(filename))
	planModel := npaMinModel(fwtypes.StringValue(display))

	// No seed → lastSent is empty → branch 3.
	var rs, ws fakeTemplateStore

	resp := callSuppressTemplateDrift(t, ctx, sch, stateModel, planModel, planModel, &rs, &ws)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics)
	}
	if got := readPlanTemplate(ctx, t, resp); got != display {
		t.Errorf("plan.template = %q, want %q (branch 3: honest diff)", got, display)
	}
}

// U7: Step 1 — writeStore.SetKey is called with the config template value.
func TestSuppressTemplateDrift_U7_WritesPrivateKey(t *testing.T) {
	ctx := context.Background()
	sch := npaRulesSchema(ctx)
	const display = "Private Apps Block Template"

	stateModel := npaMinModel(fwtypes.StringValue("5.html"))
	planModel := npaMinModel(fwtypes.StringValue(display))

	var rs, ws fakeTemplateStore
	resp := callSuppressTemplateDrift(t, ctx, sch, stateModel, planModel, planModel, &rs, &ws)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics)
	}
	written, _ := ws.GetKey(ctx, npaTemplatePrivateKey)
	if decodeStoreStr(written) != display {
		t.Errorf("writeStore[%q] = %q, want %q (step 1 must record config template)", npaTemplatePrivateKey, decodeStoreStr(written), display)
	}
}

// U7b: Step 1 — nil writeStore causes no error and plan template is unchanged.
func TestSuppressTemplateDrift_U7b_NilWriteStoreNoError(t *testing.T) {
	ctx := context.Background()
	sch := npaRulesSchema(ctx)
	const display = "Private Apps Block Template"

	stateModel := npaMinModel(fwtypes.StringValue("5.html"))
	planModel := npaMinModel(fwtypes.StringValue(display))

	var rs fakeTemplateStore
	resp := callSuppressTemplateDrift(t, ctx, sch, stateModel, planModel, planModel, &rs, nil)

	if resp.Diagnostics.HasError() {
		t.Errorf("nil writeStore should not cause errors: %v", resp.Diagnostics)
	}
	// Branch 3 (no private state seeded), but the plan should still hold the display name —
	// not the state filename and not null. Verifies the nil writeStore path doesn't skip Step 2.
	if got := readPlanTemplate(ctx, t, resp); got != display {
		t.Errorf("plan.template = %q, want %q (nil writeStore must not corrupt plan)", got, display)
	}
}

// Regression: config omits template, state holds a .html filename.
// The framework's ExplicitSuppress plan modifier keeps plan.template = state.template ("5.html").
// suppressTemplateDrift must NOT overwrite it with null (configVal) when Branch 1 applies.
// Before the fix (planmodify.go:223), the tentative plan.template = stateVal mutation
// ran unconditionally, then the Branch 1 restore set plan.template = configVal (null),
// causing "Provider produced inconsistent result after apply".
func TestSuppressTemplateDrift_Regression_ConfigNullStateFilename(t *testing.T) {
	ctx := context.Background()
	sch := npaRulesSchema(ctx)
	const filename = "5.html"

	// State: API returned filename. Config: template field omitted (null).
	stateModel := npaMinModel(fwtypes.StringValue(filename))
	// Plan: framework ExplicitSuppress keeps plan.template = state.template = filename.
	planModel := npaMinModel(fwtypes.StringValue(filename))
	// Config: null template (omitted from HCL).
	cfgModel := npaMinModel(fwtypes.StringNull())

	var rs, ws fakeTemplateStore
	resp := callSuppressTemplateDrift(t, ctx, sch, stateModel, planModel, cfgModel, &rs, &ws)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics)
	}
	// Plan must be unchanged: still "5.html" (state value). Must NOT be null.
	if got := readPlanTemplate(ctx, t, resp); got != filename {
		t.Errorf("plan.template = %q, want %q (config null must not corrupt plan with null)", got, filename)
	}
	// Step 1 must NOT have written null/empty to writeStore (config is null → guard skips).
	if len(ws.setCalls) > 0 {
		t.Errorf("writeStore.SetKey should not be called when config template is null, got calls: %v", ws.setCalls)
	}
}

// U8: Create plan — null state. Step 1 runs (key written), step 2 skipped.
func TestSuppressTemplateDrift_U8_CreatePlanEarlyReturn(t *testing.T) {
	ctx := context.Background()
	sch := npaRulesSchema(ctx)
	const display = "Private Apps Block Template"

	nullSt := nullFWState(ctx, sch)
	plan := buildFWPlan(ctx, t, sch, npaMinModel(fwtypes.StringValue(display)))
	cfg := buildFWConfig(ctx, t, sch, npaMinModel(fwtypes.StringValue(display)))

	req := resource.ModifyPlanRequest{State: nullSt, Plan: plan, Config: cfg}
	resp := &resource.ModifyPlanResponse{Plan: plan}

	var rs, ws fakeTemplateStore
	suppressTemplateDrift(ctx, req, resp, &rs, &ws)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics)
	}
	// Step 1: display name should be recorded.
	if written, _ := ws.GetKey(ctx, npaTemplatePrivateKey); decodeStoreStr(written) != display {
		t.Errorf("writeStore[%q] = %q, want %q (step 1 on create)", npaTemplatePrivateKey, decodeStoreStr(written), display)
	}
	// Step 2 skipped (state null) → plan.template unchanged.
	if got := readPlanTemplate(ctx, t, resp); got != display {
		t.Errorf("plan.template = %q, want %q (step 2 skipped on create)", got, display)
	}
}

// U9: Destroy plan — null plan. Step 1 skipped (guarded by !plan.Raw.IsNull()).
func TestSuppressTemplateDrift_U9_DestroyPlanEarlyReturn(t *testing.T) {
	ctx := context.Background()
	sch := npaRulesSchema(ctx)

	state := buildFWState(ctx, t, sch, npaMinModel(fwtypes.StringValue("5.html")))
	nullPl := nullFWPlan(ctx, sch)
	nullCfg := tfsdk.Config{Schema: sch, Raw: nullPl.Raw}

	req := resource.ModifyPlanRequest{State: state, Plan: nullPl, Config: nullCfg}
	resp := &resource.ModifyPlanResponse{Plan: nullPl}

	var rs, ws fakeTemplateStore
	suppressTemplateDrift(ctx, req, resp, &rs, &ws)

	if resp.Diagnostics.HasError() {
		t.Errorf("destroy plan should not cause errors: %v", resp.Diagnostics)
	}
	// Step 1 is guarded by !req.Plan.Raw.IsNull() — no SetKey on destroy.
	if len(ws.setCalls) > 0 {
		t.Errorf("writeStore.SetKey should not be called on destroy plan, got calls: %v", ws.setCalls)
	}
}

// U10a: otherChanges=false after list normalization.
//
// When plan and state differ only in the order of a list (normalizeStringListAttr
// sets plan = state order), suppressTemplateDrift should see no other changes
// and Branch 2 should suppress the template diff.
func TestSuppressTemplateDrift_U10a_OtherChangesFalseAfterNormalization(t *testing.T) {
	ctx := context.Background()
	sch := npaRulesSchema(ctx)
	const (
		display  = "Private Apps Block Template"
		filename = "5.html"
	)

	stateModel := npaMinModel(fwtypes.StringValue(filename))
	stateModel.RuleData.PrivateApps = []fwtypes.String{
		fwtypes.StringValue("app-b"),
		fwtypes.StringValue("app-a"),
	}
	planModel := npaMinModel(fwtypes.StringValue(display))
	planModel.RuleData.PrivateApps = []fwtypes.String{
		fwtypes.StringValue("app-a"), // different order (config order)
		fwtypes.StringValue("app-b"),
	}

	state := buildFWState(ctx, t, sch, stateModel)
	plan := buildFWPlan(ctx, t, sch, planModel)
	cfg := buildFWConfig(ctx, t, sch, planModel)

	req := resource.ModifyPlanRequest{State: state, Plan: plan, Config: cfg}
	resp := &resource.ModifyPlanResponse{Plan: plan}

	// Simulate ModifyPlan: run list normalization BEFORE suppressTemplateDrift.
	normalizeStringListAttr(ctx, req, resp, privateAppsAttrPath)

	// Now run suppressTemplateDrift (after normalization, list is equal → no other changes).
	var rs fakeTemplateStore
	rs.seed(npaTemplatePrivateKey, display)
	suppressTemplateDrift(ctx, req, resp, &rs, nil)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics)
	}
	// After list normalization, the only diff was the template (display vs filename).
	// Branch 2 suppresses it → plan.template = state.template = filename.
	if got := readPlanTemplate(ctx, t, resp); got != filename {
		t.Errorf("plan.template = %q, want %q (branch 2 after list normalization)", got, filename)
	}
}

// U10b: Full fake-store path — Branch 2 with other changes present.
//
// State has a different description than plan (another real change exists).
// After tentative suppression, plan still differs (description) → otherChanges=true
// → plan.template is restored to the display name so the PUT succeeds.
func TestSuppressTemplateDrift_U10b_FakeStoreFullPath_OtherChanges(t *testing.T) {
	ctx := context.Background()
	sch := npaRulesSchema(ctx)
	const (
		display  = "Private Apps Block Template"
		filename = "5.html"
	)

	stateModel := npaMinModelDesc(fwtypes.StringValue(filename), fwtypes.StringValue("v1"))
	planModel := npaMinModelDesc(fwtypes.StringValue(display), fwtypes.StringValue("v2"))

	var rs, ws fakeTemplateStore
	rs.seed(npaTemplatePrivateKey, display)

	resp := callSuppressTemplateDrift(t, ctx, sch, stateModel, planModel, planModel, &rs, &ws)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics)
	}
	// Other changes (description v1→v2) → plan.template = display name.
	if got := readPlanTemplate(ctx, t, resp); got != display {
		t.Errorf("plan.template = %q, want %q (branch 2 other changes, display name for PUT)", got, display)
	}
	// Step 1 should have written the display name.
	if len(ws.setCalls) == 0 {
		t.Error("expected writeStore.SetKey to be called (step 1)")
	}
}

// U11: Branch 3 — private state records an old display name; user changed template.
// Expected: plan.template = new config value (honest diff, one-time update).
func TestSuppressTemplateDrift_U11_Branch3_ChangedTemplate(t *testing.T) {
	ctx := context.Background()
	sch := npaRulesSchema(ctx)
	const (
		oldDisplay = "Old Block Template"
		newDisplay = "New Block Template"
		filename   = "5.html"
	)

	stateModel := npaMinModel(fwtypes.StringValue(filename))
	planModel := npaMinModel(fwtypes.StringValue(newDisplay))

	// Private state holds the OLD display name — template changed.
	var rs fakeTemplateStore
	rs.seed(npaTemplatePrivateKey, oldDisplay)
	var ws fakeTemplateStore

	resp := callSuppressTemplateDrift(t, ctx, sch, stateModel, planModel, planModel, &rs, &ws)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics)
	}
	// Branch 3: lastSent ("old") != cfg ("new") → honest diff.
	if got := readPlanTemplate(ctx, t, resp); got != newDisplay {
		t.Errorf("plan.template = %q, want %q (branch 3: template changed)", got, newDisplay)
	}
	// Step 1 should update writeStore with the new display name.
	if written, _ := ws.GetKey(ctx, npaTemplatePrivateKey); decodeStoreStr(written) != newDisplay {
		t.Errorf("writeStore[%q] = %q, want %q (step 1 should update)", npaTemplatePrivateKey, decodeStoreStr(written), newDisplay)
	}
}

// U12: Shared read/write store — template display name changes between steps.
//
// In terraform-plugin-framework ≥1.2, req.Private and resp.Private share the same
// underlying ProviderData object, so SetKey on resp.Private is immediately visible
// to GetKey on req.Private within the same ModifyPlan call. Before the read-before-
// write fix, the Step 2 (write) ran first, causing Step 1 (read) to see the NEWLY
// written value ("B") rather than the PREVIOUSLY committed value ("A"), which
// incorrectly triggered Branch 2 suppression and hid a genuine template change.
//
// This test uses the same fakeTemplateStore for both readStore and writeStore to
// replicate that behaviour. The fix (read before write) must produce Branch 3
// (plan.template = new display name) rather than Branch 2 (plan.template = filename).
func TestSuppressTemplateDrift_U12_SharedStore_TemplateChange(t *testing.T) {
	ctx := context.Background()
	sch := npaRulesSchema(ctx)
	const (
		oldDisplay = "Default Template"
		newDisplay = "tf-test-template"
		filename   = "block_page.html"
	)

	stateModel := npaMinModel(fwtypes.StringValue(filename))
	planModel := npaMinModel(fwtypes.StringValue(newDisplay))

	// Use the SAME store for read and write — mirrors the real framework.
	shared := &fakeTemplateStore{}
	shared.seed(npaTemplatePrivateKey, oldDisplay) // private state from previous apply

	resp := callSuppressTemplateDrift(t, ctx, sch, stateModel, planModel, planModel, shared, shared)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected errors: %v", resp.Diagnostics)
	}
	// Branch 3 must fire: lastSent ("Default Template") ≠ cfg ("tf-test-template").
	// plan.template must be the new display name, not the suppressed filename.
	if got := readPlanTemplate(ctx, t, resp); got != newDisplay {
		t.Errorf("plan.template = %q, want %q (branch 3: shared store must not suppress template change)", got, newDisplay)
	}
	// Private state should now record the new display name.
	if written, _ := shared.GetKey(ctx, npaTemplatePrivateKey); decodeStoreStr(written) != newDisplay {
		t.Errorf("store[%q] = %q, want %q (new display name must be committed)", npaTemplatePrivateKey, decodeStoreStr(written), newDisplay)
	}
}
