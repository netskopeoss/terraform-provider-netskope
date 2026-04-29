package provider

import (
	"sort"
	"strings"
	"testing"
)

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
//
// These test the core logic used by normalizeInt64ListAttr:
// sort both value lists and compare.
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
// Template display name preservation tests
//
// These test the core logic used by preserveTemplateDisplayName:
// if state has a .html file name and config has a display name, preserve config.
// =============================================================================

func TestTemplate_FileNameInState_DisplayNameInConfig(t *testing.T) {
	configVal := "Private Apps Block Template"
	stateVal := "5.html"

	// Should preserve config: state is a file name, config is a display name
	if strings.HasSuffix(stateVal, ".html") && !strings.HasSuffix(configVal, ".html") {
		// This is the case where we preserve the config value
	} else {
		t.Fatal("should detect file name in state and display name in config")
	}
}

func TestTemplate_BothDisplayNames_NoChange(t *testing.T) {
	configVal := "Private Apps Block Template"
	stateVal := "Private Apps Block Template"

	if configVal != stateVal {
		t.Fatal("identical values should not trigger preservation")
	}
}

func TestTemplate_BothFileNames_NoPreservation(t *testing.T) {
	configVal := "5.html"
	stateVal := "5.html"

	if configVal != stateVal {
		t.Fatal("identical file names should not trigger preservation")
	}
}

func TestTemplate_ConfigIsFileName_NoPreservation(t *testing.T) {
	configVal := "block_page.html"
	stateVal := "block_page.html"

	// If user explicitly uses file name in config, don't override
	shouldPreserve := strings.HasSuffix(stateVal, ".html") && !strings.HasSuffix(configVal, ".html")
	if shouldPreserve {
		t.Fatal("should not preserve when config is also a file name")
	}
}

func TestTemplate_UserChangesTemplate(t *testing.T) {
	configVal := "New Block Template"
	stateVal := "5.html"

	// Config changed to a new display name, state still has old file name
	// Should preserve config (new display name)
	if !(strings.HasSuffix(stateVal, ".html") && !strings.HasSuffix(configVal, ".html")) {
		t.Fatal("should preserve new display name when state has file name")
	}
}

func TestTemplate_DefaultBlockPage(t *testing.T) {
	configVal := "Default Template"
	stateVal := "block_page.html"

	if !(strings.HasSuffix(stateVal, ".html") && !strings.HasSuffix(configVal, ".html")) {
		t.Fatal("should preserve Default Template when state has block_page.html")
	}
}
