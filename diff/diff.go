package diff

import (
	"fmt"
	"sort"
	"strings"

	"github.com/plexusone/systemspec-architecture/sas"
)

// Diff computes the ChangeSet from base to proposed: what an author would
// need to review to understand how the architecture changed.
func Diff(base, proposed *sas.Architecture) ChangeSet {
	var changes []Change
	changes = append(changes, diffNodes(base.Nodes, proposed.Nodes)...)
	changes = append(changes, diffRelationships(base.Relationships, proposed.Relationships)...)
	changes = append(changes, diffBoundaries(base.Boundaries, proposed.Boundaries)...)
	return ChangeSet{Changes: changes}
}

func diffNodes(base, proposed []sas.Node) []Change {
	nodeID := func(n sas.Node) string { return n.ID }
	baseByID := indexByID(base, nodeID)
	proposedByID := indexByID(proposed, nodeID)

	var changes []Change
	for _, id := range unionKeys(baseByID, proposedByID) {
		b, inBase := baseByID[id]
		p, inProposed := proposedByID[id]
		switch {
		case !inBase:
			changes = append(changes, Change{Kind: ChangeKindNodeAdded, ElementType: ElementTypeNode, ElementID: id})
		case !inProposed:
			changes = append(changes, Change{Kind: ChangeKindNodeRemoved, ElementType: ElementTypeNode, ElementID: id})
		default:
			if deltas := diffNodeFields(b, p); len(deltas) > 0 {
				changes = append(changes, Change{Kind: ChangeKindNodeModified, ElementType: ElementTypeNode, ElementID: id, Deltas: deltas})
			}
			if deltas := diffBoundaryMembership(b.Boundaries, p.Boundaries); len(deltas) > 0 {
				changes = append(changes, Change{Kind: ChangeKindBoundaryMembershipChanged, ElementType: ElementTypeNode, ElementID: id, Deltas: deltas})
			}
		}
	}
	return changes
}

func diffNodeFields(b, p sas.Node) []FieldDelta {
	var deltas []FieldDelta
	addFieldDelta(&deltas, "kind", string(b.Kind), string(p.Kind))
	addFieldDelta(&deltas, "name", b.Name, p.Name)
	addFieldDelta(&deltas, "owner", b.Owner, p.Owner)
	addFieldDelta(&deltas, "technology", technologyString(b.Technology), technologyString(p.Technology))
	addFieldDelta(&deltas, "identity", identityString(b.Identity), identityString(p.Identity))
	addFieldDelta(&deltas, "criticality", string(b.Criticality), string(p.Criticality))
	return deltas
}

func diffBoundaryMembership(base, proposed []string) []FieldDelta {
	baseSet := toSet(base)
	proposedSet := toSet(proposed)

	var deltas []FieldDelta
	for _, id := range sortedKeys(proposedSet) {
		if !baseSet[id] {
			deltas = append(deltas, FieldDelta{Field: "boundaries", After: id})
		}
	}
	for _, id := range sortedKeys(baseSet) {
		if !proposedSet[id] {
			deltas = append(deltas, FieldDelta{Field: "boundaries", Before: id})
		}
	}
	return deltas
}

func diffRelationships(base, proposed []sas.Relationship) []Change {
	return diffElements(base, proposed, func(r sas.Relationship) string { return r.ID },
		ElementTypeRelationship, ChangeKindRelationshipAdded, ChangeKindRelationshipRemoved, ChangeKindRelationshipModified,
		diffRelationshipFields)
}

func diffRelationshipFields(b, p sas.Relationship) []FieldDelta {
	var deltas []FieldDelta
	addFieldDelta(&deltas, "kind", string(b.Kind), string(p.Kind))
	addFieldDelta(&deltas, "from", b.From, p.From)
	addFieldDelta(&deltas, "to", b.To, p.To)
	addFieldDelta(&deltas, "transport.protocol", transportField(b, protocol), transportField(p, protocol))
	addFieldDelta(&deltas, "transport.port", transportField(b, port), transportField(p, port))
	addFieldDelta(&deltas, "transport.applicationProtocol", transportField(b, applicationProtocol), transportField(p, applicationProtocol))
	addFieldDelta(&deltas, "transport.encryption", transportField(b, encryption), transportField(p, encryption))
	addFieldDelta(&deltas, "operations", operationSet(b.Operations), operationSet(p.Operations))
	addFieldDelta(&deltas, "identity", identityRefString(b.Identity), identityRefString(p.Identity))
	addFieldDelta(&deltas, "authorization.entitlements", entitlementSet(b.Authorization), entitlementSet(p.Authorization))
	addFieldDelta(&deltas, "data.classifications", dataClassificationSet(b.Data), dataClassificationSet(p.Data))
	addFieldDelta(&deltas, "crossesBoundaries", stringSet(b.CrossesBoundaries), stringSet(p.CrossesBoundaries))
	addFieldDelta(&deltas, "criticalPath", fmt.Sprintf("%t", b.CriticalPath), fmt.Sprintf("%t", p.CriticalPath))
	addFieldDelta(&deltas, "sync", string(b.Sync), string(p.Sync))
	addFieldDelta(&deltas, "protocolRef", b.ProtocolRef, p.ProtocolRef)
	return deltas
}

func diffBoundaries(base, proposed []sas.Boundary) []Change {
	return diffElements(base, proposed, func(b sas.Boundary) string { return b.ID },
		ElementTypeBoundary, ChangeKindBoundaryAdded, ChangeKindBoundaryRemoved, ChangeKindBoundaryModified,
		diffBoundaryFields)
}

