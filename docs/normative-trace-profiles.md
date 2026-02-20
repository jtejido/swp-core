# Profile Normative Traceability Matrix

This matrix maps key profile-level normative statements to conformance vectors.
Evidence vectors are defined in `conformance/vectors/catalog.md`.

## MCP Mapping

| ID | Profile | Statement summary | Evidence vectors |
|---|---|---|---|
| P-MCP-001 | MCP | Preserve JSON-RPC bytes in relay mode | `mcp_0001`, `mcp_0002`, `mcp_0015` |
| P-MCP-002 | MCP | Responses reuse request `msg_id` | `mcp_0006` |
| P-MCP-003 | MCP | Unsupported `msg_type` rejected | `mcp_0012` |

## A2A

| ID | Profile | Statement summary | Evidence vectors |
|---|---|---|---|
| P-A2A-001 | A2A | No event/result before task | `a2a_0006`, `a2a_0007` |
| P-A2A-002 | A2A | Terminal lifecycle closure | `a2a_0010`, `a2a_0011` |
| P-A2A-003 | A2A | Deterministic duplicate handling | `a2a_0008`, `a2a_0009` |

## SWP Foundation

| ID | Profile | Statement summary | Evidence vectors |
|---|---|---|---|
| P-RPC-001 | SWP-RPC | Terminal response/error closure | `rpc_0001`, `rpc_0004` |
| P-RPC-002 | SWP-RPC | Streaming order and terminal semantics | `rpc_0003` |
| P-RPC-003 | SWP-RPC | Idempotency-key replay behavior | `rpc_0005` |
| P-EVT-001 | SWP-EVENTS | Required fields enforced | `events_0001` |
| P-EVT-002 | SWP-EVENTS | Ordered delivery semantics | `events_0002` |
| P-EVT-003 | SWP-EVENTS | Correlation propagation | `events_0003` |
| P-OBS-001 | SWP-OBS | Valid traceparent required | `obs_0001` |
| P-OBS-002 | SWP-OBS | tracestate preservation | `obs_0003` |

## SWP Discovery

| ID | Profile | Statement summary | Evidence vectors |
|---|---|---|---|
| P-AGD-001 | SWP-AGDISC | Deterministic get/doc/not-found behavior | `agdisc_0001`, `agdisc_0002` |
| P-AGD-002 | SWP-AGDISC | Invalid card handling | `agdisc_0003` |
| P-AGD-003 | SWP-AGDISC | Cache validator semantics | `agdisc_0004`, `agdisc_0005` |
| P-TLD-001 | SWP-TOOLDISC | List/get semantics | `tooldisc_0001`, `tooldisc_0002` |
| P-TLD-002 | SWP-TOOLDISC | Descriptor/schema invariants | `tooldisc_0003`, `tooldisc_0004` |

## SWP Data and Governance

| ID | Profile | Statement summary | Evidence vectors |
|---|---|---|---|
| P-ART-001 | SWP-ARTIFACT | Integrity and corruption handling | `artifact_0003`, `artifact_0006` |
| P-ART-002 | SWP-ARTIFACT | Resume/range behavior | `artifact_0004`, `artifact_0005` |
| P-STA-001 | SWP-STATE | Parent DAG validation | `state_0003` |
| P-STA-002 | SWP-STATE | Hash consistency checks | `state_0002` |
| P-POL-001 | SWP-POLICYHINT | Unknown-key deterministic behavior | `policyhint_0002` |
| P-POL-002 | SWP-POLICYHINT | Violation payload requirements | `policyhint_0004` |
| P-CRD-001 | SWP-CRED | Chain length and expiry enforced | `cred_0001`, `cred_0002` |
| P-CRD-002 | SWP-CRED | Invalid credential handling | `cred_0003` |
| P-RLY-001 | SWP-RELAY | At-least-once with dedupe | `relay_0001`, `relay_0002` |
| P-RLY-002 | SWP-RELAY | Retry/dead-letter semantics | `relay_0003`, `relay_0004` |

