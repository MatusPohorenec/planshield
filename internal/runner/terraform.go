package runner

import (
	"bytes"
	"fmt"
	"os/exec"
)

// TerraformShowJSON invokes `terraform show -json <planfile>` and returns the
// raw JSON bytes. This is used when the user provides a binary plan file
// instead of pre-generated JSON.
func TerraformShowJSON(planPath string) ([]byte, error) {
	cmd := exec.Command("terraform", "show", "-json", planPath)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("terraform show -json failed: %w\nstderr: %s", err, stderr.String())
	}
	return stdout.Bytes(), nil
}
