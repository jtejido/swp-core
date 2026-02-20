
# SWP Security Bindings (Draft v0.1)

## 1. Scope

This document defines **security bindings** that specify the required security properties of the channel carrying
SWP frames. SWP Core intentionally does not mandate a specific security protocol.

Bindings may be combined with transport bindings (e.g., HTTP/2, QUIC) but are specified independently.

## 2. S1: Authenticated Confidential Channel (ACC)

S1 is an abstract binding. It defines required security properties and identity exposure requirements without
mandating a specific wire-security protocol.

### 2.1 Required properties

A channel conforming to S1 MUST provide:
- confidentiality of application data in transit
- integrity protection against modification
- peer authentication (client and server) sufficient to establish sender identity for authorization decisions
- replay protection appropriate to the underlying secure channel
- downgrade resistance for negotiated channel parameters

S1 does not prescribe a particular mechanism.

Endpoints MUST NOT accept plaintext SWP traffic on non-loopback interfaces.

### 2.2 Identity surface

An S1-conformant implementation MUST surface a stable peer identity to the profile layer, sufficient to apply
authorization policies.

Minimum identity requirements:
- identity value MUST be stable for the connection lifetime.
- identity value MUST be unforgeable within the chosen channel-security mechanism.
- identity value MUST be available to profile handlers before profile authorization is evaluated.

Identity format is binding-defined (for example, certificate subject or workload identity URI).

### 2.3 Channel failure behavior

- frames MUST NOT be processed until channel authentication and integrity/confidentiality negotiation succeed.
- if channel integrity/authentication fails at any point, the endpoint MUST terminate the channel.
- endpoints SHOULD emit auditable security events on channel auth/integrity failures.
- failures in this section SHOULD map to `ERR_SECURITY_POLICY` for conformance reporting.

### 2.4 Replay and timestamp expectations

If deployment policy enforces timestamp freshness using `ts_unix_ms`:
- receivers MUST define and document `MAX_CLOCK_SKEW_MS`.
- stale messages outside the freshness window MUST be rejected.
- receivers SHOULD log replay-window failures as security events.

If deployment policy does not enforce timestamp freshness, this MUST be explicitly documented.

### 2.5 Downgrade requirements

- endpoints MUST fail closed when required channel properties cannot be negotiated.
- endpoints MUST NOT silently continue with weaker confidentiality or authentication guarantees than deployment policy requires.

## 3. Threat model notes

S1 deployments SHOULD account for:
- resource exhaustion via oversized or bursty traffic
- identity misbinding between channel identity and application principal
- downgrade attempts on channel capabilities
- replay attempts near freshness-window boundaries
- weak trust anchor governance

See `docs/security-considerations.md` for a full threat and mitigation model.

## 4. S2 (Placeholder): Message Integrity

A future binding MAY define message-level signatures for brokered delivery where intermediaries terminate
transport security.

## 5. S3 (Placeholder): Message Confidentiality

A future binding MAY define message-level encryption for end-to-end privacy across relays/brokers.

## 6. Example annexes

- `docs/s1-annex-tls-mtls.md` provides a non-normative mapping example of S1 to TLS/mTLS.
