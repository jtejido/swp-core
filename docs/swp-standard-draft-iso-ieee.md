# SWP Protocol Family Standard Draft (ISO/IEEE Style)

## 1 Scope

This document specifies the SWP protocol family for profile-based interoperability over a slim binary substrate.
The family includes:

- SWP Core (framing, envelope, dispatch, baseline errors)
- MCP Mapping Profile (model/tool JSON-RPC mapping)
- A2A Profile (agent-to-agent task lifecycle)
- SWP extended profile family (AGDISC, TOOLDISC, RPC, EVENTS, ARTIFACT, CRED, POLICYHINT, STATE, OBS, RELAY)
- S1 Security Binding (abstract authenticated confidential channel requirements)

This document does not define:

- application IAM or policy language
- orchestration framework behavior
- message-level cryptography profiles (reserved for future bindings)

## 2 Normative References

- RFC 2119, Key words for use in RFCs to Indicate Requirement Levels
- RFC 8174, Ambiguity of Uppercase vs Lowercase in RFC 2119 Key Words
- JSON-RPC 2.0 Specification

## 3 Terms and Definitions

- Frame: One length-delimited stream unit containing one encoded envelope.
- Envelope: Core metadata + opaque profile payload.
- Profile: A payload semantics specification selected by `profile_id`.
- Binding: Transport/security requirement set used with Core.
- `msg_id`: Message-level correlation identifier.
- `task_id`: A2A task-lifecycle identifier.

Canonical term set: `docs/glossary.md`.

## 4 Symbols and Abbreviated Terms

- MCP: Model Context Protocol
- A2A: Agent-to-Agent
- S1: SWP Security Binding 1
- `MAX_FRAME_BYTES`, `MAX_PAYLOAD_BYTES`, `MAX_CLOCK_SKEW_MS`: deployment limits

## 5 Conformance Language

The key words MUST, MUST NOT, REQUIRED, SHALL, SHOULD, SHOULD NOT, MAY in this document are interpreted as described in RFC 2119 and RFC 8174.

## 6 Protocol Architecture

SWP is layered:

1. Core
2. Encoding binding(s)
3. Payload encoding binding(s)
4. Profile(s)
5. Binding(s)

Core remains minimal and stable; profile and binding semantics evolve independently under compatibility policy.

## 7 SWP Core Specification

### 7.1 Framing

A stream SHALL be encoded as:

1. 32-bit unsigned big-endian length prefix `N`
2. `N` bytes of encoded envelope

Receivers MUST reject invalid frames including unreadable prefix, zero-length prefix, oversized prefix, truncated body, or undecodable envelope.

### 7.2 Envelope

Required envelope fields:

- `version`
- `profile_id`
- `msg_type`
- `msg_id`
- `flags`
- `ts_unix_ms`
- `payload`

Unknown flags MUST NOT change interpretation of known fields.

### 7.2A Envelope Encoding Binding

Core v1 implementations MUST support E1 envelope encoding:

- `docs/encoding-binding-e1.md`

Additional envelope encodings MAY be supported but do not replace E1 requirement.

### 7.2B Profile Payload Encoding Binding

Profiles that do not define an alternative payload encoding MUST support P1:

- `docs/profile-payload-encoding-p1.md`

P1 uses Protocol Buffers (proto3) wire format with profile-specific normative schema annexes.

### 7.3 Validation and Dispatch

Receivers MUST validate:

- supported `version`
- known/handled `profile_id`
- `msg_id` length bounds
- payload size bounds
- timestamp freshness when policy is enabled

Dispatch MUST be based on `profile_id`.

### 7.4 Core Error Registry

Core error statuses:

- `OK`
- `INVALID_FRAME`
- `UNSUPPORTED_VERSION`
- `UNKNOWN_PROFILE`
- `INVALID_ENVELOPE`
- `INTERNAL_ERROR`

## 8 Profile Registry

Profile ID allocation and governance are specified in `docs/profile-registry.md`.
Initial assignments:

- `1` -> MCP Mapping Profile
- `2` -> A2A Profile
- `3-9` -> reserved foundational range
- `10` -> SWP-AGDISC
- `11` -> SWP-TOOLDISC
- `12` -> SWP-RPC
- `13` -> SWP-EVENTS
- `14` -> SWP-ARTIFACT
- `15` -> SWP-CRED
- `16` -> SWP-POLICYHINT
- `17` -> SWP-STATE
- `18` -> SWP-OBS
- `19` -> SWP-RELAY

