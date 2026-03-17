package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/planshield/planshield/internal/plan"
	"github.com/planshield/planshield/internal/redact"
	"github.com/planshield/planshield/internal/render"
	"github.com/planshield/planshield/internal/runner"
)

// Exit codes for CI integration.
const (
	ExitOK             = 0
	ExitHasDestroys    = 2
	ExitError          = 1
)

func main() {
	os.Exit(run())
}

func run() int {
	var (
		formatFlag       string
		redactPatternsFlag string
		noDefaultRedact  bool
		planFile         string
	)

	flag.StringVar(&formatFlag, "format", "md", "Output format: md or json")
	flag.StringVar(&redactPatternsFlag, "redact-patterns", "", "Comma-separated extra regex patterns for redaction")
	flag.BoolVar(&noDefaultRedact, "no-default-redact", false, "Disable built-in secret-detection patterns")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: planshield [flags] <plan-file>\n\n")
		fmt.Fprintf(os.Stderr, "Render a Terraform plan into a minimal, safe-to-share change list.\n\n")
		fmt.Fprintf(os.Stderr, "The <plan-file> can be:\n")
		fmt.Fprintf(os.Stderr, "  - A JSON file produced by `terraform show -json tfplan`\n")
		fmt.Fprintf(os.Stderr, "  - A binary plan file (requires `terraform` on PATH)\n")
		fmt.Fprintf(os.Stderr, "  - \"-\" to read JSON from stdin\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExit codes:\n")
		fmt.Fprintf(os.Stderr, "  0  Success, no destructive changes\n")
		fmt.Fprintf(os.Stderr, "  1  Error\n")
		fmt.Fprintf(os.Stderr, "  2  Success, but plan contains destroys or replaces\n")
	}
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: plan file argument required. Use -h for help.")
		return ExitError
	}
	planFile = flag.Arg(0)

	if formatFlag != "md" && formatFlag != "json" {
		fmt.Fprintf(os.Stderr, "Error: unsupported format %q (use md or json)\n", formatFlag)
		return ExitError
	}

	// Compile extra redact patterns.
	var extraPatterns []string
	if redactPatternsFlag != "" {
		extraPatterns = strings.Split(redactPatternsFlag, ",")
	}
	extraCompiled, err := redact.CompilePatterns(extraPatterns)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid redaction pattern: %v\n", err)
		return ExitError
	}
	redactor := redact.New(!noDefaultRedact, extraCompiled)

	// Parse plan.
	tfPlan, err := loadPlan(planFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return ExitError
	}

	// Scrub sensitive values before any rendering.
	redactor.RedactPlan(tfPlan)

	summary := plan.Summarize(tfPlan)

	// Render output.
	switch formatFlag {
	case "md":
		if err := render.Markdown(os.Stdout, summary); err != nil {
			fmt.Fprintf(os.Stderr, "Error rendering markdown: %v\n", err)
			return ExitError
		}
	case "json":
		if err := render.JSON(os.Stdout, summary); err != nil {
			fmt.Fprintf(os.Stderr, "Error rendering JSON: %v\n", err)
			return ExitError
		}
	}

	if summary.HasDestroys {
		return ExitHasDestroys
	}
	return ExitOK
}

// loadPlan loads a Terraform plan from a file path or stdin.
// It detects whether the input is JSON or binary and handles accordingly.
func loadPlan(path string) (*plan.TerraformPlan, error) {
	if path == "-" {
		return plan.Parse(os.Stdin)
	}

	// Read the first few bytes to detect if it's JSON.
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading plan file: %w", err)
	}

	// If it looks like JSON, parse directly.
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) > 0 && trimmed[0] == '{' {
		return plan.Parse(bytes.NewReader(data))
	}

	// Otherwise, try terraform show -json on the binary plan.
	jsonBytes, err := runner.TerraformShowJSON(path)
	if err != nil {
		return nil, err
	}

	// Validate it's JSON before parsing.
	if !json.Valid(jsonBytes) {
		return nil, fmt.Errorf("terraform show -json produced invalid JSON")
	}
	return plan.Parse(bytes.NewReader(jsonBytes))
}
