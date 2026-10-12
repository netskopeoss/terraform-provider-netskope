package provider

import (
	"reflect"
	"testing"
)

// applyMoves simulates the API: each "top" PATCH puts the rule above everything else.
func applyMoves(live, moves []string) []string {
	out := append([]string(nil), live...)
	for i := len(moves) - 1; i >= 0; i-- {
		id := moves[i]
		rest := make([]string, 0, len(out))
		for _, v := range out {
			if v != id {
				rest = append(rest, v)
			}
		}
		out = append([]string{id}, rest...)
	}
	return out
}

func TestComputeTopMoves(t *testing.T) {
	tests := []struct {
		name    string
		live    []string
		desired []string
		want    []string
	}{
		{"already ordered", []string{"a", "b", "c"}, []string{"a", "b", "c"}, nil},
		{"single rule", []string{"a"}, []string{"a"}, nil},
		{"empty", nil, nil, nil},
		{"last rule moved to top", []string{"a", "b", "c"}, []string{"c", "a", "b"}, []string{"c"}},
		{"middle rule moved to top", []string{"a", "b", "c", "d"}, []string{"c", "a", "b", "d"}, []string{"c"}},
		{"first rule moved to bottom", []string{"a", "b", "c"}, []string{"b", "c", "a"}, []string{"b", "c"}},
		{"fully reversed", []string{"a", "b", "c"}, []string{"c", "b", "a"}, []string{"c", "b"}},
		{"two swapped", []string{"a", "b", "c"}, []string{"b", "a", "c"}, []string{"b"}},
		{"live missing a rule", []string{"a", "c"}, []string{"a", "b", "c"}, []string{"a", "b", "c"}},
		{"live has different rule", []string{"a", "b", "x"}, []string{"a", "b", "c"}, []string{"a", "b", "c"}},
		{"duplicate in desired", []string{"a", "b", "b"}, []string{"a", "b", "b"}, []string{"a", "b", "b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeTopMoves(tt.live, tt.desired)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("computeTopMoves(%v, %v) = %v, want %v", tt.live, tt.desired, got, tt.want)
			}
		})
	}
}

// Applying the returned moves with "top" semantics must always reach the desired order.
func TestComputeTopMoves_ConvergesForAllPermutations(t *testing.T) {
	ids := []string{"a", "b", "c", "d", "e"}
	var perms [][]string
	var gen func(cur, rest []string)
	gen = func(cur, rest []string) {
		if len(rest) == 0 {
			perms = append(perms, append([]string(nil), cur...))
			return
		}
		for i := range rest {
			next := append(append([]string(nil), rest[:i]...), rest[i+1:]...)
			gen(append(cur, rest[i]), next)
		}
	}
	gen(nil, ids)

	for _, live := range perms {
		for _, desired := range perms {
			moves := computeTopMoves(live, desired)
			if got := applyMoves(live, moves); !reflect.DeepEqual(got, desired) {
				t.Fatalf("live=%v desired=%v moves=%v -> %v", live, desired, moves, got)
			}
			if len(moves) > len(desired) {
				t.Fatalf("more moves than rules: live=%v desired=%v moves=%v", live, desired, moves)
			}
		}
	}
}

// Removing an ID from rule_ids (issue 125): the API list still contains the
// removed rule, but the managed view of it must need zero moves.
func TestFilterManaged_RemovedIDNeedsNoMoves(t *testing.T) {
	apiOrder := []string{"a", "b", "c", "d", "x"} // x is unmanaged, c was just removed
	desired := []string{"a", "b", "d"}

	live := filterManaged(apiOrder, desired)
	if want := []string{"a", "b", "d"}; !reflect.DeepEqual(live, want) {
		t.Fatalf("filterManaged = %v, want %v", live, want)
	}
	if moves := computeTopMoves(live, desired); moves != nil {
		t.Fatalf("computeTopMoves = %v, want no moves", moves)
	}
}

func TestFilterManaged(t *testing.T) {
	tests := []struct {
		name     string
		apiOrder []string
		managed  []string
		want     []string
	}{
		{"keeps api order, not managed order", []string{"a", "b", "c"}, []string{"c", "a"}, []string{"a", "c"}},
		{"drops unmanaged", []string{"x", "a", "y", "b"}, []string{"a", "b"}, []string{"a", "b"}},
		{"managed rule missing from api", []string{"a", "c"}, []string{"a", "b", "c"}, []string{"a", "c"}},
		{"nothing managed", []string{"a", "b"}, nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filterManaged(tt.apiOrder, tt.managed); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("filterManaged(%v, %v) = %v, want %v", tt.apiOrder, tt.managed, got, tt.want)
			}
		})
	}
}
