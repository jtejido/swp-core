# SWP-TOOLDISC (Draft v0.1)

## 1. Scope

SWP-TOOLDISC defines generic tool discovery independent of MCP session transport.

**Profile ID:** `11`

## 2. Message model

Supported message classes:

- `TOOLDISC_LIST_REQ`
- `TOOLDISC_LIST_RESP`
- `TOOLDISC_GET_REQ`
- `TOOLDISC_GET_RESP`
- `TOOLDISC_ERR`

## 3. msg_type assignments

- `1`: `TOOLDISC_LIST_REQ`
- `2`: `TOOLDISC_LIST_RESP`
- `3`: `TOOLDISC_GET_REQ`
- `4`: `TOOLDISC_GET_RESP`
- `5`: `TOOLDISC_ERR`

Any other `msg_type` is invalid for this profile version.

Payload encoding for this profile MUST support P1:

- `docs/profile-payload-encoding-p1.md`
- Normative schema annex: `proto/swp_tooldisc.proto`

P1 opaque-bytes fields for this profile:

- `ToolDescriptor.descriptor_payload`


## 4. Tool descriptor requirements

Each descriptor MUST include:

- stable `tool_id`
- `name`
- `version`
- schema reference(s)

Hints such as rate/cost MUST be non-binding and clearly marked.
Schema references MUST be resolvable (URI or content hash form).

## 5. Behavior

- list operations SHOULD support paging/filtering.
- get operations MUST be deterministic by `tool_id` + optional version selector.

## 6. Conformance requirements

A conforming implementation MUST:

- enforce required descriptor fields.
- enforce schema-ref validity rules.
- reject unsupported `msg_type` values.
- handle missing-tool requests deterministically.

See vectors in `conformance/vectors/catalog.md` under `tooldisc_*`.
