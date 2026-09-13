package diff

import (
	"time"

	"github.com/plexusone/systemspec-architecture/sas"
)

// Baseline is a named, approved architecture version. Once approved, a
// Baseline is immutable by convention: callers must not mutate
// Architecture in place after construction, the same way a Git tag names
// a commit rather than a mutable ref. SAS does not enforce this at the
// language level (no exported struct in this codebase does), but every
// Baseline consumer treats it as a historical fact, never a working copy.
type Baseline struct {
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	Architecture sas.Architecture `json:"architecture"`
	ApprovedAt   time.Time        `json:"approvedAt"`
	ApprovedBy   string           `json:"approvedBy"`
}

// Assessment is a recorded human decision about a proposed architecture
// change relative to a Baseline, with the evidence it was based on.
//
// Decision is a plain string, not a closed enum, deliberately: the
// vocabulary of valid decisions is a compliance-profile concern (the
// FedRAMP profile's three-outcome NO_REVIEW_REQUIRED /
// CHANGE_ASSESSMENT_REQUIRED / POTENTIAL_SIGNIFICANT_CHANGE vocabulary is
// just one profile's), and compliance logic lives in profiles, not core —
// the same principle that keeps validate.Profile rules out of the sas
// package itself. A profile package defines its own typed decision
// constants and is responsible for only ever writing one of them here.
type Assessment struct {
	ID         string    `json:"id"`
	BaselineID string    `json:"baselineId"`
	Decision   string    `json:"decision"`
	Reviewer   string    `json:"reviewer"`
	DecidedAt  time.Time `json:"decidedAt"`

	// Evidence references what the decision was based on, e.g. specific
	// ChangeImpact entries, ticket links, or an external review document.
	Evidence []string `json:"evidence,omitempty"`

	Notes string `json:"notes,omitempty"`
}

// ChangeReview ties an approved Baseline to a proposed Architecture and
// the human Assessment of the change between them — the
// {baseline, proposed, assessment} triple a release references.
// Assessment is nil until a human records a decision; a ChangeReview is a
// legitimate, useful value before that (its ChangeSet/Impact can already
// be computed and circulated for review).
type ChangeReview struct {
	ID         string           `json:"id"`
	Baseline   Baseline         `json:"baseline"`
	Proposed   sas.Architecture `json:"proposed"`
	Assessment *Assessment      `json:"assessment,omitempty"`
}

// ChangeSet computes the diff between this review's baseline and
// proposed architecture.
func (r ChangeReview) ChangeSet() ChangeSet {
	return Diff(&r.Baseline.Architecture, &r.Proposed)
}

// Impact computes the semantic classification of this review's change.
func (r ChangeReview) Impact() ChangeImpact {
	proposed := r.Proposed
	return Classify(r.ChangeSet(), &proposed)
}
