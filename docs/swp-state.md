# SWP-STATE (Draft v0.1)

## 1. Scope

SWP-STATE defines immutable shared state transfer using content-addressable blobs with optional parent DAG references.

**Profile ID:** `17`

## 2. Message model

Supported message classes:

- `STATE_PUT`
- `STATE_GET`
- `STATE_DELTA` (optional)
- `STATE_ERR`

## 3. msg_type assignments

- `1`: `STATE_PUT`
- `2`: `STATE_GET`
- `3`: `STATE_DELTA`
- `4`: `STATE_ERR`

Any other `msg_type` is invalid for this profile version.

Payload encoding for this profile MUST support P1:

- `docs/profile-payload-encoding-p1.md`
- Normative schema annex: `proto/swp_state.proto`

P1 opaque-bytes fields for this profile:

- `StatePut.state_id`
- `StatePut.blob`
- `StatePut.parent_ids`
- `StatePut.metadata`
- `StateGet.state_id`
- `StateDelta.state_id`
- `StateDelta.delta`
- `StateDelta.parent_ids`


## 4. State invariants

- `state_id` SHOULD be content-addressable hash.
- parent references form a DAG; referenced parents MUST exist when declared.
- delta payload format is opaque for this profile version.

## 5. Behavior

- `STATE_PUT` stores immutable state blob + metadata.
- `STATE_GET` fetches blob by identifier.
- `STATE_DELTA` may carry opaque delta update content.

## 6. Conformance requirements

A conforming implementation MUST:

- enforce parent reference validity.
- enforce hash/id consistency rules.
- reject unsupported `msg_type` values.
- return deterministic missing-state behavior.

See vectors in `conformance/vectors/catalog.md` under `state_*`.
