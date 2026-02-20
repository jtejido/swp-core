package core

import "testing"

func TestValidateEnvelope(t *testing.T) {
	v := DefaultValidator()
	v.KnownProfiles = map[uint64]struct{}{1: {}}
	env := Envelope{
		Version:   CoreVersion,
		ProfileID: 1,
		MsgType:   1,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   []byte("ok"),
	}
	if err := v.ValidateEnvelope(env); err != nil {
		t.Fatalf("expected valid envelope, got %v", err)
	}
}

func TestValidateEnvelopeUnknownProfile(t *testing.T) {
	v := DefaultValidator()
	v.KnownProfiles = map[uint64]struct{}{1: {}}
	env := Envelope{
		Version:   CoreVersion,
		ProfileID: 99,
		MsgType:   1,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   []byte("ok"),
	}
	if err := v.ValidateEnvelope(env); err == nil {
		t.Fatalf("expected error")
	} else if CodeFromError(err) != CodeUnknownProfile {
		t.Fatalf("unexpected code: %s", CodeFromError(err))
	}
}

func TestValidateEnvelopeRejectsZeroMsgType(t *testing.T) {
	v := DefaultValidator()
	v.KnownProfiles = map[uint64]struct{}{1: {}}
	env := Envelope{
		Version:   CoreVersion,
		ProfileID: 1,
		MsgType:   0,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   []byte("ok"),
	}
	if err := v.ValidateEnvelope(env); err == nil {
		t.Fatalf("expected error")
	} else if CodeFromError(err) != CodeInvalidEnvelope {
		t.Fatalf("unexpected code: %s", CodeFromError(err))
	}
}
