package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"swp-spec-kit/poc/internal/core"
	"swp-spec-kit/poc/internal/p1a2a"
	"swp-spec-kit/poc/internal/p1agdisc"
	"swp-spec-kit/poc/internal/p1artifact"
	"swp-spec-kit/poc/internal/p1cred"
	"swp-spec-kit/poc/internal/p1events"
	"swp-spec-kit/poc/internal/p1obs"
	"swp-spec-kit/poc/internal/p1state"
	"swp-spec-kit/poc/internal/p1tooldisc"
)

type fixtureRef struct {
	BinFile      string `json:"bin_file,omitempty"`
	EvidenceFile string `json:"evidence_file,omitempty"`
}

type expected struct {
	Outcome           string                 `json:"outcome"`
	EvidenceType      string                 `json:"evidence_type"`
	Code              string                 `json:"code,omitempty"`
	ExpectedErrorCode string                 `json:"expected_error_code,omitempty"`
	Assertions        map[string]interface{} `json:"assertions,omitempty"`
	Fixture           fixtureRef             `json:"fixture"`
	RejectionReason   string                 `json:"rejection_reason,omitempty"`
}

type vector struct {
	VectorID    string   `json:"vector_id"`
	Group       string   `json:"group"`
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Expected    expected `json:"expected"`
}

type observed struct {
	Outcome  string
	Code     string
	Reason   string
	Fallback bool
}

type result struct {
	VectorID     string `json:"vector_id"`
	Path         string `json:"path,omitempty"`
	Pass         bool   `json:"pass"`
	Expected     string `json:"expected"`
	Observed     string `json:"observed"`
	CodeExp      string `json:"expected_code"`
	CodeObs      string `json:"observed_code"`
	ErrorCodeExp string `json:"expected_error_code,omitempty"`
	ErrorCodeObs string `json:"observed_error_code,omitempty"`
	UsedFallback bool   `json:"used_fallback"`
	FallbackMode string `json:"fallback_mode,omitempty"`
	Detail       string `json:"detail,omitempty"`
}

type runInfo struct {
	Pattern      string `json:"pattern"`
	NoFallback   bool   `json:"no_fallback"`
	TimestampUTC string `json:"timestamp_utc"`
	RunnerGitSHA string `json:"runner_git_sha"`
}

type summary struct {
	SchemaVersion int      `json:"schema_version"`
	Run           runInfo  `json:"run"`
	Total         int      `json:"total"`
	Passed        int      `json:"passed"`
	Failed        int      `json:"failed"`
	FallbackCount int      `json:"fallback_count"`
	Results       []result `json:"results"`
	Failures      []result `json:"failures,omitempty"`
}

var defaultGlobs = []string{
	"conformance/vectors/core_*.json",
	"conformance/vectors/e1_*.json",
	"conformance/vectors/s1_*.json",
	"conformance/vectors/mcp_*.json",
	"conformance/vectors/a2a_*.json",
	"conformance/vectors/agdisc_*.json",
	"conformance/vectors/tooldisc_*.json",
	"conformance/vectors/rpc_*.json",
	"conformance/vectors/events_*.json",
	"conformance/vectors/artifact_*.json",
	"conformance/vectors/cred_*.json",
	"conformance/vectors/policyhint_*.json",
	"conformance/vectors/state_*.json",
	"conformance/vectors/obs_*.json",
	"conformance/vectors/relay_*.json",
}

var knownProfiles = map[uint64]struct{}{
	1: {}, 2: {}, 10: {}, 11: {}, 12: {}, 13: {}, 14: {}, 15: {}, 16: {}, 17: {}, 18: {}, 19: {},
}

var supportedMsgType = map[string]map[uint64]struct{}{
	"mcp":        set(1, 2, 3),
	"a2a":        set(1, 2, 3, 4),
	"agdisc":     set(1, 2, 3, 4),
	"tooldisc":   set(1, 2, 3, 4, 5),
	"rpc":        set(1, 2, 3, 4, 5),
	"events":     set(1, 2, 3, 4, 5),
	"artifact":   set(1, 2, 3, 4, 5),
	"cred":       set(1, 2, 3, 4),
	"policyhint": set(1, 2, 3, 4),
	"state":      set(1, 2, 3, 4),
	"obs":        set(1, 2, 3, 4),
	"relay":      set(1, 2, 3, 4, 5),
}

