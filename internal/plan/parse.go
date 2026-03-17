package plan

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// ParseFile reads and parses a Terraform plan JSON file from disk.
func ParseFile(path string) (*TerraformPlan, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening plan file: %w", err)
	}
	defer f.Close()
	return Parse(f)
}

// Parse decodes a Terraform plan JSON from a reader.
// Uses streaming decode to stay efficient on large plans.
func Parse(r io.Reader) (*TerraformPlan, error) {
	var plan TerraformPlan
	dec := json.NewDecoder(r)
	if err := dec.Decode(&plan); err != nil {
		return nil, fmt.Errorf("decoding plan JSON: %w", err)
	}
	return &plan, nil
}
