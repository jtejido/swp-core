package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"swp-spec-kit/poc/internal/core"
)

type vectorDoc struct {
	VectorID    string         `json:"vector_id"`
	Group       string         `json:"group"`
	Category    string         `json:"category"`
	Description string         `json:"description"`
	Expected    map[string]any `json:"expected"`
}

type fixtureSpec struct {
	outcome         string
	evidenceType    string
	code            string
	rejectionReason string
	assertions      map[string]any
	framed          []byte
}

func main() {
	patterns := []string{"core_*.json", "e1_*.json", "s1_*.json", "mcp_*.json"}
	paths := make([]string, 0)
	for _, p := range patterns {
		matches, err := filepath.Glob(filepath.Join("conformance", "vectors", p))
		if err != nil {
			panic(err)
		}
		paths = append(paths, matches...)
	}
	sort.Strings(paths)

	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}
		var doc vectorDoc
		if err := json.Unmarshal(raw, &doc); err != nil {
			panic(err)
		}

		spec := buildSpec(doc.VectorID)
		binFile := filepath.Join("conformance", "vectors", doc.VectorID+".bin")
		if spec.evidenceType == "runtime" {
			if len(spec.framed) == 0 {
				panic(fmt.Errorf("runtime vector missing framed bytes: %s", doc.VectorID))
			}
			if err := os.WriteFile(binFile, spec.framed, 0o644); err != nil {
				panic(err)
			}
		}

		doc.Expected = map[string]any{
			"outcome":       spec.outcome,
			"evidence_type": spec.evidenceType,
			"code":          spec.code,
			"assertions":    spec.assertions,
		}
		if spec.rejectionReason != "" {
			doc.Expected["rejection_reason"] = spec.rejectionReason
		}
		if spec.evidenceType == "runtime" {
			doc.Expected["fixture"] = map[string]any{"bin_file": filepath.Base(binFile)}
		} else {
			doc.Expected["fixture"] = map[string]any{"evidence_file": doc.VectorID + ".evidence.md"}
		}

		out, err := json.MarshalIndent(doc, "", "  ")
		if err != nil {
			panic(err)
		}
		if err := os.WriteFile(path, append(out, '\n'), 0o644); err != nil {
			panic(err)
		}
	}

	fmt.Printf("generated C1 fixtures for %d vectors\n", len(paths))
}

func buildSpec(id string) fixtureSpec {
	if id == "core_0033_extensibility_backward_compatibility" {
		return fixtureSpec{
			outcome:      "process_check",
			evidenceType: "process",
			code:         "COMPATIBILITY_POLICY",
			assertions: map[string]any{
				"policy":   "backward-compatibility-review",
				"artifact": "docs/versioning-compatibility.md",
			},
		}
	}

	if strings.HasPrefix(id, "core_") {
		return coreSpec(id)
	}
	if strings.HasPrefix(id, "e1_") {
		return e1Spec(id)
	}
	if strings.HasPrefix(id, "mcp_") {
		return mcpSpec(id)
	}
	if strings.HasPrefix(id, "s1_") {
		return s1Spec(id)
	}
	panic(fmt.Errorf("unexpected vector id: %s", id))
}

