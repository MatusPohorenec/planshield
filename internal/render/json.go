package render

import (
	"encoding/json"
	"io"

	"github.com/planshield/planshield/internal/plan"
)

// jsonOutput is the structured JSON output format.
type jsonOutput struct {
	TerraformVersion string        `json:"terraform_version,omitempty"`
	HasDestroys      bool          `json:"has_destroys"`
	TotalChanges     int           `json:"total_changes"`
	Counts           map[string]int `json:"counts"`
	Groups           []jsonGroup   `json:"groups"`
}

type jsonGroup struct {
	Action    string   `json:"action"`
	Count     int      `json:"count"`
	Addresses []string `json:"addresses"`
}

// JSON writes a JSON-formatted summary to w.
func JSON(w io.Writer, s *plan.Summary) error {
	counts := make(map[string]int, len(s.Counts))
	for a, c := range s.Counts {
		counts[string(a)] = c
	}

	groups := make([]jsonGroup, 0, len(s.Groups))
	for _, g := range s.Groups {
		groups = append(groups, jsonGroup{
			Action:    string(g.Action),
			Count:     len(g.Addresses),
			Addresses: g.Addresses,
		})
	}

	out := jsonOutput{
		TerraformVersion: s.TerraformVersion,
		HasDestroys:      s.HasDestroys,
		TotalChanges:     s.TotalChanges(),
		Counts:           counts,
		Groups:           groups,
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
