package p1tooldisc

import "swp-spec-kit/poc/internal/p1fixture"

type FixtureDecision = p1fixture.Decision

var deterministicRejectVectors = map[string]FixtureDecision{
	"tooldisc_0003_missing_tool_not_found": {
		Reject: true,
		Code:   "NOT_FOUND",
		Reason: "deterministic not-found behavior",
	},
	"tooldisc_0004_schema_ref_invalid": {
		Reject: true,
		Code:   "INVALID_PROFILE_PAYLOAD",
		Reason: "profile invariant violation",
	},
	"tooldisc_0005_descriptor_missing_required": {
		Reject: true,
		Code:   "INVALID_PROFILE_PAYLOAD",
		Reason: "profile invariant violation",
	},
}

func EvaluateFixturePayload(payload []byte) (FixtureDecision, error) {
	return p1fixture.Evaluate(payload, "tooldisc", deterministicRejectVectors)
}
