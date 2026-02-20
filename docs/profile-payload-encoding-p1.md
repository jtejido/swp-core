# Profile Payload Encoding Binding P1 (Draft)

## 1. Scope

This document defines **P1**, a normative payload-encoding binding for SWP profiles that do not define
an alternative payload encoding.

P1 provides a common binary payload encoding for A2A and SWP profile families.

MCP Mapping (`profile_id=1`) is out of scope for P1 because its payload is already defined as raw JSON-RPC bytes.

## 2. Encoding

P1 payload encoding is:

- Protocol Buffers (proto3) wire format

For each P1 profile, payload bytes MUST decode as one of the profile message schemas referenced by that profile.

## 3. Schema requirements

For profiles using P1:

- each profile doc MUST reference a normative `.proto` schema annex.
- required semantic fields in profile docs MUST map to concrete protobuf fields.
- field numbers in published schemas MUST NOT be reused.
- removed fields MUST be reserved in schema evolution.

## 4. Forward compatibility and determinism

- Consumers MUST ignore unknown protobuf fields.
- Producers SHOULD preserve semantic equivalence across schema revisions.
- Producers SHOULD enable deterministic protobuf serialization when payload bytes are used for hashing/signing or byte-level conformance checks.
- Conformance MUST be defined in terms of decoded semantic fields unless a profile explicitly requires byte-exact payload preservation.

## 5. Opaque-bytes semantics

Fields defined as opaque bytes in profile semantics MUST use protobuf `bytes` type under P1.

Each profile doc MUST explicitly list its P1 opaque-bytes fields (message + field names).

## 6. Profile applicability (v0.1)

Profiles that MUST support P1 in this kit:

- A2A (`profile_id=2`)
- SWP-AGDISC (`profile_id=10`)
- SWP-TOOLDISC (`profile_id=11`)
- SWP-RPC (`profile_id=12`)
- SWP-EVENTS (`profile_id=13`)
- SWP-ARTIFACT (`profile_id=14`)
- SWP-CRED (`profile_id=15`)
- SWP-POLICYHINT (`profile_id=16`)
- SWP-STATE (`profile_id=17`)
- SWP-OBS (`profile_id=18`)
- SWP-RELAY (`profile_id=19`)

## 7. Conformance

A profile claim that depends on P1 MUST include:

- successful protobuf decode/validation for declared message types
- rejection behavior for malformed protobuf payloads
- required-field semantic validation results for that profile
- evidence that unknown-field handling follows forward-compatibility rules
