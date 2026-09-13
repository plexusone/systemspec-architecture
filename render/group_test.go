package render

import (
	"testing"

	"github.com/plexusone/systemspec-architecture/sas"
)

func testGroupArchitecture() sas.Architecture {
	return sas.Architecture{
		Nodes: []sas.Node{
			{ID: "web", Kind: sas.NodeKindService, Boundaries: []string{"vpc-a"}},
			{ID: "api", Kind: sas.NodeKindService, Boundaries: []string{"vpc-a", "trust-app"}},
			{ID: "db", Kind: sas.NodeKindDataDatabase, Boundaries: []string{"vpc-b"}},
			{ID: "external", Kind: sas.NodeKindExternalService},
		},
		Boundaries: []sas.Boundary{
			{ID: "vpc-a", Kind: sas.BoundaryKindNetwork, Name: "VPC A"},
			{ID: "vpc-b", Kind: sas.BoundaryKindNetwork, Name: "VPC B"},
			{ID: "trust-app", Kind: sas.BoundaryKindTrust, Name: "App Trust Zone"},
		},
	}
}

func TestGroupNodes_NoGroupBy(t *testing.T) {
	arch := testGroupArchitecture()
	groups := GroupNodes(arch, "")

	if len(groups) != 1 || groups[0].Boundary != nil {
		t.Fatalf("expected one ungrouped bucket, got %+v", groups)
	}
	if len(groups[0].Nodes) != 4 {
		t.Fatalf("expected all 4 nodes in the single bucket, got %d", len(groups[0].Nodes))
	}
}

func TestGroupNodes_ByNetworkBoundary(t *testing.T) {
	arch := testGroupArchitecture()
	groups := GroupNodes(arch, string(sas.BoundaryKindNetwork))

	if len(groups) != 3 {
		t.Fatalf("expected 3 groups (vpc-a, vpc-b, ungrouped), got %d: %+v", len(groups), groups)
	}

	if groups[0].Boundary == nil || groups[0].Boundary.ID != "vpc-a" {
		t.Fatalf("expected first group to be vpc-a, got %+v", groups[0])
	}
	if len(groups[0].Nodes) != 2 || groups[0].Nodes[0].ID != "web" || groups[0].Nodes[1].ID != "api" {
		t.Fatalf("expected vpc-a group to contain web, api, got %+v", groups[0].Nodes)
	}

	if groups[1].Boundary == nil || groups[1].Boundary.ID != "vpc-b" {
		t.Fatalf("expected second group to be vpc-b, got %+v", groups[1])
	}
	if len(groups[1].Nodes) != 1 || groups[1].Nodes[0].ID != "db" {
		t.Fatalf("expected vpc-b group to contain db, got %+v", groups[1].Nodes)
	}

	last := groups[len(groups)-1]
	if last.Boundary != nil {
		t.Fatalf("expected trailing ungrouped bucket, got %+v", last)
	}
	if len(last.Nodes) != 1 || last.Nodes[0].ID != "external" {
		t.Fatalf("expected ungrouped bucket to contain external, got %+v", last.Nodes)
	}
}

func TestGroupNodes_FirstMatchingBoundaryWins(t *testing.T) {
	// "api" belongs to both vpc-a (network) and trust-app (trust). When
	// grouping by network, api must land in vpc-a — trust-app should be
	// irrelevant since it isn't a network boundary.
	arch := testGroupArchitecture()
	groups := GroupNodes(arch, string(sas.BoundaryKindNetwork))

	found := false
	for _, g := range groups {
		if g.Boundary == nil || g.Boundary.ID != "vpc-a" {
			continue
		}
		for _, n := range g.Nodes {
			if n.ID == "api" {
				found = true
			}
		}
	}
	if !found {
		t.Fatal("expected api to be grouped under vpc-a")
	}
}

func TestGroupNodes_UnusedBoundaryKindOmitted(t *testing.T) {
	arch := testGroupArchitecture()
	// No node belongs to a "compliance" boundary at all — there isn't
	// even one declared — so grouping by it should yield a single
	// ungrouped bucket, not an empty compliance group.
	groups := GroupNodes(arch, string(sas.BoundaryKindCompliance))

	if len(groups) != 1 || groups[0].Boundary != nil {
		t.Fatalf("expected single ungrouped bucket, got %+v", groups)
	}
	if len(groups[0].Nodes) != 4 {
		t.Fatalf("expected all 4 nodes ungrouped, got %d", len(groups[0].Nodes))
	}
}
