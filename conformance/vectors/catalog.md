# Vector Catalog (Comprehensive Set)

This catalog defines the minimum vector set across Core, profiles, and S1 binding.

## Core vectors

1. `core_0001_valid_min_frame`: minimal valid frame with supported version/profile and empty payload.
2. `core_0002_valid_typical_frame`: valid frame with common payload and 16-byte `msg_id`.
3. `core_0003_invalid_zero_length`: length prefix is zero.
4. `core_0004_invalid_truncated_prefix`: fewer than 4 bytes available for prefix.
5. `core_0005_invalid_oversized_length`: prefix larger than `MAX_FRAME_BYTES`.
6. `core_0006_invalid_truncated_body`: prefix declares N, stream contains fewer than N bytes.
7. `core_0007_invalid_envelope_decode`: body bytes cannot decode as envelope.
8. `core_0008_unsupported_version`: envelope version not supported.
9. `core_0009_unknown_profile`: unsupported `profile_id`.
10. `core_0010_invalid_msg_id_short`: `msg_id` shorter than `MIN_MSG_ID_BYTES`.
11. `core_0011_invalid_msg_id_long`: `msg_id` longer than `MAX_MSG_ID_BYTES`.
12. `core_0012_invalid_payload_oversize`: payload exceeds `MAX_PAYLOAD_BYTES`.
13. `core_0013_unknown_flags_set`: unknown flag bits set; frame still accepted.
14. `core_0014_stale_timestamp`: timestamp outside accepted freshness window.
15. `core_0015_future_timestamp`: timestamp too far in the future.
16. `core_0016_burst_limit_exceeded`: valid frames exceeding configured rate limit.
17. `core_0017_unknown_profile_with_error_path`: receiver returns `UNKNOWN_PROFILE`.
18. `core_0018_unknown_profile_no_error_path`: receiver closes stream on unknown profile.
19. `core_0019_boundary_max_frame_exact`: frame exactly equals `MAX_FRAME_BYTES`.
20. `core_0020_boundary_max_payload_exact`: payload exactly equals `MAX_PAYLOAD_BYTES`.
21. `core_0021_missing_required_field_version`: envelope missing/invalid `version` field is rejected.
22. `core_0022_missing_required_field_profile_id`: envelope missing/invalid `profile_id` field is rejected.
23. `core_0023_missing_required_field_msg_type`: envelope missing/invalid `msg_type` field is rejected.
24. `core_0024_missing_required_field_msg_id`: envelope missing/invalid `msg_id` field is rejected.
25. `core_0025_missing_required_field_ts_unix_ms`: envelope missing/invalid timestamp is rejected when required by policy.
26. `core_0026_unknown_flags_no_reinterpretation`: unknown flags do not alter interpretation of known fields.
27. `core_0027_duplicate_inflight_msg_id`: sender/adapter conformance check detects duplicate in-flight `msg_id`.
28. `core_0028_error_mapping_invalid_frame`: when error response path exists, framing failure maps to `INVALID_FRAME`.
29. `core_0029_error_mapping_unsupported_version`: when error response path exists, unsupported version maps to `UNSUPPORTED_VERSION`.
30. `core_0030_error_mapping_invalid_envelope`: when error response path exists, envelope invariant failure maps to `INVALID_ENVELOPE`.
31. `core_0031_optional_fields_no_semantic_override`: optional/extension fields do not change required-field semantics.
32. `core_0032_profile_dispatch_known_profile`: frame dispatch follows `profile_id` for known profiles.
33. `core_0033_extensibility_backward_compatibility`: publication/process check validates compatibility policy for versioned changes.

## Encoding binding E1 vectors

1. `e1_0001_valid_min_envelope`: valid E1-encoded envelope decodes successfully.
2. `e1_0002_varint_too_long_invalid`: varint longer than 10 bytes is rejected.
3. `e1_0003_varint_overflow_invalid`: varint overflowing uint64 is rejected.
4. `e1_0004_invalid_version`: E1 envelope version not equal to 1 is rejected.
5. `e1_0005_empty_msg_id_invalid`: E1 envelope with empty msg_id is rejected.
6. `e1_0006_unknown_extension_ignored`: unknown extension TLV type is ignored without changing known field semantics.
7. `e1_0007_extensions_too_large`: extension block exceeding MAX_EXT_BYTES is rejected.
8. `e1_0008_truncated_bytes_field`: truncated bytes field encoding is rejected.

## MCP mapping vectors

