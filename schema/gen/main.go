//go:build ignore

// Command gen generates the JSON Schema document from the sas.Architecture
// Go type, which is the source of truth. Run via `go generate ./schema/...`
// from the repository root.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/invopop/jsonschema"

	"github.com/plexusone/systemspec-architecture/sas"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	return generate(&sas.Architecture{}, "architecture.schema.json")
}

func generate(v any, filename string) error {
	reflector := &jsonschema.Reflector{
		ExpandedStruct: true,
	}
	schema := reflector.Reflect(v)
	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal schema for %s: %w", filename, err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(filename, data, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", filename, err)
	}
	return nil
}
