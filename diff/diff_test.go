package diff

import (
	"testing"

	"github.com/plexusone/systemspec-architecture/sas"
)

func TestDiff_NoChanges(t *testing.T) {
	arch := &sas.Architecture{
		Nodes: []sas.Node{{ID: "api", Kind: sas.NodeKindService, Name: "API"}},
	}
	cs := Diff(arch, arch)
	if len(cs.Changes) != 0 {
		t.Fatalf("expected no changes, got %+v", cs.Changes)
	}
}

func TestDiff_NodeAddedAndRemoved(t *testing.T) {
	base := &sas.Architecture{
		Nodes: []sas.Node{{ID: "old", Kind: sas.NodeKindService, Name: "Old"}},
	}
	proposed := &sas.Architecture{
		Nodes: []sas.Node{{ID: "new", Kind: sas.NodeKindService, Name: "New"}},
	}
	cs := Diff(base, proposed)
	if len(cs.Changes) != 2 {
		t.Fatalf("expected 2 changes, got %d: %+v", len(cs.Changes), cs.Changes)
	}
	if cs.Changes[0].Kind != ChangeKindNodeAdded || cs.Changes[0].ElementID != "new" {
		t.Errorf("expected node_added for 'new' first (sorted), got %+v", cs.Changes[0])
	}
	if cs.Changes[1].Kind != ChangeKindNodeRemoved || cs.Changes[1].ElementID != "old" {
		t.Errorf("expected node_removed for 'old' second (sorted), got %+v", cs.Changes[1])
	}
}

func TestDiff_NodeModified(t *testing.T) {
	base := &sas.Architecture{
		Nodes: []sas.Node{{
			ID: "api", Kind: sas.NodeKindService, Name: "API", Owner: "team-a",
			Technology: &sas.Technology{Provider: "aws", Service: "ecs"},
		}},
	}
	proposed := &sas.Architecture{
		Nodes: []sas.Node{{
			ID: "api", Kind: sas.NodeKindService, Name: "API", Owner: "team-b",
			Technology: &sas.Technology{Provider: "aws", Service: "lambda"},
		}},
	}
	cs := Diff(base, proposed)
	if len(cs.Changes) != 1 {
		t.Fatalf("expected 1 change, got %d: %+v", len(cs.Changes), cs.Changes)
	}
	change := cs.Changes[0]
	if change.Kind != ChangeKindNodeModified {
		t.Fatalf("expected node_modified, got %s", change.Kind)
	}
	want := map[string][2]string{
		"owner":      {"team-a", "team-b"},
		"technology": {"aws/ecs", "aws/lambda"},
	}
	if len(change.Deltas) != len(want) {
		t.Fatalf("expected %d deltas, got %d: %+v", len(want), len(change.Deltas), change.Deltas)
	}
	for _, d := range change.Deltas {
		wantVals, ok := want[d.Field]
		if !ok {
			t.Errorf("unexpected field delta: %+v", d)
			continue
		}
		if d.Before != wantVals[0] || d.After != wantVals[1] {
			t.Errorf("field %s: got before=%q after=%q, want before=%q after=%q", d.Field, d.Before, d.After, wantVals[0], wantVals[1])
		}
	}
}

func TestDiff_BoundaryMembershipChanged(t *testing.T) {
	base := &sas.Architecture{
		Nodes: []sas.Node{{ID: "db", Kind: sas.NodeKindDataDatabase, Name: "DB", Boundaries: []string{"vpc-internal"}}},
	}
	proposed := &sas.Architecture{
		Nodes: []sas.Node{{ID: "db", Kind: sas.NodeKindDataDatabase, Name: "DB", Boundaries: []string{"vpc-internal", "trust-zone-external"}}},
	}
	cs := Diff(base, proposed)
	if len(cs.Changes) != 1 {
		t.Fatalf("expected 1 change, got %d: %+v", len(cs.Changes), cs.Changes)
	}
	change := cs.Changes[0]
	if change.Kind != ChangeKindBoundaryMembershipChanged {
		t.Fatalf("expected boundary_membership_changed, got %s", change.Kind)
	}
	if len(change.Deltas) != 1 || change.Deltas[0].After != "trust-zone-external" {
		t.Fatalf("expected one added-boundary delta for trust-zone-external, got %+v", change.Deltas)
	}
}

func TestDiff_NodeModifiedAndBoundaryMembershipChanged_AreDistinctChanges(t *testing.T) {
	base := &sas.Architecture{
		Nodes: []sas.Node{{ID: "db", Kind: sas.NodeKindDataDatabase, Name: "DB", Owner: "team-a", Boundaries: []string{"vpc-internal"}}},
	}
	proposed := &sas.Architecture{
		Nodes: []sas.Node{{ID: "db", Kind: sas.NodeKindDataDatabase, Name: "DB", Owner: "team-b", Boundaries: []string{"vpc-internal", "trust-zone-external"}}},
	}
	cs := Diff(base, proposed)
	if len(cs.Changes) != 2 {
		t.Fatalf("expected 2 changes (node_modified + boundary_membership_changed), got %d: %+v", len(cs.Changes), cs.Changes)
	}
	if cs.Changes[0].Kind != ChangeKindNodeModified {
		t.Errorf("expected node_modified first, got %s", cs.Changes[0].Kind)
	}
	if cs.Changes[1].Kind != ChangeKindBoundaryMembershipChanged {
		t.Errorf("expected boundary_membership_changed second, got %s", cs.Changes[1].Kind)
	}
}

