package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/plexusone/systemspec-architecture/assure"
	"github.com/plexusone/systemspec-architecture/sas"
)

// AssureOptions configures an Assure run.
type AssureOptions struct {
	// ArchitecturePath is the path to a SAS architecture JSON document.
	// Required.
	ArchitecturePath string
}

// Assure loads the architecture at opts.ArchitecturePath and computes its
// assurance-reference coverage report.
func Assure(opts AssureOptions) (assure.Report, error) {
	data, err := os.ReadFile(opts.ArchitecturePath)
	if err != nil {
		return assure.Report{}, fmt.Errorf("read %s: %w", opts.ArchitecturePath, err)
	}

	var arch sas.Architecture
	if err := json.Unmarshal(data, &arch); err != nil {
		return assure.Report{}, fmt.Errorf("parse %s: %w", opts.ArchitecturePath, err)
	}

	return assure.Assure(&arch), nil
}
