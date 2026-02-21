# Key Decisions

This file records scope and architecture decisions for SWP.

## D-001: Federated Route

Decision:
- design for cross-organization federation, not only internal mesh use.

Reason:
- federation constraints force standard-worthy rigor (identity, replay boundaries, conformance).

Implication:
- security binding and conformance requirements are first-class artifacts.

## D-002: Slim Core, Rich Profiles

Decision:
- keep Core limited to framing, envelope, dispatch, and baseline errors.

Reason:
- small core is easier to standardize and keep stable.

Implication:
- all business semantics stay in profile specs.

## D-003: MCP Mapping Is Interop-First

Decision:
- preserve MCP semantics by carrying JSON-RPC payloads as opaque bytes in mapping mode.

Reason:
- minimizes semantic drift and avoids accidental incompatibility.

Implication:
- bridge/adapters may parse JSON, but transport can relay without reinterpretation.

## D-004: A2A Naming and Scope

Decision:
- use A2A as the active agent-to-agent profile name.

Reason:
- aligns with current ecosystem terminology and avoids ACP/A2A ambiguity.

Implication:
- docs and registry use `A2A` consistently.

## D-005: PAKE Deferred

Decision:
- do not include PAKE in current baseline.

Reason:
- keep initial security surface minimal and focused on channel binding requirements.

Implication:
- onboarding/bootstrap mechanisms are future binding work, not Core.

## D-006: SWP Family Profiles Included

Decision:
- adopt the extended SWP profile family (`profile_id` 10-19) in this spec kit.

Reason:
- broadens interoperability beyond MCP/A2A into discovery, transport primitives, data movement, and governance signals.

Implication:
- conformance, publication, and traceability artifacts must include the SWP family alongside MCP/A2A.

## D-007: Conformance Artifacts Are Versioned

Decision:
- keep generated conformance artifacts under `artifacts/conformance/` as versioned project artifacts.

Reason:
- publication and reviewer workflows need reproducible JSON/log outputs tied to tagged releases.

Implication:
- CI and release packaging (`make conformance-pack`) are first-class, and artifact changes are expected when vector behavior changes.
