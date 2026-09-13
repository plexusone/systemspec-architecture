package assure

import "github.com/plexusone/systemspec-architecture/sas"

// ElementKind distinguishes a node from a relationship in a Gap.
type ElementKind string

const (
	ElementKindNode         ElementKind = "node"
	ElementKindRelationship ElementKind = "relationship"
)

// EvidenceCategory names one of the four kinds of assurance evidence a
// sas.Assurance can reference.
type EvidenceCategory string

const (
	EvidenceTests      EvidenceCategory = "tests"
	EvidenceMetrics    EvidenceCategory = "metrics"
	EvidenceDetections EvidenceCategory = "detections"
	EvidenceDeployment EvidenceCategory = "deployment"
)

// Gap identifies one element (node or relationship) and which evidence
// categories it declares no references for.
type Gap struct {
	Kind    ElementKind        `json:"kind"`
	ID      string             `json:"id"`
	Missing []EvidenceCategory `json:"missing"`
}

// Counts tracks, for one element type (nodes or relationships), the
// total count and how many declare at least one reference in each
// evidence category.
type Counts struct {
	Total      int `json:"total"`
	Tests      int `json:"tests"`
	Metrics    int `json:"metrics"`
	Detections int `json:"detections"`
	Deployment int `json:"deployment"`
}

// Coverage returns c's fraction (0.0-1.0) of elements with at least one
// reference in category. Returns 0 when there are no elements of this
// type, rather than dividing by zero.
func (c Counts) Coverage(category EvidenceCategory) float64 {
	if c.Total == 0 {
		return 0
	}
	var covered int
	switch category {
	case EvidenceTests:
		covered = c.Tests
	case EvidenceMetrics:
		covered = c.Metrics
	case EvidenceDetections:
		covered = c.Detections
	case EvidenceDeployment:
		covered = c.Deployment
	}
	return float64(covered) / float64(c.Total)
}

// Report is the result of running Assure over an architecture.
type Report struct {
	Nodes         Counts `json:"nodes"`
	Relationships Counts `json:"relationships"`
	Gaps          []Gap  `json:"gaps,omitempty"`
}

// Assure computes assurance-reference coverage for every node and
// relationship in arch.
func Assure(arch *sas.Architecture) Report {
	var report Report

	for _, n := range arch.Nodes {
		report.Nodes.Total++
		missing := tally(&report.Nodes, n.Assurance)
		if len(missing) > 0 {
			report.Gaps = append(report.Gaps, Gap{Kind: ElementKindNode, ID: n.ID, Missing: missing})
		}
	}

	for _, r := range arch.Relationships {
		report.Relationships.Total++
		missing := tally(&report.Relationships, r.Assurance)
		if len(missing) > 0 {
			report.Gaps = append(report.Gaps, Gap{Kind: ElementKindRelationship, ID: r.ID, Missing: missing})
		}
	}

	return report
}

// tally increments c's per-category counts for a's populated categories
// and returns the categories that were empty or absent.
func tally(c *Counts, a *sas.Assurance) []EvidenceCategory {
	var missing []EvidenceCategory

	if a != nil && len(a.Tests) > 0 {
		c.Tests++
	} else {
		missing = append(missing, EvidenceTests)
	}
	if a != nil && len(a.Metrics) > 0 {
		c.Metrics++
	} else {
		missing = append(missing, EvidenceMetrics)
	}
	if a != nil && len(a.Detections) > 0 {
		c.Detections++
	} else {
		missing = append(missing, EvidenceDetections)
	}
	if a != nil && len(a.Deployment) > 0 {
		c.Deployment++
	} else {
		missing = append(missing, EvidenceDeployment)
	}

	return missing
}