1. `mcp_0001_request_roundtrip`: JSON-RPC request bytes preserved exactly end-to-end.
2. `mcp_0002_response_roundtrip`: JSON-RPC response bytes preserved exactly.
3. `mcp_0003_notification_no_response`: notification accepted without response requirement.
4. `mcp_0004_error_mapping_internal`: core `INTERNAL_ERROR` mapped to JSON-RPC internal error.
5. `mcp_0005_error_mapping_unknown_profile`: `UNKNOWN_PROFILE` mapped deterministically.
6. `mcp_0006_response_msg_id_reuse`: response reuses originating request `msg_id`.
7. `mcp_0007_request_missing_id_invalid`: generated request without JSON-RPC `id` is rejected.
8. `mcp_0008_response_both_result_and_error_invalid`: response containing both `result` and `error` is rejected.
9. `mcp_0009_notification_with_response_violation`: notification path does not require or emit response.
10. `mcp_0010_invalid_utf8_payload`: invalid UTF-8 payload rejected by validating endpoint/gateway.
11. `mcp_0011_invalid_json_payload`: non-JSON payload rejected by validating endpoint/gateway.
12. `mcp_0012_unsupported_msg_type`: unsupported `msg_type` for MCP mapping is rejected.
13. `mcp_0013_stream_partial_then_terminal_success`: partial notifications followed by terminal response.
14. `mcp_0014_stream_partial_then_terminal_error`: partial notifications followed by terminal error response.
15. `mcp_0015_payload_preservation_whitespace`: relay preserves JSON payload bytes including whitespace and key order.

## A2A vectors

1. `a2a_0001_handshake_success`: handshake accepted and capabilities surfaced.
2. `a2a_0002_task_event_result_success`: valid task lifecycle completes with terminal success result.
3. `a2a_0003_task_result_failure`: task terminates with failure result.
4. `a2a_0004_event_after_terminal_result`: post-terminal event is rejected or ignored per profile rule.
5. `a2a_0005_duplicate_terminal_result`: duplicate terminal result handled idempotently.
6. `a2a_0006_event_before_task_invalid`: event referencing task before task creation is rejected.
7. `a2a_0007_result_before_task_invalid`: result referencing unknown task is rejected.
8. `a2a_0008_duplicate_task_same_payload_idempotent`: duplicate task with semantically equivalent payload is handled idempotently.
9. `a2a_0009_duplicate_task_conflicting_payload_rejected`: conflicting duplicate task is rejected.
10. `a2a_0010_post_terminal_event_rejected`: event after terminal result is rejected deterministically.
11. `a2a_0011_post_terminal_result_rejected`: additional conflicting result after terminal state is rejected.
12. `a2a_0012_task_malformed_input_failure`: malformed task input yields terminal failure result.
13. `a2a_0013_unsupported_capability_failure`: task requiring unsupported capability yields deterministic failure.
14. `a2a_0014_unsupported_msg_type`: unsupported `msg_type` value for A2A profile is rejected.
15. `a2a_0015_multi_task_interleaving_valid`: interleaved events across different task_id values remain valid per-task ordering.

## S1 binding vectors

1. `s1_0001_unauthenticated_peer_rejected`: peer without required auth is rejected before any frame is processed.
2. `s1_0002_authentication_failure_terminates`: failed peer authentication terminates channel deterministically.
3. `s1_0003_downgrade_policy_failure`: endpoint fails closed when required security parameters are not met.
4. `s1_0004_integrity_failure_terminates`: detected integrity/auth failure terminates channel and logs security event.
5. `s1_0005_timestamp_freshness_enforced`: stale frame is rejected when freshness policy is enabled.
6. `s1_0006_timestamp_freshness_disabled_documented`: behavior matches explicit policy when freshness checks are disabled.

## SWP-AGDISC vectors

1. `agdisc_0001_get_doc_success`: valid agent-card retrieval succeeds.
2. `agdisc_0002_not_found`: unknown agent yields deterministic not-found behavior.
3. `agdisc_0003_invalid_doc_rejected`: malformed/invalid agent card is rejected.
4. `agdisc_0004_etag_not_modified`: validator-based not-modified behavior is deterministic.
5. `agdisc_0005_cache_expiry_refresh`: expired cache state triggers refresh behavior.
6. `agdisc_0006_unsupported_msg_type`: unsupported AGDISC `msg_type` is rejected.

## SWP-TOOLDISC vectors

1. `tooldisc_0001_list_paging_filter`: list request supports deterministic paging/filter semantics.
2. `tooldisc_0002_get_success`: get request returns deterministic descriptor payload.
3. `tooldisc_0003_missing_tool_not_found`: unknown tool yields deterministic not-found behavior.
4. `tooldisc_0004_schema_ref_invalid`: invalid/unresolvable schema reference is rejected.
5. `tooldisc_0005_descriptor_missing_required`: descriptor missing required fields is rejected.
6. `tooldisc_0006_unsupported_msg_type`: unsupported TOOLDISC `msg_type` is rejected.

## SWP-RPC vectors