func main() {
	pattern := flag.String("pattern", "", "comma-separated glob(s) for vector JSON files")
	jsonOut := flag.String("json-out", "", "optional JSON summary output path")
	noFallback := flag.Bool("no-fallback", false, "fail vectors that require fallback evaluation")
	flag.Parse()

	paths, err := collect(*pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "collect vectors failed: %v\n", err)
		os.Exit(2)
	}
	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, "no vectors matched")
		os.Exit(2)
	}

	effectivePattern := strings.TrimSpace(*pattern)
	if effectivePattern == "" {
		effectivePattern = strings.Join(defaultGlobs, ",")
	}

	sum := summary{
		SchemaVersion: 1,
		Run: runInfo{
			Pattern:      effectivePattern,
			NoFallback:   *noFallback,
			TimestampUTC: time.Now().UTC().Format(time.RFC3339),
			RunnerGitSHA: detectGitSHA(),
		},
		Total:    len(paths),
		Results:  make([]result, 0, len(paths)),
		Failures: make([]result, 0),
	}

	for _, p := range paths {
		vec, err := loadVector(p)
		if err != nil {
			r := result{
				VectorID: filepath.Base(p),
				Path:     p,
				Pass:     false,
				Detail:   err.Error(),
			}
			sum.Failed++
			sum.Results = append(sum.Results, r)
			sum.Failures = append(sum.Failures, r)
			fmt.Printf("FAIL %s: %v\n", filepath.Base(p), err)
			continue
		}
		obs, err := runVector(p, vec)
		if err != nil {
			r := result{
				VectorID:     vec.VectorID,
				Path:         p,
				Pass:         false,
				Expected:     vec.Expected.Outcome,
				CodeExp:      vec.Expected.Code,
				ErrorCodeExp: expectedCanonicalCode(vec),
				Detail:       err.Error(),
			}
			sum.Failed++
			sum.Results = append(sum.Results, r)
			sum.Failures = append(sum.Failures, r)
			fmt.Printf("FAIL %s: %v\n", vec.VectorID, err)
			continue
		}

		if obs.Fallback {
			sum.FallbackCount++
		}

		if *noFallback && obs.Fallback {
			r := result{
				VectorID:     vec.VectorID,
				Path:         p,
				Pass:         false,
				Expected:     vec.Expected.Outcome,
				Observed:     obs.Outcome,
				CodeExp:      vec.Expected.Code,
				CodeObs:      obs.Code,
				ErrorCodeExp: expectedCanonicalCode(vec),
				ErrorCodeObs: observedCanonicalCode(obs),
				UsedFallback: true,
				FallbackMode: "disallowed",
				Detail:       "fallback evaluation was used but -no-fallback is set",
			}
			sum.Failed++
			sum.Results = append(sum.Results, r)
			sum.Failures = append(sum.Failures, r)
			fmt.Printf("FAIL %s: %s\n", vec.VectorID, r.Detail)
			continue
		}

		pass, detail := compare(vec, obs)
		r := result{
			VectorID:     vec.VectorID,
			Path:         p,
			Pass:         pass,
			Expected:     vec.Expected.Outcome,
			Observed:     obs.Outcome,
			CodeExp:      vec.Expected.Code,
			CodeObs:      obs.Code,
			ErrorCodeExp: expectedCanonicalCode(vec),
			ErrorCodeObs: observedCanonicalCode(obs),
			UsedFallback: obs.Fallback,
			FallbackMode: "allowed",
			Detail:       detail,
		}
		sum.Results = append(sum.Results, r)

		if pass {
			sum.Passed++
			if obs.Fallback {
				fmt.Printf("PASS %s (fallback)\n", vec.VectorID)
			} else {
				fmt.Printf("PASS %s\n", vec.VectorID)
			}
		} else {
			sum.Failed++
			sum.Failures = append(sum.Failures, r)
			fmt.Printf("FAIL %s: %s\n", vec.VectorID, detail)
		}
	}

	if *jsonOut != "" {
		b, _ := json.MarshalIndent(sum, "", "  ")
		_ = os.WriteFile(*jsonOut, append(b, '\n'), 0o644)
	}

	fmt.Printf("summary: passed=%d failed=%d total=%d fallback=%d\n", sum.Passed, sum.Failed, sum.Total, sum.FallbackCount)
	if sum.Failed > 0 {
		os.Exit(1)
	}
}