func TestDiff_RelationshipModified_AuthenticationMechanismChange(t *testing.T) {
	base := &sas.Architecture{
		Relationships: []sas.Relationship{{
			ID: "web-to-api", From: "web", To: "api", Kind: sas.RelationKindCalls,
			Transport: &sas.Transport{Encryption: "mtls"},
		}},
	}
	proposed := &sas.Architecture{
		Relationships: []sas.Relationship{{
			ID: "web-to-api", From: "web", To: "api", Kind: sas.RelationKindCalls,
			Transport: &sas.Transport{Encryption: "tls"},
		}},
	}
	cs := Diff(base, proposed)
	if len(cs.Changes) != 1 || cs.Changes[0].Kind != ChangeKindRelationshipModified {
		t.Fatalf("expected 1 relationship_modified change, got %+v", cs.Changes)
	}
	deltas := cs.Changes[0].Deltas
	if len(deltas) != 1 || deltas[0].Field != "transport.encryption" || deltas[0].Before != "mtls" || deltas[0].After != "tls" {
		t.Fatalf("expected single transport.encryption delta mtls->tls, got %+v", deltas)
	}
}

func TestDiff_RelationshipModified_OperationExpansion(t *testing.T) {
	base := &sas.Architecture{
		Relationships: []sas.Relationship{{
			ID: "api-to-db", From: "api", To: "db", Kind: sas.RelationKindDataAccess,
			Operations: []sas.Operation{sas.OperationRead},
		}},
	}
	proposed := &sas.Architecture{
		Relationships: []sas.Relationship{{
			ID: "api-to-db", From: "api", To: "db", Kind: sas.RelationKindDataAccess,
			Operations: []sas.Operation{sas.OperationRead, sas.OperationCreate},
		}},
	}
	cs := Diff(base, proposed)
	if len(cs.Changes) != 1 {
		t.Fatalf("expected 1 change, got %+v", cs.Changes)
	}
	deltas := cs.Changes[0].Deltas
	if len(deltas) != 1 || deltas[0].Field != "operations" || deltas[0].Before != "read" || deltas[0].After != "create,read" {
		t.Fatalf("expected operations delta read->create,read, got %+v", deltas)
	}
}

func TestDiff_BoundaryAddedRemovedModified(t *testing.T) {
	base := &sas.Architecture{
		Boundaries: []sas.Boundary{
			{ID: "old-b", Kind: sas.BoundaryKindTrust, Name: "Old"},
			{ID: "shared", Kind: sas.BoundaryKindNetwork, Name: "Shared"},
		},
	}
	proposed := &sas.Architecture{
		Boundaries: []sas.Boundary{
			{ID: "new-b", Kind: sas.BoundaryKindTrust, Name: "New"},
			{ID: "shared", Kind: sas.BoundaryKindNetwork, Name: "Shared Renamed"},
		},
	}
	cs := Diff(base, proposed)
	if len(cs.Changes) != 3 {
		t.Fatalf("expected 3 changes, got %d: %+v", len(cs.Changes), cs.Changes)
	}
	// Sorted by ID: new-b (added), old-b (removed), shared (modified).
	if cs.Changes[0].Kind != ChangeKindBoundaryAdded || cs.Changes[0].ElementID != "new-b" {
		t.Errorf("expected boundary_added for new-b, got %+v", cs.Changes[0])
	}
	if cs.Changes[1].Kind != ChangeKindBoundaryRemoved || cs.Changes[1].ElementID != "old-b" {
		t.Errorf("expected boundary_removed for old-b, got %+v", cs.Changes[1])
	}
	if cs.Changes[2].Kind != ChangeKindBoundaryModified || cs.Changes[2].ElementID != "shared" {
		t.Errorf("expected boundary_modified for shared, got %+v", cs.Changes[2])
	}
}

func TestDiff_NewExternalBoundaryCrossing_DataFlowClassificationAdded(t *testing.T) {
	// The PRD's expressiveness test: "this PR adds an external
	// trust-boundary crossing carrying customer data through a new
	// authorization mechanism" must be visible as concrete field deltas.
	base := &sas.Architecture{
		Relationships: []sas.Relationship{{
			ID: "api-to-partner", From: "api", To: "partner", Kind: sas.RelationKindCalls,
		}},
	}
	proposed := &sas.Architecture{
		Relationships: []sas.Relationship{{
			ID: "api-to-partner", From: "api", To: "partner", Kind: sas.RelationKindCalls,
			CrossesBoundaries: []string{"trust-external"},
			Data:              &sas.DataFlow{Classifications: []string{"customer_data"}},
			Identity:          &sas.IdentityRef{Identity: &sas.Identity{Type: sas.IdentityTypeWorkload, Mechanism: sas.IdentityMechanismOAuthClient}},
		}},
	}
	cs := Diff(base, proposed)
	if len(cs.Changes) != 1 {
		t.Fatalf("expected 1 change, got %+v", cs.Changes)
	}
	byField := map[string]FieldDelta{}
	for _, d := range cs.Changes[0].Deltas {
		byField[d.Field] = d
	}
	if d, ok := byField["crossesBoundaries"]; !ok || d.After != "trust-external" {
		t.Errorf("expected crossesBoundaries delta, got %+v", byField["crossesBoundaries"])
	}
	if d, ok := byField["data.classifications"]; !ok || d.After != "customer_data" {
		t.Errorf("expected data.classifications delta, got %+v", byField["data.classifications"])
	}
	if d, ok := byField["identity"]; !ok || d.After != "workload:oauth_client" {
		t.Errorf("expected identity delta, got %+v", byField["identity"])
	}
}
