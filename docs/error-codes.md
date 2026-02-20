# SWP Error Code Taxonomy (ERR_*)

This document defines the canonical conformance error taxonomy for SWP.

## 1. Scope

- These codes are normative for conformance expectations and specification text.
- Implementations MAY expose different runtime/internal codes.
- Where runtime codes differ, conformance artifacts SHOULD include both:
  - `expected_error_code` (canonical `ERR_*`)
  - `code` (implementation/runtime alias)

## 2. Canonical codes

Core and E1:
- `ERR_INVALID_FRAME`: invalid/truncated frame prefix/body, malformed frame boundary.
- `ERR_FRAME_TOO_LARGE`: frame length exceeds configured maximum.
- `ERR_INVALID_UVARINT`: malformed, truncated, overflowing, or overlong (`>10` octets) uvarint.
- `ERR_UNSUPPORTED_VERSION`: unsupported Core or binding-required version.
- `ERR_INVALID_ENVELOPE`: envelope invariant violation.
- `ERR_MSG_ID_INVALID`: message identifier missing/invalid length/invalid format.
- `ERR_PAYLOAD_TOO_LARGE`: payload exceeds configured maximum.
- `ERR_EXT_TOO_LARGE`: extension block exceeds configured maximum.
- `ERR_UNKNOWN_PROFILE`: unknown/unsupported profile identifier.

Binding and policy:
- `ERR_SECURITY_POLICY`: security binding or policy-gate rejection.
- `ERR_RATE_LIMIT_EXCEEDED`: rate/quota policy rejection.

Profile and semantic:
- `ERR_UNSUPPORTED_MSG_TYPE`: unsupported profile message type.
- `ERR_INVALID_MCP_PAYLOAD`: invalid MCP mapping payload shape/encoding.
- `ERR_INVALID_PROFILE_PAYLOAD`: invalid profile payload semantics/encoding.
- `ERR_DUPLICATE_MSG_ID`: duplicate in-flight `msg_id` policy violation.
- `ERR_NOT_FOUND`: deterministic not-found outcome where profile defines it.
- `ERR_COMPATIBILITY_POLICY`: process/policy compatibility check failure.

## 3. Legacy alias mapping (informative)

Legacy aliases observed in current vectors/runtime tooling:
- `INVALID_FRAME` -> `ERR_INVALID_FRAME` or `ERR_FRAME_TOO_LARGE` (based on rejection reason)
- `UNSUPPORTED_VERSION` -> `ERR_UNSUPPORTED_VERSION`
- `UNKNOWN_PROFILE` -> `ERR_UNKNOWN_PROFILE`
- `INVALID_ENVELOPE` -> `ERR_INVALID_ENVELOPE`
- `SECURITY_POLICY` -> `ERR_SECURITY_POLICY`
- `UNSUPPORTED_MSG_TYPE` -> `ERR_UNSUPPORTED_MSG_TYPE`
- `INVALID_MCP_PAYLOAD` -> `ERR_INVALID_MCP_PAYLOAD`
- `INVALID_PROFILE_PAYLOAD` -> `ERR_INVALID_PROFILE_PAYLOAD`
- `DUPLICATE_MSG_ID` -> `ERR_DUPLICATE_MSG_ID`
- `RATE_LIMIT_EXCEEDED` -> `ERR_RATE_LIMIT_EXCEEDED`
- `NOT_FOUND` -> `ERR_NOT_FOUND`
- `COMPATIBILITY_POLICY` -> `ERR_COMPATIBILITY_POLICY`

