
# SWP Core Specification (Draft v0.1)

## 1. Scope

This document specifies **SWP Core**, a minimal, generic framing and envelope format for exchanging
application messages on a reliable byte stream. SWP Core is designed to carry multiple **profiles**
(e.g., tool invocation, agent-to-agent messaging) without embedding profile semantics in the core.

SWP Core defines:
- stream framing
- the SWP Envelope
- the mandatory-to-implement envelope encoding binding reference
- core validation invariants
- a minimal error/status model
- mandatory parsing limits for robustness
- the profile identification mechanism

SWP Core does **not** define:
- profile payload semantics (defined in profile documents)
- application authentication/authorization models
- service discovery, routing, or orchestration frameworks
- any particular transport protocol (HTTP/2, QUIC, etc. are out of scope)

## 2. Terminology

- **Frame**: One length-delimited unit on the stream, containing a single Envelope.
- **Envelope**: Core header fields plus an opaque profile payload.
- **Profile**: A separately specified payload format identified by a stable Profile ID.
- **Binding**: A separately specified set of requirements that binds SWP to a transport and/or security properties.

Canonical repository-wide terms are listed in `docs/glossary.md`.

The terms MUST, SHOULD, MAY are to be interpreted as described in RFC 2119 / RFC 8174.

## 3. Framing (Stream)

A SWP stream SHALL be a sequence of Frames. Each Frame SHALL be encoded as:

- A 32-bit unsigned integer **N** in network byte order (big-endian), representing the number of octets that follow.
- Exactly **N** octets representing the encoded Envelope.

A receiver MUST treat any of the following as `ERR_INVALID_FRAME`:
- length prefix not fully readable
- N equal to zero
- N larger than `MAX_FRAME_BYTES`
- fewer than N trailing octets available
- envelope payload that cannot be decoded by the envelope codec

### 3.1 Parsing limits

An implementation MUST enforce a maximum allowed frame size **MAX_FRAME_BYTES**.
If a received N exceeds MAX_FRAME_BYTES, the implementation MUST reject the frame and SHOULD terminate the stream.

An implementation SHOULD impose a limit on the number of frames processed per unit time to mitigate resource exhaustion.

### 3.2 Recommended defaults (non-normative)

- MAX_FRAME_BYTES: 8 MiB
- MAX_PAYLOAD_BYTES: MAX_FRAME_BYTES minus envelope overhead
- MIN_MSG_ID_BYTES: 8
- MAX_MSG_ID_BYTES: 64
- MAX_CLOCK_SKEW_MS: 300000 (5 minutes)

### 3.3 Envelope encoding binding

SWP Core defines logical envelope fields. Wire-level interoperability requires a concrete octet encoding.

Core v1 conformance therefore requires support for the E1 encoding binding:

- `docs/encoding-binding-e1.md`

Implementations MAY support additional encodings, but MUST support E1.

### 3.4 Cross-document normative obligations

The following obligations are REQUIRED for Core conformance:

- Receivers MUST reject malformed E1 `uvarint` values (overflow, >10 octets, or truncation). See `docs/encoding-binding-e1.md` section 2.1.
- Receivers MUST reject E1 decode truncation for required envelope fields and length-delimited fields. See `docs/encoding-binding-e1.md` sections 2 and 4.
- Receivers MUST reject unsupported Core versions and MUST NOT perform implicit downgrade to older Core behavior. See `docs/versioning-compatibility.md` section 3.
- If SWP-RPC is implemented, endpoints MUST enforce single terminal lifecycle outcome per `rpc_id`, and MUST reject conflicting idempotency-key replays. See `docs/swp-rpc.md` section 5.
- If SWP-RELAY is implemented, endpoints MUST enforce at-least-once delivery with deterministic `delivery_id` deduplication semantics. See `docs/swp-relay.md` section 4.
- Frames MUST NOT be processed on non-loopback interfaces unless an authenticated confidential channel binding is active. See `docs/security-bindings.md` section 2.

