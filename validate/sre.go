package validate

import (
	"fmt"

	"github.com/plexusone/systemspec-architecture/sas"
)

// sreCriticalTiers are the Criticality values the sre profile treats as
// requiring an owner, an availability target, and assurance metrics.
var sreCriticalTiers = map[sas.Criticality]bool{
	sas.CriticalityTier0: true,
	sas.CriticalityTier1: true,
}

// checkSRE implements the sre profile: a tier0 or tier1 node must declare
// an owner, an SRE extension with an availability target, and at least
// one assurance metric reference.
func checkSRE(arch *sas.Architecture) []Finding {
	var findings []Finding

	for i, n := range arch.Nodes {
		if !sreCriticalTiers[n.Criticality] {
			continue
		}

		if n.Owner == "" {
			findings = append(findings, Finding{
				RuleID:   "sre.critical-owner-required",
				Severity: SeverityError,
				Profile:  ProfileSRE,
				Message:  fmt.Sprintf("node %q is %s but declares no owner", n.ID, n.Criticality),
				Path:     fmt.Sprintf("nodes[%d]", i),
			})
		}

		if n.Extensions == nil || n.Extensions.SRE == nil || n.Extensions.SRE.AvailabilityTarget == "" {
			findings = append(findings, Finding{
				RuleID:   "sre.critical-availability-target-required",
				Severity: SeverityError,
				Profile:  ProfileSRE,
				Message:  fmt.Sprintf("node %q is %s but declares no sre.availabilityTarget extension", n.ID, n.Criticality),
				Path:     fmt.Sprintf("nodes[%d]", i),
			})
		}

		if n.Assurance == nil || len(n.Assurance.Metrics) == 0 {
			findings = append(findings, Finding{
				RuleID:   "sre.critical-assurance-metrics-required",
				Severity: SeverityError,
				Profile:  ProfileSRE,
				Message:  fmt.Sprintf("node %q is %s but declares no assurance metrics", n.ID, n.Criticality),
				Path:     fmt.Sprintf("nodes[%d]", i),
			})
		}
	}

	return findings
}
