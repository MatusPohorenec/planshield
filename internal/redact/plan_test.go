package redact

import (
	"os"
	"testing"

	"github.com/planshield/planshield/internal/plan"
)

func TestRedactPlanScrubsSensitiveValues(t *testing.T) {
	f, err := os.Open("../../testdata/mixed_plan.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	p, err := plan.Parse(f)
	if err != nil {
		t.Fatal(err)
	}

	r := New(true, nil)
	r.RedactPlan(p)

	// The db_instance has password marked sensitive in both before and after.
	var dbChange *plan.ResourceChange
	for i := range p.ResourceChanges {
		if p.ResourceChanges[i].Address == "aws_db_instance.main" {
			dbChange = &p.ResourceChanges[i]
			break
		}
	}
	if dbChange == nil {
		t.Fatal("aws_db_instance.main not found")
	}

	if dbChange.Change.Before["password"] != "REDACTED" {
		t.Errorf("before password = %v, want REDACTED", dbChange.Change.Before["password"])
	}
	if dbChange.Change.After["password"] != "REDACTED" {
		t.Errorf("after password = %v, want REDACTED", dbChange.Change.After["password"])
	}
	// Non-sensitive fields should be untouched.
	if dbChange.Change.After["engine"] != "postgres" {
		t.Errorf("after engine = %v, want postgres", dbChange.Change.After["engine"])
	}

	// Output db_connection_string is marked after_sensitive=true and contains a connection string.
	oc, ok := p.OutputChanges["db_connection_string"]
	if !ok {
		t.Fatal("output db_connection_string not found")
	}
	if oc.After != "REDACTED" {
		t.Errorf("output after = %v, want REDACTED", oc.After)
	}
}

func TestRedactPlanCatchesPatternInUnsensitiveField(t *testing.T) {
	p := &plan.TerraformPlan{
		ResourceChanges: []plan.ResourceChange{
			{
				Address: "aws_ssm_parameter.key",
				Change: plan.Change{
					Actions:         []string{"create"},
					Before:          nil,
					After:           map[string]interface{}{"value": "AKIAIOSFODNN7EXAMPLE"},
					BeforeSensitive: false,
					AfterSensitive:  map[string]interface{}{}, // not marked sensitive!
				},
			},
		},
	}

	r := New(true, nil)
	r.RedactPlan(p)

	got := p.ResourceChanges[0].Change.After["value"]
	if got != "REDACTED" {
		t.Errorf("AWS key not caught by pattern: got %v, want REDACTED", got)
	}
}
