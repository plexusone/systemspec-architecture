package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/plexusone/systemspec-architecture/sas"
	"github.com/plexusone/systemspec-architecture/validate"
)

// ValidateOptions configures a Validate run.
type ValidateOptions struct {
	// ArchitecturePath is the path to a SAS architecture JSON document.
	// Required.
	ArchitecturePath string

	// Profiles are the profiles to validate against, in addition to the
	// always-on referential-integrity checks.
	Profiles []validate.Profile
}

// ValidateResult is the outcome of a Validate run.
type ValidateResult struct {
	// Findings lists every rule failure. Empty means clean.
	Findings []validate.Finding
}

// OK reports whether the result has no error-severity findings. Warnings
// alone do not fail validation.
func (r ValidateResult) OK() bool {
	for _, f := range r.Findings {
		if f.Severity == validate.SeverityError {
			return false
		}
	}
	return true
}

// Validate loads the architecture document at opts.ArchitecturePath and
// runs validate.Validate against it with the requested profiles.
func Validate(opts ValidateOptions) (ValidateResult, error) {
	data, err := os.ReadFile(opts.ArchitecturePath)
	if err != nil {
		return ValidateResult{}, fmt.Errorf("read %s: %w", opts.ArchitecturePath, err)
	}

	var arch sas.Architecture
	if err := json.Unmarshal(data, &arch); err != nil {
		return ValidateResult{}, fmt.Errorf("parse %s: %w", opts.ArchitecturePath, err)
	}

	findings := validate.Validate(&arch, opts.Profiles...)
	return ValidateResult{Findings: findings}, nil
}

// FormatValidateReport renders a ValidateResult as either human-readable
// console text (format == "console" or "") or JSON (format == "json").
func FormatValidateReport(result ValidateResult, format string) (string, error) {
	switch format {
	case "", "console":
		return formatValidateConsole(result), nil
	case "json":
		return formatValidateJSON(result)
	default:
		return "", fmt.Errorf("unknown format %q (want console or json)", format)
	}
}

func formatValidateConsole(result ValidateResult) string {
	if len(result.Findings) == 0 {
		return "✅ No issues found\n"
	}

	var b strings.Builder
	errorCount, warningCount := 0, 0
	for _, f := range result.Findings {
		profile := "core"
		if f.Profile != "" {
			profile = string(f.Profile)
		}
		fmt.Fprintf(&b, "[%s] %s (%s) %s: %s\n", f.Severity, f.RuleID, profile, f.Path, f.Message)
		switch f.Severity {
		case validate.SeverityError:
			errorCount++
		case validate.SeverityWarning:
			warningCount++
		}
	}
	fmt.Fprintf(&b, "\nSummary: %d error(s), %d warning(s)\n", errorCount, warningCount)
	return b.String()
}

func formatValidateJSON(result ValidateResult) (string, error) {
	findings := result.Findings
	if findings == nil {
		findings = []validate.Finding{}
	}
	data, err := json.MarshalIndent(struct {
		OK       bool               `json:"ok"`
		Findings []validate.Finding `json:"findings"`
	}{OK: result.OK(), Findings: findings}, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal validate report: %w", err)
	}
	return string(data) + "\n", nil
}
