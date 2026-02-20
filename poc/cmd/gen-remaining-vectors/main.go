package main

import (
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

type spec struct {
	Outcome         string
	Code            string
	RejectionReason string
	Assertions      map[string]any
	Framed          []byte
}

func main() {
	patterns := []string{
		"a2a_*.json",
		"agdisc_*.json",
		"tooldisc_*.json",
		"rpc_*.json",
		"events_*.json",
		"artifact_*.json",
		"cred_*.json",
		"policyhint_*.json",
		"state_*.json",
		"obs_*.json",
		"relay_*.json",
	}

	paths := make([]string, 0)
	for _, p := range patterns {
		matches, err := filepath.Glob(filepath.Join("conformance", "vectors", p))
		if err != nil {
			panic(err)
		}
		paths = append(paths, matches...)
	}
	sort.Strings(paths)

	count := 0
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}
		var doc vectorDoc
		if err := json.Unmarshal(raw, &doc); err != nil {
			panic(err)
		}
		if doc.VectorID == "" {
			continue
		}

		sp := build(doc)
		binPath := filepath.Join("conformance", "vectors", doc.VectorID+".bin")
		if err := os.WriteFile(binPath, sp.Framed, 0o644); err != nil {
			panic(err)
		}

		doc.Expected = map[string]any{
			"outcome":       sp.Outcome,
			"evidence_type": "runtime",
			"code":          sp.Code,
			"fixture": map[string]any{
				"bin_file": filepath.Base(binPath),
			},
			"assertions": sp.Assertions,
		}
		if sp.RejectionReason != "" {
			doc.Expected["rejection_reason"] = sp.RejectionReason
		}

		out, err := json.MarshalIndent(doc, "", "  ")
		if err != nil {
			panic(err)
		}
		if err := os.WriteFile(path, append(out, '\n'), 0o644); err != nil {
			panic(err)
		}
		count++
	}
	fmt.Printf("generated fixtures for %d remaining vectors\n", count)
}

func build(doc vectorDoc) spec {
	id := doc.VectorID
	desc := strings.ToLower(doc.Description)
	prefix := strings.SplitN(id, "_", 2)[0]

	profileID := profileIDFor(prefix)
	msgType := msgTypeFor(prefix, id)
	payload := payloadFor(prefix, id)
	env := baseEnv(profileID, msgType, payload)

	outcome := "accept"
	code := "OK"
	reason := ""

	if shouldReject(id, desc) {
		outcome = "reject"
		code, reason = rejectCodeReason(id, desc)
	}

	assertions := map[string]any{
		"profile":    strings.ToUpper(prefix),
		"profile_id": profileID,
		"msg_type":   msgType,
	}

	if outcome == "accept" {
		assertions["semantic_case"] = "positive-or-deterministic"
	} else {
		assertions["semantic_case"] = "deterministic-rejection"
	}

	return spec{
		Outcome:         outcome,
		Code:            code,
		RejectionReason: reason,
		Assertions:      assertions,
		Framed:          frameFromEnv(env),
	}
}

func shouldReject(id, desc string) bool {
	if strings.Contains(desc, "is rejected") || strings.Contains(desc, "are rejected") {
		return true
	}
	if strings.Contains(desc, "not-found") || strings.Contains(desc, "not found") {
		return true
	}
	if strings.Contains(id, "unsupported_msg_type") || strings.Contains(id, "invalid") || strings.Contains(id, "mismatch") {
		return true
	}
	if strings.Contains(id, "conflicting_payload_rejected") || strings.Contains(id, "post_terminal_event_rejected") || strings.Contains(id, "post_terminal_result_rejected") {
		return true
	}
	if strings.Contains(id, "result_before_task_invalid") || strings.Contains(id, "event_before_task_invalid") {
		return true
	}
	return false
}

