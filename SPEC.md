# Systems Architecture Spec Specification v0.1

This document is the normative specification of the Systems Architecture
Spec (SAS) data model, describing what is true of the `sas` package's
types and the JSON documents they marshal to and from.

SAS describes a software system as a typed graph — **nodes**, directed
**relationships**, and **boundaries** — plus **views** (queries over that
graph) and **protocol bindings** (instantiations of externally-defined
choreography). Diagrams, threat models, launch-readiness checks, and
change reports are all computed from this graph; none of them are the
graph. The primary responsibility of this specification is the semantic
model itself.

Every type in this specification conforms to the PlexusOne static-type
profile: no `oneOf`, `anyOf`, `allOf`, or other JSON Schema composition,
and no untyped or heterogeneous maps. Polymorphism is expressed with
explicit, named, optional fields (see §6, Extensions) rather than schema
composition, so the canonical Go model, its generated JSON Schema, and its
generated Zod/TypeScript types remain structurally isomorphic.

## 1. Architecture Document

An **Architecture** is the top-level SAS document.

```text
version        SAS document schema version, e.g. "0.1". Required.
metadata       descriptive facts about the system (name, description, owner)
nodes          every system, service, component, and actor. Required (may be empty).
relationships  every directed edge between nodes
boundaries     every network, trust, environment, account, region,
               compliance, and organization boundary
views          named queries over this architecture
bindings       protocol bindings instantiating PIDL protocol definitions
```

## 2. Nodes

A **Node** is a system, service, component, or actor in the architecture
graph.

```text
id            stable identity referenced by relationships, boundaries, and views. Required.
kind          classification from the closed core vocabulary (§2.1). Required.
name          human-readable name. Required.
owner         who is responsible for this node, e.g. a team name
technology    the concrete vendor product implementing this node (§2.2)
identity      who or what this node runs as (§5)
boundaries    IDs of every Boundary this node is a member of (§4)
criticality   operational importance tier: tier0 | tier1 | tier2
assurance     evidence references (§9)
refs          pointers to detail owned by other specifications (§6.1)
extensions    namespaced typed data for concerns outside the core shape (§6)
```

### 2.1 Kind taxonomy

`kind` is drawn from a small, closed, dotted vocabulary:

```text
compute.function     compute.instance     compute.container
compute.sandbox      data.database        messaging.queue
gateway              service              agent
mcp_server           external_service     actor
```

`compute.container` is a containerized workload (an OCI container or Kubernetes
Pod); `compute.sandbox` is an isolated sandboxed execution environment (a
CaaS/code-interpreter sandbox or microVM). Both remain generic — a specific
runtime is expressed via `technology`, not a new `kind`.

This vocabulary MUST NOT grow to accommodate specific vendor products.
"AWS Lambda" is never a new `kind` value — it is `kind: "compute.function"`
plus `technology: {provider: "aws", service: "lambda"}` (§2.2). This keeps
generic analysis ("find all databases", "find externally exposed compute")
working identically across clouds and on-premises systems.

### 2.2 Technology

**Technology** names the concrete vendor product implementing a node,
independently of its core `kind`.

```text
provider    vendor or platform, e.g. "aws", "gcp", "k8s". Required.
service     specific product within that provider, e.g. "lambda", "rds"
```

A `provider`/`service` pair that has no entry in a technology catalog
remains valid — catalogs are a display and analysis convenience, not a
closed vocabulary the core schema enforces.

## 3. Relationships

A **Relationship** is a directed edge between two nodes: `from` and `to`
node IDs, plus everything a security or threat-model review needs that a
plain arrow cannot carry.

```text
id                  stable identity. Required.
from                source node ID. Required.
to                  target node ID. Required.
kind                classification from the closed core vocabulary (§3.1). Required.
transport           protocol/port/encryption facts (§3.2)
operations          specific verbs performed, from the generic vocabulary (§3.3)
identity            which identity the caller acts as (§5)
authorization       entitlements exercised (§5)
data                data classifications carried across the wire
crossesBoundaries    asserted boundary crossings (§4.1)
criticalPath        whether this relationship sits on a critical execution path
sync                "sync" | "async"
assurance           evidence references (§9)
protocolRef         a PIDL protocol this relationship participates in (§8)
extensions          namespaced typed data (§6)
```

### 3.1 Kind taxonomy

```text
calls                data.access           data_flow
publishes_to         subscribes_to         authenticates_to
authorizes_via       observes              enforces
depends_on           spawns                delegates_to
```

### 3.2 Transport

```text
protocol              transport-layer protocol, e.g. "tcp", "udp"
port                  network port
applicationProtocol   application-layer protocol, e.g. "https", "postgresql", "grpc"
encryption            "tls" | "mtls" | "none" | ...
```

### 3.3 Operations

`operations` is drawn from a generic, protocol-independent verb
vocabulary:

```text
discover    read    create    update    delete    execute    administer
```

