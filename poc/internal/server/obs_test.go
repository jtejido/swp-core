package server

import (
	"context"
	"testing"

	"swp-spec-kit/poc/internal/core"
	"swp-spec-kit/poc/internal/p1obs"
)

func TestHandleSWPOBSSetAndGet(t *testing.T) {
	setPayload, err := p1obs.EncodePayloadSet(p1obs.ObsSet{
		Traceparent: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
		Tracestate:  "vendor=a",
		MsgID:       []byte("12345678abcdefgh"),
	})
	if err != nil {
		t.Fatalf("encode OBS set payload: %v", err)
	}
	setEnv := core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileSWPOBS,
		MsgType:   obsMsgTypeSet,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   setPayload,
	}
	out, err := handleSWPOBS(context.Background(), setEnv)
	if err != nil {
		t.Fatalf("set failed: %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("expected no set response, got %d", len(out))
	}

	getPayload, err := p1obs.EncodePayloadGet(p1obs.ObsGet{IncludeCurrent: true})
	if err != nil {
		t.Fatalf("encode OBS get payload: %v", err)
	}
	getEnv := core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileSWPOBS,
		MsgType:   obsMsgTypeGet,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   getPayload,
	}
	out, err = handleSWPOBS(context.Background(), getEnv)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected one get response, got %d", len(out))
	}
	if out[0].MsgType != obsMsgTypeDoc {
		t.Fatalf("expected msg_type %d, got %d", obsMsgTypeDoc, out[0].MsgType)
	}
	doc, err := p1obs.DecodePayloadDoc(out[0].Payload)
	if err != nil {
		t.Fatalf("decode OBS doc payload: %v", err)
	}
	if doc.Traceparent == "" {
		t.Fatalf("expected traceparent in obs doc")
	}
}

func TestHandleSWPOBSInvalidTraceparent(t *testing.T) {
	setPayload, err := p1obs.EncodePayloadSet(p1obs.ObsSet{
		Traceparent: "invalid",
	})
	if err != nil {
		t.Fatalf("encode OBS set payload: %v", err)
	}
	setEnv := core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileSWPOBS,
		MsgType:   obsMsgTypeSet,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   setPayload,
	}
	_, err = handleSWPOBS(context.Background(), setEnv)
	if err == nil {
		t.Fatalf("expected invalid traceparent error")
	}
	if core.CodeFromError(err) != core.CodeInvalidEnvelope {
		t.Fatalf("expected INVALID_ENVELOPE, got %s", core.CodeFromError(err))
	}
}
