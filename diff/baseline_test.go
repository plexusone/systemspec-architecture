package diff

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/plexusone/systemspec-architecture/sas"
)

func TestBaseline_JSONRoundTrip(t *testing.T) {
	original := Baseline{
		ID:   "baseline-2026-09",
		Name: "Approved 2026-09 architecture",
		Architecture: sas.Architecture{
			Version: "0.1",
			Nodes:   []sas.Node{{ID: "api", Kind: sas.NodeKindService, Name: "API"}},
		},
		ApprovedAt: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		ApprovedBy: "security-team",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var round Baseline
	if err := json.Unmarshal(data, &round); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if round.ID != original.ID || round.Name != original.Name || round.ApprovedBy != original.ApprovedBy {
		t.Fatalf("round-trip mismatch: got %+v, want %+v", round, original)
	}
	if len(round.Architecture.Nodes) != 1 || round.Architecture.Nodes[0].ID != "api" {
		t.Fatalf("architecture did not round-trip: %+v", round.Architecture)
	}
}

func TestAssessment_JSONRoundTrip_ArbitraryDecisionString(t *testing.T) {
	// Decision is intentionally a free string at this layer; a compliance
	// profile package owns its own closed vocabulary.
	original := Assessment{
		ID: "assess-1", BaselineID: "baseline-2026-09",
		Decision: "change_assessment_required", Reviewer: "jane",
		DecidedAt: time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC),
		Evidence:  []string{"impact://new_external_dependency/api-to-payments"},
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var round Assessment
	if err := json.Unmarshal(data, &round); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if round.ID != original.ID || round.BaselineID != original.BaselineID ||
		round.Decision != original.Decision || round.Reviewer != original.Reviewer ||
		!round.DecidedAt.Equal(original.DecidedAt) || len(round.Evidence) != len(original.Evidence) {
		t.Fatalf("round-trip mismatch: got %+v, want %+v", round, original)
	}
}

func TestChangeReview_ChangeSetAndImpact(t *testing.T) {
	baseline := Baseline{
		ID:   "baseline-1",
		Name: "Approved",
		Architecture: sas.Architecture{
			Nodes: []sas.Node{
				{ID: "api", Kind: sas.NodeKindService, Name: "API"},
			},
		},
	}
	proposed := sas.Architecture{
		Nodes: []sas.Node{
			{ID: "api", Kind: sas.NodeKindService, Name: "API"},
			{ID: "payments", Kind: sas.NodeKindExternalService, Name: "Payments"},
		},
		Relationships: []sas.Relationship{
			{ID: "api-to-payments", From: "api", To: "payments", Kind: sas.RelationKindCalls},
		},
	}
	review := ChangeReview{ID: "review-1", Baseline: baseline, Proposed: proposed}

	cs := review.ChangeSet()
	if len(cs.Changes) != 2 { // node_added(payments) + relationship_added(api-to-payments)
		t.Fatalf("expected 2 changes, got %d: %+v", len(cs.Changes), cs.Changes)
	}

	impact := review.Impact()
	found := false
	for _, imp := range impact.Impacts {
		if imp.Kind == ImpactKindNewExternalDependency {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected new_external_dependency impact, got %+v", impact.Impacts)
	}
}

func TestChangeReview_AssessmentNilUntilRecorded(t *testing.T) {
	review := ChangeReview{ID: "review-1"}
	if review.Assessment != nil {
		t.Fatalf("expected nil Assessment on a fresh ChangeReview, got %+v", review.Assessment)
	}
}
