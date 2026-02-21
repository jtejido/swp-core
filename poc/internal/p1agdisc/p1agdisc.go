package p1agdisc

import "swp-spec-kit/poc/internal/p1fixture"

type FixtureDecision = p1fixture.Decision

var deterministicRejectVectors = map[string]FixtureDecision{
	"agdisc_0002_not_found": {
		Reject: true,
		Code:   "NOT_FOUND",
		Reason: "deterministic not-found behavior",
	},
	"agdisc_0003_invalid_doc_rejected": {
		Reject: true,
		Code:   "INVALID_PROFILE_PAYLOAD",
		Reason: "profile invariant violation",
	},
}

func EvaluateFixturePayload(payload []byte) (FixtureDecision, error) {
	return p1fixture.Evaluate(payload, "agdisc", deterministicRejectVectors)
}
