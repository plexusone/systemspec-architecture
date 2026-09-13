package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/plexusone/systemspec-architecture/render/d2"
	"github.com/plexusone/systemspec-architecture/render/dot"
	"github.com/plexusone/systemspec-architecture/render/mermaid"
	"github.com/plexusone/systemspec-architecture/sas"
)

// ViewOptions configures a View render.
type ViewOptions struct {
	// ArchitecturePath is the path to a SAS architecture JSON document.
	// Required.
	ArchitecturePath string

	// ViewID, when non-empty, looks up a view declared in the
	// architecture's own Views list. Mutually exclusive with the ad-hoc
	// selection fields below — when ViewID is set, they are ignored.
	ViewID string

	// The remaining fields build an ad-hoc sas.View when ViewID is
	// empty; see sas.View for their meaning.
	IncludeKinds      []string
	IncludeRelations  []string
	IncludeBoundaries []string
	GroupBy           string

	// Format selects the renderer: "mermaid", "d2", or "dot".
	Format string
}

// View loads the architecture document at opts.ArchitecturePath, resolves
// the view to render (by ID lookup or from the ad-hoc selection fields),
// and renders it in the requested format.
func View(opts ViewOptions) (string, error) {
	data, err := os.ReadFile(opts.ArchitecturePath)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", opts.ArchitecturePath, err)
	}

	var arch sas.Architecture
	if err := json.Unmarshal(data, &arch); err != nil {
		return "", fmt.Errorf("parse %s: %w", opts.ArchitecturePath, err)
	}

	view, err := resolveView(&arch, opts)
	if err != nil {
		return "", err
	}

	switch opts.Format {
	case "mermaid":
		return mermaid.Render(&arch, view)
	case "d2":
		return d2.Render(&arch, view)
	case "dot":
		return dot.Render(&arch, view)
	default:
		return "", fmt.Errorf("unknown format %q (want mermaid, d2, or dot)", opts.Format)
	}
}

func resolveView(arch *sas.Architecture, opts ViewOptions) (sas.View, error) {
	if opts.ViewID != "" {
		v, ok := arch.ViewByID(opts.ViewID)
		if !ok {
			return sas.View{}, fmt.Errorf("view %q not found in %s", opts.ViewID, opts.ArchitecturePath)
		}
		return *v, nil
	}
	return sas.View{
		ID:                "ad-hoc",
		IncludeKinds:      opts.IncludeKinds,
		IncludeRelations:  opts.IncludeRelations,
		IncludeBoundaries: opts.IncludeBoundaries,
		GroupBy:           opts.GroupBy,
	}, nil
}
