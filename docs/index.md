# Systems Architecture Spec

A statically-typed-friendly, Go-first specification for describing software systems as
semantic graphs: nodes, typed relationships, boundaries, identities, entitlements, and
protocol bindings.

**Architecture is a typed graph with properties and constraints; diagrams are views over
that graph.**

## Why Systems Architecture Spec?

Architecture today lives in draw.io/Lucidchart files that preserve geometry and labels
but lose the semantics of what a node *is*, what an edge *means*, who owns it, what
trust zone it sits in, and what policy governs it. SAS makes those facts machine-checkable:

- **Nodes and relationships** — typed, not shapes and arrows. A relationship carries
  transport, operations, identity, entitlements, data classification, and boundary
  crossings, not just a label.
- **Boundaries as objects** — a node commonly belongs to an account, a region, a trust
  zone, and a compliance boundary at once; "crossing" is computed from membership, not
  a single nesting hierarchy.
- **Views, not diagram files** — one canonical model produces development, deployment,
  security, and threat-model projections without semantic drift.
- **Profiles, not levels** — semantics are optional in the core and required only by
  the profile that needs them, so a minimal architecture file stays minimal.

The primary responsibility of this project is the semantic model itself. Diagrams,
threat modeling, launch readiness, and change analysis are **use-case requirements**:
they prove the model works in practice, but they are consumers of the spec, not the
spec. See the [Specification](spec.md) for the full normative model.

## Status: v0.2

The core semantic model, validation engine, three renderers, technology catalogs,
PIDL protocol bindings, the Threat Model Spec bridge, and assurance coverage reporting
are implemented and dogfooded end to end on real PlexusOne portfolio systems.

v0.2 adds a semantic **diff engine** (typed `ChangeSet` with change-impact
classification and a `Baseline`/`Assessment`/`ChangeReview` model), a **FedRAMP
change-assessment profile** (three-outcome classification with control-mapping
hooks), and **container/sandbox** node and boundary kinds for K8s-native and
sandboxed-execution topology.

## Quick Example

```go
import "github.com/plexusone/systemspec-architecture/sas"

arch := sas.Architecture{
    Version: "0.1",
    Nodes: []sas.Node{
        {ID: "web", Kind: sas.NodeKindService, Name: "Storefront"},
        {ID: "db", Kind: sas.NodeKindDataDatabase, Name: "Orders Database"},
    },
    Relationships: []sas.Relationship{
        {ID: "web-to-db", From: "web", To: "db", Kind: sas.RelationKindDataAccess},
    },
}
```

```sh
sas validate <architecture.json> --profile security
sas view <architecture.json> --view context --format d2
```

`sas validate --profile security` is the launch-readiness gate: run it before a
system with internet-facing traffic goes to production.

## Integrations

SAS binds to specialized specs by reference rather than absorbing them:

- [PIDL](https://github.com/grokify/pidl) for protocol choreography
- [Threat Model Spec](https://github.com/grokify/threat-model-spec) for risk reasoning
- [Multi-Agent Spec](https://github.com/plexusone/multi-agent-spec) for agent definitions

## Next Steps

- Read the [Specification](spec.md) for the full data model.
- See the [PRD](specs/initiatives/INIT-SYSTEMSARCHITECTURESPEC-001/PRD.md) for the
  problem this project solves and who it's for.
- Browse the [v0.1.0 release notes](releases/v0.1.0.md) for what shipped.
