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

	// Build set of managed IDs for fast lookup
	managed := make(map[string]bool, len(managedIDs))
	for _, id := range managedIDs {
		managed[id.ValueString()] = true
	}

	// Extract managed rules in their current API order
	var liveOrder []types.String
	for _, rule := range res.NpaPolicyResponse.Data {
		if rule.ID != nil && managed[*rule.ID] {
			liveOrder = append(liveOrder, types.StringValue(*rule.ID))
		}
	}

	return liveOrder, nil
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

// applyOrder PATCHes each rule sequentially to set list position.
// First rule → "top", each subsequent rule → "after" the previous rule's ID.
// After all PATCHes, verifies the order and retries if the API hasn't
// fully committed the changes.
func (r *NPARulesOrderResource) applyOrder(ctx context.Context, ruleIDs []types.String) error {
	if len(ruleIDs) == 0 {
		return nil
	}

	// Wait for recently created rules to settle before reordering.
	// The API has eventual consistency — rules created in parallel
	// may not be fully committed when the order resource starts.
	time.Sleep(3 * time.Second)

	// Apply ordering with verify-and-retry
	const maxAttempts = 3
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if err := r.patchOrder(ctx, ruleIDs); err != nil {
			return err
		}

		// Verify the order took effect
		time.Sleep(2 * time.Second)
		if r.verifyOrder(ctx, ruleIDs) {
			return nil
		}

		// Order didn't stick — wait before retry
		time.Sleep(3 * time.Second)
	}

	return fmt.Errorf("rule ordering did not converge after %d attempts", maxAttempts)
}

// patchOrder moves rules to the top in reverse order.
// Processing [A, B, C] in reverse: C→top, B→top (pushes C down), A→top
// (pushes B and C down). Result: A, B, C from top.
// This avoids "after" references which have eventual consistency issues.
func (r *NPARulesOrderResource) patchOrder(ctx context.Context, ruleIDs []types.String) error {
	order := shared.OrderTop

	// Process in reverse — each "top" pushes previous ones down
	for i := len(ruleIDs) - 1; i >= 0; i-- {
		id := ruleIDs[i].ValueString()

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

// verifyOrder checks if the managed rules appear in the expected order
// in the API's list.
func (r *NPARulesOrderResource) verifyOrder(ctx context.Context, ruleIDs []types.String) bool {
	liveOrder, err := r.readLiveOrder(ctx, ruleIDs)
	if err != nil || len(liveOrder) != len(ruleIDs) {
		return false
	}

	for i := range ruleIDs {
		if ruleIDs[i].ValueString() != liveOrder[i].ValueString() {
			return false
		}
	}

	return true
}

func (r *NPARulesOrderResource) computeID(ruleIDs []types.String) string {
	ids := make([]string, len(ruleIDs))
	for i, id := range ruleIDs {
		ids[i] = id.ValueString()
	}
	return fmt.Sprintf("npa-rules-order:%d", len(ids))
}
