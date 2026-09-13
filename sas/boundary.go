package sas

// BoundaryKind classifies what kind of boundary a Boundary represents.
type BoundaryKind string

const (
	BoundaryKindNetwork      BoundaryKind = "network"
	BoundaryKindTrust        BoundaryKind = "trust"
	BoundaryKindEnvironment  BoundaryKind = "environment"
	BoundaryKindAccount      BoundaryKind = "account"
	BoundaryKindRegion       BoundaryKind = "region"
	BoundaryKindCompliance   BoundaryKind = "compliance"
	BoundaryKindOrganization BoundaryKind = "organization"
	// BoundaryKindSandbox is the isolation boundary around a sandboxed
	// execution environment (a CaaS/code-interpreter sandbox or microVM),
	// distinct from a generic network boundary because escaping it is a
	// distinct, higher-severity event.
	BoundaryKindSandbox BoundaryKind = "sandbox"
	// BoundaryKindContainer is the isolation boundary around a container
	// or Kubernetes Pod, distinct from the host and from the cluster
	// network it sits in.
	BoundaryKindContainer BoundaryKind = "container"
)

// ComplianceBoundary carries control-mapping detail for a Boundary that
// represents (or overlaps with) a compliance scope such as a FedRAMP
// authorization boundary.
type ComplianceBoundary struct {
	// FedRAMPBoundary marks that this boundary is (or is part of) a
	// FedRAMP authorization boundary.
	FedRAMPBoundary bool `json:"fedrampBoundary,omitempty"`

	// Controls lists the control IDs mapped to this boundary, e.g.
	// "SC-7", "AC-3".
	Controls []string `json:"controls,omitempty"`
}

// Boundary is an object, not a rectangle drawn around nodes: an AWS
// account, a VPC, a trust zone, a FedRAMP authorization boundary, and a
// PCI scope can all contain the same node simultaneously, so a resource
// belongs to Boundaries by ID membership rather than by nesting in a
// single hierarchy. Whether a relationship "crosses" a boundary is
// computable from Node.Boundaries membership: exactly one endpoint being
// a member means the edge crosses it.
type Boundary struct {
	// ID is the stable identity referenced by Node.Boundaries and
	// Relationship.CrossesBoundaries. Required.
	ID string `json:"id"`

	// Kind classifies what kind of boundary this is. Required.
	Kind BoundaryKind `json:"kind"`

	// Name is the human-readable name. Required.
	Name string `json:"name"`

	// Attributes carries simple key-value facts about the boundary, e.g.
	// {"vpc": "prod", "classification": "restricted"}.
	Attributes map[string]string `json:"attributes,omitempty"`

	// Compliance carries control-mapping detail when this boundary
	// represents a compliance scope.
	Compliance *ComplianceBoundary `json:"compliance,omitempty"`

	// Extensions carries namespaced, typed data for concerns that do not
	// belong in the core shape.
	Extensions *Extensions `json:"extensions,omitempty"`
}
