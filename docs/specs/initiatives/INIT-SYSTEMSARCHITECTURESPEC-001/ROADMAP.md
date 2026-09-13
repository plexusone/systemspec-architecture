# ROADMAP — Systems Architecture Spec — Machine-Readable System Contract (SAS v0.1)

**Initiative:** `INIT-SYSTEMSARCHITECTURESPEC-001`
**Repository:** `github.com/plexusone/systemspec-architecture`

## Phase 1 — Core IR & Schema Pipeline

**Theme:** Lock the typed graph model and the Go → JSON Schema → Zod pipeline before any tooling.

- [ ] `RMI-SYSTEMSARCHITECTURESPEC-001` Core graph types: Architecture, Node (kind taxonomy, technology, owner, criticality), Boundary with multi-membership
- [ ] `RMI-SYSTEMSARCHITECTURESPEC-002` Relationship semantics: transport (protocol/port/encryption), operations vocabulary with generic-verb mappings, data classification, critical path
- [ ] `RMI-SYSTEMSARCHITECTURESPEC-003` Identity and entitlement types: identity type/mechanism, subject-action-resource-conditions entitlements
- [ ] `RMI-SYSTEMSARCHITECTURESPEC-004` Namespaced typed extensions and external reference model (openapi, pidl, threat-model-spec, multi-agent-spec, terraform, otel)
- [ ] `RMI-SYSTEMSARCHITECTURESPEC-005` Schema pipeline: generate JSON Schemas from Go types (invopop/jsonschema), schemakit lint camelCase, go:embed, tools.go guard, Go/JSON/Zod round-trip fixtures
  - Depends on: `RMI-SYSTEMSARCHITECTURESPEC-001`
  - Depends on: `RMI-SYSTEMSARCHITECTURESPEC-002`
  - Depends on: `RMI-SYSTEMSARCHITECTURESPEC-003`
  - Depends on: `RMI-SYSTEMSARCHITECTURESPEC-004`

## Phase 2 — Views, Profiles & Validation

**Theme:** Prove optional-in-core, required-by-profile on the real portfolio, not synthetic fixtures.

- [ ] `RMI-SYSTEMSARCHITECTURESPEC-006` View model: views as queries (includeKinds/includeRelations/includeBoundaries, groupBy, abstraction level)
  - Depends on: `RMI-SYSTEMSARCHITECTURESPEC-005`
- [ ] `RMI-SYSTEMSARCHITECTURESPEC-007` Profile model: development, deployment, security, threat-model, sre profiles declaring required semantics
  - Depends on: `RMI-SYSTEMSARCHITECTURESPEC-005`
- [ ] `RMI-SYSTEMSARCHITECTURESPEC-008` Validation rules engine: referential integrity, tier0 requirements, boundary-crossing authn/encryption rules, profile-conditional severity
  - Depends on: `RMI-SYSTEMSARCHITECTURESPEC-007`
- [ ] `RMI-SYSTEMSARCHITECTURESPEC-009` sas CLI foundation: cobra scaffold, sas validate with profile selection, JSON and console output
  - Depends on: `RMI-SYSTEMSARCHITECTURESPEC-008`
- [ ] `RMI-SYSTEMSARCHITECTURESPEC-010` Fixture corpus and dogfood architectures: valid/invalid examples plus real portfolio systems (horizontal forge app and vertical web app) modeled end to end
  - Depends on: `RMI-SYSTEMSARCHITECTURESPEC-008`

## Phase 3 — Rendering & Catalogs

**Theme:** First payoff — diagrams for the portfolio; one model, many correct projections.

- [ ] `RMI-SYSTEMSARCHITECTURESPEC-011` Mermaid renderer: view-driven flowchart output with boundary grouping
  - Depends on: `RMI-SYSTEMSARCHITECTURESPEC-006`
- [ ] `RMI-SYSTEMSARCHITECTURESPEC-012` D2 renderer: primary generated-diagram target with layout and grouping hints
  - Depends on: `RMI-SYSTEMSARCHITECTURESPEC-006`
- [ ] `RMI-SYSTEMSARCHITECTURESPEC-013` Graphviz DOT renderer for large-graph auto-layout
  - Depends on: `RMI-SYSTEMSARCHITECTURESPEC-006`
