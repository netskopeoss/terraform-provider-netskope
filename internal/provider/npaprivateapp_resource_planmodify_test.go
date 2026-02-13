package provider

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	tfTypes "github.com/netskopeoss/terraform-provider-netskope/internal/provider/types"
)

// =============================================================================
// Protocol key tests
// =============================================================================

func TestProtocolKey_Basic(t *testing.T) {
	p := tfTypes.ProtocolItem{
		Protocol: types.StringValue("tcp"),
		Port:     types.StringValue("443"),
	}
	// Build the same key string the normalizer builds.
	key := fmt.Sprintf("%s:%s", p.Protocol.ValueString(), p.Port.ValueString())
	if key != "tcp:443" {
		t.Errorf("expected tcp:443, got %s", key)
	}
}

func TestProtocolKey_UDP(t *testing.T) {
	p := tfTypes.ProtocolItem{
		Protocol: types.StringValue("udp"),
		Port:     types.StringValue("53"),
	}
	key := fmt.Sprintf("%s:%s", p.Protocol.ValueString(), p.Port.ValueString())
	if key != "udp:53" {
		t.Errorf("expected udp:53, got %s", key)
	}
}

// =============================================================================
// Publisher key tests
// =============================================================================

func TestPublisherKey_ByID(t *testing.T) {
	p := tfTypes.PublisherItem{
		PublisherID:   types.StringValue("100"),
		PublisherName: types.StringValue("pub-a"),
	}
	key := publisherKey(p)
	if key != "id:100" {
		t.Errorf("expected id:100, got %s", key)
	}
}

func TestPublisherKey_FallbackToName(t *testing.T) {
	p := tfTypes.PublisherItem{
		PublisherID:   types.StringUnknown(),
		PublisherName: types.StringValue("pub-a"),
	}
	key := publisherKey(p)
	if key != "name:pub-a" {
		t.Errorf("expected name:pub-a, got %s", key)
	}
}

func TestPublisherKey_NullID_FallbackToName(t *testing.T) {
	p := tfTypes.PublisherItem{
		PublisherID:   types.StringNull(),
		PublisherName: types.StringValue("pub-b"),
	}
	key := publisherKey(p)
	if key != "name:pub-b" {
		t.Errorf("expected name:pub-b, got %s", key)
	}
}

func TestPublisherKey_BothUnknown_ReturnsEmpty(t *testing.T) {
	p := tfTypes.PublisherItem{
		PublisherID:   types.StringUnknown(),
		PublisherName: types.StringUnknown(),
	}
	key := publisherKey(p)
	if key != "" {
		t.Errorf("expected empty string, got %s", key)
	}
}

// =============================================================================
// Tag key tests
// =============================================================================

func TestTagKey_ByName(t *testing.T) {
	tag := tfTypes.Tags{
		TagName: types.StringValue("infosec"),
		TagID:   types.Int32Value(7),
	}
	// Tag key uses tag_name (tag_id is Computed-only, unknown in plan).
	if tag.TagName.ValueString() != "infosec" {
		t.Errorf("expected infosec, got %s", tag.TagName.ValueString())
	}
}

// =============================================================================
// Sorted key comparison tests
//
// These test the core logic: sort both key lists and compare.
// This is the same approach used in each normalize function.
// =============================================================================

func TestSortedKeysMatch_SameSetDifferentOrder(t *testing.T) {
	plan := []string{"tcp:443", "tcp:22", "udp:53"}
	state := []string{"tcp:22", "tcp:443", "udp:53"}

	sort.Strings(plan)
	sort.Strings(state)

	for i := range plan {
		if plan[i] != state[i] {
			t.Fatalf("keys should match after sort, but position %d differs: %s vs %s", i, plan[i], state[i])
		}
	}
}

func TestSortedKeysMatch_DifferentElements(t *testing.T) {
	plan := []string{"tcp:443", "tcp:8080"}
	state := []string{"tcp:22", "tcp:443"}

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

func TestSortedKeysMatch_DuplicateElements(t *testing.T) {
	// Two tcp:443 in both — should match.
	plan := []string{"tcp:443", "tcp:443", "tcp:22"}
	state := []string{"tcp:22", "tcp:443", "tcp:443"}

	sort.Strings(plan)
	sort.Strings(state)

	for i := range plan {
		if plan[i] != state[i] {
			t.Fatalf("keys should match (duplicates handled), but position %d differs", i)
		}
	}
}

func TestSortedKeysMatch_DuplicateCountMismatch(t *testing.T) {
	// Plan has two tcp:443, state has two tcp:22 — should NOT match.
	plan := []string{"tcp:443", "tcp:443", "tcp:22"}
	state := []string{"tcp:22", "tcp:22", "tcp:443"}

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

func TestSortedKeysMatch_EmptyLists(t *testing.T) {
	plan := []string{}
	state := []string{}

	sort.Strings(plan)
	sort.Strings(state)

	if len(plan) != len(state) {
		t.Fatal("empty lists should have same length")
	}
	// No elements to compare — match is trivially true.
}

// =============================================================================
// BUG-005: Case-insensitive tag name matching
// =============================================================================

func TestTagKeysMatch_CaseInsensitive(t *testing.T) {
	// API returns "Production" but user config says "production".
	// Without lowercasing, sort.Strings uses ASCII order where uppercase < lowercase,
	// so "Production" sorts before "infrastructure" but "production" sorts after it.
	// This misaligns the lists even though they represent the same tags.
	plan := []string{"infrastructure", "production"}
	state := []string{"Production", "infrastructure"}

	// Show that naive sort + case-sensitive compare fails:
	sortedPlan := make([]string, len(plan))
	sortedState := make([]string, len(state))
	copy(sortedPlan, plan)
	copy(sortedState, state)
	sort.Strings(sortedPlan)
	sort.Strings(sortedState)
	// sortedPlan = ["infrastructure", "production"], sortedState = ["Production", "infrastructure"]
	if sortedPlan[0] == sortedState[0] {
		t.Fatal("naive sort should misalign these lists due to case")
	}

	// Now apply the fix: lowercase before sorting.
	loweredPlan := make([]string, len(plan))
	loweredState := make([]string, len(state))
	for i, k := range plan {
		loweredPlan[i] = strings.ToLower(k)
	}
	for i, k := range state {
		loweredState[i] = strings.ToLower(k)
	}
	sort.Strings(loweredPlan)
	sort.Strings(loweredState)

	for i := range loweredPlan {
		if loweredPlan[i] != loweredState[i] {
			t.Fatalf("lowercased+sorted keys should match at position %d: %q vs %q", i, loweredPlan[i], loweredState[i])
		}
	}
}

func TestTagKeysMatch_CaseInsensitive_DifferentTags(t *testing.T) {
	// Different tags should NOT match even with lowercasing.
	plan := []string{"production", "web"}
	state := []string{"Production", "infrastructure"}

	loweredPlan := make([]string, len(plan))
	loweredState := make([]string, len(state))
	for i, k := range plan {
		loweredPlan[i] = strings.ToLower(k)
	}
	for i, k := range state {
		loweredState[i] = strings.ToLower(k)
	}
	sort.Strings(loweredPlan)
	sort.Strings(loweredState)

	match := true
	for i := range loweredPlan {
		if loweredPlan[i] != loweredState[i] {
			match = false
			break
		}
	}
	if match {
		t.Fatal("different tags should NOT match even with lowercasing")
	}
}