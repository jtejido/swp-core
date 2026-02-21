package server

import (
	"context"
	"testing"

	"swp-spec-kit/poc/internal/core"
	"swp-spec-kit/poc/internal/p1agdisc"
)

func TestHandleSWPAGDISCGetDoc(t *testing.T) {
	payload, err := p1agdisc.EncodePayloadGet(p1agdisc.AgdiscGet{
		AgentID: "agent.demo",
	})
	if err != nil {
		t.Fatalf("encode AGDISC get payload: %v", err)
	}
	env := core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileSWPAGDISC,
		MsgType:   agdiscMsgTypeGet,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   payload,
	}

	out, err := handleSWPAGDISC(context.Background(), env)
	if err != nil {
		t.Fatalf("handleSWPAGDISC failed: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 response, got %d", len(out))
	}
	if out[0].MsgType != agdiscMsgTypeDoc {
		t.Fatalf("expected AGDISC doc msg_type=%d, got %d", agdiscMsgTypeDoc, out[0].MsgType)
	}

	doc, err := p1agdisc.DecodePayloadDoc(out[0].Payload)
	if err != nil {
		t.Fatalf("decode AGDISC doc payload failed: %v", err)
	}
	if doc.AgentID != "agent.demo" {
		t.Fatalf("expected agent.demo, got %q", doc.AgentID)
	}
	if doc.ETag == "" {
		t.Fatalf("expected non-empty etag")
	}
}

func TestHandleSWPAGDISCNotModified(t *testing.T) {
	payload, err := p1agdisc.EncodePayloadGet(p1agdisc.AgdiscGet{
		AgentID:     "agent.demo",
		IfNoneMatch: "etag-agent-demo-v1",
	})
	if err != nil {
		t.Fatalf("encode AGDISC get payload: %v", err)
	}
	env := core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileSWPAGDISC,
		MsgType:   agdiscMsgTypeGet,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   payload,
	}

	out, err := handleSWPAGDISC(context.Background(), env)
	if err != nil {
		t.Fatalf("handleSWPAGDISC failed: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 response, got %d", len(out))
	}
	if out[0].MsgType != agdiscMsgTypeNotModified {
		t.Fatalf("expected AGDISC not-modified msg_type=%d, got %d", agdiscMsgTypeNotModified, out[0].MsgType)
	}
}

func TestHandleSWPAGDISCNotFound(t *testing.T) {
	payload, err := p1agdisc.EncodePayloadGet(p1agdisc.AgdiscGet{
		AgentID: "agent.missing",
	})
	if err != nil {
		t.Fatalf("encode AGDISC get payload: %v", err)
	}
	env := core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileSWPAGDISC,
		MsgType:   agdiscMsgTypeGet,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   payload,
	}

	out, err := handleSWPAGDISC(context.Background(), env)
	if err != nil {
		t.Fatalf("handleSWPAGDISC failed: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 response, got %d", len(out))
	}
	if out[0].MsgType != agdiscMsgTypeErr {
		t.Fatalf("expected AGDISC err msg_type=%d, got %d", agdiscMsgTypeErr, out[0].MsgType)
	}
	derr, err := p1agdisc.DecodePayloadErr(out[0].Payload)
	if err != nil {
		t.Fatalf("decode AGDISC err payload failed: %v", err)
	}
	if derr.Code != "NOT_FOUND" {
		t.Fatalf("expected NOT_FOUND code, got %q", derr.Code)
	}
}
