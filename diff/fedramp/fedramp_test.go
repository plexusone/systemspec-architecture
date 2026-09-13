package fedramp

import (
	"testing"

	"github.com/plexusone/systemspec-architecture/diff"
	"github.com/plexusone/systemspec-architecture/sas"
)

func TestClassify_NoImpacts_NoReviewRequired(t *testing.T) {
	got := Classify(diff.ChangeImpact{})
	if got.Decision != DecisionNoReviewRequired {
		t.Fatalf("expected no_review_required, got %s", got.Decision)
	}
	if len(got.TriggeringImpacts) != 0 || len(got.AffectedControls) != 0 {
		t.Fatalf("expected no triggering impacts or controls, got %+v", got)
	}
}

func TestClassify_NewExternalDependency_PotentialSignificantChange(t *testing.T) {
	impact := diff.ChangeImpact{Impacts: []diff.Impact{
		{Kind: diff.ImpactKindNewExternalDependency, ElementType: diff.ElementTypeRelationship, ElementID: "api-to-payments"},
	}}
	got := Classify(impact)
	if got.Decision != DecisionPotentialSignificantChange {
		t.Fatalf("expected potential_significant_change, got %s", got.Decision)
	}
	if len(got.TriggeringImpacts) != 1 {
		t.Fatalf("expected 1 triggering impact, got %+v", got.TriggeringImpacts)
	}
	wantControls := map[string]bool{"CA-3": true, "SA-9": true}
	if len(got.AffectedControls) != len(wantControls) {
		t.Fatalf("expected controls %v, got %v", wantControls, got.AffectedControls)
	}
	for _, c := range got.AffectedControls {
		if !wantControls[c] {
			t.Errorf("unexpected control %s", c)
		}
	}
}

func TestClassify_EntitlementExpansionOnly_ChangeAssessmentRequired(t *testing.T) {
	impact := diff.ChangeImpact{Impacts: []diff.Impact{
		{Kind: diff.ImpactKindEntitlementExpansion, ElementType: diff.ElementTypeRelationship, ElementID: "api-to-db"},
	}}
	got := Classify(impact)
	if got.Decision != DecisionChangeAssessmentRequired {
		t.Fatalf("expected change_assessment_required, got %s", got.Decision)
	}
}

func TestClassify_MixedSeverity_TakesMostSevere(t *testing.T) {
	impact := diff.ChangeImpact{Impacts: []diff.Impact{
		{Kind: diff.ImpactKindEntitlementExpansion, ElementID: "a"},
		{Kind: diff.ImpactKindNewBoundaryCrossing, ElementID: "b"},
		{Kind: diff.ImpactKindComputeModelChanged, ElementID: "c"},
	}}
	got := Classify(impact)
	if got.Decision != DecisionPotentialSignificantChange {
		t.Fatalf("expected potential_significant_change (most severe), got %s", got.Decision)
	}
	// Only the potential-significant-change-tier impact should be
	// "triggering" — the lower-severity ones didn't drive the outcome.
	if len(got.TriggeringImpacts) != 1 || got.TriggeringImpacts[0].ElementID != "b" {
		t.Fatalf("expected only element 'b' as triggering, got %+v", got.TriggeringImpacts)
	}
}

func TestClassify_EndToEnd_NewExternalDependencyScenario(t *testing.T) {
	base := &sas.Architecture{
		Nodes: []sas.Node{{ID: "api", Kind: sas.NodeKindService, Name: "API"}},
	}
	proposed := &sas.Architecture{
		Nodes: []sas.Node{
			{ID: "api", Kind: sas.NodeKindService, Name: "API"},
			{ID: "payments", Kind: sas.NodeKindExternalService, Name: "Payments"},
		},
		Relationships: []sas.Relationship{
			{ID: "api-to-payments", From: "api", To: "payments", Kind: sas.RelationKindCalls},
		},
	}
	changeSet := diff.Diff(base, proposed)
	impact := diff.Classify(changeSet, proposed)
	got := Classify(impact)

	if got.Decision != DecisionPotentialSignificantChange {
		t.Fatalf("expected potential_significant_change end to end, got %s: %+v", got.Decision, got)
	}
	if got.Rationale == "" {
		t.Fatal("expected a non-empty rationale")
	}
}
