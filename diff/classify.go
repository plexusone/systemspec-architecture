package diff

import (
	"fmt"
	"strings"

	"github.com/plexusone/systemspec-architecture/sas"
)

// writeOperations are the sas.Operation values that constitute a write,
// for the purposes of detecting entitlement expansion. discover/read are
// deliberately excluded; execute is deliberately excluded too since it
// does not itself imply a change to a resource's state.
var writeOperations = map[string]bool{
	string(sas.OperationCreate):     true,
	string(sas.OperationUpdate):     true,
	string(sas.OperationDelete):     true,
	string(sas.OperationAdminister): true,
}

// Classify derives the semantic ChangeImpact from a ChangeSet. proposed
// supplies the full field values for elements diff didn't need to
// materialize on Added changes (e.g. a new relationship's target node
// kind, or its full CrossesBoundaries list).
func Classify(cs ChangeSet, proposed *sas.Architecture) ChangeImpact {
	nodesByID := indexByID(proposed.Nodes, func(n sas.Node) string { return n.ID })
	relsByID := indexByID(proposed.Relationships, func(r sas.Relationship) string { return r.ID })

	var impacts []Impact
	for _, c := range cs.Changes {
		switch c.ElementType {
		case ElementTypeRelationship:
			impacts = append(impacts, classifyRelationshipChange(c, nodesByID, relsByID)...)
		case ElementTypeNode:
			impacts = append(impacts, classifyNodeChange(c, nodesByID)...)
		}
	}
	return ChangeImpact{Impacts: impacts}
}

func classifyRelationshipChange(c Change, nodesByID map[string]sas.Node, relsByID map[string]sas.Relationship) []Impact {
	var impacts []Impact
	switch c.Kind {
	case ChangeKindRelationshipAdded:
		r, ok := relsByID[c.ElementID]
		if !ok {
			return nil
		}
		if target, ok := nodesByID[r.To]; ok && target.Kind == sas.NodeKindExternalService {
			impacts = append(impacts, Impact{
				Kind: ImpactKindNewExternalDependency, ElementType: ElementTypeRelationship, ElementID: c.ElementID,
				Detail: fmt.Sprintf("new relationship to external service %q", r.To),
			})
		}
		if len(r.CrossesBoundaries) > 0 {
			impacts = append(impacts, Impact{
				Kind: ImpactKindNewBoundaryCrossing, ElementType: ElementTypeRelationship, ElementID: c.ElementID,
				AffectedBoundaries: append([]string(nil), r.CrossesBoundaries...),
				Detail:             "new relationship crosses a boundary from the start",
			})
		}
		if r.Data != nil && len(r.Data.Classifications) > 0 {
			impacts = append(impacts, Impact{
				Kind: ImpactKindDataClassificationChange, ElementType: ElementTypeRelationship, ElementID: c.ElementID,
				Detail: fmt.Sprintf("new relationship carries data classifications: %s", stringSet(r.Data.Classifications)),
			})
		}
		if identityRefString(r.Identity) != "" {
			impacts = append(impacts, Impact{
				Kind: ImpactKindAuthnMechanismChanged, ElementType: ElementTypeRelationship, ElementID: c.ElementID,
				Detail: fmt.Sprintf("new relationship established with identity %q", identityRefString(r.Identity)),
			})
		}
	case ChangeKindRelationshipModified:
		for _, d := range c.Deltas {
			impacts = append(impacts, classifyRelationshipFieldDelta(c.ElementID, d)...)
		}
	}
	return impacts
}

func classifyRelationshipFieldDelta(elementID string, d FieldDelta) []Impact {
	switch d.Field {
	case "identity":
		return []Impact{{
			Kind: ImpactKindAuthnMechanismChanged, ElementType: ElementTypeRelationship, ElementID: elementID,
			Detail: fmt.Sprintf("identity changed from %q to %q", d.Before, d.After),
		}}
	case "operations":
		added := setDiffAdded(d.Before, d.After)
		if !containsWriteOperation(added) {
			return nil
		}
		return []Impact{{
			Kind: ImpactKindEntitlementExpansion, ElementType: ElementTypeRelationship, ElementID: elementID,
			Detail: fmt.Sprintf("operations expanded from %q to %q", d.Before, d.After),
		}}
	case "authorization.entitlements":
		added := setDiffAdded(d.Before, d.After)
		if len(added) == 0 {
			return nil
		}
		return []Impact{{
			Kind: ImpactKindEntitlementExpansion, ElementType: ElementTypeRelationship, ElementID: elementID,
			Detail: fmt.Sprintf("entitlements added: %s", strings.Join(added, ",")),
		}}
	case "crossesBoundaries":
		added := setDiffAdded(d.Before, d.After)
		if len(added) == 0 {
			return nil
		}
		return []Impact{{
			Kind: ImpactKindNewBoundaryCrossing, ElementType: ElementTypeRelationship, ElementID: elementID,
			AffectedBoundaries: added,
			Detail:             fmt.Sprintf("relationship now additionally crosses: %s", strings.Join(added, ",")),
		}}
	case "data.classifications":
		return []Impact{{
			Kind: ImpactKindDataClassificationChange, ElementType: ElementTypeRelationship, ElementID: elementID,
			Detail: fmt.Sprintf("data classifications changed from %q to %q", d.Before, d.After),
		}}
	default:
		return nil
	}
}

func classifyNodeChange(c Change, nodesByID map[string]sas.Node) []Impact {
	var impacts []Impact
	switch c.Kind {
	case ChangeKindNodeAdded:
		if n, ok := nodesByID[c.ElementID]; ok && n.Kind == sas.NodeKindExternalService {
			impacts = append(impacts, Impact{
				Kind: ImpactKindNewExternalDependency, ElementType: ElementTypeNode, ElementID: c.ElementID,
				Detail: "new external_service node added",
			})
		}
	case ChangeKindNodeModified:
		for _, d := range c.Deltas {
			if d.Field != "technology" {
				continue
			}
			impacts = append(impacts, Impact{
				Kind: ImpactKindComputeModelChanged, ElementType: ElementTypeNode, ElementID: c.ElementID,
				Detail: fmt.Sprintf("technology changed from %q to %q", d.Before, d.After),
			})
		}
	}
	return impacts
}

func containsWriteOperation(ops []string) bool {
	for _, op := range ops {
		if writeOperations[op] {
			return true
		}
	}
	return false
}

// setDiffAdded returns the elements present in after (a comma-joined
// sorted set, as produced by stringSet) that are not present in before.
func setDiffAdded(before, after string) []string {
	beforeSet := toSet(splitSet(before))
	var added []string
	for _, v := range splitSet(after) {
		if !beforeSet[v] {
			added = append(added, v)
		}
	}
	return added
}

func splitSet(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}