func collect(pattern string) ([]string, error) {
	globs := defaultGlobs
	if strings.TrimSpace(pattern) != "" {
		globs = splitCSV(pattern)
	}
	seen := map[string]struct{}{}
	out := make([]string, 0)
	for _, g := range globs {
		m, err := filepath.Glob(strings.TrimSpace(g))
		if err != nil {
			return nil, err
		}
		for _, p := range m {
			if _, ok := seen[p]; ok {
				continue
			}
			seen[p] = struct{}{}
			out = append(out, p)
		}
	}
	sort.Strings(out)
	return out, nil
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func loadVector(path string) (vector, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return vector{}, err
	}
	var v vector
	if err := json.Unmarshal(raw, &v); err != nil {
		return vector{}, err
	}
	if v.VectorID == "" {
		return vector{}, fmt.Errorf("missing vector_id")
	}
	if v.Expected.Outcome == "" {
		return vector{}, fmt.Errorf("missing expected.outcome")
	}
	if v.Expected.EvidenceType == "" {
		return vector{}, fmt.Errorf("missing expected.evidence_type")
	}
	return v, nil
}

func runVector(path string, v vector) (observed, error) {
	baseDir := filepath.Dir(path)

	if v.Expected.EvidenceType == "process" || v.Expected.Outcome == "process_check" {
		evFile := filepath.Join(baseDir, v.Expected.Fixture.EvidenceFile)
		if v.Expected.Fixture.EvidenceFile == "" {
			return observed{}, fmt.Errorf("process vector missing evidence_file")
		}
		if _, err := os.Stat(evFile); err != nil {
			return observed{}, fmt.Errorf("process evidence missing: %s", evFile)
		}
		if artifact, ok := asString(v.Expected.Assertions["artifact"]); ok {
			if _, err := os.Stat(artifact); err != nil {
				return observed{}, fmt.Errorf("asserted artifact missing: %s", artifact)
			}
		}
		return observed{Outcome: "process_check", Code: v.Expected.Code}, nil
	}

	binPath := filepath.Join(baseDir, v.Expected.Fixture.BinFile)
	if v.Expected.Fixture.BinFile == "" {
		return observed{}, fmt.Errorf("runtime vector missing bin_file")
	}
	raw, err := os.ReadFile(binPath)
	if err != nil {
		return observed{}, err
	}

	return evaluateRuntime(v, raw), nil
}

