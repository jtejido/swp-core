# Security Considerations (Core + S1)

This document captures threat assumptions and baseline mitigations for SWP Core and the S1 binding.

## 1. Scope and assumptions

- SWP Core is transport-agnostic and does not provide security by itself.
- Security properties are supplied by a selected binding, with S1 as the baseline.
- Deployments may be internal, federated cross-org, or internet-exposed; threat intensity differs, controls remain similar.

Responsibility split:
- Wire-substrate scope: framing/parser safety, deterministic decoding, limits, versioning, binding requirements.
- Implementation/system scope: authorization policy, tool sandboxing, prompt-injection defenses, supply-chain controls.

## 2. Threats and mitigations

### 2.1 Replay attacks

Threat:
- attacker replays valid previously observed frames.

Mitigations:
- use S1 channels with anti-replay protections.
- enforce timestamp freshness with `ts_unix_ms` and documented `MAX_CLOCK_SKEW_MS` where policy requires.
- reject stale/future messages outside policy window.

Residual risk:
- replay may still occur within allowed freshness window unless additional nonce/session controls are used.

### 2.2 Downgrade attacks

Threat:
- attacker forces weaker channel parameters or weaker peer-auth mode.

Mitigations:
- S1 requires downgrade resistance and fail-closed negotiation.
- endpoints must not continue when required channel properties are unavailable.

Residual risk:
- operational misconfiguration of policy may allow weak settings.

### 2.3 Identity misbinding

Threat:
- channel-authenticated identity is mapped to wrong application principal/tenant.

Mitigations:
- bind authorization to surfaced S1 peer identity.
- validate tenant and policy context before profile execution.
- log principal-to-identity mapping decisions for auditability.

Residual risk:
- IAM/policy bugs outside protocol scope may still mis-authorize.

### 2.4 Resource exhaustion (DoS)

Threat:
- attacker sends oversized, malformed, or high-rate traffic to exhaust CPU/memory.

Mitigations:
- enforce `MAX_FRAME_BYTES` and `MAX_PAYLOAD_BYTES`.
- reject malformed frames early at parser boundary.
- apply connection and frame-rate limits.

Residual risk:
- volumetric network-level floods require infrastructure-level controls.

### 2.5 Parser exploitation

Threat:
- crafted payloads trigger parser bugs or unexpected behavior.

Mitigations:
- strict decoding rules and invariant checks in Core.
- mandatory negative conformance vectors and malformed cases.
- fuzzing and differential parser testing are recommended.

Residual risk:
- implementation defects remain possible even with conformance testing.

### 2.6 Trust anchor compromise

Threat:
- compromise of channel trust roots or key material allows impersonation.

Mitigations:
- protect trust stores and keys with least privilege.
- rotate credentials and revoke compromised principals.
- audit trust-anchor changes and monitor anomalous identity usage.

Residual risk:
- high-impact compromise still possible without strong ops controls.

## 3. Federation-specific guidance

- document federation trust model and identity namespace boundaries.
- require explicit tenant isolation policy at profile authorization layer.
- require auditable event logging for auth failures, replay rejections, and downgrade failures.

## 4. Out of scope

- detailed IAM/policy language design.
- full PKI governance process.
- message-level crypto (future S2/S3 bindings).
- prompt-injection prevention and tool-output trust decisions.
- local code-execution sandboxing and endpoint hardening implementation details.