Protocol-specific operations (an HTTP method, a SQL statement, an MCP
action) specialize exactly one of these verbs via catalog-supplied
mappings (e.g. HTTP `GET` → `read`, SQL `SELECT` → `read`, MCP
`tools/call` → `execute`). A security-level view can therefore show
`read` while an implementation-level view shows `HTTPS GET
/v1/orders` for the same underlying fact.

## 4. Boundaries

A **Boundary** is an object, not a rectangle drawn around nodes. An AWS
account, a VPC, a trust zone, a FedRAMP authorization boundary, and a PCI
scope can all contain the same node simultaneously, so membership is
many-to-many by node ID reference (`Node.boundaries`), never a single
nesting hierarchy.

```text
id            stable identity referenced by Node.boundaries and
              Relationship.crossesBoundaries. Required.
kind          classification from the closed core vocabulary (below). Required.
name          human-readable name. Required.
attributes    simple key-value facts, e.g. {"vpc": "prod"}
compliance    control-mapping detail (fedrampBoundary, controls[])
extensions    namespaced typed data (§6)
```

Kind taxonomy:

```text
network    trust    environment    account    region
compliance    organization    sandbox    container
```

`sandbox` is the isolation boundary around a sandboxed execution environment
(a CaaS/code-interpreter sandbox or microVM); `container` is the isolation
boundary around a container or Kubernetes Pod, distinct from the host and the
cluster.

### 4.1 Boundary crossing

A relationship **crosses** a boundary when exactly one of its `from`/`to`
endpoints is a member of it (per `Node.boundaries`). This is a derivable
fact, computed from node membership independently of whether an author
asserted it via `Relationship.crossesBoundaries` — the derived and
asserted sets MAY diverge, and a validator MAY report that divergence, but
the model itself does not require them to agree.

## 5. Identity & Entitlement

**Identity** describes who or what acts as a node, or which identity a
relationship's caller assumes.

```text
type         human | workload | service | agent | device | external. Required.
mechanism    aws_iam_role | spiffe | oauth_client | oidc | aauth |
             k8s_service_account | x509 | api_key
ref          the concrete identity, e.g. a SPIFFE ID or an IAM role ARN
```

An **IdentityRef** (used on `Relationship.identity`) either names a node
whose declared `Identity` applies (`nodeId`) or describes an identity
inline (`identity`).

An **Entitlement** is a subject-action-resource authorization tuple,
shaped to map cleanly onto ReBAC systems (subject→user, action→relation,
resource→object) without SAS itself evaluating or enforcing it:

```text
subject       identity or identity class, e.g. a node ID. Required.
action        the entitled operation, e.g. "orders.insert". Required.
resource      what the action applies to. Required.
conditions    human/agent-readable qualifiers narrowing the grant
```

`Authorization.entitlements` carries the entitlements a relationship
exercises.

## 6. Extensions

**Extensions** carries namespaced, typed data for concerns that do not
belong in the core shape (node/relationship/boundary structure). Each
namespace is an explicit, statically typed struct field — `security`,
`sre`, `compliance`, `agent` in v0.1 — never a generic or dynamic map.
Adding a new extension namespace means adding a new named field and type
to the specification, not widening a dynamic bag: this is the mechanism
by which SAS remains extensible while staying within the static-type
profile.

```text
security      { trustZone }
sre           { availabilityTarget }
compliance    { fedrampBoundary }
agent         { runtime }
```

### 6.1 External references

An **ExternalRef** points from a node or relationship to detail owned by
another specification, so SAS never duplicates that content:

```text
type    openapi | pidl | threat-model-spec | multi-agent-spec |
        terraform | otel. Required.
ref     locator within that specification. Required.
```

## 7. Views

A **View** is a query over an Architecture, not a separate diagram file.

```text
id                    stable identity. Required.
name                  human-readable name
includeKinds          restrict to nodes of these NodeKind values
includeRelations      restrict to relationships of these RelationKind values
includeBoundaries     restrict to these boundary IDs
groupBy               a BoundaryKind to nest selected nodes by (rendering hint)
level                 "generic" | "specific" (rendering hint)
```

`Select(architecture, view)` is a pure function producing the
sub-architecture a view denotes:

1. **Nodes**: every node whose `kind` is in `includeKinds`, or every node
   when `includeKinds` is empty.
2. **Relationships**: every relationship whose `kind` is in
   `includeRelations` (or every relationship, when `includeRelations` is
   empty) **and** whose `from` and `to` both survived step 1.
3. **Boundaries**: every boundary whose ID is in `includeBoundaries`, or —
   when `includeBoundaries` is empty — every boundary referenced by a
   selected node's `boundaries`.

`groupBy` and `level` are rendering hints: they are carried on the View
but do not affect `Select`'s output.

## 8. Protocol Bindings

A **ProtocolBinding** instantiates a PIDL protocol's abstract entities as
concrete nodes in this architecture. SAS never duplicates PIDL's
choreography.

```text
id              stable identity. Required.
protocolRef     "pidl://<protocol-id>", where <protocol-id> matches the
                referenced PIDL document's protocol.id field. Required.
participants    map of PIDL entity ID -> SAS node ID. Required.
```

