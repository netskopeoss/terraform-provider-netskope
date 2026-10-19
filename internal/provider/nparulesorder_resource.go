package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/netskopeoss/terraform-provider-netskope/internal/sdk"
	"github.com/netskopeoss/terraform-provider-netskope/internal/sdk/models/operations"
	"github.com/netskopeoss/terraform-provider-netskope/internal/sdk/models/shared"
)

var (
	_ resource.Resource              = &NPARulesOrderResource{}
	_ resource.ResourceWithConfigure = &NPARulesOrderResource{}
)

func NewNPARulesOrderResource() resource.Resource {
	return &NPARulesOrderResource{}
}

type NPARulesOrderResource struct {
	client *sdk.TerraformProviderNs
}

type NPARulesOrderResourceModel struct {
	RuleIDs []types.String `tfsdk:"rule_ids"`
	ID      types.String   `tfsdk:"id"`
}

func (r *NPARulesOrderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_npa_rules_order"
}

func (r *NPARulesOrderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages the list position of NPA policy rules.

Takes a list of rule IDs and reorders them so that the first ID in the list
appears first (top), and each subsequent ID appears after the previous one.

**Note:** NPA rules use most-specific-match evaluation. List position does not
determine which rule wins — it is for organizational purposes only.

Rules should be created separately with ` + "`netskope_npa_rules`" + ` using
` + "`rule_order = { order = \"bottom\" }`" + ` (safe default), then pass their
IDs to this resource to set the list position.`,

		Attributes: map[string]schema.Attribute{
			"rule_ids": schema.ListAttribute{
				Required:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Ordered list of rule IDs. Position in list = display order in the Netskope UI.",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Resource identifier.",
			},
		},
	}
}

