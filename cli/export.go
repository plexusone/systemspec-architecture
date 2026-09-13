package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/plexusone/systemspec-architecture/bridge/threatmodel"
	"github.com/plexusone/systemspec-architecture/sas"
)

// ExportThreatModelOptions configures an ExportThreatModel run.
type ExportThreatModelOptions struct {
	// ArchitecturePath is the path to a SAS architecture JSON document.
	// Required.
	ArchitecturePath string
}

// ExportThreatModel loads the architecture at opts.ArchitecturePath and
// returns its threat-model-spec DiagramIR system-under-analysis export as
// indented JSON.
func ExportThreatModel(opts ExportThreatModelOptions) (string, error) {
	data, err := os.ReadFile(opts.ArchitecturePath)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", opts.ArchitecturePath, err)
	}

	var arch sas.Architecture
	if err := json.Unmarshal(data, &arch); err != nil {
		return "", fmt.Errorf("parse %s: %w", opts.ArchitecturePath, err)
	}

	diagram := threatmodel.Export(&arch)
	out, err := json.MarshalIndent(diagram, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal threat model export: %w", err)
	}
	return string(out) + "\n", nil
}