A binding is not required to assign a participant for every entity the
referenced protocol declares. An unassigned entity is not an error — an
architecture may legitimately not model every protocol participant (e.g.
an ephemeral browser redirect) — but it is a fact tooling MAY surface.

## 9. Assurance

**Assurance** carries evidence references for a node or relationship:
where to find proof that the declared element exists and behaves
correctly. Assurance stores pointers to evidence, never the evidence
itself, and this specification does not define reconciliation against
live evidence.

```text
tests         test suite references, e.g. "test://integration/orders-db"
metrics       observability references, e.g. "otel://service/orders-api"
detections    security monitoring references, e.g. "siem://rule/..."
deployment    deployment records, e.g. "aws://lambda/orders-api"
```

## 10. Profiles

A **profile** declares which optional semantics become mandatory for a
given use case. Requesting a profile does not change what an Architecture
document is structurally allowed to contain — the JSON Schema is
identical regardless of profile — it changes which fields a validator
requires to consider the document conformant to that profile.

| Profile | Requires |
|---|---|
| `development` | Nodes and relationships exist. No additional fields. |
| `deployment` | Every node declares `technology` and belongs to an `account`, `region`, or `network` boundary; every relationship declares `transport`. |
| `security` | Every boundary-crossing relationship (§4.1) declares `transport.encryption` and `identity`; every relationship with a write operation (`create`, `update`, `delete`, `administer`) declares `authorization.entitlements`; every **internet-facing** relationship (§10.1) additionally declares `data` classifications and its target node declares an `owner`. |
| `threat-model` | Every relationship touching an `external_service` node declares `data` classifications. |
| `sre` | Every `tier0`/`tier1` node declares `owner`, an `sre.availabilityTarget` extension, and at least one `assurance.metrics` reference. |

### 10.1 Internet-facing relationships

A relationship is **internet-facing** when its source node's `kind` is
`actor` or `external_service` **and** it crosses a boundary (§4.1) — i.e.
traffic entering the architecture's own trust zone from a human actor or
an unmanaged external system, as distinct from the architecture calling
out to a third-party dependency. The `security` profile's internet-facing
requirements are the go-live gate: a conformant result under `security`
is intended to be the single check a launch process runs before a system
ships to production.

## 11. Threat Model Bridge

An Architecture MAY be exported as a **system-under-analysis** for
`github.com/grokify/threat-model-spec`'s DiagramIR format: `type: "dfd"`,
`elements` (from nodes), `boundaries` (from boundaries), and `flows`
(from relationships). This export carries structural facts only — threat
reasoning (attacks, STRIDE classification, mitigations, detections,
response actions) is not part of this specification and is not produced
by this export; element, boundary, and flow IDs are preserved unchanged
so a threat model can reference SAS IDs directly.

The mapping from `NodeKind`/`BoundaryKind` onto DiagramIR's narrower
`ElementType`/`BoundaryType` enums is lossy in specific, documented cases
(an `actor` node has no DFD equivalent and maps to `external-entity`; a
`trust`, `compliance`, or `organization` boundary has no DFD equivalent
and maps to `network`). `Relationship.data.classifications` (a free-text
list) is not mapped onto DiagramIR's single-value, four-tier
`AssetClassification` enum — there is no honest, general mapping between
the two — and is left for a human to set during threat-model authoring.

## 12. Non-goals (V1)

- Runtime reconciliation (comparing a declared Architecture against
  observed OpenTelemetry traces or cloud inventory). §9's Assurance model
  reserves the evidence-reference fields; no v1 tooling reconciles them.
- Semantic diff and change-impact classification, including FedRAMP
  significant-change assessment. Reserved for a future revision of this
  specification.
- CALM (FINOS) import/export. CALM's extensibility model depends on JSON
  Schema composition (`allOf`/`anyOf`), which this specification's static-
  type profile excludes by design; interop is demand-driven, not part of
  v1.
- A `Flow`/choreography concept duplicating PIDL. `ProtocolBinding` plus
  `Relationship.protocolRef` is the entire integration surface with PIDL;
  ordered protocol steps remain PIDL's responsibility.
- Enforcement. This specification and its reference implementation
  validate and report; they do not block, proxy, or intercept traffic.

## Prior art

The node/relationship/boundary model draws its "typed graph, not a
drawing" philosophy from architecture-as-code efforts generally (C4,
Structurizr, FINOS CALM) while deliberately diverging from CALM's
JSON-Schema-composition-based extensibility to keep the canonical model
representable as ordinary structs in Go, Rust, Swift, Java, and C# without
custom union-type machinery. The identity/entitlement model is shaped to
map cleanly onto ReBAC systems (OpenFGA, SpiceDB) for future authorization
backends without making this specification a ReBAC engine itself. Protocol
choreography is deliberately delegated to `github.com/grokify/pidl` rather
than reinvented; threat-model detail is deliberately delegated to
`github.com/grokify/threat-model-spec`.