## 4. Envelope

The Envelope comprises a small fixed set of fields required for routing, correlation, and bounded validity.

### 4.1 Required fields

An Envelope MUST include:

- **version**: protocol version of SWP Core
- **profile_id**: stable identifier of the profile that defines the payload format
- **msg_type**: profile-defined message type discriminator
- **msg_id**: message correlation identifier (opaque bytes)
- **flags**: bitfield for core-level behavior and future extensibility
- **ts_unix_ms**: sender timestamp (milliseconds since Unix epoch)
- **payload**: opaque bytes, interpreted according to profile_id

### 4.2 Optional fields

If present, optional fields MUST NOT alter the meaning of required fields.
Optional fields SHOULD be specified via future extensions to preserve backward compatibility.

### 4.3 Flags behavior

This version does not define any required core flag bits.

Receivers:
- MUST ignore unknown flag bits unless a profile explicitly requires them.
- MUST NOT reinterpret known fields based only on unknown flag bits.
- SHOULD surface unknown flag bits to observability.

### 4.4 Validation invariants

A receiver MUST validate:

1. **version** is supported.
2. **profile_id** is known or handled as "unknown profile" (see error model).
3. **msg_id** length is within [MIN_MSG_ID_BYTES, MAX_MSG_ID_BYTES]. (Recommended 16 bytes.)
4. **ts_unix_ms** is within a receiver-defined skew window if the deployment requires replay protections.
5. **payload** length is within [0, MAX_PAYLOAD_BYTES], where MAX_PAYLOAD_BYTES <= MAX_FRAME_BYTES - overhead.

If an invariant fails, the receiver MUST treat the message as invalid and respond/terminate per binding requirements.

Senders SHOULD ensure `msg_id` uniqueness across concurrently outstanding interactions on a connection.

## 5. Error / Status model

SWP Core uses canonical conformance error codes defined in:

- `docs/error-codes.md`

Core-level rejects MUST map to canonical `ERR_*` categories.
Runtime implementations MAY expose additional or legacy aliases, but conformance artifacts MUST provide canonical mappings.

The encoding of errors remains profile-defined unless a binding mandates a core error frame.

If the binding supports error responses:
- framing failures SHOULD map to `ERR_INVALID_FRAME` (or `ERR_FRAME_TOO_LARGE` when size-bound specific).
- unsupported core version SHOULD map to `ERR_UNSUPPORTED_VERSION`.
- unknown profile SHOULD map to `ERR_UNKNOWN_PROFILE`.
- envelope invariant failures SHOULD map to `ERR_INVALID_ENVELOPE`.

## 6. Profile dispatch

A receiver MUST dispatch frames based on **profile_id**.
If profile_id is unknown, the receiver MUST handle per binding (e.g., close stream, return error, or ignore).

For interoperable behavior, implementations SHOULD return `UNKNOWN_PROFILE` when response paths exist.

## 7. Core constants and registries

- Core version for this document: `1`.
- Core error registry: defined in Section 5.
- Flag registry: no assigned bits in this version.
- MTI envelope encoding binding: E1 (`docs/encoding-binding-e1.md`).

Profile IDs are governed by `docs/profile-registry.md`.

## 8. Extensibility

Future versions MUST preserve backward compatibility by:
- not reinterpreting existing required fields
- introducing new fields as optional extensions
- using version increments when backward incompatible changes are introduced

## 9. Security considerations (core)

SWP Core does not prescribe security mechanisms. Deployments that require confidentiality, integrity, and peer
authentication MUST select an appropriate **security binding** (see `docs/security-bindings.md`).

Threat model and mitigation guidance is documented in `docs/security-considerations.md`.

SWP Core responsibilities are limited to protocol-layer determinism and parser/resource safety.
Authorization policy, prompt-injection handling, tool sandboxing, and supply-chain controls are implementation responsibilities outside Core scope.
