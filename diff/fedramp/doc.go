// Package fedramp is a compliance profile layered on top of the generic
// diff package: it classifies a diff.ChangeImpact into FedRAMP's
// three-outcome change-assessment vocabulary (no review required, change
// assessment required, potential significant change).
//
// This package produces decision support only. It never itself decides
// that a change is significant — that is a recorded human decision
// (diff.Assessment), with this package's Classification as one input to
// that decision, not a replacement for it. Keeping this profile in its
// own package (rather than in diff itself) keeps the generic diff engine
// free of any one compliance framework's vocabulary, the same principle
// that keeps profile-specific rules out of the sas package and in
// validate instead.
package fedramp
