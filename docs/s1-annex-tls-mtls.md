# Annex: S1 Mapping Example Using TLS/mTLS (Non-Normative)

This annex shows one way to satisfy S1 requirements with TLS/mTLS.
It is an example, not a protocol mandate.

## 1. Channel properties mapping

S1 requirement to TLS/mTLS mapping:

- confidentiality in transit:
  - provided by TLS record encryption.
- integrity protection:
  - provided by TLS authenticated encryption.
- peer authentication:
  - server authentication via server certificate validation.
  - client authentication via client certificate (mTLS mode).
- downgrade resistance:
  - enforce policy for acceptable protocol versions and ciphers; fail closed otherwise.

## 2. Identity surface mapping

Example identity extraction:
- primary identity: certificate subject alternative name (SAN) URI or DNSName.
- optional attributes: issuer, subject, and certificate fingerprint for audit context.

Profile/runtime guidance:
- expose one canonical peer identity string to profile authorization.
- avoid ambiguous identity precedence rules across multiple cert fields.

## 3. Failure behavior mapping

- handshake failure:
  - do not process SWP frames.
  - terminate connection.
- cert validation failure:
  - terminate connection and emit audit log.
- post-handshake integrity/auth failure:
  - terminate connection and emit audit log.

## 4. Replay and freshness

- TLS includes transport-level anti-replay properties.
- if deployment additionally requires message freshness:
  - enforce `ts_unix_ms` with `MAX_CLOCK_SKEW_MS`.
  - reject stale/future frames per policy.

## 5. Operational notes

- rotate server/client certificates periodically.
- support revocation checks according to deployment policy.
- isolate trust anchors between environments and tenants where required.
