# SWP-CRED (Draft v0.1)

## 1. Scope

SWP-CRED defines opaque credential and delegation-chain carriage for federated operations.

**Profile ID:** `15`

## 2. Message model

Supported message classes:

- `CRED_PRESENT`
- `CRED_DELEGATE`
- `CRED_REVOKE` (optional)
- `CRED_ERR`

## 3. msg_type assignments

- `1`: `CRED_PRESENT`
- `2`: `CRED_DELEGATE`
- `3`: `CRED_REVOKE`
- `4`: `CRED_ERR`

Any other `msg_type` is invalid for this profile version.

Payload encoding for this profile MUST support P1:

- `docs/profile-payload-encoding-p1.md`
- Normative schema annex: `proto/swp_cred.proto`

P1 opaque-bytes fields for this profile:

- `CredPresent.credential`
- `CredPresent.chain_id`
- `CredDelegate.chain_id`
- `CredDelegate.delegation`
- `CredRevoke.chain_id`


## 4. Credential invariants

- credential bytes are opaque and tagged by `cred_type`.
- delegation chains MUST enforce max length.
- delegation entries MUST include bounded expiry.
- replay/expiry checks MUST be binding-aware and deterministic.

## 5. Behavior

- `CRED_PRESENT` provides credential material/context.
- `CRED_DELEGATE` extends delegation chain under policy.
- `CRED_REVOKE` signals invalidation intent when supported.

## 6. Conformance requirements

A conforming implementation MUST:

- enforce chain length and expiry limits.
- reject unsupported credential types per policy.
- reject unsupported `msg_type` values.
- provide deterministic invalid-credential behavior.

See vectors in `conformance/vectors/catalog.md` under `cred_*`.
