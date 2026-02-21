# SWP Profile Implementation Roadmap

This roadmap is for runtime implementation in `poc/` with conformance-first execution.

## 1. Current Baseline (from `artifacts/conformance/all.default.json`)

Current suite status:

- default: `total=138 passed=138 failed=0 fallback=0`
- strict: `total=138 passed=138 failed=0 fallback=0`

All namespaces are now strict-clean in the spec-vector runner.

## 2. Execution Order (runtime handler completion)

Strict burn-down is complete; next work is moving profile semantics from runner-native checks into full server/runtime handlers:

1. `tooldisc`
2. `events`
3. `agdisc`
4. `artifact`
5. `state`
6. `cred`
7. `obs`
8. `core` rate-limit/duplicate in-flight behavior in server path

### Status snapshot

Implemented in `poc/internal/server/`:

- A2A (`profile_id=2`)
- SWP-AGDISC (`10`)
- SWP-TOOLDISC (`11`)
- SWP-RPC (`12`)
- SWP-EVENTS (`13`)
- SWP-ARTIFACT (`14`)
- SWP-CRED (`15`)
- SWP-POLICYHINT (`16`)
- SWP-STATE (`17`)
- SWP-OBS (`18`)
- SWP-RELAY (`19`)

Core server-path policy checks are also implemented for:

- burst/rate limiting
- duplicate in-flight `msg_id` rejection window
- pluggable runtime backends with per-server option injection (`server.New(...opts)`)
- mock-driven backend injection tests through real router dispatch
- AGDISC and TOOLDISC backend injection (catalog/card sources no longer hard-coded in handlers)
- RPC and EVENTS backend injection (execution/event adapters no longer hard-coded in handlers)
- shared runtime utility packages in `poc/internal/runtime/`:
  - `errors` (alias to canonical `ERR_*` mapping)
  - `clock` (clock abstraction helpers)
  - `context` (request metadata + correlation propagation)
  - `validate` (shared field/severity/trace checks)
- OBS context is attached to per-request dispatch context and reused for correlation fallback.
- MCP and SWP-RPC server paths now emit SWP-EVENTS via injected `EventsBackend` with OBS-aware correlation.

## 3. Test Matrix (Backend Interfaces)

This matrix maps each backend interface in `poc/internal/server/runtime_backends.go` to normal-path and fault-path test files for reviewer traceability.

| Backend interface | Normal-path test files | Fault-path test files |
| --- | --- | --- |
| `A2ABackend` | `poc/internal/server/a2a_test.go`, `poc/internal/server/runtime_backends_injection_test.go` | `poc/internal/server/a2a_test.go`, `poc/internal/server/runtime_backends_injection_test.go` |
| `ArtifactBackend` | `poc/internal/server/artifact_state_cred_test.go`, `poc/internal/server/runtime_backends_injection_test.go` | `poc/internal/server/artifact_state_cred_test.go`, `poc/internal/server/runtime_backends_injection_test.go` |
| `StateBackend` | `poc/internal/server/artifact_state_cred_test.go`, `poc/internal/server/runtime_backends_injection_test.go` | `poc/internal/server/artifact_state_cred_test.go`, `poc/internal/server/runtime_backends_injection_test.go` |
| `AGDISCBackend` | `poc/internal/server/agdisc_test.go`, `poc/internal/server/runtime_backends_injection_test.go` | `poc/internal/server/agdisc_test.go` |
| `ToolDiscBackend` | `poc/internal/server/tooldisc_events_test.go`, `poc/internal/server/runtime_backends_injection_test.go` | `poc/internal/server/tooldisc_events_test.go` |
| `RPCBackend` | `poc/internal/server/server_test.go`, `poc/internal/server/runtime_backends_injection_test.go` | `poc/internal/server/runtime_backends_injection_test.go` |
| `EventsBackend` | `poc/internal/server/tooldisc_events_test.go`, `poc/internal/server/runtime_backends_injection_test.go` | `poc/internal/server/tooldisc_events_test.go`, `poc/internal/server/runtime_backends_injection_test.go` |
| `CredBackend` | `poc/internal/server/artifact_state_cred_test.go`, `poc/internal/server/runtime_backends_injection_test.go` | `poc/internal/server/artifact_state_cred_test.go` |
| `PolicyHintBackend` | `poc/internal/server/policyhint_relay_test.go`, `poc/internal/server/runtime_backends_injection_test.go` | `poc/internal/server/policyhint_relay_test.go`, `poc/internal/server/runtime_backends_injection_test.go` |
| `RelayBackend` | `poc/internal/server/policyhint_relay_test.go`, `poc/internal/server/runtime_backends_injection_test.go` | `poc/internal/server/policyhint_relay_test.go`, `poc/internal/server/runtime_backends_injection_test.go` |
| `OBSBackend` | `poc/internal/server/obs_test.go`, `poc/internal/server/runtime_backends_injection_test.go` | `poc/internal/server/obs_test.go` |

Notes:
- `runtime_backends_injection_test.go` is the canonical injected-backend dispatch verification file.
- For `AGDISCBackend`, `ToolDiscBackend`, `CredBackend`, and `OBSBackend`, current fault-path coverage is semantic/profile-level in dedicated profile tests; there is no dedicated injected-backend fault harness yet.

## 4. Per-Profile Implementation Checklist

For each namespace:

1. Implement P1 payload decode/encode in `poc/internal/p1<profile>/`.
2. Implement deterministic handler validation/execution.
3. Register routing in `poc/internal/server/server.go`.
4. Add unit tests (`codec`, `handler`, and state-machine tests where applicable).
5. Run vectors in default mode.
6. Run vectors in strict mode and remove fallback paths.

Rules:

- Do not edit specs unless a vector reveals true ambiguity.
- Keep canonical `ERR_*` + alias code mapping consistent.
- Preserve existing spec-vector-runner semantics.

## 5. Milestones

### Milestone A: Federation Control Plane Runtime

- `a2a`, `events`, `agdisc`, `tooldisc`, `obs`
- Exit criterion: server dispatch path implements profile handlers (not only runner-native semantics).

### Milestone B: Data Plane Runtime

- `artifact`, `state`
- Exit criterion: integrity and parent/hash failure paths are implemented in runtime handler/store paths.

### Milestone C: Federation Governance Runtime

- `cred` (and harden `core` server-path edge cases)
- Exit criterion: profile behavior is enforced end-to-end in runtime server flow with deterministic codes.

### Milestone D: Backend Abstraction Hardening

- backend interfaces for stateful profiles and observability context
- backend interfaces for RPC and EVENTS execution paths
- per-server backend injection options
- mock-backed routing tests verifying injected backend usage
- fault-injection tests for injected backends (conflict/not-found/internal paths)
- Exit criterion: handlers are not hard-coupled to package-level in-memory state.

## 6. Daily Commands

Default runs:

```bash
make vectors-a2a
make vectors-events
make vectors-agdisc
make vectors-tooldisc
make vectors-artifact
make vectors-state
make vectors-cred
make vectors-obs
make vectors-core
```

Strict runs:

```bash
make vectors-a2a-strict
make vectors-events-strict
make vectors-agdisc-strict
make vectors-tooldisc-strict
make vectors-artifact-strict
make vectors-state-strict
make vectors-cred-strict
make vectors-obs-strict
make vectors-core-strict
```

## 7. Scoreboard Recommendation

Track this per namespace:

- default `total/passed/failed/fallback_count`
- strict `total/passed/failed/fallback_count`
- remaining fallback vector IDs

Use this file as the runtime implementation gate now that strict-clean status is achieved.
