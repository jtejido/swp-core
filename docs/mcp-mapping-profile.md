
# MCP Mapping Profile (Draft v0.1)

## 1. Scope

This document defines a SWP profile that carries MCP-compatible JSON-RPC messages.
The objective is interoperability with minimal semantic drift.

This profile does not replace MCP-native transport expectations; it defines a bridge/mapping profile over SWP.

## 2. Payload and encoding model

Envelope `payload` contains raw JSON-RPC bytes in UTF-8 encoding.

Senders:
- MUST emit UTF-8 JSON bytes.
- MUST preserve JSON-RPC semantics.

Relays:
- MAY forward payload bytes without parsing.
- MUST preserve payload bytes exactly.

Gateways that generate messages:
- MUST produce valid JSON-RPC message structures.

## 3. msg_type assignments

- `1`: JSON-RPC Request
- `2`: JSON-RPC Response
- `3`: JSON-RPC Notification

Any other `msg_type` value is invalid for this profile version.

## 4. JSON-RPC structure requirements

For generated messages:
- Request (`msg_type=1`) MUST include `id`, `method`, and `jsonrpc`.
- Response (`msg_type=2`) MUST include `id` and exactly one of `result` or `error`.
- Notification (`msg_type=3`) MUST include `method` and MUST NOT require a response.

For pass-through relay mode:
- payload bytes may be unparsed by the relay.
- semantic validation is expected at endpoints/gateways that originate or consume messages.

Batch JSON-RPC arrays are out of scope for this profile version.

## 5. Correlation model

Two correlation layers exist:

- SWP layer: `msg_id`
- JSON-RPC layer: `id`

Rules:
- Requests (`msg_type=1`) MUST use a fresh `msg_id` per in-flight interaction on a connection.
- Responses (`msg_type=2`) MUST reuse the originating request `msg_id`.
- Gateways MUST preserve JSON-RPC `id` values unchanged.
- Notifications (`msg_type=3`) SHOULD use unique `msg_id`, and MUST NOT require response correlation.

## 6. Error behavior and deterministic mapping

Core errors remain SWP concerns. When a gateway must surface them as JSON-RPC errors, deterministic mapping SHOULD be used:

- `INVALID_FRAME` -> `-32700` (parse error)
- `UNSUPPORTED_VERSION` -> `-32600` (invalid request)
- `UNKNOWN_PROFILE` -> `-32601` (method not found)
- `INVALID_ENVELOPE` -> `-32600` (invalid request)
- `INTERNAL_ERROR` -> `-32603` (internal error)

MCP/tool-level failures MUST remain JSON-RPC `error` payloads without reinterpretation by SWP Core.

## 7. Optional streaming behavior

If incremental output is supported:
- partial outputs SHOULD be emitted as JSON-RPC notifications (`msg_type=3`).
- partial output notifications SHOULD include request correlation in JSON payload fields defined by the MCP runtime.
- terminal completion MUST be represented by a final response (`msg_type=2`) for request/response flows.
- if execution fails after partial output, terminal response SHOULD contain JSON-RPC `error`.

## 8. Security and policy notes

- this profile inherits channel and identity requirements from S1 when S1 is selected.
- authorization decisions SHOULD bind both surfaced channel identity and requested MCP method/tool.

## 9. Conformance requirements

A conforming implementation MUST:
- pass required MCP vectors in `conformance/vectors/catalog.md`.
- preserve payload bytes exactly in relay mode.
- implement deterministic `msg_id` request/response correlation.
