package validate

import (
	"fmt"

	"github.com/plexusone/systemspec-architecture/sas"
)

// checkThreatModel implements the threat-model profile: a relationship
// touching an external_service node must declare its data
// classifications, so a threat model always knows what data leaves the
// system's own trust boundary.
func checkThreatModel(arch *sas.Architecture) []Finding {
	var findings []Finding

	for i, r := range arch.Relationships {
		if !touchesExternalNode(arch, r) {
			continue
		}
		if r.Data == nil || len(r.Data.Classifications) == 0 {
			findings = append(findings, Finding{
				RuleID:   "threat-model.external-edge-data-classification-required",
				Severity: SeverityError,
				Profile:  ProfileThreatModel,
				Message:  fmt.Sprintf("relationship %q touches an external service but declares no data classifications", r.ID),
				Path:     fmt.Sprintf("relationships[%d]", i),
			})
		}
	}

	return findings
}

func touchesExternalNode(arch *sas.Architecture, r sas.Relationship) bool {
	for _, id := range []string{r.From, r.To} {
		if n, ok := arch.NodeByID(id); ok && n.Kind == sas.NodeKindExternalService {
			return true
		}
	}
	return false
}