func coreSpec(id string) fixtureSpec {
	env := baseEnv(1, 1, []byte(`{"k":"v"}`))
	outcome := "accept"
	code := "OK"
	rejectReason := ""
	assertions := map[string]any{"envelope": map[string]any{"version": 1, "profile_id": 1, "msg_type": 1, "min_msg_id_bytes": 8}}
	framed := frameFromEnv(env)

	switch id {
	case "core_0001_valid_min_frame":
		env.Payload = []byte{}
		framed = frameFromEnv(env)
		assertions["envelope"] = map[string]any{"version": 1, "profile_id": 1, "msg_type": 1, "payload_len": 0}
	case "core_0002_valid_typical_frame":
		framed = frameFromEnv(env)
	case "core_0003_invalid_zero_length":
		outcome, code, rejectReason = "reject", "INVALID_FRAME", "zero-length frame prefix"
		framed = []byte{0, 0, 0, 0}
	case "core_0004_invalid_truncated_prefix":
		outcome, code, rejectReason = "reject", "INVALID_FRAME", "truncated length prefix"
		framed = []byte{0, 0}
	case "core_0005_invalid_oversized_length":
		outcome, code, rejectReason = "reject", "INVALID_FRAME", "frame length exceeds MAX_FRAME_BYTES"
		framed = make([]byte, 4)
		binary.BigEndian.PutUint32(framed, core.DefaultLimits().MaxFrameBytes+1)
	case "core_0006_invalid_truncated_body":
		outcome, code, rejectReason = "reject", "INVALID_FRAME", "frame body truncated"
		framed = append([]byte{0, 0, 0, 10}, []byte{1, 2, 3}...)
	case "core_0007_invalid_envelope_decode":
		outcome, code, rejectReason = "reject", "INVALID_FRAME", "malformed envelope encoding"
		body := []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80}
		framed = frameFromBody(body)
	case "core_0008_unsupported_version":
		outcome, code, rejectReason = "reject", "UNSUPPORTED_VERSION", "unsupported core version"
		env.Version = 2
		framed = frameFromEnv(env)
	case "core_0009_unknown_profile":
		outcome, code, rejectReason = "reject", "UNKNOWN_PROFILE", "unknown profile_id"
		env.ProfileID = 9999
		framed = frameFromEnv(env)
	case "core_0010_invalid_msg_id_short":
		outcome, code, rejectReason = "reject", "INVALID_ENVELOPE", "msg_id shorter than MIN_MSG_ID_BYTES"
		env.MsgID = []byte("1234")
		framed = frameFromEnv(env)
	case "core_0011_invalid_msg_id_long":
		outcome, code, rejectReason = "reject", "INVALID_ENVELOPE", "msg_id longer than MAX_MSG_ID_BYTES"
		env.MsgID = bytes.Repeat([]byte{'a'}, 65)
		framed = frameFromEnv(env)
	case "core_0012_invalid_payload_oversize":
		outcome, code, rejectReason = "reject", "INVALID_ENVELOPE", "payload exceeds configured MAX_PAYLOAD_BYTES"
		env.Payload = bytes.Repeat([]byte{'p'}, 1025)
		framed = frameFromEnv(env)
		assertions["limits"] = map[string]any{"max_payload_bytes": 1024}
	case "core_0013_unknown_flags_set":
		env.Flags = 1 << 20
		framed = frameFromEnv(env)
	case "core_0014_stale_timestamp":
		outcome, code, rejectReason = "reject", "INVALID_ENVELOPE", "timestamp older than freshness window"
		env.TsUnixMs = uint64(time.Now().Add(-24 * time.Hour).UnixMilli())
		framed = frameFromEnv(env)
	case "core_0015_future_timestamp":
		outcome, code, rejectReason = "reject", "INVALID_ENVELOPE", "timestamp too far in the future"
		env.TsUnixMs = uint64(time.Now().Add(24 * time.Hour).UnixMilli())
		framed = frameFromEnv(env)
	case "core_0016_burst_limit_exceeded":
		outcome, code, rejectReason = "reject", "RATE_LIMIT_EXCEEDED", "burst limit exceeded by valid frames"
		framed = frameFromEnv(env)
	case "core_0017_unknown_profile_with_error_path":
		outcome, code, rejectReason = "reject", "UNKNOWN_PROFILE", "receiver returns unknown profile error"
		env.ProfileID = 9999
		framed = frameFromEnv(env)
	case "core_0018_unknown_profile_no_error_path":
		outcome, code, rejectReason = "reject", "UNKNOWN_PROFILE", "receiver closes stream for unknown profile without response"
		env.ProfileID = 9999
		framed = frameFromEnv(env)
	case "core_0019_boundary_max_frame_exact":
		env.Payload = bytes.Repeat([]byte{'f'}, 2048)
		framed = frameFromEnv(env)
		assertions["limits"] = map[string]any{"max_frame_bytes": len(framed) - 4}
	case "core_0020_boundary_max_payload_exact":
		env.Payload = bytes.Repeat([]byte{'p'}, 2048)
		framed = frameFromEnv(env)
		assertions["limits"] = map[string]any{"max_payload_bytes": len(env.Payload)}
	case "core_0021_missing_required_field_version":
		outcome, code, rejectReason = "reject", "UNSUPPORTED_VERSION", "missing or invalid version"
		env.Version = 0
		framed = frameFromEnv(env)
	case "core_0022_missing_required_field_profile_id":
		outcome, code, rejectReason = "reject", "UNKNOWN_PROFILE", "missing or invalid profile_id"
		env.ProfileID = 0
		framed = frameFromEnv(env)
	case "core_0023_missing_required_field_msg_type":
		outcome, code, rejectReason = "reject", "INVALID_ENVELOPE", "missing or invalid msg_type"
		env.MsgType = 0
		framed = frameFromEnv(env)
	case "core_0024_missing_required_field_msg_id":
		outcome, code, rejectReason = "reject", "INVALID_ENVELOPE", "missing or empty msg_id"
		env.MsgID = nil
		framed = frameFromEnv(env)
	case "core_0025_missing_required_field_ts_unix_ms":
		outcome, code, rejectReason = "reject", "INVALID_ENVELOPE", "missing/invalid timestamp when policy requires it"
		env.TsUnixMs = 0
		framed = frameFromEnv(env)
		assertions["policy"] = map[string]any{"timestamp_required": true}
	case "core_0026_unknown_flags_no_reinterpretation":
		env.Flags = 1 << 27
		framed = frameFromEnv(env)
	case "core_0027_duplicate_inflight_msg_id":
		outcome, code, rejectReason = "reject", "DUPLICATE_MSG_ID", "duplicate in-flight msg_id detected"
		framed = frameFromEnv(env)
	case "core_0028_error_mapping_invalid_frame":
		outcome, code, rejectReason = "reject", "INVALID_FRAME", "framing failure mapped to INVALID_FRAME"
		framed = []byte{0, 0, 0, 0}
	case "core_0029_error_mapping_unsupported_version":
		outcome, code, rejectReason = "reject", "UNSUPPORTED_VERSION", "unsupported version mapped deterministically"
		env.Version = 2
		framed = frameFromEnv(env)
	case "core_0030_error_mapping_invalid_envelope":
		outcome, code, rejectReason = "reject", "INVALID_ENVELOPE", "envelope invariant failure mapped deterministically"
		env.MsgID = []byte{}
		framed = frameFromEnv(env)
	case "core_0031_optional_fields_no_semantic_override":
		env.Extensions = []core.Extension{{Type: 2001, Value: []byte("x")}}
		framed = frameFromEnv(env)
	case "core_0032_profile_dispatch_known_profile":
		env.ProfileID = 1
		framed = frameFromEnv(env)
	default:
		panic(fmt.Errorf("unknown core vector: %s", id))
	}

	return fixtureSpec{outcome: outcome, evidenceType: "runtime", code: code, rejectionReason: rejectReason, assertions: assertions, framed: framed}
}

