package threatmodel

import (
	"testing"

	"github.com/plexusone/systemspec-architecture/sas"
)

// This suite asserts against threat-model-spec's real, documented
// enums (github.com/grokify/threat-model-spec, schema/diagram.schema.json:
// ElementType, BoundaryType, FlowType) rather than depending on that
// sibling repository's filesystem path, which would make this test
// non-portable. Export's output was interactively verified to validate
// against the real diagram.schema.json for both dogfood architectures
// before this test suite was written; see the RMI-026 commit message.

func testArchitecture() *sas.Architecture {
	return &sas.Architecture{
		Metadata: sas.Metadata{Name: "Example System", Description: "An example."},
		Nodes: []sas.Node{
			{ID: "api", Kind: sas.NodeKindService, Name: "API", Boundaries: []string{"trust-app"}},
			{ID: "db", Kind: sas.NodeKindDataDatabase, Name: "DB", Boundaries: []string{"trust-app"}},
			{ID: "gw", Kind: sas.NodeKindGateway, Name: "Gateway", Boundaries: []string{"trust-app"}},
			{ID: "agent", Kind: sas.NodeKindAgent, Name: "Agent", Boundaries: []string{"trust-app"}},
			{ID: "mcp", Kind: sas.NodeKindMCPServer, Name: "MCP Server", Boundaries: []string{"trust-app"}},
			{ID: "actor", Kind: sas.NodeKindActor, Name: "User"},
			{ID: "external", Kind: sas.NodeKindExternalService, Name: "External"},
		},
		Relationships: []sas.Relationship{
			{
				ID: "actor-to-api", From: "actor", To: "api", Kind: sas.RelationKindCalls,
				Transport: &sas.Transport{ApplicationProtocol: "https", Encryption: "tls"},
				Identity:  &sas.IdentityRef{NodeID: "actor"},
			},
			{
				ID: "api-to-db", From: "api", To: "db", Kind: sas.RelationKindDataAccess,
				Operations: []sas.Operation{sas.OperationRead, sas.OperationCreate},
			},
			{
				ID: "api-to-external", From: "api", To: "external", Kind: sas.RelationKindCalls,
				Transport: &sas.Transport{ApplicationProtocol: "https", Encryption: "none"},
			},
		},
		Boundaries: []sas.Boundary{
			{ID: "trust-app", Kind: sas.BoundaryKindTrust, Name: "App Trust Zone"},
		},
	}
}

func TestExport_TypeAndTitle(t *testing.T) {
	diagram := Export(testArchitecture())
	if diagram.Type != "dfd" {
		t.Errorf("expected type dfd, got %q", diagram.Type)
	}
	if diagram.Title != "Example System" {
		t.Errorf("expected title Example System, got %q", diagram.Title)
	}
	if diagram.Description != "An example." {
		t.Errorf("expected description to carry through, got %q", diagram.Description)
	}
}

func TestExport_TitleFallback(t *testing.T) {
	diagram := Export(&sas.Architecture{})
	if diagram.Title == "" {
		t.Error("expected a non-empty fallback title when Metadata.Name is unset")
	}
}

func TestExport_ElementTypeMapping(t *testing.T) {
	diagram := Export(testArchitecture())

	want := map[string]string{
		"api":      "process",
		"db":       "datastore",
		"gw":       "gateway",
		"agent":    "agent",
		"mcp":      "api",
		"actor":    "external-entity",
		"external": "external-entity",
	}

	byID := make(map[string]Element, len(diagram.Elements))
	for _, e := range diagram.Elements {
		byID[e.ID] = e
	}

	for id, wantType := range want {
		e, ok := byID[id]
		if !ok {
			t.Errorf("expected element %q in export", id)
			continue
		}
		if e.Type != wantType {
			t.Errorf("element %q: got type %q, want %q", id, e.Type, wantType)
		}
	}
}

func TestExport_ElementParentID(t *testing.T) {
	diagram := Export(testArchitecture())
	byID := make(map[string]Element, len(diagram.Elements))
	for _, e := range diagram.Elements {
		byID[e.ID] = e
	}

	if byID["api"].ParentID != "trust-app" {
		t.Errorf("expected api parentId trust-app, got %q", byID["api"].ParentID)
	}
	if byID["actor"].ParentID != "" {
		t.Errorf("expected actor (no boundary membership) to have empty parentId, got %q", byID["actor"].ParentID)
	}
}

