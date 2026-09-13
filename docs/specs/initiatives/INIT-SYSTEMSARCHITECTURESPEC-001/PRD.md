# PRD — Systems Architecture Spec — Machine-Readable System Contract (SAS v0.1)

**Initiative:** `INIT-SYSTEMSARCHITECTURESPEC-001`
**Home repo:** `github.com/plexusone/systemspec-architecture`
**Status:** proposed

## Problem

Systems architecture today lives in draw.io/Lucidchart files that preserve geometry and labels but lose the semantics of what a node *is*, what an edge *means*, who owns it, what trust zone it sits in, and what policy governs it. The same system is re-described independently by every discipline — developers (code + README), architects (diagrams), DevOps (Terraform), SRE (OTel dashboards), security (threat models), SOC (detection rules), compliance (SSP evidence) — and none of these representations can answer machine-checkable questions like "what changed architecturally in this PR?", "does every trust-boundary crossing define authentication?", or "is this a FedRAMP significant change?". As agentic development compresses many specialist roles into one product builder coordinating multiple agents, the absence of a shared, machine-readable system contract becomes the bottleneck: every agent (implementation, threat-model, SRE, compliance) reconstructs the system independently.

## Vision

**Architecture is a typed graph with properties and constraints; diagrams are views over that graph.** Systems Architecture Spec (SAS) is a vendor-neutral, machine-readable semantic model for systems, relationships, boundaries, identities, entitlements, and views — designed to generate architecture diagrams, validate architectural policy, compute semantic change impact, and (eventually) reconcile declared architecture with runtime reality.

SAS is a **software-system contract**, not an architecture description language. Where FINOS CALM asks "how should architecture be described?", SAS asks "what must be known about a software system for architecture to become executable, testable, observable, and continuously verifiable?" The one-line test for every field added to the core: *can this information help generate diagrams, analyze security architecture, determine change impact, or reconcile the declared system with reality?*

**Responsibility hierarchy:** the primary responsibility of this project is the semantic model itself — a spec whose typed graph carries enough meaning that machines can reason over the architecture without human interpretation. Diagrams, threat modeling, launch readiness, and change analysis are **use-case requirements**: they determine what semantics the core must be able to express and prove the model works in practice, but they are consumers of the spec, not the spec. When a use case needs something the model can't express, the model changes; when a use case wants something only it needs, that lives in the consumer, a profile, or an extension — never in the core.

## Users

- **Product builders (primary):** one person responsible for product, engineering, cloud, security, and operations — needs one boring, obvious representation of the whole system that both humans and coding agents can read and write. **The first product builder is us:** the growing PlexusOne portfolio (`plexusone/{agentforge,actionforge,dashforge}`) is the dogfood corpus. These repos already share a consistent structure (specs, `CHANGELOG.json` via structured-changelog) — SAS extends that same specs-as-infrastructure practice to architecture.
- **Systems engineers / software architects:** author the architecture as code, review changes as semantic diffs rather than picture diffs.
- **Security architects:** get first-class identity, entitlement, operation (read vs read+write), and trust-boundary semantics; feed Threat Model Spec without re-describing the system.
- **Compliance owners:** use versioned baselines and semantic diffs as decision support for FedRAMP significant-change assessment (and later SOC 2 / ISO 27001 profiles).
- **AI agents:** consume SAS as shared machine-readable system context; generate and modify architectures within a statically-typed, unambiguous schema.

## Goals

1. **Semantic core graph:** Node, Relationship, Boundary, Interface as typed domain objects. Edges carry protocol, port, operations, identity, entitlements, encryption, data classification, and critical-path — the security-significant facts most diagram formats collapse into an arrow.
2. **Statically Typed Friendly:** Go structs are the source of truth; JSON Schema is generated (invopop/jsonschema, linted by schemakit) and stays within the static-type profile — no `oneOf`/`anyOf`/`allOf`, no schema composition, explicit discriminated fields instead. TypeScript/Zod conform downstream. JSON is the interoperability contract, not a second type system.
3. **Views as queries:** one canonical model produces development, deployment, security, and threat-model projections without semantic drift. Renderers (Mermaid, D2, Graphviz DOT) are adapters, never sources of truth.
4. **Profiles, not levels:** optional-in-core, required-by-profile semantics. A developer can author `frontend → api → db` without IAM data; a security profile validates that boundary crossings define identity and encryption.
5. **Launch readiness and threat modeling:** as portfolio apps go live on the web, SAS supplies the system-under-analysis for Threat Model Spec (nodes, boundary crossings, data flows, identities) and enforces launch-readiness rules (internet-facing endpoints must define authentication, TLS, owner, data classification). PIDL `ProtocolBinding` instantiates proven protocol definitions (OAuth/AAuth — already used in threat-model-spec and aistandardsio/agent-protocols) as concrete systems.
6. **Semantic change analysis (post-v0.1):** the IR must be able to express "this PR adds an external trust-boundary crossing carrying customer data through a new authorization mechanism" — that expressiveness test governs v0.1 design, but the diff engine, baselines, and FedRAMP change-assessment profile (`NO_REVIEW_REQUIRED` / `CHANGE_ASSESSMENT_REQUIRED` / `POTENTIAL_SIGNIFICANT_CHANGE`, decision support only) ship after the v0.1 release.
7. **Integration layer, not universal schema:** SAS binds to specialized specs by reference — PIDL for protocol choreography, Threat Model Spec for risk reasoning, Multi-Agent Spec for agent definitions, OpenAPI for API detail.

