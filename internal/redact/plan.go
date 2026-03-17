package redact

import (
	"github.com/planshield/planshield/internal/plan"
)

// RedactPlan scrubs sensitive values from all resource changes and output
// changes in the plan, in place. This ensures that even if output formats
// are extended to include values, no secrets leak.
func (r *Redactor) RedactPlan(p *plan.TerraformPlan) {
	for i := range p.ResourceChanges {
		rc := &p.ResourceChanges[i]
		rc.Change.Before = r.RedactMap(rc.Change.Before, rc.Change.BeforeSensitive)
		rc.Change.After = r.RedactMap(rc.Change.After, rc.Change.AfterSensitive)
	}

	for name, oc := range p.OutputChanges {
		oc.Before = r.RedactValue(oc.Before, oc.BeforeSensitive)
		oc.After = r.RedactValue(oc.After, oc.AfterSensitive)
		p.OutputChanges[name] = oc
	}
}
