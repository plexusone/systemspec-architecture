package mermaid

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

	if !strings.HasPrefix(out, "flowchart TD\n") {
		t.Fatalf("expected flowchart TD header, got %q", out)
	}
	if !strings.Contains(out, `web["Web"]`) {
		t.Errorf("expected rectangle node for web, got:\n%s", out)
	}
	if !strings.Contains(out, `db[("DB")]`) {
		t.Errorf("expected cylinder node for db, got:\n%s", out)
	}
	if !strings.Contains(out, `user("User")`) {
		t.Errorf("expected stadium node for user, got:\n%s", out)
	}
	if !strings.Contains(out, `stripe{{"Stripe"}}`) {
		t.Errorf("expected hexagon node for stripe, got:\n%s", out)
	}
	if !strings.Contains(out, "web -->|read, create| db") {
		t.Errorf("expected operations edge label, got:\n%s", out)
	}
	if !strings.Contains(out, "user -->|calls| web") {
		t.Errorf("expected kind-fallback edge label, got:\n%s", out)
	}
}

func TestRender_GroupByBoundary(t *testing.T) {
	out, err := Render(testArchitecture(), sas.View{ID: "grouped", GroupBy: string(sas.BoundaryKindNetwork)})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	if !strings.Contains(out, `subgraph vpc-prod["Production VPC"]`) {
		t.Errorf("expected subgraph for vpc-prod, got:\n%s", out)
	}
	if !strings.Contains(out, "  end\n") {
		t.Errorf("expected subgraph end, got:\n%s", out)
	}

	// web and db must appear inside the subgraph block (indented deeper
	// than nodes outside it); user and stripe belong to no network
	// boundary and must appear outside it.
	subgraphStart := strings.Index(out, "subgraph vpc-prod")
	subgraphEnd := strings.Index(out, "  end\n")
	inside := out[subgraphStart:subgraphEnd]
	if !strings.Contains(inside, "web[") || !strings.Contains(inside, "db[") {
		t.Errorf("expected web and db inside subgraph, got:\n%s", inside)
	}
	if strings.Contains(inside, "user(") || strings.Contains(inside, "stripe{{") {
		t.Errorf("did not expect user or stripe inside subgraph, got:\n%s", inside)
	}
}

func TestRender_ViewFiltersNodes(t *testing.T) {
	out, err := Render(testArchitecture(), sas.View{ID: "databases-only", IncludeKinds: []string{string(sas.NodeKindDataDatabase)}})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	if !strings.Contains(out, `db[("DB")]`) {
		t.Errorf("expected db node, got:\n%s", out)
	}
	if strings.Contains(out, `web[`) {
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
	if strings.Contains(out, `"Big"`) {
		t.Errorf("expected embedded quotes to be escaped, got:\n%s", out)
	}
	if !strings.Contains(out, "#quot;Big#quot;") {
		t.Errorf("expected mermaid #quot; escape, got:\n%s", out)
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
