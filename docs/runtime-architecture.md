# SWP Runtime Architecture (Go POC)

This note documents the runtime flow implemented in `poc/internal/server/` after backend injection and telemetry wiring.

## 1. Request flow

1. `Server.handleConn` reads frame bytes, decodes E1 envelope, and validates Core invariants.
2. Per-request runtime context is attached before dispatch:
   - message metadata (`profile_id`, `msg_id`)
   - correlation snapshot from OBS backend (`traceparent`, `tracestate`, `msg_id`, `task_id`, `rpc_id`)
3. Router dispatches to profile handler.
4. Handler logic uses injected backends from `server.New(...options)` (or default in-memory backends).
5. Response envelopes are encoded and written back on the same connection.

## 2. Runtime utility packages

Cross-cutting helpers are under `poc/internal/runtime/`:

- `clock`: timestamp helpers for deterministic time access points.
- `context`: typed context helpers for request metadata and correlation.
- `errors`: alias-to-canonical (`ERR_*`) code mapping.
- `validate`: shared validation primitives (required fields, severity, traceparent shape).

## 3. OBS and EVENTS correlation behavior

- OBS context is persisted in the OBS backend and also attached to dispatch context.
- EVENTS publish path enriches correlation in this order:
  1. explicit event payload fields,
  2. request context correlation,
  3. OBS backend fallback.
- EVENTS validation enforces at least one correlation key (`msg_id`, `task_id`, or `rpc_id`).

## 4. Automatic telemetry emission

Server-path MCP and SWP-RPC handlers emit EVENTS via injected `EventsBackend`:

- MCP:
  - `swp.mcp.request`
  - `swp.mcp.notification`
  - `swp.mcp.response`
- SWP-RPC:
  - `swp.rpc.request`
  - `swp.rpc.stream`
  - `swp.rpc.response`
  - `swp.rpc.cancel`
  - `swp.rpc.error`

Emission uses OBS-aware correlation fallback so telemetry records remain linkable when request payloads omit `task_id` or `rpc_id`.

## 5. Test coverage anchors

- Backend injection and fault paths: `poc/internal/server/runtime_backends_injection_test.go`
- Profile behavior tests:
  - `poc/internal/server/a2a_test.go`
  - `poc/internal/server/agdisc_test.go`
  - `poc/internal/server/tooldisc_events_test.go`
  - `poc/internal/server/artifact_state_cred_test.go`
  - `poc/internal/server/policyhint_relay_test.go`
  - `poc/internal/server/obs_test.go`
  - `poc/internal/server/server_test.go`
