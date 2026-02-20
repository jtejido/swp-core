# SWP-POLICYHINT (Draft v0.1)

## 1. Scope

SWP-POLICYHINT carries portable constraints and violation signals. It is not a full policy language.

**Profile ID:** `16`

## 2. Message model

Supported message classes:

- `POLICY_HINT_SET`
- `POLICY_HINT_ACK`
- `POLICY_VIOLATION`
- `POLICY_ERR`

## 3. msg_type assignments

- `1`: `POLICY_HINT_SET`
- `2`: `POLICY_HINT_ACK`
- `3`: `POLICY_VIOLATION`
- `4`: `POLICY_ERR`

Any other `msg_type` is invalid for this profile version.

Payload encoding for this profile MUST support P1:

- `docs/profile-payload-encoding-p1.md`
- Normative schema annex: `proto/swp_policyhint.proto`

P1 opaque-bytes fields for this profile:

- none


## 4. Constraint model

Each constraint entry includes:

- `key`
- `value`
- `mode` (`MUST`, `SHOULD`, `MAY`)
- `scope_ref` (optional)

Unknown key behavior MUST be explicitly defined by implementation policy.

## 5. Violation reporting

`POLICY_VIOLATION` MUST include:

- violated key
- effective scope reference
- violation reason code

## 6. Conformance requirements

A conforming implementation MUST:

- enforce deterministic unknown-key behavior.
- preserve mode semantics.
- emit structured violation reports.
- reject unsupported `msg_type` values.

See vectors in `conformance/vectors/catalog.md` under `policyhint_*`.