func (r *NPARulesOrderResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*sdk.TerraformProviderNs)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *sdk.TerraformProviderNs, got: %T", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *NPARulesOrderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NPARulesOrderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.applyOrder(ctx, data.RuleIDs); err != nil {
		resp.Diagnostics.AddError("Failed to set rule order", err.Error())
		return
	}

	data.ID = types.StringValue(r.computeID(data.RuleIDs))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NPARulesOrderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NPARulesOrderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Query the live rule order from the API
	liveOrder, err := r.readLiveOrder(ctx, data.RuleIDs)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read rule order from API", err.Error())
		return
	}

	// Update state with the live order — if someone reordered via UI,
	// Terraform will detect the diff and plan a correction
	if liveOrder != nil {
		data.RuleIDs = liveOrder
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// readLiveOrder queries the rules list API and returns the managed rule IDs
// in their current list order. Rules deleted out-of-band are removed.
func (r *NPARulesOrderResource) readLiveOrder(ctx context.Context, managedIDs []types.String) ([]types.String, error) {
	res, err := r.client.ListNPARules(ctx, operations.ListNPARulesRequest{})
	if err != nil {
		return nil, fmt.Errorf("GET rules failed: %w", err)
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("GET rules returned %d", res.StatusCode)
	}

	if res.NpaPolicyResponse == nil {
		return nil, fmt.Errorf("GET rules returned empty response")
	}

	apiOrder := make([]string, 0, len(res.NpaPolicyResponse.Data))
	for _, rule := range res.NpaPolicyResponse.Data {
		if rule.ID != nil {
			apiOrder = append(apiOrder, *rule.ID)
		}
	}

	// Extract managed rules in their current API order
	var liveOrder []types.String
	for _, id := range filterManaged(apiOrder, stringValues(managedIDs)) {
		liveOrder = append(liveOrder, types.StringValue(id))
	}

	return liveOrder, nil
}

// filterManaged returns the IDs from apiOrder that are in managed, keeping
// their API list order. Unmanaged rules (including one just removed from
// rule_ids) are dropped, so the result is directly comparable with desired.
func filterManaged(apiOrder, managed []string) []string {
	set := make(map[string]bool, len(managed))
	for _, id := range managed {
		set[id] = true
	}

	var out []string
	for _, id := range apiOrder {
		if set[id] {
			out = append(out, id)
		}
	}
	return out
}

func (r *NPARulesOrderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data NPARulesOrderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.applyOrder(ctx, data.RuleIDs); err != nil {
		resp.Diagnostics.AddError("Failed to update rule order", err.Error())
		return
	}

	data.ID = types.StringValue(r.computeID(data.RuleIDs))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NPARulesOrderResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Deleting the order resource does not delete or reorder the rules.
}

// applyOrder brings the managed rules into the requested relative order while
// sending as few PATCHes as possible.
//
// It reads the live order first and only moves the rules that are out of place
// (see computeTopMoves). When the live order already matches, nothing is sent
// and no sleeps are taken. After PATCHing, the live order is read again; that
// read doubles as the verification, and anything still out of place is moved
// on the next attempt.
// See: https://github.com/netskopeoss/terraform-provider-netskope/issues/125
func (r *NPARulesOrderResource) applyOrder(ctx context.Context, ruleIDs []types.String) error {
	if len(ruleIDs) == 0 {
		return nil
	}

	desired := make([]string, len(ruleIDs))
	for i, id := range ruleIDs {
		desired[i] = id.ValueString()
	}

	live, err := r.readLiveOrder(ctx, ruleIDs)
	if err != nil {
		return err
	}
	// Recently created rules may not be listed yet — the API has eventual
	// consistency. Give them a moment to settle before deciding what to move.
	if len(live) != len(desired) {
		time.Sleep(3 * time.Second)
		if live, err = r.readLiveOrder(ctx, ruleIDs); err != nil {
			return err
		}
	}

	const maxAttempts = 3
	for attempt := 0; attempt < maxAttempts; attempt++ {
		moves := computeTopMoves(stringValues(live), desired)
		if len(moves) == 0 {
			return nil
		}

		if err := r.patchOrder(ctx, moves); err != nil {
			return err
		}

		// Let the new positions commit, then re-read to verify.
		time.Sleep(2 * time.Second)
		if live, err = r.readLiveOrder(ctx, ruleIDs); err != nil {
			return err
		}
	}

	if len(computeTopMoves(stringValues(live), desired)) == 0 {
		return nil
	}
	return fmt.Errorf("rule ordering did not converge after %d attempts", maxAttempts)
}

// computeTopMoves returns the rule IDs that must be moved to the top, in the
// order they should be listed in desired, so that the managed rules end up in
// desired order. live is the current relative order of the same rules.
//
// A "top" PATCH puts a rule above everything else, so every rule that is not
// moved must end up below all moved ones: the unmoved rules are a suffix of
// desired. The result is therefore the prefix of desired in front of the
// longest suffix that is already in the right relative order in live.
//
// It returns nil when live already matches desired. When live and desired do
// not describe the same set of rules (a rule is not listed yet, or desired has
// duplicates) it returns all of desired, which is always a correct fallback.
func computeTopMoves(live, desired []string) []string {
	if len(desired) == 0 {
		return nil
	}
	if len(live) != len(desired) {
		return desired
	}

	pos := make(map[string]int, len(live))
	for i, id := range live {
		pos[id] = i
	}
	if len(pos) != len(desired) {
		return desired
	}
	for _, id := range desired {
		if _, ok := pos[id]; !ok {
			return desired
		}
	}

	// Walk left from the end while the rules are already in ascending live order.
	j := len(desired) - 1
	for j > 0 && pos[desired[j-1]] < pos[desired[j]] {
		j--
	}
	if j == 0 {
		return nil
	}
	return desired[:j]
}

func stringValues(vals []types.String) []string {
	out := make([]string, len(vals))
	for i, v := range vals {
		out[i] = v.ValueString()
	}
	return out
}

// patchOrder moves the given rules to the top in reverse order.
// Processing [A, B, C] in reverse: C→top, B→top (pushes C down), A→top
// (pushes B and C down). Result: A, B, C from top.
// This avoids "after" references which have eventual consistency issues.
func (r *NPARulesOrderResource) patchOrder(ctx context.Context, ids []string) error {
	order := shared.OrderTop

	// Process in reverse — each "top" pushes previous ones down
	for i := len(ids) - 1; i >= 0; i-- {
		id := ids[i]

		res, err := r.client.UpdateNPARules(ctx, operations.UpdateNPARulesRequest{
			ID: id,
			NpaPolicyRequest: shared.NpaPolicyRequest{
				RuleOrder: &shared.RuleOrder{
					Order: &order,
				},
			},
		})
		if err != nil {
			return fmt.Errorf("PATCH failed for rule %s: %w", id, err)
		}
		if res.StatusCode != 200 {
			return fmt.Errorf("PATCH rule %s returned %d", id, res.StatusCode)
		}

		// Wait for the position to commit before moving the next rule
		time.Sleep(500 * time.Millisecond)
	}

	return nil
}

func (r *NPARulesOrderResource) computeID(ruleIDs []types.String) string {
	ids := make([]string, len(ruleIDs))
	for i, id := range ruleIDs {
		ids[i] = id.ValueString()
	}
	return fmt.Sprintf("npa-rules-order:%d", len(ids))
}