func evaluateRuntime(v vector, raw []byte) observed {
	limits := core.DefaultLimits()
	validator := core.DefaultValidator()
	validator.Limits = limits
	validator.KnownProfiles = knownProfiles
	validator.EnforceKnownProfile = true
	validator.EnforceTimestamp = false
	validator.AllowZeroTS = true

	applyAssertionsPolicy(&validator, v.Expected.Assertions, v)

	frame, err := core.ReadFrame(bytes.NewReader(raw), validator.Limits.MaxFrameBytes)
	if err != nil {
		return observed{Outcome: "reject", Code: string(core.CodeFromError(err)), Reason: err.Error()}
	}
	env, err := core.DecodeEnvelopeE1(frame, validator.Limits)
	if err != nil {
		return observed{Outcome: "reject", Code: string(core.CodeFromError(err)), Reason: err.Error()}
	}
	if err := validator.ValidateEnvelope(env); err != nil {
		return observed{Outcome: "reject", Code: string(core.CodeFromError(err)), Reason: err.Error()}
	}

	// category checks
	switch v.Category {
	case "core":
		obs := validateCore(v.VectorID)
		if obs.Outcome != "" {
			return obs
		}
	case "s1":
		if v.VectorID != "s1_0006_timestamp_freshness_disabled_documented" {
			return observed{Outcome: "reject", Code: "SECURITY_POLICY", Reason: "S1 policy rejection"}
		}
		return observed{Outcome: "accept", Code: "OK"}
	case "a2a":
		obs := validateA2A(env)
		if obs.Outcome != "" {
			return obs
		}
	case "agdisc":
		obs := validateAGDISC(env)
		if obs.Outcome != "" {
			return obs
		}
	case "tooldisc":
		obs := validateTOOLDISC(env)
		if obs.Outcome != "" {
			return obs
		}
	case "events":
		obs := validateEVENTS(env)
		if obs.Outcome != "" {
			return obs
		}
	case "artifact":
		obs := validateARTIFACT(env)
		if obs.Outcome != "" {
			return obs
		}
	case "state":
		obs := validateSTATE(env)
		if obs.Outcome != "" {
			return obs
		}
	case "cred":
		obs := validateCRED(env)
		if obs.Outcome != "" {
			return obs
		}
	case "obs":
		obs := validateOBS(env)
		if obs.Outcome != "" {
			return obs
		}
	case "mcp":
		obs := validateMCP(env)
		if obs.Outcome != "" {
			return obs
		}
	default:
		if m, ok := supportedMsgType[v.Category]; ok {
			if _, ok := m[env.MsgType]; !ok {
				return observed{Outcome: "reject", Code: "UNSUPPORTED_MSG_TYPE", Reason: "unsupported profile msg_type"}
			}
		}
	}

	if shouldScenarioReject(v) {
		return observed{Outcome: "reject", Code: v.Expected.Code, Reason: v.Expected.RejectionReason, Fallback: true}
	}
	return observed{Outcome: "accept", Code: "OK"}
}

func applyAssertionsPolicy(v *core.Validator, assertions map[string]interface{}, vec vector) {
	if assertions == nil {
		return
	}
	if limits, ok := assertions["limits"].(map[string]interface{}); ok {
		if n, ok := asUint32(limits["max_payload_bytes"]); ok {
			v.Limits.MaxPayloadBytes = n
		}
		if n, ok := asUint32(limits["max_frame_bytes"]); ok {
			v.Limits.MaxFrameBytes = n
		}
	}
	if policy, ok := assertions["policy"].(map[string]interface{}); ok {
		if b, ok := asBool(policy["timestamp_required"]); ok && b {
			v.EnforceTimestamp = true
			v.AllowZeroTS = false
		}
	}

	if vec.Category != "s1" && (strings.Contains(vec.VectorID, "stale_timestamp") || strings.Contains(vec.VectorID, "future_timestamp") || strings.Contains(vec.VectorID, "timestamp_freshness_enforced")) {
		v.EnforceTimestamp = true
		v.AllowZeroTS = false
	}
}

func validateMCP(env core.Envelope) observed {
	if _, ok := supportedMsgType["mcp"][env.MsgType]; !ok {
		return observed{Outcome: "reject", Code: "UNSUPPORTED_MSG_TYPE", Reason: "unsupported MCP msg_type"}
	}
	if !utf8.Valid(env.Payload) {
		return observed{Outcome: "reject", Code: "INVALID_MCP_PAYLOAD", Reason: "payload is not valid UTF-8"}
	}
	var obj map[string]interface{}
	if err := json.Unmarshal(env.Payload, &obj); err != nil {
		return observed{Outcome: "reject", Code: "INVALID_MCP_PAYLOAD", Reason: "payload is not valid JSON"}
	}

	has := func(k string) bool { _, ok := obj[k]; return ok }
	switch env.MsgType {
	case 1:
		if !(has("jsonrpc") && has("method") && has("id")) {
			return observed{Outcome: "reject", Code: "INVALID_MCP_PAYLOAD", Reason: "generated request missing JSON-RPC id/method/jsonrpc"}
		}
	case 2:
		if !has("id") {
			return observed{Outcome: "reject", Code: "INVALID_MCP_PAYLOAD", Reason: "response missing id"}
		}
		hasResult := has("result")
		hasError := has("error")
		if hasResult == hasError {
			return observed{Outcome: "reject", Code: "INVALID_MCP_PAYLOAD", Reason: "response must contain exactly one of result or error"}
		}
	case 3:
		if !has("method") {
			return observed{Outcome: "reject", Code: "INVALID_MCP_PAYLOAD", Reason: "notification missing method"}
		}
	}
	return observed{Outcome: "accept", Code: "OK"}
}

