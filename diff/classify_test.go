package diff

import (
	"testing"

	"github.com/plexusone/systemspec-architecture/sas"
)

// The four scenarios PLAN.md's M4 milestone names as canonical, plus the
// two additional ImpactKinds the TRD specifies.

func TestClassify_NewExternalDependency(t *testing.T) {
	base := &sas.Architecture{
		Nodes: []sas.Node{{ID: "api", Kind: sas.NodeKindService, Name: "API"}},
	}
	proposed := &sas.Architecture{
		Nodes: []sas.Node{
			{ID: "api", Kind: sas.NodeKindService, Name: "API"},
			{ID: "payments", Kind: sas.NodeKindExternalService, Name: "Payments Provider"},
		},
		Relationships: []sas.Relationship{
			{ID: "api-to-payments", From: "api", To: "payments", Kind: sas.RelationKindCalls},
		},
	}
	impact := Classify(Diff(base, proposed), proposed)

	found := false
	for _, imp := range impact.Impacts {
		if imp.Kind == ImpactKindNewExternalDependency && imp.ElementID == "api-to-payments" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected new_external_dependency for api-to-payments, got %+v", impact.Impacts)
	}
}

func TestClassify_AuthnMechanismChanged(t *testing.T) {
	base := &sas.Architecture{
		Relationships: []sas.Relationship{{
			ID: "web-to-api", From: "web", To: "api", Kind: sas.RelationKindCalls,
			Identity: &sas.IdentityRef{Identity: &sas.Identity{Type: sas.IdentityTypeWorkload, Mechanism: sas.IdentityMechanismX509}},
		}},
	}
	proposed := &sas.Architecture{
		Relationships: []sas.Relationship{{
			ID: "web-to-api", From: "web", To: "api", Kind: sas.RelationKindCalls,
			Identity: &sas.IdentityRef{Identity: &sas.Identity{Type: sas.IdentityTypeWorkload, Mechanism: sas.IdentityMechanismOAuthClient}},
		}},
	}
	impact := Classify(Diff(base, proposed), proposed)
	if len(impact.Impacts) != 1 || impact.Impacts[0].Kind != ImpactKindAuthnMechanismChanged {
		t.Fatalf("expected 1 authn_mechanism_changed impact, got %+v", impact.Impacts)
	}
}

func TestClassify_EntitlementExpansion_ReadToReadWrite(t *testing.T) {
	base := &sas.Architecture{
		Relationships: []sas.Relationship{{
			ID: "api-to-db", From: "api", To: "db", Kind: sas.RelationKindDataAccess,
			Operations: []sas.Operation{sas.OperationRead},
		}},
	}
	proposed := &sas.Architecture{
		Relationships: []sas.Relationship{{
			ID: "api-to-db", From: "api", To: "db", Kind: sas.RelationKindDataAccess,
			Operations: []sas.Operation{sas.OperationRead, sas.OperationUpdate},
		}},
	}
	impact := Classify(Diff(base, proposed), proposed)
	if len(impact.Impacts) != 1 || impact.Impacts[0].Kind != ImpactKindEntitlementExpansion {
		t.Fatalf("expected 1 entitlement_expansion impact, got %+v", impact.Impacts)
	}
}

func TestClassify_OperationChange_WithoutWriteExpansion_IsNotEntitlementExpansion(t *testing.T) {
	// read -> read+discover should NOT be entitlement expansion: discover
	// is not a write operation.
	base := &sas.Architecture{
		Relationships: []sas.Relationship{{
			ID: "api-to-db", From: "api", To: "db", Kind: sas.RelationKindDataAccess,
			Operations: []sas.Operation{sas.OperationRead},
		}},
	}
	proposed := &sas.Architecture{
		Relationships: []sas.Relationship{{
			ID: "api-to-db", From: "api", To: "db", Kind: sas.RelationKindDataAccess,
			Operations: []sas.Operation{sas.OperationRead, sas.OperationDiscover},
		}},
	}
	impact := Classify(Diff(base, proposed), proposed)
	if len(impact.Impacts) != 0 {
		t.Fatalf("expected no impacts for a non-write operation addition, got %+v", impact.Impacts)
	}
}

