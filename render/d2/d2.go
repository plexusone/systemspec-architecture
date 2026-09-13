// Package d2 renders a sas.Architecture, filtered through a sas.View, as
// D2 diagram source (https://d2lang.com). Render is a pure function: the
// same (architecture, view) input always produces the same output
// string.
package d2

import (
	"fmt"
	"strings"

	"github.com/plexusone/systemspec-architecture/render"
	"github.com/plexusone/systemspec-architecture/sas"
)

// Render produces D2 source for view's selection over arch. Nodes are
// nested inside a container block when view.GroupBy names a boundary
// kind. Edges are always declared at the top level using fully qualified
// paths for grouped nodes, which sidesteps D2's container-scoping rules
// entirely rather than relying on them.
func Render(arch *sas.Architecture, view sas.View) (string, error) {
	sel := arch.Select(view)
	groups := render.GroupNodes(sel, view.GroupBy)

	var b strings.Builder
	paths := make(map[string]string, len(sel.Nodes))

	for _, g := range groups {
		if g.Boundary != nil {
			fmt.Fprintf(&b, "%s: %s {\n", d2ID(g.Boundary.ID), d2Quoted(g.Boundary.Name))
			for _, n := range g.Nodes {
				writeNode(&b, n, "  ")
				paths[n.ID] = g.Boundary.ID + "." + n.ID
			}
			b.WriteString("}\n")
		} else {
			for _, n := range g.Nodes {
				writeNode(&b, n, "")
				paths[n.ID] = n.ID
			}
		}
	}

	for _, r := range sel.Relationships {
		fmt.Fprintf(&b, "%s -> %s: %s\n", d2ID(paths[r.From]), d2ID(paths[r.To]), render.EdgeLabel(r))
	}

	return b.String(), nil
}

func writeNode(b *strings.Builder, n sas.Node, indent string) {
	shape := render.ShapeForKind(n.Kind)
	if shape == render.ShapeRectangle {
		fmt.Fprintf(b, "%s%s: %s\n", indent, d2ID(n.ID), d2Quoted(n.Name))
		return
	}
	fmt.Fprintf(b, "%s%s: %s {\n%s  shape: %s\n%s}\n", indent, d2ID(n.ID), d2Quoted(n.Name), indent, d2ShapeName(shape), indent)
}

func d2ShapeName(shape render.Shape) string {
	switch shape {
	case render.ShapeCylinder:
		return "cylinder"
	case render.ShapeStadium:
		return "person"
	case render.ShapeHexagon:
		return "hexagon"
	default:
		return "rectangle"
	}
}

// d2ID passes node/boundary IDs (and their dotted qualified paths)
// through unchanged: D2 accepts hyphenated bare identifiers directly,
// which matches the kebab-case IDs used throughout SAS examples.
func d2ID(id string) string {
	return id
}

func d2Quoted(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
}
