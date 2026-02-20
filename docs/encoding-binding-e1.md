# Encoding Binding E1: Varint+Bytes (Normative)

## 1. Scope

This document defines **E1**, a mandatory-to-implement envelope encoding binding for SWP Core v1.
E1 provides compact binary encoding with deterministic parsing and a minimal extension mechanism.

E1 uses:

- frame prefix: 32-bit big-endian length `N`
- envelope body: ordered fields encoded using `uvarint` and `bytes`

## 2. Primitive encodings

### 2.1 `uvarint`

`uvarint` uses unsigned LEB128:

- lower 7 bits per octet carry payload
- MSB indicates continuation
- least-significant group first

Receivers MUST reject `uvarint` values that:

- exceed 10 octets
- overflow 64-bit unsigned range
- are truncated before termination

### 2.2 `bytes`

`bytes` is encoded as:

1. `uvarint` length
2. `length` octets

## 3. Envelope v1 encoding

Within frame body (`N` octets), Envelope v1 SHALL be encoded in this exact order:

1. `version`: `uvarint` (MUST be `1`)
2. `profile_id`: `uvarint`
3. `msg_type`: `uvarint`
4. `flags`: `uvarint`
5. `ts_unix_ms`: `uvarint` (`0` MAY be used when sender timestamp is unavailable)
6. `msg_id`: `bytes`
7. `extensions`: `bytes` (TLV block; MAY be empty)
8. `payload`: `bytes`

### 3.1 Extensions block

`extensions` contains TLV entries:

- `ext_type`: `uvarint`
- `ext_val`: `bytes`

Rules:

- receivers MUST ignore unknown `ext_type` values
- extension types `0-15` are reserved for core/bindings
- extension types `16+` are profile-defined

## 4. Limits and rejection rules

Implementations MUST enforce:

- `MAX_FRAME_BYTES`
- `MAX_MSG_ID_BYTES` (recommended: 32)
- `MAX_EXT_BYTES` (recommended: 4096)
- `MAX_PAYLOAD_BYTES` (<= `MAX_FRAME_BYTES` minus overhead)

Receivers MUST reject a frame if:

- `version != 1`
- `msg_id` length is `0` or exceeds `MAX_MSG_ID_BYTES`
- `extensions` length exceeds `MAX_EXT_BYTES`
- `payload` length exceeds `MAX_PAYLOAD_BYTES`
- parse is truncated or contains malformed varints

These rejections SHOULD map to canonical codes in `docs/error-codes.md`:

- malformed/overflow/truncated varints -> `ERR_INVALID_UVARINT`
- extension-size violations -> `ERR_EXT_TOO_LARGE`
- payload-size violations -> `ERR_PAYLOAD_TOO_LARGE`
- other decode truncation/format failures -> `ERR_INVALID_FRAME` or `ERR_INVALID_ENVELOPE`

Implementations SHOULD parse incrementally and MUST NOT perform unbounded allocations based on untrusted length data.

## 5. Interoperability notes

E1 is mandatory-to-implement for Core v1 conformance.
Additional encoding bindings MAY be defined, but E1 support is required for baseline interoperability.
