# SWP Go POC

This POC implements the minimal recommended scope:

- Core only first: framing + E1 envelope decode/encode + validation + profile dispatch.
- Profiles: MCP Mapping (profile `1`) and SWP-RPC (profile `12`).
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

## A2A note

A2A profile behavior is still spec-only in this repo. This POC uses SWP-RPC streaming as the second executable profile while preserving the same Core transport/validation path.
