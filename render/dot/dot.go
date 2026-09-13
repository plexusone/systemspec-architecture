// Package dot renders a sas.Architecture, filtered through a sas.View, as
// Graphviz DOT source. Render is a pure function: the same (architecture,
// view) input always produces the same output string.
package dot

import (
	"fmt"
	"strings"

	"github.com/plexusone/systemspec-architecture/render"
	"github.com/plexusone/systemspec-architecture/sas"
)

// Render produces DOT source for view's selection over arch, suitable
// for Graphviz's automatic layout on large graphs. Nodes are grouped into
// a "cluster_"-prefixed subgraph — the prefix Graphviz requires to draw a
// visible boundary around it — when view.GroupBy names a boundary kind.
// Every identifier is quoted: SAS IDs are commonly kebab-case, and
// unquoted DOT identifiers cannot contain hyphens.
func Render(arch *sas.Architecture, view sas.View) (string, error) {
	sel := arch.Select(view)

	var b strings.Builder
	b.WriteString("digraph architecture {\n")
	b.WriteString("  rankdir=TD;\n")

	for _, g := range render.GroupNodes(sel, view.GroupBy) {
		if g.Boundary != nil {
			fmt.Fprintf(&b, "  subgraph %s {\n", dotQuoted("cluster_"+g.Boundary.ID))
			fmt.Fprintf(&b, "    label=%s;\n", dotQuoted(g.Boundary.Name))
			for _, n := range g.Nodes {
				writeNode(&b, n, "    ")
			}
			b.WriteString("  }\n")
		} else {
			for _, n := range g.Nodes {
				writeNode(&b, n, "  ")
			}
		}
	}

	for _, r := range sel.Relationships {
		fmt.Fprintf(&b, "  %s -> %s [label=%s];\n", dotQuoted(r.From), dotQuoted(r.To), dotQuoted(render.EdgeLabel(r)))
	}

	b.WriteString("}\n")
	return b.String(), nil
}

func writeNode(b *strings.Builder, n sas.Node, indent string) {
	fmt.Fprintf(b, "%s%s [label=%s, shape=%s];\n", indent, dotQuoted(n.ID), dotQuoted(n.Name), dotShapeName(render.ShapeForKind(n.Kind)))
}

func dotShapeName(shape render.Shape) string {
	switch shape {
	case render.ShapeCylinder:
		return "cylinder"
	case render.ShapeStadium:
		return "ellipse"
	case render.ShapeHexagon:
		return "hexagon"
	default:
		return "box"
	}
}

// dotQuoted quotes a DOT identifier or label, escaping embedded double
// quotes. Every identifier is quoted unconditionally: unquoted DOT IDs
// cannot contain hyphens, which SAS IDs commonly do.
func dotQuoted(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
}