- [ ] `RMI-SYSTEMSARCHITECTURESPEC-014` Technology catalogs: AWS, GCP, Kubernetes kind/service mappings, operation specializations, display labels
- [ ] `RMI-SYSTEMSARCHITECTURESPEC-015` sas view command: render any view in any format; three distinct projections of the dogfood architectures verified
  - Depends on: `RMI-SYSTEMSARCHITECTURESPEC-011`
  - Depends on: `RMI-SYSTEMSARCHITECTURESPEC-012`
  - Depends on: `RMI-SYSTEMSARCHITECTURESPEC-013`
  - Depends on: `RMI-SYSTEMSARCHITECTURESPEC-010`

## Phase 4 — Threat Modeling, PIDL & Launch Readiness

**Theme:** Second payoff — going-live security: PIDL bindings, threat-model handoff, launch-readiness checks. Closes v0.1.

- [ ] `RMI-SYSTEMSARCHITECTURESPEC-021` ProtocolBinding: bind PIDL protocol participants to concrete SAS nodes; sas bind resolution checks
  - Depends on: `RMI-SYSTEMSARCHITECTURESPEC-005`
- [ ] `RMI-SYSTEMSARCHITECTURESPEC-026` Threat Model Spec bridge: export system-under-analysis (nodes, boundary crossings, data flows, identities) consumable by threat-model-spec
  - Depends on: `RMI-SYSTEMSARCHITECTURESPEC-005`
- [ ] `RMI-SYSTEMSARCHITECTURESPEC-027` Launch-readiness validation rules: internet-facing exposure checks (public endpoints require authn, TLS, owner, data classification)
  - Depends on: `RMI-SYSTEMSARCHITECTURESPEC-008`
- [ ] `RMI-SYSTEMSARCHITECTURESPEC-022` Assurance references and coverage: per-node/edge test/metric/detection/deployment refs, sas assure coverage report
  - Depends on: `RMI-SYSTEMSARCHITECTURESPEC-008`
- [ ] `RMI-SYSTEMSARCHITECTURESPEC-028` Launch dogfood: model and threat-model the first live web properties end to end (vertical apps going to production)
  - Depends on: `RMI-SYSTEMSARCHITECTURESPEC-015`
  - Depends on: `RMI-SYSTEMSARCHITECTURESPEC-026`
  - Depends on: `RMI-SYSTEMSARCHITECTURESPEC-027`
- [ ] `RMI-SYSTEMSARCHITECTURESPEC-025` v0.1 release: SPEC.md normative document, README with statically-typed-friendly positioning, CI (test/lint/schema-lint/conformance), tag
  - Depends on: `RMI-SYSTEMSARCHITECTURESPEC-015`
  - Depends on: `RMI-SYSTEMSARCHITECTURESPEC-021`
  - Depends on: `RMI-SYSTEMSARCHITECTURESPEC-022`
  - Depends on: `RMI-SYSTEMSARCHITECTURESPEC-028`

## Phase 5 — Semantic Diff & Change Impact

**Theme:** Architecture change as typed, classifiable data — post-v0.1 (v0.2 candidate).

- [ ] `RMI-SYSTEMSARCHITECTURESPEC-016` Semantic diff engine: typed ChangeSet with node/edge/boundary and field-level deltas
  - Depends on: `RMI-SYSTEMSARCHITECTURESPEC-005`
- [ ] `RMI-SYSTEMSARCHITECTURESPEC-017` Change impact classification: new external dependency, authn change, entitlement expansion, new boundary crossing, data classification change
  - Depends on: `RMI-SYSTEMSARCHITECTURESPEC-016`
- [ ] `RMI-SYSTEMSARCHITECTURESPEC-018` Baseline and assessment model: approved immutable baselines, proposed versions, recorded human decisions with evidence
  - Depends on: `RMI-SYSTEMSARCHITECTURESPEC-016`
- [ ] `RMI-SYSTEMSARCHITECTURESPEC-019` FedRAMP change-assessment profile: three-outcome classification with control-mapping hooks and decision-support language
  - Depends on: `RMI-SYSTEMSARCHITECTURESPEC-017`
  - Depends on: `RMI-SYSTEMSARCHITECTURESPEC-018`
- [ ] `RMI-SYSTEMSARCHITECTURESPEC-020` sas diff command: human and JSON output, CI-gate exit codes, four canonical scenarios verified on fixtures
  - Depends on: `RMI-SYSTEMSARCHITECTURESPEC-019`

Out of scope for this initiative: CALM interop adapters are demand-driven — CALM is not yet widely adopted and is FINOS/finance-domain-centric, so adapters will be scoped as new RMIs only if and when someone asks for CALM.
