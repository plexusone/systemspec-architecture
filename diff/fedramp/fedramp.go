package fedramp

import (
	"fmt"
	"sort"
	"strings"

	"github.com/plexusone/systemspec-architecture/diff"
)

// Decision is FedRAMP's three-outcome change-assessment classification.
type Decision string

const (
	DecisionNoReviewRequired           Decision = "no_review_required"
	DecisionChangeAssessmentRequired   Decision = "change_assessment_required"
	DecisionPotentialSignificantChange Decision = "potential_significant_change"
)

// rank orders Decision by severity so the overall Classify result is the
// most severe Decision triggered by any single Impact.
var rank = map[Decision]int{
	DecisionNoReviewRequired:           0,
	DecisionChangeAssessmentRequired:   1,
	DecisionPotentialSignificantChange: 2,
}

// severityByKind maps each diff.ImpactKind to the Decision it triggers on
// its own. New external interconnections and new trust-boundary
// crossings are treated as potential significant changes, matching
// FedRAMP guidance's emphasis on new system interconnections; the
// remaining kinds warrant an assessment but are not, by themselves,
// treated as a signal of significant change.
var severityByKind = map[diff.ImpactKind]Decision{
	diff.ImpactKindNewExternalDependency:    DecisionPotentialSignificantChange,
	diff.ImpactKindNewBoundaryCrossing:      DecisionPotentialSignificantChange,
	diff.ImpactKindAuthnMechanismChanged:    DecisionChangeAssessmentRequired,
	diff.ImpactKindEntitlementExpansion:     DecisionChangeAssessmentRequired,
	diff.ImpactKindDataClassificationChange: DecisionChangeAssessmentRequired,
	diff.ImpactKindComputeModelChanged:      DecisionChangeAssessmentRequired,
}

// controlMappings is an illustrative starting point for the
// control-mapping hook, not authoritative compliance guidance. Every
// System Security Plan differs; a real deployment should review and
// replace or extend this table against its own control baseline rather
// than treat these IDs as a compliance determination.
var controlMappings = map[diff.ImpactKind][]string{
	diff.ImpactKindNewExternalDependency:    {"CA-3", "SA-9"},
	diff.ImpactKindNewBoundaryCrossing:      {"SC-7"},
	diff.ImpactKindAuthnMechanismChanged:    {"IA-2", "IA-5"},
	diff.ImpactKindEntitlementExpansion:     {"AC-3", "AC-6"},
	diff.ImpactKindDataClassificationChange: {"RA-2"},
	diff.ImpactKindComputeModelChanged:      {"CM-3", "CM-8"},
}

// Classification is this profile's decision-support output for one
// diff.ChangeImpact.
type Classification struct {
	Decision Decision `json:"decision"`

	// TriggeringImpacts are the specific Impacts that produced Decision,
	// so a reviewer can see exactly why without re-deriving it.
	TriggeringImpacts []diff.Impact `json:"triggeringImpacts,omitempty"`

	// AffectedControls names control IDs plausibly relevant to the
	// triggering impacts, via the illustrative controlMappings hook.
	AffectedControls []string `json:"affectedControls,omitempty"`

	// Rationale is a short, human-readable explanation of Decision,
	// suitable for surfacing directly in a review tool.
	Rationale string `json:"rationale"`
}

// Classify produces decision support for impact: a suggested Decision,
// the impacts that triggered it, plausibly affected controls, and a
// human-readable rationale. It never records an actual assessment
// decision — see diff.Assessment for that.
func Classify(impact diff.ChangeImpact) Classification {
	decision := DecisionNoReviewRequired
	for _, imp := range impact.Impacts {
		if d, ok := severityByKind[imp.Kind]; ok && rank[d] > rank[decision] {
			decision = d
		}
	}
	if decision == DecisionNoReviewRequired {
		return Classification{
			Decision:  decision,
			Rationale: "No changes matched a FedRAMP change-assessment trigger.",
		}
	}

	var triggering []diff.Impact
	for _, imp := range impact.Impacts {
		if d, ok := severityByKind[imp.Kind]; ok && d == decision {
			triggering = append(triggering, imp)
		}
	}

	return Classification{
		Decision:          decision,
		TriggeringImpacts: triggering,
		AffectedControls:  affectedControls(triggering),
		Rationale:         rationale(decision, triggering),
	}
}

func affectedControls(impacts []diff.Impact) []string {
	seen := make(map[string]bool)
	var controls []string
	for _, imp := range impacts {
		for _, c := range controlMappings[imp.Kind] {
			if seen[c] {
				continue
			}
			seen[c] = true
			controls = append(controls, c)
		}
	}
	sort.Strings(controls)
	return controls
}

func rationale(decision Decision, triggering []diff.Impact) string {
	kinds := make([]string, len(triggering))
	for i, imp := range triggering {
		kinds[i] = string(imp.Kind)
	}
	switch decision {
	case DecisionPotentialSignificantChange:
		return fmt.Sprintf(
			"This change includes %d potential-significant-change impact(s): %s. "+
				"FedRAMP guidance treats new external interconnections and new trust-boundary "+
				"crossings as significant-change candidates; an authorized reviewer should confirm.",
			len(triggering), strings.Join(kinds, ", "))
	case DecisionChangeAssessmentRequired:
		return fmt.Sprintf(
			"This change includes %d impact(s) requiring assessment: %s. "+
				"None independently indicate a significant change, but each should be reviewed.",
			len(triggering), strings.Join(kinds, ", "))
	default:
		return "No changes matched a FedRAMP change-assessment trigger."
	}
}
