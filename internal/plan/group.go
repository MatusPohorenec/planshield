package plan

import "sort"

// Action is a normalized action label for grouping.
type Action string

const (
	Create  Action = "Create"
	Update  Action = "Update"
	Delete  Action = "Delete"
	Replace Action = "Replace"
	Read    Action = "Read"
	NoOp    Action = "No-Op"
)

// GroupedChange holds the addresses for a single action bucket.
type GroupedChange struct {
	Action    Action
	Addresses []string
}

// Summary is the final structured output of a plan analysis.
type Summary struct {
	TerraformVersion string
	Groups           []GroupedChange
	Counts           map[Action]int
	HasDestroys      bool
}

// Summarize groups resource changes by action and returns a Summary.
func Summarize(p *TerraformPlan) *Summary {
	buckets := make(map[Action][]string)

	for _, rc := range p.ResourceChanges {
		action := classifyActions(rc.Change.Actions)
		if action == NoOp {
			continue
		}
		buckets[action] = append(buckets[action], rc.Address)
	}

	// Sort addresses within each bucket for deterministic output.
	for _, addrs := range buckets {
		sort.Strings(addrs)
	}

	// Build ordered groups: Delete and Replace first (highest risk), then Update, Create, Read.
	order := []Action{Delete, Replace, Update, Create, Read}
	var groups []GroupedChange
	counts := make(map[Action]int)
	hasDestroys := false

	for _, a := range order {
		addrs, ok := buckets[a]
		if !ok {
			continue
		}
		groups = append(groups, GroupedChange{Action: a, Addresses: addrs})
		counts[a] = len(addrs)
		if a == Delete || a == Replace {
			hasDestroys = true
		}
	}

	return &Summary{
		TerraformVersion: p.TerraformVersion,
		Groups:           groups,
		Counts:           counts,
		HasDestroys:      hasDestroys,
	}
}

// classifyActions maps the Terraform actions array to a single Action label.
// A replace is modeled as ["delete","create"] or ["create","delete"].
func classifyActions(actions []string) Action {
	if len(actions) == 0 {
		return NoOp
	}
	if len(actions) == 2 {
		has := func(a string) bool {
			return actions[0] == a || actions[1] == a
		}
		if has(ActionCreate) && has(ActionDelete) {
			return Replace
		}
	}
	switch actions[0] {
	case ActionCreate:
		return Create
	case ActionUpdate:
		return Update
	case ActionDelete:
		return Delete
	case ActionRead:
		return Read
	default:
		return NoOp
	}
}

// TotalChanges returns the total number of non-noop resource changes.
func (s *Summary) TotalChanges() int {
	total := 0
	for _, c := range s.Counts {
		total += c
	}
	return total
}
