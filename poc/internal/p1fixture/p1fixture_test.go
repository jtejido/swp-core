package p1fixture

import (
	"encoding/binary"
	"testing"
)

const wireTypeBytesTest = 2

func TestEvaluate(t *testing.T) {
	rejectRules := map[string]Decision{
		"v_reject": {Reject: true, Code: "INVALID_PROFILE_PAYLOAD", Reason: "reject"},
	}

	d, err := Evaluate(encodeFixturePayload("events", "v_ok"), "events", rejectRules)
	if err != nil {
		t.Fatalf("Evaluate() unexpected error: %v", err)
	}
	if d.Reject || d.Code != "OK" {
		t.Fatalf("expected accept OK, got %+v", d)
	}

	d, err = Evaluate(encodeFixturePayload("events", "v_reject"), "events", rejectRules)
	if err != nil {
		t.Fatalf("Evaluate() unexpected error: %v", err)
	}
	if !d.Reject || d.Code != "INVALID_PROFILE_PAYLOAD" {
		t.Fatalf("expected reject INVALID_PROFILE_PAYLOAD, got %+v", d)
	}

	d, err = Evaluate(encodeFixturePayload("wrong", "v_ok"), "events", rejectRules)
	if err != nil {
		t.Fatalf("Evaluate() unexpected error: %v", err)
	}
	if !d.Reject || d.Code != "INVALID_PROFILE_PAYLOAD" {
		t.Fatalf("expected marker reject, got %+v", d)
	}
}

func TestDecodeProfileAndVectorIDInvalid(t *testing.T) {
	if _, _, err := DecodeProfileAndVectorID([]byte{0x0a, 0x03, 'a', 'b', 'c'}); err == nil {
		t.Fatalf("expected error for missing vector id")
	}
}

func encodeFixturePayload(profile, vectorID string) []byte {
	var out []byte
	out = appendKey(out, 1, wireTypeBytesTest)
	out = appendBytes(out, []byte(profile))
	out = appendKey(out, 2, wireTypeBytesTest)
	out = appendBytes(out, []byte(vectorID))
	return out
}

func appendKey(dst []byte, field, wt uint64) []byte {
	return binary.AppendUvarint(dst, (field<<3)|wt)
}

func appendBytes(dst []byte, b []byte) []byte {
	dst = binary.AppendUvarint(dst, uint64(len(b)))
	return append(dst, b...)
}
