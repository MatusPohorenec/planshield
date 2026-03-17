package render

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/planshield/planshield/internal/plan"
)

func testSummary() *plan.Summary {
	return &plan.Summary{
		TerraformVersion: "1.7.4",
		Groups: []plan.GroupedChange{
			{Action: plan.Delete, Addresses: []string{"aws_iam_role.old_role"}},
			{Action: plan.Replace, Addresses: []string{"aws_db_instance.main"}},
			{Action: plan.Update, Addresses: []string{"aws_security_group.allow_tls"}},
			{Action: plan.Create, Addresses: []string{"aws_instance.web", "aws_s3_bucket.data"}},
		},
		Counts: map[plan.Action]int{
			plan.Create:  2,
			plan.Update:  1,
			plan.Delete:  1,
			plan.Replace: 1,
		},
		HasDestroys: true,
	}
}

func TestMarkdownOutput(t *testing.T) {
	var buf bytes.Buffer
	s := testSummary()

	if err := Markdown(&buf, s); err != nil {
		t.Fatalf("Markdown() error: %v", err)
	}

	out := buf.String()

	// Check header
	if !strings.Contains(out, "## PlanShield Summary") {
		t.Error("missing header")
	}

	// Check version
	if !strings.Contains(out, "1.7.4") {
		t.Error("missing terraform version")
	}

	// Check warning for destroys
	if !strings.Contains(out, "Destructive changes detected") {
		t.Error("missing destructive changes warning")
	}

	// Check all resource addresses appear
	for _, addr := range []string{
		"aws_iam_role.old_role",
		"aws_db_instance.main",
		"aws_security_group.allow_tls",
		"aws_instance.web",
		"aws_s3_bucket.data",
	} {
		if !strings.Contains(out, addr) {
			t.Errorf("missing address %q in markdown output", addr)
		}
	}

	// Check action headers
	for _, heading := range []string{"Delete", "Replace", "Update", "Create"} {
		if !strings.Contains(out, heading) {
			t.Errorf("missing action heading %q", heading)
		}
	}
}

func TestMarkdownNoChanges(t *testing.T) {
	var buf bytes.Buffer
	s := &plan.Summary{
		TerraformVersion: "1.7.4",
		Counts:           map[plan.Action]int{},
	}

	if err := Markdown(&buf, s); err != nil {
		t.Fatalf("Markdown() error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "No changes") {
		t.Error("expected 'No changes' message")
	}
}

func TestJSONOutput(t *testing.T) {
	var buf bytes.Buffer
	s := testSummary()

	if err := JSON(&buf, s); err != nil {
		t.Fatalf("JSON() error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("JSON output is not valid JSON: %v", err)
	}

	if result["has_destroys"] != true {
		t.Error("has_destroys should be true")
	}

	totalChanges, ok := result["total_changes"].(float64)
	if !ok || int(totalChanges) != 5 {
		t.Errorf("total_changes = %v, want 5", result["total_changes"])
	}

	groups, ok := result["groups"].([]interface{})
	if !ok {
		t.Fatal("groups should be an array")
	}
	if len(groups) != 4 {
		t.Errorf("groups count = %d, want 4", len(groups))
	}
}
