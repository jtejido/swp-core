
# SWP Profile Registry (Draft)

## 1. Purpose

Profiles define the semantics and payload encoding carried inside SWP Envelope.payload. SWP Core
identifies a profile using a stable numeric **profile_id**.

## 2. Profile ID allocation

- profile_id is an unsigned integer.
- The registry MUST avoid re-use of assigned IDs.
- A profile specification MUST define:
  - profile name
  - profile_id
  - profile versioning scheme
  - payload encoding
  - msg_type enumeration for the profile

Recommended allocation policy:
- 0: reserved
- 1-1023: standards-track allocations
- 1024-4095: provisional/experimental allocations
- 4096 and above: private-use allocations

Reserved foundational range:
- 3-9: reserved for future foundational protocol profiles/bindings.
- IDs 3-9 MUST NOT be assigned by provisional/private processes in this version line.

## 3. Suggested initial allocations (non-normative)

- 1: MCP Mapping Profile (SWP.MCPMAP)
- 2: A2A Payload Profile (SWP.A2A)
- 10: SWP-AGDISC
- 11: SWP-TOOLDISC
- 12: SWP-RPC
- 13: SWP-EVENTS
- 14: SWP-ARTIFACT
- 15: SWP-CRED
- 16: SWP-POLICYHINT
- 17: SWP-STATE
- 18: SWP-OBS
- 19: SWP-RELAY

## 4. Versioning

Each profile MUST specify compatibility rules.
SWP Core version changes are independent of profile version changes.

Profile change guidance:
- additive message fields: backward-compatible
- new `msg_type` values: backward-compatible if unknown types are safely rejectable
- semantic reinterpretation of existing fields: backward-incompatible
- removal of required fields: backward-incompatible
