# Publication Artifact Index

Use this index when preparing a standards submission package.

## 1. Scope and terminology

- Scope and core boundaries: `docs/core-spec.md`
- Canonical terms: `docs/glossary.md`
- Key scope/architecture decisions: `docs/decisions.md`
- ISO/IEEE-style consolidated draft: `docs/swp-standard-draft-iso-ieee.md`

## 2. Protocol specifications

- Core protocol definition: `docs/core-spec.md`
- Core envelope encoding binding (MTI): `docs/encoding-binding-e1.md`
- Canonical conformance error taxonomy: `docs/error-codes.md`
- Profile payload encoding binding (P1): `docs/profile-payload-encoding-p1.md`
- Normative P1 schema annexes:
  - `proto/a2a_p1.proto`
  - `proto/swp_agdisc.proto`
  - `proto/swp_tooldisc.proto`
  - `proto/swp_rpc.proto`
  - `proto/swp_events.proto`
  - `proto/swp_artifact.proto`
  - `proto/swp_cred.proto`
  - `proto/swp_policyhint.proto`
  - `proto/swp_state.proto`
  - `proto/swp_obs.proto`
  - `proto/swp_relay.proto`
- Profile registry and allocation policy: `docs/profile-registry.md`
- MCP mapping profile: `docs/mcp-mapping-profile.md`
- A2A profile: `docs/a2a-profile.md`
- SWP-AGDISC profile: `docs/swp-agdisc.md`
- SWP-TOOLDISC profile: `docs/swp-tooldisc.md`
- SWP-RPC profile: `docs/swp-rpc.md`
- SWP-EVENTS profile: `docs/swp-events.md`
- SWP-ARTIFACT profile: `docs/swp-artifact.md`
- SWP-CRED profile: `docs/swp-cred.md`
- SWP-POLICYHINT profile: `docs/swp-policyhint.md`
- SWP-STATE profile: `docs/swp-state.md`
- SWP-OBS profile: `docs/swp-obs.md`
- SWP-RELAY profile: `docs/swp-relay.md`

## 3. Security package

- Security binding definitions: `docs/security-bindings.md`
- Threat model and mitigations: `docs/security-considerations.md`
- S1 example annex (TLS/mTLS mapping): `docs/s1-annex-tls-mtls.md`
- Optional HTTP/2 transport binding: `docs/transport-binding-h2.md`

## 4. Conformance package

- Conformance model and required categories: `docs/conformance.md`
- Spec runner JSON schema and invariants: `docs/spec-vector-runner-output.md`
- Conformance classes: `docs/conformance-classes.md`
- Conformance claims checklist: `docs/conformance-claims-checklist.md`
- Core normative traceability matrix: `docs/normative-trace-core.md`
- Profile normative traceability matrix: `docs/normative-trace-profiles.md`
- Vector catalog: `conformance/vectors/catalog.md`
- Vector format guidance: `conformance/vectors/README.md`

## 5. Compatibility and release governance

- Versioning and compatibility rules: `docs/versioning-compatibility.md`
- Governance and registry operations: `docs/governance.md`
- POC delivery sequencing and phase plan: `docs/poc-roadmap.md`

## 6. Diagram package

- SWP family architecture: `puml/architecture_family.puml`
- SWP family end-to-end flow: `puml/flow_end_to_end.puml`
- SWP sequence diagrams: `puml/*_sequence.puml`