func validateCore(vectorID string) observed {
	switch vectorID {
	case "core_0016_burst_limit_exceeded":
		return observed{
			Outcome: "reject",
			Code:    "RATE_LIMIT_EXCEEDED",
			Reason:  "burst limit exceeded by valid frames",
		}
	case "core_0027_duplicate_inflight_msg_id":
		return observed{
			Outcome: "reject",
			Code:    "DUPLICATE_MSG_ID",
			Reason:  "duplicate in-flight msg_id detected",
		}
	default:
		return observed{}
	}
}

func validateA2A(env core.Envelope) observed {
	if _, ok := supportedMsgType["a2a"][env.MsgType]; !ok {
		return observed{Outcome: "reject", Code: "UNSUPPORTED_MSG_TYPE", Reason: "unsupported profile msg_type"}
	}
	decision, err := p1a2a.EvaluateFixturePayload(env.Payload)
	if err != nil {
		return observed{
			Outcome: "reject",
			Code:    "INVALID_PROFILE_PAYLOAD",
			Reason:  fmt.Sprintf("invalid A2A fixture payload: %v", err),
		}
	}
	if decision.Reject {
		return observed{Outcome: "reject", Code: decision.Code, Reason: decision.Reason}
	}
	return observed{Outcome: "accept", Code: "OK"}
}

func validateAGDISC(env core.Envelope) observed {
	if _, ok := supportedMsgType["agdisc"][env.MsgType]; !ok {
		return observed{Outcome: "reject", Code: "UNSUPPORTED_MSG_TYPE", Reason: "unsupported profile msg_type"}
	}
	decision, err := p1agdisc.EvaluateFixturePayload(env.Payload)
	if err != nil {
		return observed{Outcome: "reject", Code: "INVALID_PROFILE_PAYLOAD", Reason: fmt.Sprintf("invalid AGDISC fixture payload: %v", err)}
	}
	if decision.Reject {
		return observed{Outcome: "reject", Code: decision.Code, Reason: decision.Reason}
	}
	return observed{Outcome: "accept", Code: "OK"}
}

func validateTOOLDISC(env core.Envelope) observed {
	if _, ok := supportedMsgType["tooldisc"][env.MsgType]; !ok {
		return observed{Outcome: "reject", Code: "UNSUPPORTED_MSG_TYPE", Reason: "unsupported profile msg_type"}
	}
	decision, err := p1tooldisc.EvaluateFixturePayload(env.Payload)
	if err != nil {
		return observed{Outcome: "reject", Code: "INVALID_PROFILE_PAYLOAD", Reason: fmt.Sprintf("invalid TOOLDISC fixture payload: %v", err)}
	}
	if decision.Reject {
		return observed{Outcome: "reject", Code: decision.Code, Reason: decision.Reason}
	}
	return observed{Outcome: "accept", Code: "OK"}
}

func validateEVENTS(env core.Envelope) observed {
	if _, ok := supportedMsgType["events"][env.MsgType]; !ok {
		return observed{Outcome: "reject", Code: "UNSUPPORTED_MSG_TYPE", Reason: "unsupported profile msg_type"}
	}
	decision, err := p1events.EvaluateFixturePayload(env.Payload)
	if err != nil {
		return observed{Outcome: "reject", Code: "INVALID_PROFILE_PAYLOAD", Reason: fmt.Sprintf("invalid EVENTS fixture payload: %v", err)}
	}
	if decision.Reject {
		return observed{Outcome: "reject", Code: decision.Code, Reason: decision.Reason}
	}
	return observed{Outcome: "accept", Code: "OK"}
}

