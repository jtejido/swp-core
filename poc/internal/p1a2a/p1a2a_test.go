package p1a2a

import (
	"encoding/binary"
	"testing"
)

const fixtureWtBytes = 2

func TestEvaluateFixturePayload(t *testing.T) {
	tests := []struct {
		name       string
		vectorID   string
		wantReject bool
		wantCode   string
	}{
		{
			name:       "accept-handshake",
			vectorID:   "a2a_0001_handshake_success",
			wantReject: false,
			wantCode:   "OK",
		},
		{
			name:       "accept-duplicate-terminal-result-idempotent",
			vectorID:   "a2a_0005_duplicate_terminal_result",
			wantReject: false,
			wantCode:   "OK",
		},
		{
			name:       "reject-event-after-terminal",
			vectorID:   "a2a_0004_event_after_terminal_result",
			wantReject: true,
			wantCode:   "INVALID_PROFILE_PAYLOAD",
		},
		{
			name:       "reject-result-before-task",
			vectorID:   "a2a_0007_result_before_task_invalid",
			wantReject: true,
			wantCode:   "INVALID_PROFILE_PAYLOAD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := encodeFixturePayload("a2a", tt.vectorID)
			decision, err := EvaluateFixturePayload(payload)
			if err != nil {
				t.Fatalf("EvaluateFixturePayload() error = %v", err)
			}
			if decision.Reject != tt.wantReject {
				t.Fatalf("decision.Reject = %v, want %v", decision.Reject, tt.wantReject)
			}
			if decision.Code != tt.wantCode {
				t.Fatalf("decision.Code = %q, want %q", decision.Code, tt.wantCode)
			}
		})
	}
}

func TestEvaluateFixturePayloadInvalid(t *testing.T) {
	_, err := EvaluateFixturePayload([]byte{0x0a, 0x03, 'a', '2', 'a'})
	if err == nil {
		t.Fatalf("expected missing vector id error")
	}
}

func encodeFixturePayload(profile, vectorID string) []byte {
	var out []byte
	out = fixtureAppendKey(out, 1, fixtureWtBytes)
	out = fixtureAppendBytes(out, []byte(profile))
	out = fixtureAppendKey(out, 2, fixtureWtBytes)
	out = fixtureAppendBytes(out, []byte(vectorID))
	return out
}

func fixtureAppendKey(dst []byte, field, wt uint64) []byte {
	return binary.AppendUvarint(dst, (field<<3)|wt)
}

func fixtureAppendBytes(dst []byte, b []byte) []byte {
	dst = binary.AppendUvarint(dst, uint64(len(b)))
	return append(dst, b...)
}
