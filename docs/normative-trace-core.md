# Core Normative Traceability Matrix

This matrix maps normative statements in `docs/core-spec.md` to conformance vectors.
Statement IDs are local to this matrix for traceability and publication packaging.

## Legend

- Type: `MUST`, `SHOULD`, or `SHALL`
- Evidence: vector IDs in `conformance/vectors/catalog.md`

## Matrix

| ID | Section | Type | Normative statement (summary) | Evidence vectors |
|---|---|---|---|---|
| C-001 | 3 | SHALL | Stream is encoded as sequence of length-delimited frames | `core_0001`, `core_0002` |
| C-002 | 3 | MUST | Length prefix fully readable | `core_0004` |
| C-003 | 3 | MUST | Zero length prefix is invalid | `core_0003` |
| C-004 | 3 | MUST | Prefix above `MAX_FRAME_BYTES` is invalid | `core_0005` |
| C-005 | 3 | MUST | Truncated frame body is invalid | `core_0006` |
| C-006 | 3 | MUST | Undecodable envelope payload is invalid | `core_0007` |
| C-007 | 3.1 | MUST | Enforce `MAX_FRAME_BYTES` | `core_0005`, `core_0019` |
| C-008 | 3.1 | SHOULD | Terminate stream on oversized frame reject | `core_0005` |
| C-009 | 3.1 | SHOULD | Apply frame-rate/resource limiting | `core_0016` |
| C-010 | 4.1 | MUST | Envelope includes required core fields | `core_0021`, `core_0022`, `core_0023`, `core_0024`, `core_0025` |
| C-011 | 4.2 | MUST | Optional fields do not alter required-field meaning | `core_0031` |
| C-012 | 4.3 | MUST | Unknown flags ignored unless profile requires otherwise | `core_0013` |
| C-013 | 4.3 | MUST | Unknown flags do not reinterpret known fields | `core_0026` |
| C-014 | 4.3 | SHOULD | Unknown flags surfaced to observability | `core_0013` |
| C-015 | 4.4 | MUST | Reject unsupported core version | `core_0008` |
| C-016 | 4.4 | MUST | Validate known/handled `profile_id` | `core_0009`, `core_0017`, `core_0018` |
| C-017 | 4.4 | MUST | Enforce `msg_id` length bounds | `core_0010`, `core_0011` |
| C-018 | 4.4 | MUST | Enforce timestamp/skew policy when enabled | `core_0014`, `core_0015` |
| C-019 | 4.4 | MUST | Enforce payload length bound | `core_0012`, `core_0020` |
| C-020 | 4.4 | MUST | Invalid invariants treated as invalid message | `core_0008`, `core_0009`, `core_0010`, `core_0011`, `core_0012` |
| C-021 | 4.4 | SHOULD | Sender maintains unique in-flight `msg_id` | `core_0027` |
| C-022 | 5 | SHOULD | Map framing failures to `INVALID_FRAME` | `core_0028` |
| C-023 | 5 | SHOULD | Map unsupported version to `UNSUPPORTED_VERSION` | `core_0029` |
| C-024 | 5 | SHOULD | Map unknown profile to `UNKNOWN_PROFILE` | `core_0017` |
| C-025 | 5 | SHOULD | Map envelope invariant failure to `INVALID_ENVELOPE` | `core_0030` |
| C-026 | 6 | MUST | Dispatch based on `profile_id` | `core_0032` |
| C-027 | 6 | MUST | Unknown profile handled per binding | `core_0017`, `core_0018` |
| C-028 | 6 | SHOULD | Return `UNKNOWN_PROFILE` when response path exists | `core_0017` |
| C-029 | 8 | MUST | Preserve backward compatibility rules for extensibility | `core_0033` |
| C-030 | 3.3 | MUST | Core v1 supports E1 envelope encoding binding | `e1_0001` |
| C-031 | 3.3 | MUST | Additional encodings cannot replace required E1 support | `e1_0001` |
| C-032 | E1 | MUST | Reject malformed varints (too long/overflow) | `e1_0002`, `e1_0003` |
| C-033 | E1 | MUST | Reject E1 envelope versions other than 1 | `e1_0004` |
| C-034 | E1 | MUST | Reject empty/invalid msg_id in E1 encoding | `e1_0005` |
| C-035 | E1 | MUST | Ignore unknown extension TLV types | `e1_0006` |
| C-036 | E1 | MUST | Reject extension block above configured limit | `e1_0007` |
| C-037 | E1 | MUST | Reject truncated bytes-field encodings | `e1_0008` |

## Notes

- `core_0033` is a publication/process conformance vector (spec-level), not a runtime wire vector.
- Where statements include conditional behavior ("if binding supports error responses"), evidence vectors are scoped accordingly.
