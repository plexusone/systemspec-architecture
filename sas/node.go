package sas

// NodeKind is the small, closed core vocabulary for what a node is.
// Concrete vendor products are never new NodeKind values — they are
// expressed via Technology (kind "compute.function" + provider "aws" +
// service "lambda"), so generic analysis ("find all databases") works
// across clouds without knowing every vendor product.
type NodeKind string

const (
	NodeKindComputeFunction NodeKind = "compute.function"
	NodeKindComputeInstance NodeKind = "compute.instance"
	// NodeKindContainer is a containerized workload — an OCI container or
	// a Kubernetes Pod. Vendor/orchestrator specifics (e.g. k8s Pod) live
	// in Technology, not in the kind.
	NodeKindContainer NodeKind = "compute.container"
	// NodeKindSandbox is an isolated sandboxed execution environment — a
	// CaaS/code-interpreter sandbox or microVM in which untrusted or
	// agent-generated code runs under strong isolation.
	NodeKindSandbox         NodeKind = "compute.sandbox"
	NodeKindDataDatabase    NodeKind = "data.database"
	NodeKindMessagingQueue  NodeKind = "messaging.queue"
	NodeKindGateway         NodeKind = "gateway"
	NodeKindService         NodeKind = "service"
	NodeKindAgent           NodeKind = "agent"
	NodeKindMCPServer       NodeKind = "mcp_server"
	NodeKindExternalService NodeKind = "external_service"
	NodeKindActor           NodeKind = "actor"
)

// Criticality names a node's operational importance tier.
type Criticality string

const (
	CriticalityTier0 Criticality = "tier0"
	CriticalityTier1 Criticality = "tier1"
	CriticalityTier2 Criticality = "tier2"
)

// Node is a system, service, component, or actor in the architecture
// graph. What makes it a specific vendor product is Technology, not Kind;
// what makes it a specific instance is Identity; what makes it governed is
// the Boundaries it belongs to.
type Node struct {
	// ID is the stable identity referenced by relationships and views.
	// Required.
	ID string `json:"id"`

	// Kind classifies what this node is, from the small core vocabulary.
	// Required.
	Kind NodeKind `json:"kind"`

	// Name is the human-readable name. Required.
	Name string `json:"name"`

	// Owner identifies who is responsible for this node, e.g. a team
	// name. Required by the sre profile for tier0/tier1 nodes.
	Owner string `json:"owner,omitempty"`

	// Technology names the concrete vendor product implementing this
	// node. Required by the deployment profile.
	Technology *Technology `json:"technology,omitempty"`

	// Identity describes who or what this node runs as.
	Identity *Identity `json:"identity,omitempty"`

	// Boundaries lists the IDs of every Boundary this node is a member
	// of. Membership is many-to-many by design: a single node commonly
	// belongs to an account, a region, a VPC, a trust zone, and a
	// compliance boundary at once — there is no single hierarchy.
	Boundaries []string `json:"boundaries,omitempty"`

	// Criticality is this node's operational importance tier. Required
	// by the sre profile for nodes that need an availability target.
	Criticality Criticality `json:"criticality,omitempty"`

	// Assurance references evidence that this node exists and works as
	// declared.
	Assurance *Assurance `json:"assurance,omitempty"`

	// Refs points at detail owned by other specifications: an OpenAPI
	// document, a Multi-Agent Spec definition, a Threat Model Spec
	// entry, a Terraform resource.
	Refs []ExternalRef `json:"refs,omitempty"`

	// Extensions carries namespaced, typed data for concerns (security,
	// SRE, compliance, agent) that do not belong in the core shape.
	Extensions *Extensions `json:"extensions,omitempty"`
}
