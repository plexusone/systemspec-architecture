package validate

import "github.com/plexusone/systemspec-architecture/sas"

// Validate checks arch against the always-on referential-integrity rules
// plus the rules for every requested profile, and returns every finding.
// A clean architecture returns an empty (non-nil-length-zero) slice.
// Unrecognized profiles are ignored by the check dispatch below but
// reported as findings themselves, so a typo'd profile name is visible
// rather than silently a no-op.
func Validate(arch *sas.Architecture, profiles ...Profile) []Finding {
	var findings []Finding

	findings = append(findings, checkReferentialIntegrity(arch)...)

	requested := make(map[Profile]bool, len(profiles))
	for _, p := range profiles {
		if requested[p] {
			continue
		}
		requested[p] = true

		if !p.IsKnown() {
			findings = append(findings, Finding{
				RuleID:   "core.unknown-profile",
				Severity: SeverityError,
				Message:  "unknown profile \"" + string(p) + "\"",
				Path:     "$",
			})
			continue
		}

		switch p {
		case ProfileDevelopment:
			// No additional rules: development is the base profile.
		case ProfileDeployment:
			findings = append(findings, checkDeployment(arch)...)
		case ProfileSecurity:
			findings = append(findings, checkSecurity(arch)...)
			findings = append(findings, checkLaunchReadiness(arch)...)
		case ProfileThreatModel:
			findings = append(findings, checkThreatModel(arch)...)
		case ProfileSRE:
			findings = append(findings, checkSRE(arch)...)
		}
	}

	return findings
}
