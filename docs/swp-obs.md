# SWP-OBS (Draft v0.1)

## 1. Scope

SWP-OBS defines observability context propagation. This profile can be used directly or paired with binding-level
context transport.

**Profile ID:** `18`

## 2. Message model

Supported message classes:

- `OBS_SET`
- `OBS_GET`
- `OBS_DOC`
- `OBS_ERR`

## 3. msg_type assignments

- `1`: `OBS_SET`
- `2`: `OBS_GET`
- `3`: `OBS_DOC`
- `4`: `OBS_ERR`

Any other `msg_type` is invalid for this profile version.

Payload encoding for this profile MUST support P1:

- `docs/profile-payload-encoding-p1.md`
- Normative schema annex: `proto/swp_obs.proto`

P1 opaque-bytes fields for this profile:

- `ObsSet.msg_id`
- `ObsSet.task_id`
- `ObsSet.rpc_id`
- `ObsDoc.msg_id`
- `ObsDoc.task_id`
- `ObsDoc.rpc_id`


## 4. Context requirements

Trace context fields:

- `traceparent` (required when trace context is present)
- `tracestate` (optional)

Invariants:

- `traceparent` format MUST be preserved and validated.
- unknown `tracestate` entries MUST NOT be mutated.
- correlation SHOULD include one or more of `msg_id`, `task_id`, `rpc_id`.

## 5. Behavior

- `OBS_SET` attaches/propagates context.
- `OBS_GET` requests current context snapshot.
- `OBS_DOC` returns context document.

## 6. Conformance requirements

A conforming implementation MUST:

- preserve valid trace context exactly.
- reject invalid traceparent format.
- avoid mutation of unknown tracestate entries.
- reject unsupported `msg_type` values.

See vectors in `conformance/vectors/catalog.md` under `obs_*`.
