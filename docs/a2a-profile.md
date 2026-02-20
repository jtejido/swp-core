
# A2A Payload Profile (Draft v0.1)

## 1. Scope

This document defines a minimal A2A profile over SWP for interoperable task delegation.
The profile focuses on a strict lifecycle: handshake, task, progress events, and terminal result.

## 2. Message model

This profile defines four message classes:

- Handshake / Capability Advertisement
- Task / Delegation
- Event / Progress
- Result / Terminal outcome

Payload encoding for this profile MUST support P1:

- `docs/profile-payload-encoding-p1.md`
- Normative schema annex: `proto/a2a_p1.proto`

P1 opaque-bytes fields for this profile:

- `Task.task_id`
- `Task.input`
- `Event.task_id`
- `Event.event_payload`
- `Result.task_id`
- `Result.output`


Endpoints that originate or consume messages MUST enforce the required field semantics in this document.

## 3. msg_type assignments

- `1`: Handshake
- `2`: Task
- `3`: Event
- `4`: Result

Any other `msg_type` value is invalid for this profile version.

## 4. Required message fields

Required semantic fields by message class:

- Handshake:
  - `agent_id`
  - `capabilities` (may be empty)
- Task:
  - `task_id`
  - `kind`
  - `input` (may be empty bytes)
- Event:
  - `task_id`
  - event content (`message` and/or implementation-defined event payload)
- Result:
  - `task_id`
  - terminal state (`ok=true` success or `ok=false` failure)
  - success output or failure reason

## 5. Correlation model

Two correlation layers exist:

- SWP layer: `msg_id`
- A2A task layer: `task_id`

Rules:

- `msg_id` identifies an individual message exchange.
- `task_id` identifies a task lifecycle across Task, Event, and Result.
- Receivers MUST treat `task_id` as the primary lifecycle key.

## 6. Lifecycle and ordering rules

- A Task starts a lifecycle for a unique `task_id`.
- Event and Result MUST reference an existing Task `task_id`.
- Event MUST NOT appear before Task for the same `task_id`.
- Result MUST be terminal for `task_id`.
- After terminal Result, senders MUST NOT emit additional Event or Result for that `task_id`.

Per-task ordering:

- senders MUST preserve message order for a single `task_id` on a channel.
- receivers SHOULD process per-task messages in arrival order.

## 7. Idempotency and duplicates

- Duplicate Task with the same `task_id` and semantically equivalent content SHOULD be handled idempotently.
- Duplicate Task with conflicting content for the same `task_id` MUST be rejected.
- Duplicate terminal Result for the same `task_id` SHOULD be ignored if semantically equivalent.
- Duplicate terminal Result with conflicting terminal state SHOULD be rejected.

## 8. Error behavior

Profile-level failures SHOULD be surfaced as terminal Result with `ok=false`.

Recommended profile error conditions:

- unsupported capability
- malformed task input
- unknown task reference
- execution failed
- authorization denied (if surfaced at profile layer)

When an Event or Result references unknown `task_id`, receivers MUST reject the message.

## 9. Security and policy notes

- This profile inherits channel and identity properties from S1 when S1 is selected.
- Authorization SHOULD bind surfaced channel identity and requested task/capability.
- Multi-tenant deployments SHOULD scope `task_id` uniqueness to tenant context.

## 10. Conformance requirements

A conforming implementation MUST:

- pass required A2A vectors in `conformance/vectors/catalog.md`.
- enforce lifecycle closure after terminal Result.
- enforce deterministic handling of duplicates and unknown `task_id` references.
