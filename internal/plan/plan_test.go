package plan

import (
	"os"
	"testing"
)

func TestParseMixedPlan(t *testing.T) {
	f, err := os.Open("../../testdata/mixed_plan.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	p, err := Parse(f)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if p.TerraformVersion != "1.7.4" {
		t.Errorf("TerraformVersion = %q, want %q", p.TerraformVersion, "1.7.4")
	}

	if len(p.ResourceChanges) != 6 {
		t.Errorf("ResourceChanges count = %d, want 6", len(p.ResourceChanges))
	}
}

func TestSummarizeMixedPlan(t *testing.T) {
	f, err := os.Open("../../testdata/mixed_plan.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	p, err := Parse(f)
	if err != nil {
		t.Fatal(err)
	}

	s := Summarize(p)

	// Expected: 2 create, 1 update, 1 delete, 1 replace = 5 total (no-op excluded)
	if s.TotalChanges() != 5 {
		t.Errorf("TotalChanges = %d, want 5", s.TotalChanges())
	}

	if !s.HasDestroys {
		t.Error("HasDestroys should be true")
	}

	expectedCounts := map[Action]int{
		Create:  2,
		Update:  1,
		Delete:  1,
		Replace: 1,
	}
	for action, want := range expectedCounts {
		got := s.Counts[action]
		if got != want {
			t.Errorf("Count[%s] = %d, want %d", action, got, want)
		}
	}

	// Verify ordering: Delete first, then Replace, then Update, then Create.
	if len(s.Groups) != 4 {
		t.Fatalf("Groups count = %d, want 4", len(s.Groups))
	}
	expectedOrder := []Action{Delete, Replace, Update, Create}
	for i, want := range expectedOrder {
		if s.Groups[i].Action != want {
			t.Errorf("Groups[%d].Action = %s, want %s", i, s.Groups[i].Action, want)
		}
	}
}

func TestSummarizeNoChanges(t *testing.T) {
	f, err := os.Open("../../testdata/no_changes.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	p, err := Parse(f)
	if err != nil {
		t.Fatal(err)
	}

	s := Summarize(p)

	if s.TotalChanges() != 0 {
		t.Errorf("TotalChanges = %d, want 0", s.TotalChanges())
	}

	if s.HasDestroys {
		t.Error("HasDestroys should be false")
	}
}

func TestClassifyActions(t *testing.T) {
	tests := []struct {
		name    string
		actions []string
		want    Action
	}{
		{"create", []string{"create"}, Create},
		{"update", []string{"update"}, Update},
		{"delete", []string{"delete"}, Delete},
		{"replace delete-create", []string{"delete", "create"}, Replace},
		{"replace create-delete", []string{"create", "delete"}, Replace},
		{"no-op", []string{"no-op"}, NoOp},
		{"empty", []string{}, NoOp},
		{"read", []string{"read"}, Read},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyActions(tt.actions)
			if got != tt.want {
				t.Errorf("classifyActions(%v) = %s, want %s", tt.actions, got, tt.want)
			}
		})
	}
}
