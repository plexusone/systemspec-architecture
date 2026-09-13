// Package threatmodel exports a sas.Architecture as the
// system-under-analysis for github.com/grokify/threat-model-spec: a DFD
// (data flow diagram) of components, boundaries, and flows, so a threat
// model never re-describes the system it analyzes. Threat reasoning —
// attacks, STRIDE threats, mitigations, detections, response actions —
// is intentionally not exported; that stays in Threat Model Spec, which
// references SAS IDs directly since Export preserves them unchanged.
//
// SAS does not take a Go module dependency on threat-model-spec. The
// types here are a narrow, local mirror of the fields
// threat-model-spec's schema/diagram.schema.json (DiagramIR) actually
// requires from a system-under-analysis, verified against that schema
// directly, so Export's output is valid (a subset of) a DiagramIR
// document without importing threat-model-spec's Go types.
package threatmodel

import (
	"strings"

	"github.com/plexusone/systemspec-architecture/sas"
)

// DiagramIR mirrors the system-under-analysis fields of
// threat-model-spec's DiagramIR: type, title, elements, boundaries, and
// flows. Threat-modeling fields (attacks, threats, mitigations,
// detections, responseActions, actors, phases, messages, attackTree) are
// out of scope for this bridge and therefore absent from this type.
type DiagramIR struct {
	// Type is always "dfd": SAS's static architecture model maps
	// naturally onto a data flow diagram, not an attack-chain, sequence,
	// or attack-tree diagram.
	Type string `json:"type"`

	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Elements    []Element  `json:"elements,omitempty"`
	Boundaries  []Boundary `json:"boundaries,omitempty"`
	Flows       []Flow     `json:"flows,omitempty"`
}

// Element mirrors threat-model-spec's Element: a component in the
// diagram (process, datastore, external-entity, gateway, browser, agent,
// or api).
type Element struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Type     string `json:"type"`
	ParentID string `json:"parentId,omitempty"`
}

// Boundary mirrors threat-model-spec's Boundary: a trust or network
// boundary (browser, localhost, network, cloud, breached, container,
// sandbox, agent, or origin).
type Boundary struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type"`
}

// Flow mirrors threat-model-spec's Flow: a directed edge between
// elements.
type Flow struct {
	From          string `json:"from"`
	To            string `json:"to"`
	Label         string `json:"label,omitempty"`
	Type          string `json:"type,omitempty"`
	Protocol      string `json:"protocol,omitempty"`
	Encrypted     *bool  `json:"encrypted,omitempty"`
	Authenticated *bool  `json:"authenticated,omitempty"`
}

// Export translates arch into a DiagramIR system-under-analysis. Every
// element, boundary, and flow ID is preserved unchanged from the SAS
// document, so a threat model built from this export can reference SAS
// IDs directly.
func Export(arch *sas.Architecture) DiagramIR {
	title := arch.Metadata.Name
	if title == "" {
		title = "SAS Architecture Export"
	}

	diagram := DiagramIR{
		Type:        "dfd",
		Title:       title,
		Description: arch.Metadata.Description,
	}

	for _, n := range arch.Nodes {
		diagram.Elements = append(diagram.Elements, Element{
			ID:       n.ID,
			Label:    n.Name,
			Type:     elementType(n.Kind),
			ParentID: primaryTrustBoundary(arch, n),
		})
	}

	for _, b := range arch.Boundaries {
		diagram.Boundaries = append(diagram.Boundaries, Boundary{
			ID:    b.ID,
			Label: b.Name,
			Type:  boundaryType(b.Kind),
		})
	}

	for _, r := range arch.Relationships {
		flow := Flow{
			From:  r.From,
			To:    r.To,
			Label: flowLabel(r),
			Type:  "normal",
		}
		if r.Transport != nil {
			flow.Protocol = r.Transport.ApplicationProtocol
			encrypted := r.Transport.Encryption != "" && r.Transport.Encryption != "none"
			flow.Encrypted = &encrypted
		}
		if r.Identity != nil {
			authenticated := true
			flow.Authenticated = &authenticated
		}
		diagram.Flows = append(diagram.Flows, flow)
	}

	return diagram
}

// elementType maps a NodeKind to threat-model-spec's ElementType enum
// (process, datastore, external-entity, gateway, browser, agent, api).
// NodeKindActor has no direct match in that enum — DFD conventions treat
// a human user as outside the system under analysis, which
// "external-entity" captures better than any other option.
func elementType(kind sas.NodeKind) string {
	switch kind {
	case sas.NodeKindDataDatabase:
		return "datastore"
	case sas.NodeKindExternalService, sas.NodeKindActor:
		return "external-entity"
	case sas.NodeKindGateway:
		return "gateway"
	case sas.NodeKindAgent:
		return "agent"
	case sas.NodeKindMCPServer:
		return "api"
	default:
		return "process"
	}
}

// boundaryType maps a BoundaryKind to threat-model-spec's BoundaryType
// enum (browser, localhost, network, cloud, breached, container,
// sandbox, agent, origin). That enum is narrower than SAS's boundary
// vocabulary and has no generic "trust zone" option, so trust,
// compliance, and organization boundaries fall back to "network" as the
// most general infrastructure-boundary type. This is a lossy,
// best-effort translation — SAS remains the source of truth for what the
// boundary actually represents; consult the SAS document, not this
// export, for the real boundary kind.
func boundaryType(kind sas.BoundaryKind) string {
	switch kind {
	case sas.BoundaryKindNetwork:
		return "network"
	case sas.BoundaryKindAccount, sas.BoundaryKindRegion, sas.BoundaryKindEnvironment:
		return "cloud"
	default:
		return "network"
	}
}

// primaryTrustBoundary picks the first Trust-kind boundary (in
// Architecture.Boundaries declaration order) node n belongs to, since
// DiagramIR nests an element under a single parent boundary while SAS
// allows many-to-many membership. This mirrors the same
// first-declared-wins pragma used by render.GroupNodes for visual
// grouping. Returns "" if n belongs to no trust boundary.
func primaryTrustBoundary(arch *sas.Architecture, n sas.Node) string {
	membership := make(map[string]bool, len(n.Boundaries))
	for _, id := range n.Boundaries {
		membership[id] = true
	}
	for _, b := range arch.Boundaries {
		if b.Kind == sas.BoundaryKindTrust && membership[b.ID] {
			return b.ID
		}
	}
	return ""
}

// flowLabel shows a relationship's operations when declared, falling
// back to its kind — the same choice render.EdgeLabel makes, duplicated
// here in three lines rather than importing the render package, since a
// threat-model export and a diagram renderer are unrelated consumers of
// the same underlying fact.
func flowLabel(r sas.Relationship) string {
	if len(r.Operations) > 0 {
		ops := make([]string, len(r.Operations))
		for i, op := range r.Operations {
			ops[i] = string(op)
		}
		return strings.Join(ops, ", ")
	}
	return string(r.Kind)
}
