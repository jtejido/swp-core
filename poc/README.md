# SWP Go POC

This POC implements the minimal recommended scope:

- Core only first: framing + E1 envelope decode/encode + validation + profile dispatch.
- Profiles: MCP Mapping (`1`), A2A (`2`), and SWP family (`10`-`19`).
- Demo flows: one request/response (MCP `tools/list`) and one streaming flow (SWP-RPC `demo.stream.count`).
- Vector tooling: `*.json` + real `*.bin` fixtures for a starter POC vector set.

This POC uses protobuf payload encoding for SWP-RPC (P1), aligned with `docs/profile-payload-encoding-p1.md` and `proto/swp_rpc.proto`.

## Prereqs

- Go `1.22+`
- `make`
- Optional: `podman` with compose plugin (`podman compose`)
- Optional: `jq` (for pretty JSON in curl demo)

## Local quickstart

1. Generate starter vectors:

```bash
make gen-vectors
```

To materialize concrete fixtures for full C1 class (`core_*`, `e1_*`, `s1_*`, `mcp_*`):

```bash
make gen-c1-vectors
```

To materialize concrete fixtures for the remaining profile namespaces (`a2a_*`, `rpc_*`, `events_*`, `agdisc_*`, etc.):

```bash
make gen-remaining-vectors
```

2. Run tests:

```bash
make test
make test-race
```

3. Validate POC vectors:

```bash
make poc-vectors
```

Run full spec vectors:

```bash
make vectors
```

Optional subset/example output file:

```bash
make vectors SPEC_VECTOR_ARGS="-pattern conformance/vectors/core_*.json,conformance/vectors/e1_*.json -json-out /tmp/spec-vectors-summary.json"
```

Strict mode (disallow fallback evaluation):

```bash
make vectors-strict SPEC_VECTOR_ARGS="-pattern 'conformance/vectors/core_*.json' -json-out /tmp/spec-vectors-strict-summary.json"
```

`json-out` summaries now include per-vector `used_fallback` for auditability.

4. Run TCP server and demo client:

```bash
make run-server
# in another terminal
make run-client
```

Or one-command demo:

```bash
make demo
```

## MCP JSON gateway (for external/OSS clients)

The POC includes `mcp-json-gateway` (`POST /mcp`), which converts JSON-RPC requests into SWP MCP profile frames.
Gateway implementation now reuses a persistent SWP TCP connection (with reconnect on failure) rather than opening a new connection per request.

Run locally:

```bash
make run-server
# new terminal
make run-gateway
# new terminal
make mcp-curl
```

Any JSON-RPC-capable client can call `http://127.0.0.1:8080/mcp`.

## Podman compose flows

Bring up server + gateway:

```bash
make podman-up
```

Run MCP request through gateway:

```bash
make mcp-curl
```

Run demo client against server via compose:

```bash
make podman-demo
```

Run vector runner in container:

```bash
make podman-poc-vectors
```

Run full spec vectors in container:

```bash
make podman-vectors
```

Shutdown:

```bash
make podman-down
```

## Runtime backend injection

Server runtime state is now pluggable via `server.New(logger, ...options)`:

- `server.WithA2ABackend(...)`
- `server.WithAGDISCBackend(...)`
- `server.WithToolDiscBackend(...)`
- `server.WithRPCBackend(...)`
- `server.WithEventsBackend(...)`
- `server.WithArtifactBackend(...)`
- `server.WithStateBackend(...)`
- `server.WithCredBackend(...)`
- `server.WithPolicyHintBackend(...)`
- `server.WithRelayBackend(...)`
- `server.WithOBSBackend(...)`

If no options are provided, the server uses built-in in-memory backends.

Runtime cross-cutting helpers live in `poc/internal/runtime/`:

- `clock`: reusable clock abstraction helpers
- `context`: request metadata + correlation propagation helpers
- `errors`: alias/runtime code to canonical `ERR_*` mapping helper
- `validate`: shared field/severity/trace validation helpers

MCP and SWP-RPC server paths emit SWP-EVENTS through the injected `EventsBackend`, using OBS context as correlation fallback when `task_id`/`rpc_id` are absent on emitted events.

### Backend contract matrix

These profile handlers use backend interfaces in `poc/internal/server/runtime_backends.go`.
Reject paths should map to canonical `ERR_*` taxonomy in `docs/error-codes.md`.

| Profile | Handler | Option | Backend interface | Typical canonical reject codes |
| --- | --- | --- | --- | --- |
| A2A (`2`) | `handleA2A` | `WithA2ABackend` | `A2ABackend` | `ERR_INVALID_PROFILE_PAYLOAD`, `ERR_INVALID_FRAME`, `ERR_UNSUPPORTED_MSG_TYPE` |
| SWP-AGDISC (`10`) | `handleSWPAGDISC` | `WithAGDISCBackend` | `AGDISCBackend` | `ERR_NOT_FOUND`, `ERR_INVALID_PROFILE_PAYLOAD`, `ERR_UNSUPPORTED_MSG_TYPE` |
| SWP-TOOLDISC (`11`) | `handleSWPToolDisc` | `WithToolDiscBackend` | `ToolDiscBackend` | `ERR_NOT_FOUND`, `ERR_INVALID_PROFILE_PAYLOAD`, `ERR_UNSUPPORTED_MSG_TYPE` |
| SWP-RPC (`12`) | `handleSWPRPC` | `WithRPCBackend` | `RPCBackend` | `ERR_INVALID_PROFILE_PAYLOAD`, `ERR_UNSUPPORTED_MSG_TYPE`, `ERR_COMPATIBILITY_POLICY` |
| SWP-EVENTS (`13`) | `handleSWPEvents` | `WithEventsBackend` | `EventsBackend` | `ERR_INVALID_PROFILE_PAYLOAD`, `ERR_UNSUPPORTED_MSG_TYPE`, `ERR_NOT_FOUND` |
| SWP-ARTIFACT (`14`) | `handleSWPArtifact` | `WithArtifactBackend` | `ArtifactBackend` | `ERR_NOT_FOUND`, `ERR_INVALID_PROFILE_PAYLOAD`, `ERR_COMPATIBILITY_POLICY` |
| SWP-CRED (`15`) | `handleSWPCred` | `WithCredBackend` | `CredBackend` | `ERR_INVALID_PROFILE_PAYLOAD`, `ERR_SECURITY_POLICY`, `ERR_COMPATIBILITY_POLICY` |
| SWP-POLICYHINT (`16`) | `handleSWPPolicyHint` | `WithPolicyHintBackend` | `PolicyHintBackend` | `ERR_INVALID_PROFILE_PAYLOAD`, `ERR_COMPATIBILITY_POLICY` |
| SWP-STATE (`17`) | `handleSWPState` | `WithStateBackend` | `StateBackend` | `ERR_NOT_FOUND`, `ERR_INVALID_PROFILE_PAYLOAD`, `ERR_COMPATIBILITY_POLICY` |
| SWP-OBS (`18`) | `handleSWPOBS` | `WithOBSBackend` | `OBSBackend` | `ERR_INVALID_PROFILE_PAYLOAD`, `ERR_COMPATIBILITY_POLICY` |
| SWP-RELAY (`19`) | `handleSWPRelay` | `WithRelayBackend` | `RelayBackend` | `ERR_NOT_FOUND`, `ERR_INVALID_PROFILE_PAYLOAD`, `ERR_RATE_LIMIT_EXCEEDED` |
