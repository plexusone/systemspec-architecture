package render

import "github.com/plexusone/systemspec-architecture/sas"

// Group is a bucket of nodes rendered together, either because they share
// membership in a boundary of the requested kind, or (Boundary == nil)
// because they do not belong to any boundary of that kind.
type Group struct {
	Boundary *sas.Boundary
	Nodes    []sas.Node
}

// GroupNodes buckets arch's nodes by their membership in a boundary of
// kind groupBy, for renderers that nest nodes into visual containers
// (Mermaid subgraph, D2 nested block, DOT cluster). Group order follows
// Architecture.Boundaries declaration order; a node belonging to more
// than one boundary of the same kind is placed in the first such
// boundary declared there, since a diagram needs one visual home per
// node even though the semantic model allows many-to-many membership.
// Nodes belonging to no boundary of that kind land in a trailing group
// with Boundary == nil. When groupBy is empty, GroupNodes returns a
// single ungrouped bucket containing every node in its original order.
func GroupNodes(arch sas.Architecture, groupBy string) []Group {
	if groupBy == "" {
		return []Group{{Nodes: arch.Nodes}}
	}

	var groups []Group
	index := make(map[string]int, len(arch.Boundaries)) // boundary ID -> index into groups

	for _, b := range arch.Boundaries {
		if string(b.Kind) != groupBy {
			continue
		}
		b := b
		index[b.ID] = len(groups)
		groups = append(groups, Group{Boundary: &b})
	}

	var ungrouped []sas.Node
	for _, n := range arch.Nodes {
		placed := false
		for _, boundaryID := range n.Boundaries {
			if i, ok := index[boundaryID]; ok {
				groups[i].Nodes = append(groups[i].Nodes, n)
				placed = true
				break
			}
		}
		if !placed {
			ungrouped = append(ungrouped, n)
		}
	}

	// Drop declared boundary groups that no selected node ended up in, so
	// renderers never emit an empty container.
	nonEmpty := groups[:0]
	for _, g := range groups {
		if len(g.Nodes) > 0 {
			nonEmpty = append(nonEmpty, g)
		}
	}
	groups = nonEmpty

	if len(ungrouped) > 0 {
		groups = append(groups, Group{Nodes: ungrouped})
	}

	return groups
}