func e1Spec(id string) fixtureSpec {
	env := baseEnv(1, 1, []byte("e1"))
	outcome := "accept"
	code := "OK"
	rejectReason := ""
	assertions := map[string]any{"encoding": "E1"}
	framed := frameFromEnv(env)

	switch id {
	case "e1_0001_valid_min_envelope":
		framed = frameFromEnv(baseEnv(1, 1, []byte{}))
	case "e1_0002_varint_too_long_invalid":
		outcome, code, rejectReason = "reject", "INVALID_FRAME", "varint longer than 10 bytes"
		body := []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80}
		framed = frameFromBody(body)
	case "e1_0003_varint_overflow_invalid":
		outcome, code, rejectReason = "reject", "INVALID_FRAME", "uvarint overflow"
		body := []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x02}
		framed = frameFromBody(body)
	case "e1_0004_invalid_version":
		outcome, code, rejectReason = "reject", "UNSUPPORTED_VERSION", "E1 version must equal 1"
		env.Version = 2
		framed = frameFromEnv(env)
	case "e1_0005_empty_msg_id_invalid":
		outcome, code, rejectReason = "reject", "INVALID_ENVELOPE", "msg_id is empty"
		env.MsgID = []byte{}
		framed = frameFromEnv(env)
	case "e1_0006_unknown_extension_ignored":
		env.Extensions = []core.Extension{{Type: 4097, Value: []byte("opaque")}}
		framed = frameFromEnv(env)
	case "e1_0007_extensions_too_large":
		outcome, code, rejectReason = "reject", "INVALID_ENVELOPE", "extensions exceed MAX_EXT_BYTES"
		env.Extensions = []core.Extension{{Type: 16, Value: bytes.Repeat([]byte{'x'}, 5000)}}
		framed = frameFromEnv(env)
	case "e1_0008_truncated_bytes_field":
		outcome, code, rejectReason = "reject", "INVALID_FRAME", "truncated bytes field"
		body, _ := core.EncodeEnvelopeE1(env)
		framed = frameFromBody(body[:len(body)-1])
	default:
		panic(fmt.Errorf("unknown e1 vector: %s", id))
	}

	return fixtureSpec{outcome: outcome, evidenceType: "runtime", code: code, rejectionReason: rejectReason, assertions: assertions, framed: framed}
}

