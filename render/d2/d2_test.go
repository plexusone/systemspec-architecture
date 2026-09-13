package d2

import (
	"strings"
	"testing"

	"github.com/plexusone/systemspec-architecture/sas"
)

func testArchitecture() *sas.Architecture {
	return &sas.Architecture{
		Version: "0.1",
		Nodes: []sas.Node{
			{ID: "web", Kind: sas.NodeKindService, Name: "Web", Boundaries: []string{"vpc-prod"}},
			{ID: "db", Kind: sas.NodeKindDataDatabase, Name: "DB", Boundaries: []string{"vpc-prod"}},
			{ID: "user", Kind: sas.NodeKindActor, Name: "User"},
			{ID: "stripe", Kind: sas.NodeKindExternalService, Name: "Stripe"},
		},
		Relationships: []sas.Relationship{
			{ID: "user-to-web", From: "user", To: "web", Kind: sas.RelationKindCalls},
			{ID: "web-to-db", From: "web", To: "db", Kind: sas.RelationKindDataAccess, Operations: []sas.Operation{sas.OperationRead, sas.OperationCreate}},
			{ID: "web-to-stripe", From: "web", To: "stripe", Kind: sas.RelationKindCalls},
		},
		Boundaries: []sas.Boundary{
			{ID: "vpc-prod", Kind: sas.BoundaryKindNetwork, Name: "Production VPC"},
		},
	}
}

func TestRender_NodesAndEdges(t *testing.T) {
	out, err := Render(testArchitecture(), sas.View{ID: "all"})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	if !strings.Contains(out, `web: "Web"`) {
		t.Errorf("expected default-shape node for web, got:\n%s", out)
	}
	if !strings.Contains(out, `db: "DB" {`) || !strings.Contains(out, "shape: cylinder") {
		t.Errorf("expected cylinder-shaped db node, got:\n%s", out)
	}
	if !strings.Contains(out, "shape: person") {
		t.Errorf("expected person-shaped user node, got:\n%s", out)
	}
	if !strings.Contains(out, "shape: hexagon") {
		t.Errorf("expected hexagon-shaped stripe node, got:\n%s", out)
	}
	if !strings.Contains(out, "web -> db: read, create") {
		t.Errorf("expected operations edge label, got:\n%s", out)
	}
	if !strings.Contains(out, "user -> web: calls") {
		t.Errorf("expected kind-fallback edge label, got:\n%s", out)
	}
}

func TestRender_GroupByBoundary_QualifiesEdgePaths(t *testing.T) {
	out, err := Render(testArchitecture(), sas.View{ID: "grouped", GroupBy: string(sas.BoundaryKindNetwork)})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	if !strings.Contains(out, `vpc-prod: "Production VPC" {`) {
		t.Errorf("expected vpc-prod container, got:\n%s", out)
	}
	// web and db are inside vpc-prod, so the edge between them must use
	// qualified paths.
	if !strings.Contains(out, "vpc-prod.web -> vpc-prod.db: read, create") {
		t.Errorf("expected qualified edge path within the container, got:\n%s", out)
	}
	// user (ungrouped) -> web (grouped): the grouped endpoint must be
	// qualified, the ungrouped endpoint must not be.
	if !strings.Contains(out, "user -> vpc-prod.web: calls") {
		t.Errorf("expected mixed qualified/unqualified edge path, got:\n%s", out)
	}
}

func TestRender_ViewFiltersNodes(t *testing.T) {
	out, err := Render(testArchitecture(), sas.View{ID: "databases-only", IncludeKinds: []string{string(sas.NodeKindDataDatabase)}})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(out, `db: "DB"`) {
		t.Errorf("expected db node, got:\n%s", out)
	}
	if strings.Contains(out, `web: "Web"`) {
		t.Errorf("did not expect web node in a databases-only view, got:\n%s", out)
	}
}

func TestRender_EscapesQuotesInNames(t *testing.T) {
	arch := &sas.Architecture{
		Nodes: []sas.Node{{ID: "a", Kind: sas.NodeKindService, Name: `The "Big" App`}},
	}
	out, err := Render(arch, sas.View{ID: "all"})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(out, `\"Big\"`) {
		t.Errorf("expected escaped quotes, got:\n%s", out)
	}
}

func TestRender_Deterministic(t *testing.T) {
	arch := testArchitecture()
	view := sas.View{ID: "all", GroupBy: string(sas.BoundaryKindNetwork)}

	first, err := Render(arch, view)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	second, err := Render(arch, view)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if first != second {
		t.Fatalf("expected deterministic output, got:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}
