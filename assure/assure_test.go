package assure

import (
	"testing"

	"github.com/plexusone/systemspec-architecture/sas"
)

func testArchitecture() *sas.Architecture {
	return &sas.Architecture{
		Nodes: []sas.Node{
			{
				ID: "api", Kind: sas.NodeKindService, Name: "API",
				Assurance: &sas.Assurance{
					Tests:   []string{"test://integration/api"},
					Metrics: []string{"otel://service/api"},
				},
			},
			{ID: "db", Kind: sas.NodeKindDataDatabase, Name: "DB"}, // no assurance at all
		},
		Relationships: []sas.Relationship{
			{
				ID: "api-to-db", From: "api", To: "db",
				Assurance: &sas.Assurance{
					Tests:      []string{"test://integration/api-db"},
					Metrics:    []string{"otel://edge/api-db"},
					Detections: []string{"siem://rule/api-db"},
					Deployment: []string{"aws://rds/db"},
				},
			},
		},
	}
}

func TestAssure_Counts(t *testing.T) {
	report := Assure(testArchitecture())

	if report.Nodes.Total != 2 {
		t.Fatalf("expected 2 nodes, got %d", report.Nodes.Total)
	}
	if report.Nodes.Tests != 1 || report.Nodes.Metrics != 1 {
		t.Errorf("expected 1 node with tests and 1 with metrics, got tests=%d metrics=%d", report.Nodes.Tests, report.Nodes.Metrics)
	}
	if report.Nodes.Detections != 0 || report.Nodes.Deployment != 0 {
		t.Errorf("expected 0 nodes with detections/deployment, got detections=%d deployment=%d", report.Nodes.Detections, report.Nodes.Deployment)
	}

	if report.Relationships.Total != 1 {
		t.Fatalf("expected 1 relationship, got %d", report.Relationships.Total)
	}
	if report.Relationships.Tests != 1 || report.Relationships.Metrics != 1 || report.Relationships.Detections != 1 || report.Relationships.Deployment != 1 {
		t.Errorf("expected fully-covered relationship, got %+v", report.Relationships)
	}
}

func TestAssure_Gaps(t *testing.T) {
	report := Assure(testArchitecture())

	var apiGap, dbGap *Gap
	for i := range report.Gaps {
		switch report.Gaps[i].ID {
		case "api":
			apiGap = &report.Gaps[i]
		case "db":
			dbGap = &report.Gaps[i]
		}
	}

	if apiGap == nil {
		t.Fatal("expected a gap for api (missing detections/deployment)")
	}
	if len(apiGap.Missing) != 2 {
		t.Errorf("expected api to be missing exactly 2 categories, got %v", apiGap.Missing)
	}

	if dbGap == nil {
		t.Fatal("expected a gap for db (no assurance at all)")
	}
	if len(dbGap.Missing) != 4 {
		t.Errorf("expected db to be missing all 4 categories, got %v", dbGap.Missing)
	}

	// api-to-db is fully covered, so it must not appear in Gaps at all.
	for _, g := range report.Gaps {
		if g.ID == "api-to-db" {
			t.Errorf("did not expect a gap for the fully-covered relationship, got %+v", g)
		}
	}
}

func TestCounts_Coverage(t *testing.T) {
	c := Counts{Total: 4, Tests: 3, Metrics: 2, Detections: 0, Deployment: 4}

	if got := c.Coverage(EvidenceTests); got != 0.75 {
		t.Errorf("Coverage(tests) = %v, want 0.75", got)
	}
	if got := c.Coverage(EvidenceMetrics); got != 0.5 {
		t.Errorf("Coverage(metrics) = %v, want 0.5", got)
	}
	if got := c.Coverage(EvidenceDetections); got != 0 {
		t.Errorf("Coverage(detections) = %v, want 0", got)
	}
	if got := c.Coverage(EvidenceDeployment); got != 1 {
		t.Errorf("Coverage(deployment) = %v, want 1", got)
	}
}

func TestCounts_Coverage_ZeroTotal(t *testing.T) {
	c := Counts{}
	if got := c.Coverage(EvidenceTests); got != 0 {
		t.Errorf("Coverage on zero-total Counts = %v, want 0 (not NaN/divide-by-zero)", got)
	}
}

func TestAssure_EmptyArchitecture(t *testing.T) {
	report := Assure(&sas.Architecture{})
	if report.Nodes.Total != 0 || report.Relationships.Total != 0 {
		t.Errorf("expected zero counts for an empty architecture, got %+v", report)
	}
	if len(report.Gaps) != 0 {
		t.Errorf("expected no gaps for an empty architecture, got %+v", report.Gaps)
	}
}