func TestClassify_NewBoundaryCrossing(t *testing.T) {
	base := &sas.Architecture{
		Relationships: []sas.Relationship{{ID: "api-to-partner", From: "api", To: "partner", Kind: sas.RelationKindCalls}},
	}
	proposed := &sas.Architecture{
		Relationships: []sas.Relationship{{
			ID: "api-to-partner", From: "api", To: "partner", Kind: sas.RelationKindCalls,
			CrossesBoundaries: []string{"trust-external"},
		}},
	}
	impact := Classify(Diff(base, proposed), proposed)
	if len(impact.Impacts) != 1 || impact.Impacts[0].Kind != ImpactKindNewBoundaryCrossing {
		t.Fatalf("expected 1 new_boundary_crossing impact, got %+v", impact.Impacts)
	}
	if len(impact.Impacts[0].AffectedBoundaries) != 1 || impact.Impacts[0].AffectedBoundaries[0] != "trust-external" {
		t.Fatalf("expected AffectedBoundaries=[trust-external], got %+v", impact.Impacts[0].AffectedBoundaries)
	}
}

func TestClassify_DataClassificationChange(t *testing.T) {
	base := &sas.Architecture{
		Relationships: []sas.Relationship{{
			ID: "api-to-partner", From: "api", To: "partner", Kind: sas.RelationKindCalls,
			Data: &sas.DataFlow{Classifications: []string{"metadata"}},
		}},
	}
	proposed := &sas.Architecture{
		Relationships: []sas.Relationship{{
			ID: "api-to-partner", From: "api", To: "partner", Kind: sas.RelationKindCalls,
			Data: &sas.DataFlow{Classifications: []string{"metadata", "customer_data"}},
		}},
	}
	impact := Classify(Diff(base, proposed), proposed)
	if len(impact.Impacts) != 1 || impact.Impacts[0].Kind != ImpactKindDataClassificationChange {
		t.Fatalf("expected 1 data_classification_change impact, got %+v", impact.Impacts)
	}
}

func TestClassify_ComputeModelChanged(t *testing.T) {
	base := &sas.Architecture{
		Nodes: []sas.Node{{ID: "api", Kind: sas.NodeKindComputeInstance, Name: "API", Technology: &sas.Technology{Provider: "aws", Service: "ecs"}}},
	}
	proposed := &sas.Architecture{
		Nodes: []sas.Node{{ID: "api", Kind: sas.NodeKindComputeFunction, Name: "API", Technology: &sas.Technology{Provider: "aws", Service: "lambda"}}},
	}
	impact := Classify(Diff(base, proposed), proposed)
	found := false
	for _, imp := range impact.Impacts {
		if imp.Kind == ImpactKindComputeModelChanged {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected compute_model_changed impact, got %+v", impact.Impacts)
	}
}

func TestClassify_PRDExpressivenessTest_CombinedScenario(t *testing.T) {
	// "This PR adds an external trust-boundary crossing carrying customer
	// data through a new authorization mechanism" — the PRD's literal
	// expressiveness test, verified end to end from Diff through Classify.
	base := &sas.Architecture{
		Nodes: []sas.Node{
			{ID: "api", Kind: sas.NodeKindService, Name: "API"},
			{ID: "partner", Kind: sas.NodeKindExternalService, Name: "Partner"},
		},
	}
	proposed := &sas.Architecture{
		Nodes: []sas.Node{
			{ID: "api", Kind: sas.NodeKindService, Name: "API"},
			{ID: "partner", Kind: sas.NodeKindExternalService, Name: "Partner"},
		},
		Relationships: []sas.Relationship{{
			ID: "api-to-partner", From: "api", To: "partner", Kind: sas.RelationKindCalls,
			CrossesBoundaries: []string{"trust-external"},
			Data:              &sas.DataFlow{Classifications: []string{"customer_data"}},
			Identity:          &sas.IdentityRef{Identity: &sas.Identity{Type: sas.IdentityTypeWorkload, Mechanism: sas.IdentityMechanismOAuthClient}},
		}},
	}
	impact := Classify(Diff(base, proposed), proposed)

	kinds := map[ImpactKind]bool{}
	for _, imp := range impact.Impacts {
		kinds[imp.Kind] = true
	}
	for _, want := range []ImpactKind{
		ImpactKindNewExternalDependency,
		ImpactKindNewBoundaryCrossing,
		ImpactKindDataClassificationChange,
		ImpactKindAuthnMechanismChanged,
	} {
		if !kinds[want] {
			t.Errorf("expected impact kind %s, got %+v", want, impact.Impacts)
		}
	}
}