## Non-Goals (V1)

- **Runtime reconciliation engines** (OTel trace ↔ declared flow, cloud-inventory drift). The IR reserves reconciliation states (`DECLARED_AND_OBSERVED`, `DECLARED_NOT_OBSERVED`, `OBSERVED_NOT_DECLARED`, `OBSERVED_DIFFERENTLY`, `NOT_OBSERVABLE`) and assurance evidence references, but live reconciliation is a separate follow-on project consuming SAS.
- **Detailed protocol choreography** — PIDL owns ordered interactions; SAS binds, never duplicates.
- **Threat analysis semantics** — Threat Model Spec owns threats/mitigations; SAS supplies the system-under-analysis.
- **Chaos/resilience experiment execution** — semantic failure-mode fields are reserved; planners/runners come later.
- **CALM (out of scope; demand-driven)** — CALM's extensibility model is JSON-Schema-composition-first (`allOf` extensions, `anyOf` in core), violating the static-type profile, so it cannot be the foundation. Import/export adapters are not part of this initiative at all: CALM is not yet widely adopted and is FINOS/finance-domain-centric, so adapters get scoped as new work only if someone actually asks for CALM. The lesson from the spec family: structured-changelog and structured-evaluation earned their place through constant use; interop with external standards follows demand, not speculation.
- **A hundred-type UML/ArchiMate-style metamodel** — extensibility comes from namespaced typed extensions and catalogs, not core-type proliferation.
- **Cost budgets, SOC detection coverage, maturity scoring** — assurance dimensions beyond evidence references are follow-on profiles.

## Key Product Decisions

- **Core ontology vs technology catalogs.** "AWS Lambda" is not a core type; it is `kind: compute.function` + `technology: {provider: aws, service: lambda}`. Generic analysis ("find all databases", "find externally exposed compute") works across clouds; renderers still show familiar icons. Catalogs (AWS, GCP, Azure, Kubernetes, MCP) live outside the core schema.
- **Edges matter more than nodes for security.** `READ ONLY via reporting-role` and `READ+WRITE via application-role` on the same arrow are different relationships from a threat-model perspective; the Relationship type makes that distinction mandatory to express and cheap to author.
- **Boundaries are objects, not rectangles.** A resource belongs to multiple overlapping boundaries (AWS account, VPC, trust zone, FedRAMP authorization boundary, PCI scope) — there is no single hierarchy.
- **Operations vocabulary with mappings.** Generic verbs (`discover/read/create/update/delete/execute/administer`) with per-protocol specializations (HTTP GET → read, SQL SELECT → read, MCP invoke_tool → execute), so a security view shows `READ` while an implementation view shows `HTTPS GET /v1/...` from the same model.
- **Views are queries, not files.** No `security.drawio` alongside `overview.drawio`; a view declares included kinds/relations/boundaries and grouping over the canonical graph.
- **Three-outcome change classification.** Change assessment yields decision support plus evidence for an authorized human decision — deliberately not a boolean `significantChange: true`.
- **Compliance logic lives in profiles, not core.** FedRAMP is a change-assessment profile; the same semantic diff serves SOC 2, ISO 27001, and internal change policy.
- **Repository home: PlexusOne.** SAS is a foundational specification in the PlexusOne "specs family" (alongside Multi-Agent Spec, Design System Spec), following the shared pattern: Go-first, generated schema, statically typed friendly, spec + reference implementation in one repo. CLI name: `sas`.

## Success Criteria

- SAS fully describes at least one horizontal forge app and one vertical web app from the real portfolio and generates correct Mermaid/D2 diagrams for at least three distinct views from one model — replacing hand-drawn diagrams for those repos.
- A live-site launch (vertical web app) uses the SAS → Threat Model Spec bridge for its threat model, and `sas validate` launch-readiness rules pass before go-live.
- A PIDL OAuth protocol definition binds to concrete SAS nodes via `ProtocolBinding` without duplicating choreography in SAS.
- Generated JSON Schema passes `schemakit lint --property-case camelCase` with zero violations; Go ↔ JSON ↔ Zod round-trip fixtures pass in CI.
- Post-v0.1: `sas diff` on a hand-authored before/after pair correctly flags new external dependency, authentication mechanism change, operation expansion (`read` → `read+write`), and new trust-boundary crossing.
