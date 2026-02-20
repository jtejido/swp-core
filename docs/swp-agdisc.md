# SWP-AGDISC (Draft v0.1)

## 1. Scope

SWP-AGDISC defines agent discovery using an Agent Card retrieval model with cache-aware behavior.

**Profile ID:** `10`

## 2. Message model

Supported message classes:

- `AGDISC_GET`
- `AGDISC_DOC`
- `AGDISC_NOT_MODIFIED` (optional)
- `AGDISC_ERR`

## 3. msg_type assignments

- `1`: `AGDISC_GET`
- `2`: `AGDISC_DOC`
- `3`: `AGDISC_NOT_MODIFIED`
- `4`: `AGDISC_ERR`

Any other `msg_type` is invalid for this profile version.

Payload encoding for this profile MUST support P1:

- `docs/profile-payload-encoding-p1.md`
- Normative schema annex: `proto/swp_agdisc.proto`

P1 opaque-bytes fields for this profile:

- `AgdiscDoc.card_payload`


## 4. Agent Card requirements

An Agent Card MUST declare:

- schema/profile revision
- stable agent identifier
- endpoint set
- capability declarations

If card payload is opaque bytes, bytes MUST be preserved end-to-end.

## 5. Caching semantics

- Implementations SHOULD support cache validators (for example ETag semantics).
- Cache freshness policy SHOULD be explicit (for example max-age equivalent).
- `AGDISC_NOT_MODIFIED` indicates cached representation remains valid.

## 6. Error model

`AGDISC_ERR` SHOULD include deterministic reasons (not found, invalid card, unauthorized, internal).

## 7. Conformance requirements

A conforming implementation MUST:

- reject unsupported `msg_type` values.
- enforce Agent Card revision presence.
- preserve opaque card bytes when using opaque mode.
- implement deterministic not-found and invalid-card handling.

See vectors in `conformance/vectors/catalog.md` under `agdisc_*`.
