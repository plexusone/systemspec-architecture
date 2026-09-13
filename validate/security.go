package validate

import (
	"fmt"

	"github.com/plexusone/systemspec-architecture/sas"
)

// writeOperations are the Operation values the security profile treats
// as requiring explicit authorization.
var writeOperations = map[sas.Operation]bool{
	sas.OperationCreate:     true,
	sas.OperationUpdate:     true,
	sas.OperationDelete:     true,
	sas.OperationAdminister: true,
}

// checkSecurity implements the security profile: a relationship that
// crosses a trust boundary (derived from node membership via
// sas.Architecture.CrossedBoundaries) must declare transport encryption
// and an identity; a relationship performing a write operation must
// declare its authorization entitlements.
func checkSecurity(arch *sas.Architecture) []Finding {
	var findings []Finding

	for i, r := range arch.Relationships {
		if crossed := arch.CrossedBoundaries(r); len(crossed) > 0 {
			if r.Transport == nil || r.Transport.Encryption == "" {
				findings = append(findings, Finding{
					RuleID:   "security.boundary-crossing-encryption-required",
					Severity: SeverityError,
					Profile:  ProfileSecurity,
					Message:  fmt.Sprintf("relationship %q crosses boundaries %v but declares no transport encryption", r.ID, crossed),
					Path:     fmt.Sprintf("relationships[%d]", i),
				})
			}
			if r.Identity == nil {
				findings = append(findings, Finding{
					RuleID:   "security.boundary-crossing-identity-required",
					Severity: SeverityError,
					Profile:  ProfileSecurity,
					Message:  fmt.Sprintf("relationship %q crosses boundaries %v but declares no identity", r.ID, crossed),
					Path:     fmt.Sprintf("relationships[%d]", i),
				})
			}
		}

		if hasWriteOperation(r.Operations) {
			if r.Authorization == nil || len(r.Authorization.Entitlements) == 0 {
				findings = append(findings, Finding{
					RuleID:   "security.write-entitlements-required",
					Severity: SeverityError,
					Profile:  ProfileSecurity,
					Message:  fmt.Sprintf("relationship %q performs a write operation but declares no authorization entitlements", r.ID),
					Path:     fmt.Sprintf("relationships[%d]", i),
				})
			}
		}
	}

	return findings
}

func hasWriteOperation(ops []sas.Operation) bool {
	for _, op := range ops {
		if writeOperations[op] {
			return true
		}
	}
	return false
}
