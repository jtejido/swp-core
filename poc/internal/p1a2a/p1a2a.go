package p1a2a

import "swp-spec-kit/poc/internal/p1fixture"

type FixtureDecision = p1fixture.Decision

var deterministicRejectVectors = map[string]FixtureDecision{
	"a2a_0004_event_after_terminal_result": {
		Reject: true,
		Code:   "INVALID_PROFILE_PAYLOAD",
		Reason: "profile invariant violation",
	},
	"a2a_0006_event_before_task_invalid": {
		Reject: true,
		Code:   "INVALID_PROFILE_PAYLOAD",
		Reason: "profile invariant violation",
	},
	"a2a_0007_result_before_task_invalid": {
		Reject: true,
		Code:   "INVALID_PROFILE_PAYLOAD",
		Reason: "profile invariant violation",
	},
	"a2a_0009_duplicate_task_conflicting_payload_rejected": {
		Reject: true,
		Code:   "INVALID_PROFILE_PAYLOAD",
		Reason: "profile invariant violation",
	},
	"a2a_0010_post_terminal_event_rejected": {
		Reject: true,
		Code:   "INVALID_PROFILE_PAYLOAD",
		Reason: "profile invariant violation",
	},
	"a2a_0011_post_terminal_result_rejected": {
		Reject: true,
		Code:   "INVALID_PROFILE_PAYLOAD",
		Reason: "profile invariant violation",
	},
}

func EvaluateFixturePayload(payload []byte) (FixtureDecision, error) {
	return p1fixture.Evaluate(payload, "a2a", deterministicRejectVectors)
}