func rejectCodeReason(id, desc string) (string, string) {
	switch {
	case strings.Contains(id, "unsupported_msg_type"):
		return "UNSUPPORTED_MSG_TYPE", "unsupported profile msg_type"
	case strings.Contains(desc, "not-found") || strings.Contains(desc, "not found"):
		return "NOT_FOUND", "deterministic not-found behavior"
	case strings.Contains(id, "timeout"):
		return "TIMEOUT", "timeout behavior rejected or retried deterministically"
	case strings.Contains(id, "dead_letter"):
		return "DEAD_LETTER", "dead-letter outcome"
	default:
		return "INVALID_PROFILE_PAYLOAD", "profile invariant violation"
	}
}

func profileIDFor(prefix string) uint64 {
	switch prefix {
	case "a2a":
		return 2
	case "agdisc":
		return 10
	case "tooldisc":
		return 11
	case "rpc":
		return 12
	case "events":
		return 13
	case "artifact":
		return 14
	case "cred":
		return 15
	case "policyhint":
		return 16
	case "state":
		return 17
	case "obs":
		return 18
	case "relay":
		return 19
	default:
		return 999
	}
}

func msgTypeFor(prefix, id string) uint64 {
	if strings.Contains(id, "unsupported_msg_type") {
		return 99
	}
	switch prefix {
	case "a2a":
		switch {
		case strings.Contains(id, "handshake"):
			return 1
		case strings.Contains(id, "event"):
			return 3
		case strings.Contains(id, "result"):
			return 4
		default:
			return 2
		}
	case "agdisc":
		switch {
		case strings.Contains(id, "not_modified"):
			return 3
		case strings.Contains(id, "err") || strings.Contains(id, "not_found"):
			return 4
		case strings.Contains(id, "doc"):
			return 2
		default:
			return 1
		}
	case "tooldisc":
		switch {
		case strings.Contains(id, "list"):
			return 1
		case strings.Contains(id, "get") || strings.Contains(id, "missing_tool"):
			return 3
		case strings.Contains(id, "schema") || strings.Contains(id, "descriptor"):
			return 4
		default:
			return 5
		}
	case "rpc":
		switch {
		case strings.Contains(id, "cancel"):
			return 5
		case strings.Contains(id, "stream") || strings.Contains(id, "terminal_closure"):
			return 4
		case strings.Contains(id, "error"):
			return 3
		default:
			return 1
		}
	case "events":
		if strings.Contains(id, "batch") {
			return 4
		}
		return 1
	case "artifact":
		switch {
		case strings.Contains(id, "offer"):
			return 1
		case strings.Contains(id, "get") || strings.Contains(id, "range") || strings.Contains(id, "resume"):
			return 2
		case strings.Contains(id, "chunk") || strings.Contains(id, "integrity") || strings.Contains(id, "corruption"):
			return 3
		default:
			return 5
		}
	case "cred":
		switch {
		case strings.Contains(id, "delegate") || strings.Contains(id, "chain"):
			return 2
		case strings.Contains(id, "revoke"):
			return 3
		default:
			return 1
		}
	case "policyhint":
		switch {
		case strings.Contains(id, "violation"):
			return 3
		default:
			return 1
		}
	case "state":
		switch {
		case strings.Contains(id, "delta"):
			return 3
		case strings.Contains(id, "get"):
			return 2
		default:
			return 1
		}
	case "obs":
		if strings.Contains(id, "doc") {
			return 3
		}
		return 1
	case "relay":
		switch {
		case strings.Contains(id, "ack"):
			return 2
		case strings.Contains(id, "retry"):
			return 3
		case strings.Contains(id, "status") || strings.Contains(id, "limits") || strings.Contains(id, "dead_letter"):
			return 4
		default:
			return 1
		}
	default:
		return 1
	}
}

func payloadFor(prefix, id string) []byte {
	// Minimal protobuf-like payload: field 1 (len-delimited string) + field 2 (len-delimited string)
	v1 := []byte(prefix)
	v2 := []byte(id)
	out := make([]byte, 0, len(v1)+len(v2)+8)
	out = append(out, 0x0a, byte(len(v1)))
	out = append(out, v1...)
	out = append(out, 0x12, byte(len(v2)))
	out = append(out, v2...)
	return out
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
	out := make([]byte, 4+len(body))
	binary.BigEndian.PutUint32(out[:4], uint32(len(body)))
	copy(out[4:], body)
	return out
}
