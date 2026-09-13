package dot

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

func TestRender_HeaderAndFooter(t *testing.T) {
	out, err := Render(testArchitecture(), sas.View{ID: "all"})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.HasPrefix(out, "digraph architecture {\n") {
		t.Errorf("expected digraph header, got:\n%s", out)
	}
	if !strings.HasSuffix(out, "}\n") {
		t.Errorf("expected closing brace, got:\n%s", out)
	}
}

func TestRender_NodesAndEdges(t *testing.T) {
	out, err := Render(testArchitecture(), sas.View{ID: "all"})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	if !strings.Contains(out, `"web" [label="Web", shape=box];`) {
		t.Errorf("expected box-shaped web node, got:\n%s", out)
	}
	if !strings.Contains(out, `"db" [label="DB", shape=cylinder];`) {
		t.Errorf("expected cylinder-shaped db node, got:\n%s", out)
	}
	if !strings.Contains(out, `"user" [label="User", shape=ellipse];`) {
		t.Errorf("expected ellipse-shaped user node, got:\n%s", out)
	}
	if !strings.Contains(out, `"stripe" [label="Stripe", shape=hexagon];`) {
		t.Errorf("expected hexagon-shaped stripe node, got:\n%s", out)
	}
	if !strings.Contains(out, `"web" -> "db" [label="read, create"];`) {
		t.Errorf("expected operations edge label, got:\n%s", out)
	}
	if !strings.Contains(out, `"user" -> "web" [label="calls"];`) {
		t.Errorf("expected kind-fallback edge label, got:\n%s", out)
	}
}

func TestRender_GroupByBoundary(t *testing.T) {
	out, err := Render(testArchitecture(), sas.View{ID: "grouped", GroupBy: string(sas.BoundaryKindNetwork)})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	if !strings.Contains(out, `subgraph "cluster_vpc-prod" {`) {
		t.Errorf("expected cluster_-prefixed subgraph, got:\n%s", out)
	}
	if !strings.Contains(out, `label="Production VPC";`) {
		t.Errorf("expected cluster label, got:\n%s", out)
	}

	start := strings.Index(out, `subgraph "cluster_vpc-prod"`)
	end := strings.Index(out[start:], "  }\n") + start
	inside := out[start:end]
	if !strings.Contains(inside, `"web"`) || !strings.Contains(inside, `"db"`) {
		t.Errorf("expected web and db inside the cluster, got:\n%s", inside)
	}
	if strings.Contains(inside, `"user"`) || strings.Contains(inside, `"stripe"`) {
		t.Errorf("did not expect user or stripe inside the cluster, got:\n%s", inside)
	}
}

func TestRender_ViewFiltersNodes(t *testing.T) {
	out, err := Render(testArchitecture(), sas.View{ID: "databases-only", IncludeKinds: []string{string(sas.NodeKindDataDatabase)}})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(out, `"db" [label="DB"`) {
		t.Errorf("expected db node, got:\n%s", out)
	}
	if strings.Contains(out, `"web" [label="Web"`) {
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
