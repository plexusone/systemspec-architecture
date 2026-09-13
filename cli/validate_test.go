package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/plexusone/systemspec-architecture/validate"
)

func writeTempArchitecture(t *testing.T, contents string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "architecture.json")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write temp architecture: %v", err)
	}
	return path
}

const cleanArchitectureJSON = `{
  "version": "0.1",
  "nodes": [
    {"id": "a", "kind": "service", "name": "A"},
    {"id": "b", "kind": "service", "name": "B"}
  ],
  "relationships": [
    {"id": "a-to-b", "from": "a", "to": "b", "kind": "calls"}
  ]
}`

const invalidReferenceArchitectureJSON = `{
  "version": "0.1",
  "nodes": [{"id": "a", "kind": "service", "name": "A"}],
  "relationships": [
    {"id": "a-to-missing", "from": "a", "to": "missing", "kind": "calls"}
  ]
}`

func TestValidate_CleanArchitecture(t *testing.T) {
	path := writeTempArchitecture(t, cleanArchitectureJSON)

	result, err := Validate(ValidateOptions{ArchitecturePath: path})
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if !result.OK() {
		t.Errorf("expected OK, got findings %+v", result.Findings)
	}
}

func TestValidate_DanglingReference(t *testing.T) {
	path := writeTempArchitecture(t, invalidReferenceArchitectureJSON)

	result, err := Validate(ValidateOptions{ArchitecturePath: path})
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if result.OK() {
		t.Fatal("expected dangling reference to fail validation")
	}
	if len(result.Findings) == 0 || result.Findings[0].RuleID != "core.relationship-endpoints-resolve" {
		t.Errorf("expected core.relationship-endpoints-resolve finding, got %+v", result.Findings)
	}
}

func TestValidate_MissingFile(t *testing.T) {
	_, err := Validate(ValidateOptions{ArchitecturePath: "/nonexistent/architecture.json"})
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestValidate_ProfileApplied(t *testing.T) {
	path := writeTempArchitecture(t, cleanArchitectureJSON)

	result, err := Validate(ValidateOptions{ArchitecturePath: path, Profiles: []validate.Profile{validate.ProfileDeployment}})
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if result.OK() {
		t.Fatal("expected deployment profile to flag missing technology/boundaries")
	}
}

func TestFormatValidateReport_ConsoleClean(t *testing.T) {
	out, err := FormatValidateReport(ValidateResult{}, "console")
	if err != nil {
		t.Fatalf("FormatValidateReport: %v", err)
	}
	if !strings.Contains(out, "No issues found") {
		t.Errorf("expected clean-report message, got %q", out)
	}
}

func TestFormatValidateReport_ConsoleWithFindings(t *testing.T) {
	result := ValidateResult{Findings: []validate.Finding{
		{RuleID: "core.relationship-endpoints-resolve", Severity: validate.SeverityError, Path: "relationships[0]", Message: "boom"},
	}}
	out, err := FormatValidateReport(result, "console")
	if err != nil {
		t.Fatalf("FormatValidateReport: %v", err)
	}
	if !strings.Contains(out, "core.relationship-endpoints-resolve") || !strings.Contains(out, "1 error(s)") {
		t.Errorf("expected finding and summary in output, got %q", out)
	}
}

func TestFormatValidateReport_JSON(t *testing.T) {
	result := ValidateResult{Findings: []validate.Finding{
		{RuleID: "core.relationship-endpoints-resolve", Severity: validate.SeverityError, Path: "relationships[0]", Message: "boom"},
	}}
	out, err := FormatValidateReport(result, "json")
	if err != nil {
		t.Fatalf("FormatValidateReport: %v", err)
	}
	if !strings.Contains(out, `"ok": false`) || !strings.Contains(out, "core.relationship-endpoints-resolve") {
		t.Errorf("expected JSON with ok:false and finding, got %q", out)
	}
}

func TestFormatValidateReport_JSON_Clean(t *testing.T) {
	out, err := FormatValidateReport(ValidateResult{}, "json")
	if err != nil {
		t.Fatalf("FormatValidateReport: %v", err)
	}
	if !strings.Contains(out, `"ok": true`) || !strings.Contains(out, `"findings": []`) {
		t.Errorf("expected JSON with ok:true and empty findings array, got %q", out)
	}
}

func TestFormatValidateReport_UnknownFormat(t *testing.T) {
	if _, err := FormatValidateReport(ValidateResult{}, "yaml"); err == nil {
		t.Fatal("expected error for unknown format")
	}
}