// diffElements is the shared added/removed/modified shape used by
// diffRelationships and diffBoundaries. diffNodes does not use it: a node
// can produce two distinct Change entries (a field modification and a
// boundary-membership change), which this single-change-per-element
// helper doesn't model.
func diffElements[T any](
	base, proposed []T,
	id func(T) string,
	elementType ElementType,
	addedKind, removedKind, modifiedKind ChangeKind,
	fields func(base, proposed T) []FieldDelta,
) []Change {
	baseByID := indexByID(base, id)
	proposedByID := indexByID(proposed, id)

	var changes []Change
	for _, eid := range unionKeys(baseByID, proposedByID) {
		b, inBase := baseByID[eid]
		p, inProposed := proposedByID[eid]
		switch {
		case !inBase:
			changes = append(changes, Change{Kind: addedKind, ElementType: elementType, ElementID: eid})
		case !inProposed:
			changes = append(changes, Change{Kind: removedKind, ElementType: elementType, ElementID: eid})
		default:
			if deltas := fields(b, p); len(deltas) > 0 {
				changes = append(changes, Change{Kind: modifiedKind, ElementType: elementType, ElementID: eid, Deltas: deltas})
			}
		}
	}
	return changes
}

func indexByID[T any](items []T, id func(T) string) map[string]T {
	m := make(map[string]T, len(items))
	for _, item := range items {
		m[id(item)] = item
	}
	return m
}

func diffBoundaryFields(b, p sas.Boundary) []FieldDelta {
	var deltas []FieldDelta
	addFieldDelta(&deltas, "kind", string(b.Kind), string(p.Kind))
	addFieldDelta(&deltas, "name", b.Name, p.Name)
	addFieldDelta(&deltas, "attributes", attributeMapString(b.Attributes), attributeMapString(p.Attributes))
	addFieldDelta(&deltas, "compliance.fedrampBoundary", complianceFedRAMPString(b.Compliance), complianceFedRAMPString(p.Compliance))
	addFieldDelta(&deltas, "compliance.controls", complianceControlsString(b.Compliance), complianceControlsString(p.Compliance))
	return deltas
}

// addFieldDelta appends a FieldDelta for field only when before and after
// differ, so an unchanged field produces no noise in the ChangeSet.
func addFieldDelta(deltas *[]FieldDelta, field, before, after string) {
	if before == after {
		return
	}
	*deltas = append(*deltas, FieldDelta{Field: field, Before: before, After: after})
}

func unionKeys[V any](base, proposed map[string]V) []string {
	seen := make(map[string]bool, len(base)+len(proposed))
	for id := range base {
		seen[id] = true
	}
	for id := range proposed {
		seen[id] = true
	}
	ids := make([]string, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func toSet(values []string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, v := range values {
		set[v] = true
	}
	return set
}

func sortedKeys(set map[string]bool) []string {
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func stringSet(values []string) string {
	if len(values) == 0 {
		return ""
	}
	sorted := append([]string(nil), values...)
	sort.Strings(sorted)
	return strings.Join(sorted, ",")
}

func operationSet(ops []sas.Operation) string {
	values := make([]string, len(ops))
	for i, op := range ops {
		values[i] = string(op)
	}
	return stringSet(values)
}

func technologyString(t *sas.Technology) string {
	if t == nil {
		return ""
	}
	if t.Service == "" {
		return t.Provider
	}
	return t.Provider + "/" + t.Service
}

func identityString(id *sas.Identity) string {
	if id == nil {
		return ""
	}
	parts := []string{string(id.Type)}
	if id.Mechanism != "" {
		parts = append(parts, string(id.Mechanism))
	}
	if id.Ref != "" {
		parts = append(parts, id.Ref)
	}
	return strings.Join(parts, ":")
}

func identityRefString(ref *sas.IdentityRef) string {
	if ref == nil {
		return ""
	}
	if ref.NodeID != "" {
		return "node:" + ref.NodeID
	}
	return identityString(ref.Identity)
}

func entitlementSet(auth *sas.Authorization) string {
	if auth == nil || len(auth.Entitlements) == 0 {
		return ""
	}
	values := make([]string, len(auth.Entitlements))
	for i, e := range auth.Entitlements {
		values[i] = fmt.Sprintf("%s:%s:%s", e.Subject, e.Action, e.Resource)
	}
	return stringSet(values)
}

func dataClassificationSet(data *sas.DataFlow) string {
	if data == nil {
		return ""
	}
	return stringSet(data.Classifications)
}

func attributeMapString(attrs map[string]string) string {
	if len(attrs) == 0 {
		return ""
	}
	values := make([]string, 0, len(attrs))
	for k, v := range attrs {
		values = append(values, k+"="+v)
	}
	return stringSet(values)
}

func complianceFedRAMPString(c *sas.ComplianceBoundary) string {
	if c == nil {
		return ""
	}
	return fmt.Sprintf("%t", c.FedRAMPBoundary)
}

func complianceControlsString(c *sas.ComplianceBoundary) string {
	if c == nil {
		return ""
	}
	return stringSet(c.Controls)
}

// transportField is a small enum used only to keep the four
// transportField calls in diffRelationshipFields free of repeated nil
// checks on Relationship.Transport.
type transportSubfield int

const (
	protocol transportSubfield = iota
	port
	applicationProtocol
	encryption
)

func transportField(r sas.Relationship, sub transportSubfield) string {
	if r.Transport == nil {
		return ""
	}
	switch sub {
	case protocol:
		return r.Transport.Protocol
	case port:
		if r.Transport.Port == 0 {
			return ""
		}
		return fmt.Sprintf("%d", r.Transport.Port)
	case applicationProtocol:
		return r.Transport.ApplicationProtocol
	case encryption:
		return r.Transport.Encryption
	default:
		return ""
	}
}
