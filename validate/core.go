package validate

import (
	"fmt"

	"github.com/plexusone/systemspec-architecture/sas"
)

// checkReferentialIntegrity runs regardless of requested profile: every
// reference an author writes into the document must resolve to something
// that exists in it. A dangling reference is not a profile-optional
// concern — it means the document is internally inconsistent.
func checkReferentialIntegrity(arch *sas.Architecture) []Finding {
	var findings []Finding

	for i, n := range arch.Nodes {
		for _, boundaryID := range n.Boundaries {
			if _, ok := arch.BoundaryByID(boundaryID); !ok {
				findings = append(findings, Finding{
					RuleID:   "core.node-boundary-resolves",
					Severity: SeverityError,
					Message:  fmt.Sprintf("node %q references unknown boundary %q", n.ID, boundaryID),
					Path:     fmt.Sprintf("nodes[%d]", i),
				})
			}
		}
	}

	for i, r := range arch.Relationships {
		if _, ok := arch.NodeByID(r.From); !ok {
			findings = append(findings, Finding{
				RuleID:   "core.relationship-endpoints-resolve",
				Severity: SeverityError,
				Message:  fmt.Sprintf("relationship %q references unknown source node %q", r.ID, r.From),
				Path:     fmt.Sprintf("relationships[%d]", i),
			})
		}
		if _, ok := arch.NodeByID(r.To); !ok {
			findings = append(findings, Finding{
				RuleID:   "core.relationship-endpoints-resolve",
				Severity: SeverityError,
				Message:  fmt.Sprintf("relationship %q references unknown target node %q", r.ID, r.To),
				Path:     fmt.Sprintf("relationships[%d]", i),
			})
		}
		for _, boundaryID := range r.CrossesBoundaries {
			if _, ok := arch.BoundaryByID(boundaryID); !ok {
				findings = append(findings, Finding{
					RuleID:   "core.relationship-crosses-resolves",
					Severity: SeverityError,
					Message:  fmt.Sprintf("relationship %q asserts crossing unknown boundary %q", r.ID, boundaryID),
					Path:     fmt.Sprintf("relationships[%d]", i),
				})
			}
		}
	}

	return findings
}
