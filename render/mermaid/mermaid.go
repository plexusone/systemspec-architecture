// Package mermaid renders a sas.Architecture, filtered through a
// sas.View, as a Mermaid flowchart. Render is a pure function: the same
// (architecture, view) input always produces the same output string.
package mermaid

import (
	"fmt"
	"strings"

	"github.com/plexusone/systemspec-architecture/render"
	"github.com/plexusone/systemspec-architecture/sas"
)

// Render produces Mermaid flowchart source for view's selection over
// arch. Nodes are grouped into subgraphs when view.GroupBy names a
// boundary kind.
func Render(arch *sas.Architecture, view sas.View) (string, error) {
	sel := arch.Select(view)

	var b strings.Builder
	b.WriteString("flowchart TD\n")

	for _, g := range render.GroupNodes(sel, view.GroupBy) {
		if g.Boundary != nil {
			fmt.Fprintf(&b, "  subgraph %s[%s]\n", mermaidID(g.Boundary.ID), mermaidQuoted(g.Boundary.Name))
			for _, n := range g.Nodes {
				writeNode(&b, n, "    ")
			}
			b.WriteString("  end\n")
		} else {
			for _, n := range g.Nodes {
				writeNode(&b, n, "  ")
			}
		}
	}

	for _, r := range sel.Relationships {
		fmt.Fprintf(&b, "  %s -->|%s| %s\n", mermaidID(r.From), mermaidEscape(render.EdgeLabel(r)), mermaidID(r.To))
	}

	return b.String(), nil
}

func writeNode(b *strings.Builder, n sas.Node, indent string) {
	open, close := shapeDelims(render.ShapeForKind(n.Kind))
	fmt.Fprintf(b, "%s%s%s%s%s\n", indent, mermaidID(n.ID), open, mermaidQuoted(n.Name), close)
}

func shapeDelims(shape render.Shape) (open, close string) {
	switch shape {
	case render.ShapeCylinder:
		return "[(", ")]"
	case render.ShapeStadium:
		return "(", ")"
	case render.ShapeHexagon:
		return "{{", "}}"
	default:
		return "[", "]"
	}
}

// mermaidID passes node/boundary IDs through unchanged: Mermaid accepts
// alphanumeric, hyphen, and underscore identifiers directly, which
// matches the kebab-case IDs used throughout SAS examples.
func mermaidID(id string) string {
	return id
}

func mermaidQuoted(s string) string {
	return `"` + mermaidEscape(s) + `"`
}

// mermaidEscape replaces double quotes with Mermaid's HTML-entity escape,
// since raw quotes inside a quoted label break Mermaid's parser.
func mermaidEscape(s string) string {
	return strings.ReplaceAll(s, `"`, "#quot;")
}
