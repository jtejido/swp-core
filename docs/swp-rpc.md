# SWP-RPC (Draft v0.1)

## 1. Scope

SWP-RPC defines a minimal generic RPC profile over SWP for request/response interactions, deterministic errors,
optional streaming items, and cancellation.

**Profile ID:** `12`

## 2. Message model

Supported message classes:

- `RPC_REQ`
- `RPC_RESP`
- `RPC_ERR`
- `RPC_STREAM_ITEM` (optional)
- `RPC_CANCEL` (optional)

## 3. msg_type assignments

- `1`: `RPC_REQ`
- `2`: `RPC_RESP`
- `3`: `RPC_ERR`
- `4`: `RPC_STREAM_ITEM`
- `5`: `RPC_CANCEL`

Any other `msg_type` is invalid for this profile version.

Payload encoding for this profile MUST support P1:

- `docs/profile-payload-encoding-p1.md`
- Normative schema annex: `proto/swp_rpc.proto`

P1 opaque-bytes fields for this profile:

- `RpcReq.rpc_id`
- `RpcReq.params`
- `RpcResp.rpc_id`
- `RpcResp.result`
- `RpcErr.rpc_id`
- `RpcStreamItem.rpc_id`
- `RpcStreamItem.item`
- `RpcCancel.rpc_id`


## 4. Required fields

### `RPC_REQ`

- `rpc_id` (required)
- `method` (required)
- `params` (opaque bytes, optional)
- `idempotency_key` (optional but recommended for retriable methods)

### `RPC_RESP`

- `rpc_id` (required)
- `result` (opaque bytes)

### `RPC_ERR`

- `rpc_id` (required)
- `error_code` (required)
- `retryable` (required)
- `error_message` (required)

### `RPC_STREAM_ITEM`

- `rpc_id` (required)
- `seq_no` (required)
- `item` (opaque bytes, required)
- `is_terminal` (required)

### `RPC_CANCEL`

- `rpc_id` (required)
- `reason` (optional)

## 5. Correlation and lifecycle

- `rpc_id` is the operation-lifecycle key.
- Core-level request/response correlation uses `msg_id`; profile-level lifecycle correlation uses `rpc_id`.
- `RPC_RESP` or `RPC_ERR` terminates an RPC lifecycle.
- After terminal response/error, additional stream items MUST NOT be emitted.
- `RPC_CANCEL` MAY be issued prior to terminal completion; receivers SHOULD emit terminal `RPC_ERR` with cancellation code.
- Exactly one terminal outcome (`RPC_RESP` or `RPC_ERR`) MUST exist for each completed `rpc_id`.

## 6. Invariants

- `method` MUST be globally unique (reverse-DNS or URI format).
- `idempotency_key` semantics MUST be deterministic within implementation-defined replay window.
- If an `idempotency_key` is replayed with semantically identical request inputs, the endpoint MUST return the original deterministic terminal outcome.
- If an `idempotency_key` is replayed with conflicting request inputs, the endpoint MUST reject with deterministic terminal error.
- If streaming is used, `seq_no` MUST be strictly increasing per `rpc_id`.
- The replay window MUST be documented by implementations (recommended default: 24 hours).

## 7. Error model

Recommended error classes:

- invalid request
- unknown method
- unauthorized
- timeout
- unavailable
- internal
- cancelled

`retryable` MUST be explicitly set for each terminal error.

Conformance reporting SHOULD map profile-rejection outcomes to:
- `ERR_UNSUPPORTED_MSG_TYPE`
- `ERR_INVALID_PROFILE_PAYLOAD`
- `ERR_SECURITY_POLICY` (policy-gated rejection)

## 8. Conformance requirements

A conforming implementation MUST:

- reject unsupported `msg_type` values.
- enforce terminal lifecycle closure.
- enforce streaming ordering per `rpc_id`.
- implement deterministic idempotency-key handling policy.

See vectors in `conformance/vectors/catalog.md` under `rpc_*`.
