package p1artifact

import "swp-spec-kit/poc/internal/p1fixture"

type FixtureDecision = p1fixture.Decision

var deterministicRejectVectors = map[string]FixtureDecision{
	"artifact_0003_integrity_mismatch": {
		Reject: true,
		Code:   "INVALID_PROFILE_PAYLOAD",
		Reason: "profile invariant violation",
	},
	"artifact_0006_corruption_rejected": {
		Reject: true,
		Code:   "INVALID_PROFILE_PAYLOAD",
		Reason: "profile invariant violation",
	},
}

func EvaluateFixturePayload(payload []byte) (FixtureDecision, error) {
	return p1fixture.Evaluate(payload, "artifact", deterministicRejectVectors)
}
