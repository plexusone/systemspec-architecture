package render

import (
	"strings"

	"github.com/plexusone/systemspec-architecture/sas"
)

// EdgeLabel chooses what text to show on a rendered relationship: the
// comma-joined operations when the author supplied them (the
// security-relevant fact — "read, create" tells a reader more than an
// unlabeled arrow), falling back to the relationship kind when no
// operations are declared.
//
// View.Level distinguishes a generic-verb abstraction from
// protocol-specific detail, but sas.Relationship does not yet carry
// protocol-specific operation text (e.g. "HTTPS GET /v1/orders")
// separately from the generic verb — so every renderer currently ignores
// Level and always shows the generic form. Specific-level rendering will
// use it once that data exists on the model.
func EdgeLabel(r sas.Relationship) string {
	if len(r.Operations) > 0 {
		ops := make([]string, len(r.Operations))
		for i, op := range r.Operations {
			ops[i] = string(op)
		}
		return strings.Join(ops, ", ")
	}
	return string(r.Kind)
}
