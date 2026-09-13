package cli

import (
	"strings"
	"testing"

	"github.com/plexusone/systemspec-architecture/assure"
)

const architectureForAssureJSON = `{
  "version": "0.1",
  "nodes": [
    {
      "id": "api", "kind": "service", "name": "API",
      "assurance": {"tests": ["test://a"], "metrics": ["otel://a"]}
    },
    {"id": "db", "kind": "data.database", "name": "DB"}
  ],
  "relationships": [
    {"id": "api-to-db", "from": "api", "to": "db"}
  ]
}`

func TestAssure(t *testing.T) {
	path := writeTempArchitecture(t, architectureForAssureJSON)

	report, err := Assure(AssureOptions{ArchitecturePath: path})
	if err != nil {
		t.Fatalf("Assure: %v", err)
	}
	if report.Nodes.Total != 2 || report.Nodes.Tests != 1 {
		t.Errorf("unexpected node counts: %+v", report.Nodes)
	}
	if len(report.Gaps) != 3 { // api (2 categories), db (4), api-to-db (4) -> 3 elements with gaps
		t.Errorf("expected 3 elements with gaps, got %d: %+v", len(report.Gaps), report.Gaps)
	}
}

func TestAssure_MissingFile(t *testing.T) {
	if _, err := Assure(AssureOptions{ArchitecturePath: "/nonexistent/architecture.json"}); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestFormatAssureReport_Console(t *testing.T) {
	report := assure.Report{
		Nodes: assure.Counts{Total: 2, Tests: 1},
		Gaps:  []assure.Gap{{Kind: assure.ElementKindNode, ID: "db", Missing: []assure.EvidenceCategory{assure.EvidenceTests}}},
	}
	out, err := FormatAssureReport(report, "console")
	if err != nil {
		t.Fatalf("FormatAssureReport: %v", err)
	}
	if !strings.Contains(out, "50.0%") {
		t.Errorf("expected 50%% node test coverage in output, got %q", out)
	}
	if !strings.Contains(out, `"db"`) {
		t.Errorf("expected gap for db in output, got %q", out)
	}
}

func TestFormatAssureReport_ConsoleNoGaps(t *testing.T) {
	out, err := FormatAssureReport(assure.Report{}, "console")
	if err != nil {
		t.Fatalf("FormatAssureReport: %v", err)
	}
	if !strings.Contains(out, "No gaps found") {
		t.Errorf("expected no-gaps message, got %q", out)
	}
	if !strings.Contains(out, "n/a") {
		t.Errorf("expected n/a for zero-element counts, got %q", out)
	}
}

func TestFormatAssureReport_JSON(t *testing.T) {
	out, err := FormatAssureReport(assure.Report{Nodes: assure.Counts{Total: 1}}, "json")
	if err != nil {
		t.Fatalf("FormatAssureReport: %v", err)
	}
	if !strings.Contains(out, `"total": 1`) {
		t.Errorf("expected total in JSON output, got %q", out)
	}
}

func TestFormatAssureReport_UnknownFormat(t *testing.T) {
	if _, err := FormatAssureReport(assure.Report{}, "yaml"); err == nil {
		t.Fatal("expected error for unknown format")
	}
}