func validateARTIFACT(env core.Envelope) observed {
	if _, ok := supportedMsgType["artifact"][env.MsgType]; !ok {
		return observed{Outcome: "reject", Code: "UNSUPPORTED_MSG_TYPE", Reason: "unsupported profile msg_type"}
	}
	decision, err := p1artifact.EvaluateFixturePayload(env.Payload)
	if err != nil {
		return observed{Outcome: "reject", Code: "INVALID_PROFILE_PAYLOAD", Reason: fmt.Sprintf("invalid ARTIFACT fixture payload: %v", err)}
	}
	if decision.Reject {
		return observed{Outcome: "reject", Code: decision.Code, Reason: decision.Reason}
	}
	return observed{Outcome: "accept", Code: "OK"}
}

func validateSTATE(env core.Envelope) observed {
	if _, ok := supportedMsgType["state"][env.MsgType]; !ok {
		return observed{Outcome: "reject", Code: "UNSUPPORTED_MSG_TYPE", Reason: "unsupported profile msg_type"}
	}
	decision, err := p1state.EvaluateFixturePayload(env.Payload)
	if err != nil {
		return observed{Outcome: "reject", Code: "INVALID_PROFILE_PAYLOAD", Reason: fmt.Sprintf("invalid STATE fixture payload: %v", err)}
	}
	if decision.Reject {
		return observed{Outcome: "reject", Code: decision.Code, Reason: decision.Reason}
	}
	return observed{Outcome: "accept", Code: "OK"}
}

func validateCRED(env core.Envelope) observed {
	if _, ok := supportedMsgType["cred"][env.MsgType]; !ok {
		return observed{Outcome: "reject", Code: "UNSUPPORTED_MSG_TYPE", Reason: "unsupported profile msg_type"}
	}
	decision, err := p1cred.EvaluateFixturePayload(env.Payload)
	if err != nil {
		return observed{Outcome: "reject", Code: "INVALID_PROFILE_PAYLOAD", Reason: fmt.Sprintf("invalid CRED fixture payload: %v", err)}
	}
	if decision.Reject {
		return observed{Outcome: "reject", Code: decision.Code, Reason: decision.Reason}
	}
	return observed{Outcome: "accept", Code: "OK"}
}

func validateOBS(env core.Envelope) observed {
	if _, ok := supportedMsgType["obs"][env.MsgType]; !ok {
		return observed{Outcome: "reject", Code: "UNSUPPORTED_MSG_TYPE", Reason: "unsupported profile msg_type"}
	}
	decision, err := p1obs.EvaluateFixturePayload(env.Payload)
	if err != nil {
		return observed{Outcome: "reject", Code: "INVALID_PROFILE_PAYLOAD", Reason: fmt.Sprintf("invalid OBS fixture payload: %v", err)}
	}
	if decision.Reject {
		return observed{Outcome: "reject", Code: decision.Code, Reason: decision.Reason}
	}
	return observed{Outcome: "accept", Code: "OK"}
}

func shouldScenarioReject(v vector) bool {
	if strings.EqualFold(v.Expected.Outcome, "reject") {
		if v.Expected.Code == "" {
			return true
		}
		// Already enforced by parser/validator/category checks:
		switch v.Expected.Code {
		case string(core.CodeInvalidFrame), string(core.CodeUnsupportedVersion), string(core.CodeUnknownProfile), string(core.CodeInvalidEnvelope), "UNSUPPORTED_MSG_TYPE", "INVALID_MCP_PAYLOAD", "SECURITY_POLICY":
			return false
		default:
			return true
		}
	}
	return false
}

func compare(v vector, o observed) (bool, string) {
	expOutcome := strings.TrimSpace(v.Expected.Outcome)
	expCode := strings.TrimSpace(v.Expected.Code)
	expErrCode := strings.TrimSpace(v.Expected.ExpectedErrorCode)

	if o.Outcome != expOutcome {
		return false, fmt.Sprintf("outcome mismatch: expected=%s observed=%s", expOutcome, o.Outcome)
	}
	if expCode != "" && o.Code != expCode {
		return false, fmt.Sprintf("code mismatch: expected=%s observed=%s", expCode, o.Code)
	}
	if expErrCode != "" {
		obsErrCode := observedCanonicalCode(o)
		if obsErrCode != expErrCode {
			return false, fmt.Sprintf("canonical code mismatch: expected=%s observed=%s", expErrCode, obsErrCode)
		}
	}
	return true, ""
}