func mcpSpec(id string) fixtureSpec {
	env := baseEnv(1, 1, []byte(`{"jsonrpc":"2.0","id":"1","method":"tools/list","params":{}}`))
	outcome := "accept"
	code := "OK"
	rejectReason := ""
	assertions := map[string]any{"profile": "MCPMAP", "msg_type": env.MsgType}
	framed := frameFromEnv(env)

	switch id {
	case "mcp_0001_request_roundtrip":
		env.MsgType = 1
		framed = frameFromEnv(env)
	case "mcp_0002_response_roundtrip":
		env.MsgType = 2
		env.Payload = []byte(`{"jsonrpc":"2.0","id":"1","result":{"ok":true}}`)
		framed = frameFromEnv(env)
	case "mcp_0003_notification_no_response":
		env.MsgType = 3
		env.Payload = []byte(`{"jsonrpc":"2.0","method":"notify","params":{}}`)
		framed = frameFromEnv(env)
	case "mcp_0004_error_mapping_internal":
		env.MsgType = 2
		env.Payload = []byte(`{"jsonrpc":"2.0","id":"1","error":{"code":-32603,"message":"internal"}}`)
		framed = frameFromEnv(env)
		assertions["mapping"] = "INTERNAL_ERROR -> -32603"
	case "mcp_0005_error_mapping_unknown_profile":
		env.MsgType = 2
		env.Payload = []byte(`{"jsonrpc":"2.0","id":"1","error":{"code":-32601,"message":"method not found"}}`)
		framed = frameFromEnv(env)
		assertions["mapping"] = "UNKNOWN_PROFILE -> -32601"
	case "mcp_0006_response_msg_id_reuse":
		env.MsgType = 2
		env.Payload = []byte(`{"jsonrpc":"2.0","id":"reuse","result":{"ok":true}}`)
		framed = frameFromEnv(env)
	case "mcp_0007_request_missing_id_invalid":
		outcome, code, rejectReason = "reject", "INVALID_MCP_PAYLOAD", "generated request missing JSON-RPC id"
		env.MsgType = 1
		env.Payload = []byte(`{"jsonrpc":"2.0","method":"tools/list","params":{}}`)
		framed = frameFromEnv(env)
	case "mcp_0008_response_both_result_and_error_invalid":
		outcome, code, rejectReason = "reject", "INVALID_MCP_PAYLOAD", "response contains both result and error"
		env.MsgType = 2
		env.Payload = []byte(`{"jsonrpc":"2.0","id":"1","result":{},"error":{"code":-32000,"message":"x"}}`)
		framed = frameFromEnv(env)
	case "mcp_0009_notification_with_response_violation":
		env.MsgType = 3
		env.Payload = []byte(`{"jsonrpc":"2.0","method":"notify","params":{"k":"v"}}`)
		framed = frameFromEnv(env)
	case "mcp_0010_invalid_utf8_payload":
		outcome, code, rejectReason = "reject", "INVALID_MCP_PAYLOAD", "payload is not valid UTF-8"
		env.MsgType = 1
		env.Payload = []byte{0xff, 0xfe, 0xfd}
		framed = frameFromEnv(env)
	case "mcp_0011_invalid_json_payload":
		outcome, code, rejectReason = "reject", "INVALID_MCP_PAYLOAD", "payload is not valid JSON"
		env.MsgType = 1
		env.Payload = []byte(`{"jsonrpc":`)
		framed = frameFromEnv(env)
	case "mcp_0012_unsupported_msg_type":
		outcome, code, rejectReason = "reject", "UNSUPPORTED_MSG_TYPE", "unsupported MCP msg_type"
		env.MsgType = 99
		framed = frameFromEnv(env)
	case "mcp_0013_stream_partial_then_terminal_success":
		env.MsgType = 3
		env.Payload = []byte(`{"jsonrpc":"2.0","method":"partial","params":{"seq":1}}`)
		framed = frameFromEnv(env)
		assertions["scenario"] = "partial notifications followed by terminal success response"
	case "mcp_0014_stream_partial_then_terminal_error":
		env.MsgType = 3
		env.Payload = []byte(`{"jsonrpc":"2.0","method":"partial","params":{"seq":1}}`)
		framed = frameFromEnv(env)
		assertions["scenario"] = "partial notifications followed by terminal error response"
	case "mcp_0015_payload_preservation_whitespace":
		env.MsgType = 1
		env.Payload = []byte("{\n  \"jsonrpc\" : \"2.0\", \"id\" : \"1\", \"method\": \"tools/list\", \"params\": {}\n}\n")
		framed = frameFromEnv(env)
	default:
		panic(fmt.Errorf("unknown mcp vector: %s", id))
	}
	assertions["msg_type"] = env.MsgType
	return fixtureSpec{outcome: outcome, evidenceType: "runtime", code: code, rejectionReason: rejectReason, assertions: assertions, framed: framed}
}