1. `rpc_0001_req_resp_success`: valid request receives valid terminal response.
2. `rpc_0002_error_retryability`: terminal error includes deterministic retryability signal.
3. `rpc_0003_stream_ordering`: stream items preserve sequence order per rpc_id.
4. `rpc_0004_terminal_closure`: no stream item accepted after terminal response/error.
5. `rpc_0005_idempotency_replay`: idempotency-key replay behavior is deterministic.
6. `rpc_0006_cancel_semantics`: cancellation yields deterministic terminal cancellation outcome.
7. `rpc_0007_unsupported_msg_type`: unsupported RPC `msg_type` is rejected.

## SWP-EVENTS vectors

1. `events_0001_required_fields`: publish event missing required fields is rejected.
2. `events_0002_ordering`: per-stream event ordering is preserved.
3. `events_0003_correlation_propagation`: correlation keys are preserved end-to-end.
4. `events_0004_batch_order`: batch message preserves internal event order.
5. `events_0005_invalid_severity`: invalid severity value is rejected.
6. `events_0006_unsupported_msg_type`: unsupported EVENTS `msg_type` is rejected.

## SWP-ARTIFACT vectors

1. `artifact_0001_offer_get_success`: artifact offer and get lifecycle succeeds.
2. `artifact_0002_chunk_ordering`: out-of-order chunk behavior is deterministic.
3. `artifact_0003_integrity_mismatch`: hash/integrity mismatch is rejected.
4. `artifact_0004_resume_token`: resume token behavior is deterministic.
5. `artifact_0005_range_request`: range request behavior is deterministic.
6. `artifact_0006_corruption_rejected`: corrupted chunk content is rejected.
7. `artifact_0007_unsupported_msg_type`: unsupported ARTIFACT `msg_type` is rejected.

## SWP-CRED vectors

1. `cred_0001_expiry_enforced`: expired credential/delegation is rejected.
2. `cred_0002_chain_length_enforced`: delegation chain length limits are enforced.
3. `cred_0003_invalid_credential`: invalid credential bytes/type are rejected.
4. `cred_0004_delegation_propagation`: valid delegation chain propagation is deterministic.
5. `cred_0005_revoke_behavior`: revocation behavior is deterministic when supported.
6. `cred_0006_unsupported_msg_type`: unsupported CRED `msg_type` is rejected.

## SWP-POLICYHINT vectors

1. `policyhint_0001_propagation`: policy hints propagate with deterministic scoping.
2. `policyhint_0002_unknown_keys`: unknown-key behavior follows documented policy.
3. `policyhint_0003_conflicts`: conflicting constraints are resolved deterministically.
4. `policyhint_0004_violation_payload`: violation payload includes key and scope references.
5. `policyhint_0005_mode_semantics`: MUST/SHOULD/MAY mode behavior is preserved.
6. `policyhint_0006_unsupported_msg_type`: unsupported POLICYHINT `msg_type` is rejected.

## SWP-STATE vectors

1. `state_0001_put_get_success`: state put/get lifecycle succeeds.
2. `state_0002_hash_mismatch`: hash/state_id mismatch is rejected.
3. `state_0003_parent_missing`: missing declared parent reference is rejected.
4. `state_0004_dag_valid`: valid parent DAG passes validation.
5. `state_0005_delta_opaque`: delta payload is treated as opaque and preserved.
6. `state_0006_unsupported_msg_type`: unsupported STATE `msg_type` is rejected.

## SWP-OBS vectors

1. `obs_0001_traceparent_validity`: invalid traceparent format is rejected.
2. `obs_0002_propagation`: trace context propagation is preserved across hops.
3. `obs_0003_tracestate_preservation`: unknown tracestate entries are not mutated.
4. `obs_0004_correlation_linkage`: trace context linkage to msg/task/rpc correlation is deterministic.
5. `obs_0005_unsupported_msg_type`: unsupported OBS `msg_type` is rejected.

## SWP-RELAY vectors

1. `relay_0001_at_least_once`: at-least-once delivery behavior is satisfied.
2. `relay_0002_dedupe_delivery_id`: duplicate deliveries are deduped deterministically by delivery_id.
3. `relay_0003_ack_timeout_retry`: ack timeout triggers deterministic retry behavior.
4. `relay_0004_dead_letter_reason`: dead-letter behavior includes deterministic reason code.
5. `relay_0005_limits`: relay limits behavior is deterministic under pressure.
6. `relay_0006_unsupported_msg_type`: unsupported RELAY `msg_type` is rejected.

## POC vectors

1. `poc_0001_valid_mcp_request`: valid E1 + MCP request frame decodes and validates.
2. `poc_0002_valid_swprpc_request`: valid E1 + SWP-RPC request frame decodes and validates.
3. `poc_0003_invalid_version`: unsupported core version is rejected deterministically.
4. `poc_0004_invalid_empty_msg_id`: empty `msg_id` fails envelope validation.
5. `poc_0005_invalid_truncated_frame`: truncated frame body is rejected as `INVALID_FRAME`.
6. `poc_0006_invalid_varint_overflow`: malformed/overflow varint is rejected as `INVALID_FRAME`.
