package core

import "testing"

func TestE1RoundTrip(t *testing.T) {
	limits := DefaultLimits()
	in := Envelope{
		Version:   CoreVersion,
		ProfileID: 1,
		MsgType:   2,
		Flags:     0,
		TsUnixMs:  123456789,
		MsgID:     []byte("12345678abcdefgh"),
		Extensions: []Extension{
			{Type: 16, Value: []byte("x")},
		},
		Payload: []byte("hello"),
	}

	body, err := EncodeEnvelopeE1(in)
	if err != nil {
		t.Fatalf("EncodeEnvelopeE1 failed: %v", err)
	}
	out, err := DecodeEnvelopeE1(body, limits)
	if err != nil {
		t.Fatalf("DecodeEnvelopeE1 failed: %v", err)
	}

	if out.Version != in.Version || out.ProfileID != in.ProfileID || out.MsgType != in.MsgType {
		t.Fatalf("header mismatch: got %+v want %+v", out, in)
	}
	if string(out.MsgID) != string(in.MsgID) {
		t.Fatalf("msg_id mismatch")
	}
	if string(out.Payload) != string(in.Payload) {
		t.Fatalf("payload mismatch")
	}
	if len(out.Extensions) != 1 || out.Extensions[0].Type != 16 || string(out.Extensions[0].Value) != "x" {
		t.Fatalf("extensions mismatch: %+v", out.Extensions)
	}
}
