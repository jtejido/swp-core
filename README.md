
# SWP Spec Kit
[![DOI](https://zenodo.org/badge/1162329148.svg)](https://doi.org/10.5281/zenodo.18708208)

This repository contains the SWP (SlimWire Protocol) specifications, conformance artifacts, and a Go reference implementation for a slim binary wire substrate that can carry:

- an MCP mapping profile (bridge from JSON-RPC ecosystems), and
- an A2A profile (agent-to-agent interoperability),
- plus an extended SWP profile family (discovery, RPC/events/obs, artifact/state, policy/cred/relay).

A Go reference implementation exists under `poc/`.

## Start Here

1. Read `docs/decisions.md`.
2. Read `docs/core-spec.md`.
3. Read `docs/glossary.md` and keep term usage strict.
4. Maintain registry/versioning in `docs/profile-registry.md`.
5. Fill vectors in `conformance/vectors/` and validate against `docs/conformance.md`.
6. Run the reference implementation in `poc/README.md`.

## Repository Contents

- `docs/core-spec.md`: SWP core framing, envelope, invariants, limits.
- `docs/encoding-binding-e1.md`: Mandatory-to-implement envelope wire encoding binding.
- `docs/profile-payload-encoding-p1.md`: Payload encoding binding for A2A + SWP profiles.
- `docs/security-bindings.md`: S1 binding and placeholders for future bindings.
- `docs/security-considerations.md`: Threat model and mitigations for Core + S1.
- `docs/s1-annex-tls-mtls.md`: Non-normative TLS/mTLS example mapping for S1.
- `docs/transport-binding-h2.md`: Optional HTTP/2 transport binding.
- `docs/profile-registry.md`: Profile IDs and versioning policy.
- `docs/mcp-mapping-profile.md`: MCP mapping profile rules.
- `docs/a2a-profile.md`: Minimal A2A profile rules.
- `docs/conformance.md`: Required test coverage and vector format.
- `docs/conformance-classes.md`: Class-based conformance taxonomy (C0-C5).
- `docs/conformance-claims-checklist.md`: Claim criteria and evidence requirements.
- `docs/versioning-compatibility.md`: Compatibility and versioning policy.
- `docs/governance.md`: Registry and governance operations guidance.
- `docs/normative-trace-core.md`: Core MUST/SHOULD traceability to vectors.
- `docs/normative-trace-profiles.md`: Profile MUST/SHOULD traceability to vectors.
- `docs/publication-artifact-index.md`: Standards submission artifact map.
- `docs/swp-standard-draft-iso-ieee.md`: ISO/IEEE-style consolidated draft.
- `docs/swp-*.md`: Extended SWP profile family specs.
- `docs/decisions.md`: Architecture and scope decisions.
- `docs/glossary.md`: Canonical protocol terms.
- `puml/`: Component and sequence diagrams.
- `proto/`: Mixed schema set (non-normative examples + normative P1 annexes for A2A/SWP).
- `conformance/vectors/`: Golden vectors (runtime fixtures are concrete; process vectors use `.evidence.md` artifacts).
- `conformance/vectors/catalog.md`: Comprehensive required vector set.
- `poc/`: Go POC implementation (core + MCP mapping + SWP-RPC + vector tooling).
- `poc/README.md`: POC run/test instructions.
- `podman-compose.yml`: Containerized demo/test workflow.
- `Makefile`: Local and podman targets for build/test/demo.

## Standardization Position

SWP is intentionally narrow: framing, envelope, profile dispatch, and conformance rigor.
Federation value comes from:

- stable profile registry governance,
- strict parser/resource limits,
- explicit correlation and replay boundaries,
- independent implementation conformance.

## Conformance Reporting

For publication-ready reporting, use the convenience targets:

1. Core-only summary and artifacts:

```bash
make conformance-core
```

This writes:
- `artifacts/conformance/core.default.json`
- `artifacts/conformance/core.strict.json`

And prints two paper-friendly one-line summaries:
- `CORE default: total=... passed=... failed=... fallback=... json=artifacts/conformance/core.default.json`
- `CORE strict: total=... passed=... failed=... fallback=... json=artifacts/conformance/core.strict.json`

2. Optional full-suite summary and artifacts:

```bash
make conformance-all
```

This writes:
- `artifacts/conformance/all.default.json`
- `artifacts/conformance/all.strict.json`

3. Appendix-ready 3-line core block:

```bash
make conformance-summary
```

4. Publication bundle (JSON + logs + defining docs):

```bash
make conformance-pack
```

This writes:
- `artifacts/conformance/swp-conformance-bundle.tar.gz`
