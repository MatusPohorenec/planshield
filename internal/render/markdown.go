package render

import (
	"fmt"
	"io"
	"strings"

	"github.com/planshield/planshield/internal/plan"
)

// actionEmoji maps actions to visual indicators for Markdown output.
var actionEmoji = map[plan.Action]string{
	plan.Create:  "+",
	plan.Update:  "~",
	plan.Delete:  "-",
	plan.Replace: "!",
	plan.Read:    "<=",
}

// Markdown writes a Markdown-formatted summary to w.
func Markdown(w io.Writer, s *plan.Summary) error {
	fmt.Fprintf(w, "## PlanShield Summary\n\n")

	if s.TerraformVersion != "" {
		fmt.Fprintf(w, "Terraform version: `%s`\n\n", s.TerraformVersion)
	}

	if s.TotalChanges() == 0 {
		fmt.Fprintf(w, "**No changes.** Infrastructure is up-to-date.\n")
		return nil
	}

	// Counts line
	parts := make([]string, 0, 4)
	for _, a := range []plan.Action{plan.Create, plan.Update, plan.Delete, plan.Replace} {
		if c, ok := s.Counts[a]; ok && c > 0 {
			parts = append(parts, fmt.Sprintf("**%d %s**", c, a))
		}
	}
	fmt.Fprintf(w, "%s\n\n", strings.Join(parts, " · "))

	if s.HasDestroys {
		fmt.Fprintf(w, "> **⚠ Destructive changes detected.** Review carefully before applying.\n\n")
	}

	// Resource table per action group
	for _, g := range s.Groups {
		sym := actionEmoji[g.Action]
		fmt.Fprintf(w, "### %s %s (%d)\n\n", sym, g.Action, len(g.Addresses))
		fmt.Fprintf(w, "| Resource Address |\n")
		fmt.Fprintf(w, "|------------------|\n")
		for _, addr := range g.Addresses {
			fmt.Fprintf(w, "| `%s` |\n", addr)
		}
		fmt.Fprintln(w)
	}

	return nil
}
