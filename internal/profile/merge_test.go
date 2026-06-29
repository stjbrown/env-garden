package profile

import "testing"

func TestMergeConcatenatesInOrder(t *testing.T) {
	a := &Profile{Name: "a", Vars: []Var{{Key: "A", Raw: "1"}, {Key: "B", Raw: "2"}}}
	b := &Profile{Name: "b", Vars: []Var{{Key: "C", Raw: "3"}}}
	m, conflicts := Merge("a+b", []*Profile{a, b})
	if m.Name != "a+b" {
		t.Errorf("name = %q", m.Name)
	}
	if len(conflicts) != 0 {
		t.Errorf("unexpected conflicts: %+v", conflicts)
	}
	want := []Var{{Key: "A", Raw: "1"}, {Key: "B", Raw: "2"}, {Key: "C", Raw: "3"}}
	if !varsEqual(m.Vars, want) {
		t.Errorf("vars = %+v, want %+v", m.Vars, want)
	}
}

func TestMergeLastWinsKeepsPosition(t *testing.T) {
	a := &Profile{Name: "a", Vars: []Var{{Key: "TOKEN", Raw: "old"}, {Key: "X", Raw: "1"}}}
	b := &Profile{Name: "b", Vars: []Var{{Key: "TOKEN", Raw: "new"}}}
	m, conflicts := Merge("a+b", []*Profile{a, b})

	// TOKEN keeps its original (first-seen) position but takes b's value.
	want := []Var{{Key: "TOKEN", Raw: "new"}, {Key: "X", Raw: "1"}}
	if !varsEqual(m.Vars, want) {
		t.Errorf("vars = %+v, want %+v", m.Vars, want)
	}
	if len(conflicts) != 1 {
		t.Fatalf("got %d conflicts, want 1", len(conflicts))
	}
	if c := conflicts[0]; c.Key != "TOKEN" || c.Winner != "b" || c.Loser != "a" {
		t.Errorf("conflict = %+v", c)
	}
}

func TestMergePreservesRefFlag(t *testing.T) {
	a := &Profile{Name: "a", Vars: []Var{{Key: "K", Raw: "op://Vault/item/field", IsRef: true}}}
	b := &Profile{Name: "b", Vars: []Var{{Key: "L", Raw: "lit"}}}
	m, _ := Merge("a+b", []*Profile{a, b})
	if !m.HasRefs() {
		t.Error("expected merged profile to report refs")
	}
	if m.Vars[0].IsRef != true || m.Vars[1].IsRef != false {
		t.Errorf("ref flags not preserved: %+v", m.Vars)
	}
}

func TestMergeJoinsDescriptions(t *testing.T) {
	a := &Profile{Name: "a", Desc: "dev vertex"}
	b := &Profile{Name: "b", Desc: ""}
	c := &Profile{Name: "c", Desc: "slack creds"}
	m, _ := Merge("a+b+c", []*Profile{a, b, c})
	if m.Desc != "dev vertex + slack creds" {
		t.Errorf("desc = %q", m.Desc)
	}
}

func varsEqual(a, b []Var) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