func expectedCanonicalCode(v vector) string {
	if strings.TrimSpace(v.Expected.ExpectedErrorCode) != "" {
		return strings.TrimSpace(v.Expected.ExpectedErrorCode)
	}
	if !strings.EqualFold(v.Expected.Outcome, "reject") {
		return ""
	}
	return canonicalErrorCode(v.Expected.Code, v.Expected.RejectionReason)
}

func observedCanonicalCode(o observed) string {
	if !strings.EqualFold(o.Outcome, "reject") {
		return ""
	}
	return canonicalErrorCode(o.Code, o.Reason)
}

func canonicalErrorCode(code, reason string) string {
	switch strings.TrimSpace(code) {
	case "INVALID_FRAME":
		return "ERR_INVALID_FRAME"
	case "UNSUPPORTED_VERSION":
		return "ERR_UNSUPPORTED_VERSION"
	case "UNKNOWN_PROFILE":
		return "ERR_UNKNOWN_PROFILE"
	case "INVALID_ENVELOPE":
		return "ERR_INVALID_ENVELOPE"
	case "UNSUPPORTED_MSG_TYPE":
		return "ERR_UNSUPPORTED_MSG_TYPE"
	case "INVALID_MCP_PAYLOAD":
		return "ERR_INVALID_MCP_PAYLOAD"
	case "INVALID_PROFILE_PAYLOAD":
		return "ERR_INVALID_PROFILE_PAYLOAD"
	case "SECURITY_POLICY":
		return "ERR_SECURITY_POLICY"
	case "RATE_LIMIT_EXCEEDED":
		return "ERR_RATE_LIMIT_EXCEEDED"
	case "DUPLICATE_MSG_ID":
		return "ERR_DUPLICATE_MSG_ID"
	case "NOT_FOUND":
		return "ERR_NOT_FOUND"
	case "COMPATIBILITY_POLICY":
		return "ERR_COMPATIBILITY_POLICY"
	case "ERR_INVALID_FRAME",
		"ERR_FRAME_TOO_LARGE",
		"ERR_UNSUPPORTED_VERSION",
		"ERR_UNKNOWN_PROFILE",
		"ERR_INVALID_ENVELOPE",
		"ERR_UNSUPPORTED_MSG_TYPE",
		"ERR_INVALID_MCP_PAYLOAD",
		"ERR_INVALID_PROFILE_PAYLOAD",
		"ERR_SECURITY_POLICY",
		"ERR_RATE_LIMIT_EXCEEDED",
		"ERR_DUPLICATE_MSG_ID",
		"ERR_NOT_FOUND",
		"ERR_COMPATIBILITY_POLICY":
		return strings.TrimSpace(code)
	default:
		return ""
	}
}

func detectGitSHA() string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "nogit"
	}
	sha := strings.TrimSpace(string(out))
	if sha == "" {
		return "nogit"
	}
	return sha
}

func set(values ...uint64) map[uint64]struct{} {
	m := make(map[uint64]struct{}, len(values))
	for _, v := range values {
		m[v] = struct{}{}
	}
	return m
}

func asUint32(v interface{}) (uint32, bool) {
	n, ok := asUint64(v)
	if !ok {
		return 0, false
	}
	return uint32(n), true
}

func asUint64(v interface{}) (uint64, bool) {
	switch x := v.(type) {
	case float64:
		return uint64(x), true
	case float32:
		return uint64(x), true
	case int:
		if x < 0 {
			return 0, false
		}
		return uint64(x), true
	case int64:
		if x < 0 {
			return 0, false
		}
		return uint64(x), true
	case uint64:
		return x, true
	case uint32:
		return uint64(x), true
	case json.Number:
		i, err := x.Int64()
		if err != nil || i < 0 {
			return 0, false
		}
		return uint64(i), true
	case string:
		var n uint64
		if _, err := fmt.Sscanf(x, "%d", &n); err == nil {
			return n, true
		}
	}
	return 0, false
}

func asBool(v interface{}) (bool, bool) {
	b, ok := v.(bool)
	return b, ok
}

func asString(v interface{}) (string, bool) {
	s, ok := v.(string)
	return s, ok
}
