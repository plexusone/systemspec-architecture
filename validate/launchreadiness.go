package validate

import (
	"fmt"

	"github.com/plexusone/systemspec-architecture/sas"
)

// checkLaunchReadiness implements the launch-readiness rules: the go-live
// gate for a portfolio web app. It is dispatched under the security
// profile, not a separate one — `sas validate --profile security` is
// meant to be the single command a launch checklist runs.
//
// A relationship is treated as internet-facing when its source is a
// human actor or an unmanaged external system (not one of our own
// services calling out) and it crosses into a boundary the source does
// not belong to — i.e. traffic entering the architecture's own trust
// zone from outside it, not the architecture calling out to a
// third-party dependency.
func checkLaunchReadiness(arch *sas.Architecture) []Finding {
	var findings []Finding

	for i, r := range arch.Relationships {
		if !isInternetFacing(arch, r) {
			continue
		}

		if r.Transport == nil || r.Transport.Encryption == "" {
			findings = append(findings, Finding{
				RuleID:   "launch-readiness.internet-facing-tls-required",
				Severity: SeverityError,
				Profile:  ProfileSecurity,
				Message:  fmt.Sprintf("relationship %q is internet-facing but declares no transport encryption", r.ID),
				Path:     fmt.Sprintf("relationships[%d]", i),
			})
		}
		if r.Identity == nil {
			findings = append(findings, Finding{
				RuleID:   "launch-readiness.internet-facing-authn-required",
				Severity: SeverityError,
				Profile:  ProfileSecurity,
				Message:  fmt.Sprintf("relationship %q is internet-facing but declares no identity", r.ID),
				Path:     fmt.Sprintf("relationships[%d]", i),
			})
		}
		if r.Data == nil || len(r.Data.Classifications) == 0 {
			findings = append(findings, Finding{
				RuleID:   "launch-readiness.internet-facing-data-classification-required",
				Severity: SeverityError,
				Profile:  ProfileSecurity,
				Message:  fmt.Sprintf("relationship %q is internet-facing but declares no data classifications", r.ID),
				Path:     fmt.Sprintf("relationships[%d]", i),
			})
		}

		target, ok := arch.NodeByID(r.To)
		if ok && target.Owner == "" {
			findings = append(findings, Finding{
				RuleID:   "launch-readiness.internet-facing-owner-required",
				Severity: SeverityError,
				Profile:  ProfileSecurity,
				Message:  fmt.Sprintf("node %q is internet-facing (via relationship %q) but declares no owner", target.ID, r.ID),
				Path:     fmt.Sprintf("relationships[%d]", i),
			})
		}
	}

	return findings
}

// isInternetFacing reports whether r represents traffic entering the
// architecture's own trust zone from a human actor or an unmanaged
// external system, as distinct from the architecture calling out to a
// third-party dependency (e.g. our service calling an OAuth provider is
// not internet-facing in this sense; a user's browser calling our web
// app is).
func isInternetFacing(arch *sas.Architecture, r sas.Relationship) bool {
	from, ok := arch.NodeByID(r.From)
	if !ok {
		return false
	}
	if from.Kind != sas.NodeKindActor && from.Kind != sas.NodeKindExternalService {
		return false
	}
	return len(arch.CrossedBoundaries(r)) > 0
}
