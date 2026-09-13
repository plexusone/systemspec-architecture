package validate

import (
	"fmt"

	"github.com/plexusone/systemspec-architecture/sas"
)

// deploymentBoundaryKinds are the boundary kinds that satisfy the
// deployment profile's "every node belongs to an account/region/network
// boundary" requirement.
var deploymentBoundaryKinds = map[sas.BoundaryKind]bool{
	sas.BoundaryKindAccount: true,
	sas.BoundaryKindRegion:  true,
	sas.BoundaryKindNetwork: true,
}

// checkDeployment implements the deployment profile: every node must
// declare its Technology and belong to an account, region, or network
// boundary; every relationship must declare its Transport.
func checkDeployment(arch *sas.Architecture) []Finding {
	var findings []Finding

	for i, n := range arch.Nodes {
		if n.Technology == nil {
			findings = append(findings, Finding{
				RuleID:   "deployment.node-technology-required",
				Severity: SeverityError,
				Profile:  ProfileDeployment,
				Message:  fmt.Sprintf("node %q has no technology declared", n.ID),
				Path:     fmt.Sprintf("nodes[%d]", i),
			})
		}

		if !nodeInDeploymentBoundary(arch, n) {
			findings = append(findings, Finding{
				RuleID:   "deployment.node-boundary-required",
				Severity: SeverityError,
				Profile:  ProfileDeployment,
				Message:  fmt.Sprintf("node %q does not belong to an account, region, or network boundary", n.ID),
				Path:     fmt.Sprintf("nodes[%d]", i),
			})
		}
	}

	for i, r := range arch.Relationships {
		if r.Transport == nil {
			findings = append(findings, Finding{
				RuleID:   "deployment.relationship-transport-required",
				Severity: SeverityError,
				Profile:  ProfileDeployment,
				Message:  fmt.Sprintf("relationship %q has no transport declared", r.ID),
				Path:     fmt.Sprintf("relationships[%d]", i),
			})
		}
	}

	return findings
}

func nodeInDeploymentBoundary(arch *sas.Architecture, n sas.Node) bool {
	for _, boundaryID := range n.Boundaries {
		b, ok := arch.BoundaryByID(boundaryID)
		if ok && deploymentBoundaryKinds[b.Kind] {
			return true
		}
	}
	return false
}
