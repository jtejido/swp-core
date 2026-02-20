# SWP-EVENTS (Draft v0.1)

## 1. Scope

SWP-EVENTS defines structured event transport with bounded required fields and deterministic correlation behavior.

**Profile ID:** `13`

## 2. Message model

Supported message classes:

- `EVT_PUBLISH`
- `EVT_SUBSCRIBE` (optional)
- `EVT_UNSUBSCRIBE` (optional)
- `EVT_BATCH` (optional)
- `EVT_ERR`

## 3. msg_type assignments

- `1`: `EVT_PUBLISH`
- `2`: `EVT_SUBSCRIBE`
- `3`: `EVT_UNSUBSCRIBE`
- `4`: `EVT_BATCH`
- `5`: `EVT_ERR`

Any other `msg_type` is invalid for this profile version.

Payload encoding for this profile MUST support P1:

- `docs/profile-payload-encoding-p1.md`
- Normative schema annex: `proto/swp_events.proto`

P1 opaque-bytes fields for this profile:

- `EventRecord.msg_id`
- `EventRecord.task_id`
- `EventRecord.rpc_id`
- `EventRecord.body`


## 4. Event record requirements

A published event MUST contain:

- `event_id`
- `event_type`
- `severity`
- `ts_unix_ms`
- at least one correlation key (`msg_id`, `task_id`, or `rpc_id`)
- `body` (opaque bytes)

Required fields SHOULD remain at 7 or fewer for portability.

## 5. Ordering and correlation

- Ordering MUST be defined per stream by sender and honored by receiver.
- `event_id` MUST be unique per producer scope.
- Correlation keys MUST be preserved end-to-end when present.

## 6. Invariants

- `severity` MUST use a finite implementation-documented enumeration.
- `EVT_BATCH` ordering MUST preserve internal event order.
- Unknown event extensions MUST NOT invalidate known required fields.

## 7. Error model

`EVT_ERR` SHOULD be used for subscription/auth/filter failures.

## 8. Conformance requirements

A conforming implementation MUST:

- enforce required event fields.
- enforce correlation presence.
- preserve per-stream order.
- reject unsupported `msg_type` values.

See vectors in `conformance/vectors/catalog.md` under `events_*`.
