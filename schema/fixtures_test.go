package schema

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/plexusone/systemspec-architecture/sas"
	"github.com/plexusone/systemspec-architecture/validate"
)

func compileArchitectureSchema(t *testing.T) *jsonschema.Schema {
	t.Helper()

	compiler := jsonschema.NewCompiler()

	var schemaDoc any
	if err := json.Unmarshal(ArchitectureJSON, &schemaDoc); err != nil {
		t.Fatalf("unmarshal embedded schema: %v", err)
	}
	if err := compiler.AddResource("architecture.schema.json", schemaDoc); err != nil {
		t.Fatalf("add schema resource: %v", err)
	}
	compiled, err := compiler.Compile("architecture.schema.json")
	if err != nil {
		t.Fatalf("compile schema: %v", err)
	}
	return compiled
}

func globFixtures(t *testing.T, dir string) []string {
	t.Helper()
	fixtures, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		t.Fatalf("glob %s: %v", dir, err)
	}
	if len(fixtures) == 0 {
		t.Fatalf("no fixtures found under %s", dir)
	}
	return fixtures
}

// TestValidFixtures proves the shared fixture corpus round-trips through
// the full Go/JSON Schema conformance chain: every file under
// examples/fixtures/valid must validate against the generated JSON
// Schema and unmarshal cleanly into sas.Architecture. This is the
// behavioral conformance contract referenced by the TypeScript/Zod side
// (ts/), which validates the same fixtures.
func TestValidFixtures(t *testing.T) {
	compiled := compileArchitectureSchema(t)

	for _, path := range globFixtures(t, filepath.Join("..", "examples", "fixtures", "valid")) {
		path := path
		t.Run(filepath.Base(path), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}

			var doc any
			if err := json.Unmarshal(data, &doc); err != nil {
				t.Fatalf("fixture is not valid JSON: %v", err)
			}
			if err := compiled.Validate(doc); err != nil {
				t.Fatalf("fixture does not validate against architecture.schema.json: %v", err)
			}

			var arch sas.Architecture
			if err := json.Unmarshal(data, &arch); err != nil {
				t.Fatalf("fixture does not unmarshal into sas.Architecture: %v", err)
			}
			if arch.Version == "" {
				t.Error("expected non-empty Version after unmarshal")
			}
			if len(arch.Nodes) == 0 {
				t.Error("expected at least one node after unmarshal")
			}
		})
	}
}

// TestInvalidFixtures proves examples/fixtures/invalid documents are
// correctly rejected by JSON Schema validation, so the schema's required
// fields and types are actually enforced rather than merely documented.
func TestInvalidFixtures(t *testing.T) {
	compiled := compileArchitectureSchema(t)

	for _, path := range globFixtures(t, filepath.Join("..", "examples", "fixtures", "invalid")) {
		path := path
		t.Run(filepath.Base(path), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}

			var doc any
			if err := json.Unmarshal(data, &doc); err != nil {
				t.Fatalf("fixture is not valid JSON: %v", err)
			}
			if err := compiled.Validate(doc); err == nil {
				t.Fatal("expected schema validation to reject this fixture, but it passed")
			}
		})
	}
}

// TestDogfoodArchitectures proves the real portfolio architectures under
// examples/dogfood validate against the JSON Schema, unmarshal cleanly,
// and have no dangling references — the correctness bar any real
// architecture should meet regardless of how complete its profile-level
// metadata (deployment, sre) is.
func TestDogfoodArchitectures(t *testing.T) {
	compiled := compileArchitectureSchema(t)

	for _, path := range globFixtures(t, filepath.Join("..", "examples", "dogfood")) {
		path := path
		t.Run(filepath.Base(path), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}

			var doc any
			if err := json.Unmarshal(data, &doc); err != nil {
				t.Fatalf("fixture is not valid JSON: %v", err)
			}
			if err := compiled.Validate(doc); err != nil {
				t.Fatalf("dogfood architecture does not validate against architecture.schema.json: %v", err)
			}

			var arch sas.Architecture
			if err := json.Unmarshal(data, &arch); err != nil {
				t.Fatalf("dogfood architecture does not unmarshal into sas.Architecture: %v", err)
			}
			if len(arch.Nodes) == 0 {
				t.Fatal("expected at least one node")
			}
			if len(arch.Relationships) == 0 {
				t.Fatal("expected at least one relationship")
			}

			for _, f := range validate.Validate(&arch) {
				t.Errorf("referential integrity finding: %s %s: %s", f.RuleID, f.Path, f.Message)
			}
		})
	}
}
