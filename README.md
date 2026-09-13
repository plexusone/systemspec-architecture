# Systems Architecture Spec (SAS)

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![Docs][docs-mkdoc-svg]][docs-mkdoc-url]
[![Visualization][viz-svg]][viz-url]
[![License][license-svg]][license-url]

 [go-ci-svg]: https://github.com/plexusone/systemspec-architecture/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/plexusone/systemspec-architecture/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/plexusone/systemspec-architecture/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/plexusone/systemspec-architecture/actions/workflows/go-lint.yaml
 [go-sast-svg]: https://github.com/plexusone/systemspec-architecture/actions/workflows/go-sast-codeql.yaml/badge.svg?branch=main
 [go-sast-url]: https://github.com/plexusone/systemspec-architecture/actions/workflows/go-sast-codeql.yaml
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/plexusone/systemspec-architecture
 [docs-godoc-url]: https://pkg.go.dev/github.com/plexusone/systemspec-architecture
 [docs-mkdoc-svg]: https://img.shields.io/badge/docs-guide-blue.svg
 [docs-mkdoc-url]: https://plexusone.dev/systems-architecture-spec
 [viz-svg]: https://img.shields.io/badge/repo-visualization-blue.svg
 [viz-url]: https://mango-dune-07a8b7110.1.azurestaticapps.net/?repo=plexusone%2Fsystems-architecture-spec
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/plexusone/systemspec-architecture/blob/main/LICENSE

A statically-typed-friendly, Go-first specification for describing software systems as
semantic graphs: nodes, typed relationships, boundaries, identities, entitlements, and
protocol bindings.

**Architecture is a typed graph with properties and constraints; diagrams are views over
that graph.**

## Primary responsibility

The primary responsibility of this project is the semantic model itself — a spec whose
typed graph carries enough meaning that machines can reason over an architecture without
human interpretation. Diagrams, threat modeling, launch readiness, and change analysis
are **use-case requirements**: they determine what semantics the core must express and
prove the model works in practice, but they are consumers of the spec, not the spec.

If a use case needs something the model can't express, the model changes. If only one
use case wants something, it belongs in a consumer, a profile, or a namespaced extension
— never in the core.

## Status

**v0.1.** The core semantic model, validation engine, three renderers, technology
catalogs, PIDL protocol bindings, the Threat Model Spec bridge, and assurance
coverage reporting are implemented and dogfooded end to end. See
[`SPEC.md`](SPEC.md) for the normative specification and
[`docs/specs/initiatives/INIT-SYSTEMSARCHITECTURESPEC-001/`](docs/specs/initiatives/INIT-SYSTEMSARCHITECTURESPEC-001/)
for the PRD, TRD, PLAN, and ROADMAP. Semantic diff and change-impact analysis
(Phase 5) are deferred to v0.2; see SPEC.md §12 for the full non-goals list.

## Packages

| Package | Responsibility |
|---|---|
| `sas` | the core semantic graph — Architecture, Node, Relationship, Boundary, Identity, Entitlement, View, ProtocolBinding, Assurance, Extensions |
| `validate` | referential-integrity checks plus profile-conditional rules (`development`, `deployment`, `security`, `threat-model`, `sre`) |
| `render` | shared rendering foundation (node grouping, shape/label mapping) used by every renderer |
| `render/mermaid`, `render/d2`, `render/dot` | view → diagram renderers, each verified against the real `mmdc`/`d2`/`dot` compilers |
| `catalog` | AWS/GCP/Kubernetes technology display data and HTTP/SQL/MCP operation mappings |
| `bridge/threatmodel` | exports an Architecture as the system-under-analysis for [Threat Model Spec](https://github.com/grokify/threat-model-spec), verified against its real JSON Schema |
| `assure` | assurance-reference coverage reporting (tests, metrics, detections, deployment) |
| `cli` | business logic shared by every CLI command |
| `cmd/sas` | thin Cobra adapter over `cli` |
| `schema` | generated, embedded JSON Schema (`//go:embed`), linted with `schemakit --property-case camelCase` |
| `ts` | generated Zod/TypeScript types, conforming to the same fixture corpus as the Go model |

## CLI

```text
sas validate <architecture.json> [--profile development,deployment,security,threat-model,sre] [--format console|json]
sas view <architecture.json> [--view <id> | --group-by --include-kinds --include-relations --include-boundaries] --format mermaid|d2|dot
sas bind <architecture.json> --pidl <protocol.json> [--format console|json]
sas export threat-model <architecture.json>
sas assure <architecture.json> [--format console|json]
```

`sas validate --profile security` is the launch-readiness gate: run it before
a system with internet-facing traffic goes to production (see SPEC.md §10.1).

## Example

[`examples/dogfood/acme-widgets.json`](examples/dogfood/acme-widgets.json) is a
fictional storefront (browser actor, API, orders database, payment provider,
GitHub OAuth login) exercising the full field surface — boundaries, views, a
PIDL binding, and launch-readiness-clean relationships:

```sh
go run ./cmd/sas validate examples/dogfood/acme-widgets.json --profile security
go run ./cmd/sas view examples/dogfood/acme-widgets.json --view context --format d2
```

## Design principles

- **Statically Typed Friendly.** Go structs are the source of truth. JSON Schema is
  generated (never hand-written) and stays within the PlexusOne static-type profile:
  no `oneOf` / `anyOf` / `allOf`, no schema composition. Explicit discriminated fields
  instead of unions. TypeScript/Zod conform downstream. JSON is the interoperability
  contract — it never becomes a second type system.
- **Views are queries, not files.** One canonical model produces development,
  deployment, security, and threat-model projections without semantic drift.
  Renderers (Mermaid, D2, Graphviz DOT) are adapters, never sources of truth.
- **Profiles, not levels.** Semantics are optional in the core and required only by
  the profile that needs them, so a minimal architecture file stays minimal.
- **Core ontology, not a technology catalog.** "AWS Lambda" is not a core type — it's
  `kind: compute.function` + `technology: {provider: aws, service: lambda}`. Concrete
  vendor semantics live in catalogs outside the core schema.
- **Integration layer, not universal schema.** SAS binds to specialized specs by
  reference — [PIDL](https://github.com/grokify/pidl) for protocol choreography,
  [Threat Model Spec](https://github.com/grokify/threat-model-spec) for risk reasoning,
  [Multi-Agent Spec](https://github.com/plexusone/multi-agent-spec) for agent
  definitions — rather than absorbing them.

## License

MIT — see [LICENSE](LICENSE).
