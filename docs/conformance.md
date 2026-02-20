
# Conformance Suite (Draft)

Canonical reject taxonomy for conformance is defined in:

- `docs/error-codes.md`

## Core vectors (required)

Implementations claiming SWP Core conformance MUST pass:
1) Valid framing + envelope decoding
2) Invalid length prefixes (0, oversized, truncated)
3) Oversized frame rejection (MAX_FRAME_BYTES)
4) Unsupported version handling
5) Unknown profile handling
6) Invalid msg_id length handling
7) Payload length limit handling

## Encoding binding vectors (required for Core v1 claim)

E1:
- valid envelope decode using E1 field order
- malformed/overflow varint rejection
- invalid E1 version rejection
- invalid/empty msg_id rejection
- extension TLV handling (unknown types ignored)
- extension size limit enforcement

## Profile vectors (required)

MCP Mapping:
- JSON-RPC request roundtrip (opaque bytes preserved)
- JSON-RPC response roundtrip
- JSON-RPC notification behavior (no response required)
- response correlation via reused request `msg_id`
- deterministic core-error to JSON-RPC error mapping
- invalid JSON payload handling
- invalid UTF-8 payload handling
- invalid request/response/notification shape handling
- unsupported MCP mapping `msg_type` handling

A2A:
- handshake capability advertisement
- task acceptance and rejection
- event ordering (no event before task)
- terminal result success and failure
- unknown task reference handling
- duplicate task idempotency behavior
- conflicting duplicate task rejection
- terminal result closure (no post-terminal events)
- duplicate terminal result handling
- unsupported A2A `msg_type` handling

SWP-RPC:
- request/response/error lifecycle closure
- idempotency-key replay behavior
- deterministic retryability signaling
- streaming ordering and terminal semantics
- unsupported RPC `msg_type` handling

SWP-EVENTS:
- required event field enforcement
- per-stream ordering
- correlation propagation (`msg_id`/`task_id`/`rpc_id`)
- invalid event and batch handling
- unsupported EVENTS `msg_type` handling

SWP-OBS:
- traceparent validity and preservation
- tracestate unknown-entry preservation
- correlation linkage behavior
- unsupported OBS `msg_type` handling

SWP-AGDISC:
- discovery get/doc/not-modified behavior
- invalid agent card handling
- deterministic not-found behavior
- cache validator behavior
- unsupported AGDISC `msg_type` handling

SWP-TOOLDISC:
- list/get behavior (including paging/filter)
- descriptor required-field enforcement
- schema-ref validity behavior
- deterministic missing-tool handling
- unsupported TOOLDISC `msg_type` handling

SWP-ARTIFACT:
- offer/get/chunk lifecycle
- chunk/range ordering and bounds
- integrity mismatch handling
- resume behavior
- unsupported ARTIFACT `msg_type` handling

SWP-STATE:
- put/get lifecycle
- parent DAG validation
- state hash/id consistency
- missing parent/missing state handling
- unsupported STATE `msg_type` handling

SWP-POLICYHINT:
- hint propagation
- unknown-key behavior
- conflict/precedence behavior
- violation payload requirements
- unsupported POLICYHINT `msg_type` handling

SWP-CRED:
- credential type handling
- chain length and expiry enforcement
- deterministic invalid credential handling
- delegation/revocation behavior
- unsupported CRED `msg_type` handling

SWP-RELAY:
- at-least-once delivery semantics
- delivery-id dedup behavior
- ack timeout and retry behavior
- dead-letter reason behavior
- unsupported RELAY `msg_type` handling

## Binding vectors (required when claiming S1)

S1:
- unauthenticated peer is rejected before frame processing
- integrity/auth failure terminates channel
- downgrade policy failure terminates channel
- timestamp freshness policy behavior is deterministic (enforced or documented as disabled)

## Conformance classes

Class taxonomy and required vector namespaces are defined in:

- `docs/conformance-classes.md`

Recommended publication form is `SWP Cx` (for example `SWP C1`).

## Vector format

See `conformance/vectors/README.md`.
Required cases are listed in `conformance/vectors/catalog.md`.
Template expected metadata files are pre-generated under `conformance/vectors/*.json`.
Runtime/mixed vectors use concrete `.bin` fixtures.
Process vectors use `.evidence.md` artifacts where runtime bytes are not applicable.

Executable conformance claims MUST use concrete fixtures and assertions.

Reject-vector expectations:
- runtime/process vectors with `expected.outcome = "reject"` MUST include `expected.expected_error_code` using canonical `ERR_*` values from `docs/error-codes.md`.
- `expected.code` MAY be retained as runtime/legacy alias for existing tooling and transition compatibility.
- when runner tooling supports fallback evaluation, reports SHOULD include per-vector `used_fallback`.
- strict implementation-only runs SHOULD disable fallback evaluation (for example `-no-fallback`) and treat fallback-required vectors as non-pass.

JSON summary artifact expectations (for runner-produced reports):
- top-level `schema_version` MUST be present.
- top-level `run` metadata SHOULD include:
  - `pattern`
  - `no_fallback`
  - `timestamp_utc`
  - `runner_git_sha` (or `nogit` when unavailable)
- top-level `results` MUST contain exactly one entry per executed vector.
- top-level `failures`, when present, MUST be a subset of `results`.
- per-result entries SHOULD include `path` and `used_fallback` for reproducibility.
- when present, `fallback_mode = "disallowed"` indicates policy failure caused by `-no-fallback`, not protocol-level semantic mismatch.

Formal output schema and invariants:
- `docs/spec-vector-runner-output.md`

## Claims and traceability artifacts

- Claim criteria and required evidence: `docs/conformance-claims-checklist.md`
- Core normative trace matrix: `docs/normative-trace-core.md`
- Profile normative trace matrix: `docs/normative-trace-profiles.md`

## Minimum pass criteria

- Core conformance claim:
  - 100% pass on required core vectors.
- Profile conformance claim:
  - 100% pass on required profile vectors for that profile.
- Federation-ready claim:
  - core + at least one profile + required S1 binding vectors.

- Class claim:
  - 100% pass for all namespaces required by claimed class in `docs/conformance-classes.md`.