func s1Spec(id string) fixtureSpec {
	env := baseEnv(1, 1, []byte(`{"jsonrpc":"2.0","id":"s1","method":"tools/list"}`))
	outcome := "reject"
	code := "SECURITY_POLICY"
	rejectReason := "channel security policy rejection"
	assertions := map[string]any{"binding": "S1"}
	framed := frameFromEnv(env)

	switch id {
	case "s1_0001_unauthenticated_peer_rejected":
		rejectReason = "peer is unauthenticated before frame processing"
	case "s1_0002_authentication_failure_terminates":
		rejectReason = "authentication failure terminates channel"
	case "s1_0003_downgrade_policy_failure":
		rejectReason = "downgrade policy violation terminates channel"
	case "s1_0004_integrity_failure_terminates":
		rejectReason = "integrity/auth failure terminates channel"
	case "s1_0005_timestamp_freshness_enforced":
		rejectReason = "stale timestamp rejected when freshness policy is enabled"
		env.TsUnixMs = uint64(time.Now().Add(-24 * time.Hour).UnixMilli())
		framed = frameFromEnv(env)
	case "s1_0006_timestamp_freshness_disabled_documented":
		outcome = "accept"
		code = "OK"
		rejectReason = ""
		assertions["policy"] = "timestamp freshness disabled and documented"
		env.TsUnixMs = uint64(time.Now().Add(-24 * time.Hour).UnixMilli())
		framed = frameFromEnv(env)
	default:
		panic(fmt.Errorf("unknown s1 vector: %s", id))
	}

	return fixtureSpec{outcome: outcome, evidenceType: "runtime", code: code, rejectionReason: rejectReason, assertions: assertions, framed: framed}
}

func baseEnv(profileID, msgType uint64, payload []byte) core.Envelope {
	return core.Envelope{
		Version:   1,
		ProfileID: profileID,
		MsgType:   msgType,
		Flags:     0,
		TsUnixMs:  uint64(time.Now().UnixMilli()),
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   payload,
	}
}

func frameFromEnv(env core.Envelope) []byte {
	body, err := core.EncodeEnvelopeE1(env)
	if err != nil {
		panic(err)
	}
	return frameFromBody(body)
}

func frameFromBody(body []byte) []byte {
	out := make([]byte, 4+len(body))
	binary.BigEndian.PutUint32(out[:4], uint32(len(body)))
	copy(out[4:], body)
	return out
}