func TestExport_BoundaryTypeMapping(t *testing.T) {
	diagram := Export(testArchitecture())
	if len(diagram.Boundaries) != 1 {
		t.Fatalf("expected 1 boundary, got %d", len(diagram.Boundaries))
	}
	b := diagram.Boundaries[0]
	if b.ID != "trust-app" || b.Label != "App Trust Zone" {
		t.Errorf("unexpected boundary %+v", b)
	}
	// BoundaryKindTrust has no direct match in threat-model-spec's
	// BoundaryType enum; it falls back to "network".
	if b.Type != "network" {
		t.Errorf("expected trust boundary to fall back to type network, got %q", b.Type)
	}
}

func TestExport_BoundaryTypeSandboxContainerBreached(t *testing.T) {
	arch := &sas.Architecture{
		Metadata: sas.Metadata{Name: "Isolation System"},
		Nodes: []sas.Node{
			{ID: "sbx-code", Kind: sas.NodeKindSandbox, Name: "Code Sandbox", Boundaries: []string{"b-sandbox"}},
			{ID: "svc", Kind: sas.NodeKindContainer, Name: "Worker", Boundaries: []string{"b-container"}},
		},
		Boundaries: []sas.Boundary{
			{ID: "b-sandbox", Kind: sas.BoundaryKindSandbox, Name: "CaaS Sandbox"},
			{ID: "b-container", Kind: sas.BoundaryKindContainer, Name: "Pod"},
			// state=breached overrides the kind mapping (kind is network,
			// but the zone is compromised).
			{ID: "b-breached", Kind: sas.BoundaryKindNetwork, Name: "Compromised VPC",
				Attributes: map[string]string{"state": "breached"}},
		},
	}

	diagram := Export(arch)
	byID := make(map[string]Boundary, len(diagram.Boundaries))
	for _, b := range diagram.Boundaries {
		byID[b.ID] = b
	}

	want := map[string]string{
		"b-sandbox":   "sandbox",
		"b-container": "container",
		"b-breached":  "breached",
	}
	for id, wantType := range want {
		b, ok := byID[id]
		if !ok {
			t.Errorf("expected boundary %q in export", id)
			continue
		}
		if b.Type != wantType {
			t.Errorf("boundary %q: got type %q, want %q", id, b.Type, wantType)
		}
	}
}

func TestExport_FlowsCarryProtocolEncryptionAuth(t *testing.T) {
	diagram := Export(testArchitecture())
	byRoute := make(map[string]Flow, len(diagram.Flows))
	for _, f := range diagram.Flows {
		byRoute[f.From+"->"+f.To] = f
	}

	actorToAPI := byRoute["actor->api"]
	if actorToAPI.Protocol != "https" {
		t.Errorf("expected protocol https, got %q", actorToAPI.Protocol)
	}
	if actorToAPI.Encrypted == nil || !*actorToAPI.Encrypted {
		t.Errorf("expected encrypted true, got %v", actorToAPI.Encrypted)
	}
	if actorToAPI.Authenticated == nil || !*actorToAPI.Authenticated {
		t.Errorf("expected authenticated true, got %v", actorToAPI.Authenticated)
	}

	apiToExternal := byRoute["api->external"]
	if apiToExternal.Encrypted == nil || *apiToExternal.Encrypted {
		t.Errorf("expected encrypted false for encryption=none, got %v", apiToExternal.Encrypted)
	}
	if apiToExternal.Authenticated != nil {
		t.Errorf("expected authenticated to be unset (no Identity declared), got %v", *apiToExternal.Authenticated)
	}

	apiToDB := byRoute["api->db"]
	if apiToDB.Label != "read, create" {
		t.Errorf("expected operations-joined label, got %q", apiToDB.Label)
	}
	if apiToDB.Protocol != "" {
		t.Errorf("expected no protocol when Transport is nil, got %q", apiToDB.Protocol)
	}
	if apiToDB.Type != "normal" {
		t.Errorf("expected flow type normal, got %q", apiToDB.Type)
	}
}

func TestExport_FlowFallsBackToKindLabel(t *testing.T) {
	diagram := Export(testArchitecture())
	for _, f := range diagram.Flows {
		if f.From == "actor" && f.To == "api" {
			if f.Label != "calls" {
				t.Errorf("expected fallback label 'calls' (no operations declared), got %q", f.Label)
			}
		}
	}
}

func TestExport_IDsPreservedUnchanged(t *testing.T) {
	arch := testArchitecture()
	diagram := Export(arch)

	nodeIDs := make(map[string]bool, len(arch.Nodes))
	for _, n := range arch.Nodes {
		nodeIDs[n.ID] = true
	}
	for _, e := range diagram.Elements {
		if !nodeIDs[e.ID] {
			t.Errorf("exported element ID %q does not match any SAS node ID", e.ID)
		}
	}
}
