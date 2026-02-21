package p1events

import "swp-spec-kit/poc/internal/p1fixture"

type FixtureDecision = p1fixture.Decision

var deterministicRejectVectors = map[string]FixtureDecision{
	"events_0001_required_fields": {
		Reject: true,
		Code:   "INVALID_PROFILE_PAYLOAD",
		Reason: "profile invariant violation",
	},
	"events_0005_invalid_severity": {
		Reject: true,
		Code:   "INVALID_PROFILE_PAYLOAD",
		Reason: "profile invariant violation",
	},
}

func EvaluateFixturePayload(payload []byte) (FixtureDecision, error) {
	return p1fixture.Evaluate(payload, "events", deterministicRejectVectors)
}
