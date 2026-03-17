// Package plan defines the data model for Terraform plan JSON and provides
// parsing, grouping, and redaction of resource changes.
package plan

// TerraformPlan represents the top-level structure of `terraform show -json`.
type TerraformPlan struct {
	FormatVersion   string          `json:"format_version"`
	TerraformVersion string         `json:"terraform_version"`
	PlannedValues   *PlannedValues  `json:"planned_values,omitempty"`
	ResourceChanges []ResourceChange `json:"resource_changes"`
	OutputChanges   map[string]OutputChange `json:"output_changes,omitempty"`
}

// PlannedValues holds the planned state after apply.
type PlannedValues struct {
	RootModule *Module `json:"root_module,omitempty"`
}

// Module represents a Terraform module in planned values.
type Module struct {
	Resources    []PlannedResource `json:"resources,omitempty"`
	ChildModules []Module          `json:"child_modules,omitempty"`
}

// PlannedResource is a resource inside planned values.
type PlannedResource struct {
	Address string `json:"address"`
	Type    string `json:"type"`
	Name    string `json:"name"`
}

// ResourceChange represents a single resource change entry.
type ResourceChange struct {
	Address      string `json:"address"`
	ModuleAddress string `json:"module_address,omitempty"`
	Mode         string `json:"mode"`   // "managed" or "data"
	Type         string `json:"type"`
	Name         string `json:"name"`
	ProviderName string `json:"provider_name"`
	Change       Change `json:"change"`
	ActionReason string `json:"action_reason,omitempty"`
}

// Change describes what happens to a resource.
type Change struct {
	Actions         []string               `json:"actions"`
	Before          map[string]interface{} `json:"before"`
	After           map[string]interface{} `json:"after"`
	AfterUnknown    map[string]interface{} `json:"after_unknown,omitempty"`
	BeforeSensitive interface{}            `json:"before_sensitive,omitempty"`
	AfterSensitive  interface{}            `json:"after_sensitive,omitempty"`
}

// OutputChange represents a change to a Terraform output.
type OutputChange struct {
	Actions         []string    `json:"actions"`
	Before          interface{} `json:"before"`
	After           interface{} `json:"after"`
	BeforeSensitive bool        `json:"before_sensitive,omitempty"`
	AfterSensitive  bool        `json:"after_sensitive,omitempty"`
}

// Action constants matching Terraform's plan JSON action strings.
const (
	ActionCreate = "create"
	ActionUpdate = "update"
	ActionDelete = "delete"
	ActionRead   = "read"
	ActionNoop   = "no-op"
)
