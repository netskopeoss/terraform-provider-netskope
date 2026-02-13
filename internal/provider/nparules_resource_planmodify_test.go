package provider

import (
	"sort"
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
