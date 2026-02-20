# SWP-RELAY (Draft v0.1)

## 1. Scope

SWP-RELAY defines store-and-forward delivery with at-least-once guarantees, acknowledgements, retries, and dead-letter signaling.

**Profile ID:** `19`

## 2. Message model

Supported message classes:

- `RELAY_PUBLISH`
- `RELAY_ACK`
- `RELAY_NACK` (optional)
- `RELAY_STATUS` (optional)
- `RELAY_ERR`

## 3. msg_type assignments

- `1`: `RELAY_PUBLISH`
- `2`: `RELAY_ACK`
- `3`: `RELAY_NACK`
- `4`: `RELAY_STATUS`
- `5`: `RELAY_ERR`

Any other `msg_type` is invalid for this profile version.

Payload encoding for this profile MUST support P1:

- `docs/profile-payload-encoding-p1.md`
- Normative schema annex: `proto/swp_relay.proto`

P1 opaque-bytes fields for this profile:

- `RelayPublish.delivery_id`
- `RelayPublish.payload`
- `RelayAck.delivery_id`
- `RelayNack.delivery_id`
- `RelayStatus.delivery_id`


## 4. Delivery invariants

- delivery semantics MUST be at-least-once.
- `delivery_id` MUST be present on every `RELAY_PUBLISH`.
- publishers/consumers MUST coordinate dedupe via stable `delivery_id`.
- the same `delivery_id` observed within the dedupe retention window MUST be treated as duplicate delivery.
- dedupe retention window MUST be deployment-defined and documented (recommended default: 24 hours).
- retry policy MUST define backoff/limits and timeout behavior.
- dead-letter reasons MUST use a finite enumerated set.
- missing/empty `delivery_id` MUST be rejected.

## 5. Behavior

- `RELAY_PUBLISH` submits message for relay.
- `RELAY_ACK` confirms delivery processing.
- `RELAY_NACK` signals retryable/non-retryable failure when supported.
- `RELAY_STATUS` exposes relay state when supported.

## 6. Conformance requirements

A conforming implementation MUST:

- enforce at-least-once behavior.
- enforce deterministic dedupe behavior.
- enforce retry policy semantics.
- reject unsupported `msg_type` values.
- reject publish requests with missing/invalid `delivery_id`.

Conformance reporting SHOULD map profile-rejection outcomes to:
- `ERR_UNSUPPORTED_MSG_TYPE`
- `ERR_INVALID_PROFILE_PAYLOAD`
- `ERR_SECURITY_POLICY` (policy-gated rejection)

See vectors in `conformance/vectors/catalog.md` under `relay_*`.