## 9 S1 Security Binding

S1 specifies an abstract authenticated confidential channel requirement set:

- confidentiality
- integrity
- peer authentication
- replay resistance
- downgrade resistance

Endpoints MUST NOT process frames until channel security is established.
Endpoints MUST fail closed on authentication/integrity failures.
Endpoints MUST surface a stable peer identity to profile authorization logic.

Threat model and mitigations: `docs/security-considerations.md`.
Example mapping (non-normative): `docs/s1-annex-tls-mtls.md`.
Optional HTTP/2 transport binding (informative): `docs/transport-binding-h2.md`.

## 10 MCP Mapping Profile (`profile_id=1`)

### 10.1 Message Model

Payload contains UTF-8 JSON-RPC bytes.
Relay mode MUST preserve payload bytes exactly.

`msg_type` assignments:

- `1` request
- `2` response
- `3` notification

Unsupported `msg_type` values are invalid.

### 10.2 Correlation

- Request/response correlation at SWP layer uses `msg_id`.
- Response MUST reuse originating request `msg_id`.
- JSON-RPC `id` MUST be preserved.
- Notifications MUST NOT require responses.

### 10.3 Error Mapping (Gateway Mode)

Deterministic core-to-JSON-RPC mappings are defined in `docs/mcp-mapping-profile.md`.

## 11 A2A Profile (`profile_id=2`)

### 11.1 Message Model

`msg_type` assignments:

- `1` Handshake
- `2` Task
- `3` Event
- `4` Result

Unsupported `msg_type` values are invalid.

### 11.1A Payload encoding

A2A payloads MUST support P1 (`docs/profile-payload-encoding-p1.md`).

### 11.2 Lifecycle

- Task creates lifecycle for `task_id`.
- Event/Result MUST reference existing `task_id`.
- Event before Task for same `task_id` is invalid.
- Result is terminal.
- Post-terminal Event/Result is invalid.

### 11.3 Idempotency

- Equivalent duplicate Task SHOULD be idempotent.
- Conflicting duplicate Task MUST be rejected.
- Equivalent duplicate terminal Result SHOULD be ignored.
- Conflicting duplicate terminal Result MUST be rejected.

## 12 Conformance

Conformance categories:

- Core
- E1 envelope encoding
- Profile (MCP, A2A, and/or SWP family profiles)
- S1 (when binding claim is made)

Required artifacts:

- Vector catalog: `conformance/vectors/catalog.md`
- Claims checklist: `docs/conformance-claims-checklist.md`
- Conformance classes: `docs/conformance-classes.md`
- Core normative trace: `docs/normative-trace-core.md`
- Profile normative trace: `docs/normative-trace-profiles.md`

## 12.1 Extended profile family references

- AGDISC (`profile_id=10`): `docs/swp-agdisc.md`
- TOOLDISC (`profile_id=11`): `docs/swp-tooldisc.md`
- RPC (`profile_id=12`): `docs/swp-rpc.md`
- EVENTS (`profile_id=13`): `docs/swp-events.md`
- ARTIFACT (`profile_id=14`): `docs/swp-artifact.md`
- CRED (`profile_id=15`): `docs/swp-cred.md`
- POLICYHINT (`profile_id=16`): `docs/swp-policyhint.md`
- STATE (`profile_id=17`): `docs/swp-state.md`
- OBS (`profile_id=18`): `docs/swp-obs.md`
- RELAY (`profile_id=19`): `docs/swp-relay.md`

Unless explicitly overridden by a profile, SWP and A2A payloads use P1:

- `docs/profile-payload-encoding-p1.md`

## 13 Versioning and Compatibility

Compatibility policy is normative per `docs/versioning-compatibility.md`.

Key rules:

- backward-incompatible Core changes require Core version increment
- profile breaking changes require major revision or new `profile_id`
- registry lifecycle controls profile ID governance

Governance details are described in `docs/governance.md`.

## Annex A (Informative): Document Map

- Publication artifact index: `docs/publication-artifact-index.md`
- POC roadmap: `docs/poc-roadmap.md`
