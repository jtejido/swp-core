package server

import (
	"context"
	"testing"

	"swp-spec-kit/poc/internal/core"
	"swp-spec-kit/poc/internal/p1events"
	"swp-spec-kit/poc/internal/p1tooldisc"
)

func TestHandleSWPToolDiscList(t *testing.T) {
	payload, err := p1tooldisc.EncodePayloadListReq(p1tooldisc.TooldiscListReq{
		PageSize: 1,
		Filter:   "echo",
	})
	if err != nil {
		t.Fatalf("encode list req: %v", err)
	}
	env := core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileSWPToolDisc,
		MsgType:   tooldiscMsgTypeListReq,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   payload,
	}

	out, err := handleSWPToolDisc(context.Background(), env)
	if err != nil {
		t.Fatalf("handleSWPToolDisc list failed: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 response, got %d", len(out))
	}
	if out[0].MsgType != tooldiscMsgTypeListResp {
		t.Fatalf("expected list response msg_type=%d, got %d", tooldiscMsgTypeListResp, out[0].MsgType)
	}

	resp, err := p1tooldisc.DecodePayloadListResp(out[0].Payload)
	if err != nil {
		t.Fatalf("decode list response failed: %v", err)
	}
	if len(resp.Tools) != 1 || resp.Tools[0].ToolID != "echo" {
		t.Fatalf("unexpected list response tools: %+v", resp.Tools)
	}
}

func TestHandleSWPToolDiscGetNotFound(t *testing.T) {
	payload, err := p1tooldisc.EncodePayloadGetReq(p1tooldisc.TooldiscGetReq{
		ToolID: "does-not-exist",
	})
	if err != nil {
		t.Fatalf("encode get req: %v", err)
	}
	env := core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileSWPToolDisc,
		MsgType:   tooldiscMsgTypeGetReq,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   payload,
	}

	out, err := handleSWPToolDisc(context.Background(), env)
	if err != nil {
		t.Fatalf("handleSWPToolDisc get failed: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 response, got %d", len(out))
	}
	if out[0].MsgType != tooldiscMsgTypeErr {
		t.Fatalf("expected error msg_type=%d, got %d", tooldiscMsgTypeErr, out[0].MsgType)
	}
	derr, err := p1tooldisc.DecodePayloadErr(out[0].Payload)
	if err != nil {
		t.Fatalf("decode error payload failed: %v", err)
	}
	if derr.Code != "NOT_FOUND" {
		t.Fatalf("expected NOT_FOUND, got %q", derr.Code)
	}
}

func TestHandleSWPEventsPublishAndSubscribe(t *testing.T) {
	pubPayload, err := p1events.EncodePayloadPublish(p1events.EvtPublish{
		Event: p1events.EventRecord{
			EventID:   "ev-1",
			EventType: "test.event",
			Severity:  "info",
			TsUnixMs:  1,
			MsgID:     []byte("12345678abcdefgh"),
		},
	})
	if err != nil {
		t.Fatalf("encode publish payload: %v", err)
	}
	pubEnv := core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileSWPEvents,
		MsgType:   eventsMsgTypePublish,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   pubPayload,
	}

	pubOut, err := handleSWPEvents(context.Background(), pubEnv)
	if err != nil {
		t.Fatalf("publish failed: %v", err)
	}
	if len(pubOut) != 0 {
		t.Fatalf("expected no publish response, got %d", len(pubOut))
	}

	subPayload, err := p1events.EncodePayloadSubscribe(p1events.EvtSubscribe{Filter: "type:test.event"})
	if err != nil {
		t.Fatalf("encode subscribe payload: %v", err)
	}
	subEnv := core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileSWPEvents,
		MsgType:   eventsMsgTypeSubscribe,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   subPayload,
	}
	subOut, err := handleSWPEvents(context.Background(), subEnv)
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}
	if len(subOut) != 1 {
		t.Fatalf("expected 1 subscribe response, got %d", len(subOut))
	}
	if subOut[0].MsgType != eventsMsgTypeBatch {
		t.Fatalf("expected batch response msg_type=%d, got %d", eventsMsgTypeBatch, subOut[0].MsgType)
	}
	batch, err := p1events.DecodePayloadBatch(subOut[0].Payload)
	if err != nil {
		t.Fatalf("decode batch response failed: %v", err)
	}
	if len(batch.Events) != 0 {
		t.Fatalf("expected empty initial batch, got %d events", len(batch.Events))
	}
}

func TestHandleSWPEventsInvalidSeverity(t *testing.T) {
	pubPayload, err := p1events.EncodePayloadPublish(p1events.EvtPublish{
		Event: p1events.EventRecord{
			EventID:   "ev-1",
			EventType: "test.event",
			Severity:  "fatal",
		},
	})
	if err != nil {
		t.Fatalf("encode publish payload: %v", err)
	}
	pubEnv := core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileSWPEvents,
		MsgType:   eventsMsgTypePublish,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   pubPayload,
	}

	_, err = handleSWPEvents(context.Background(), pubEnv)
	if err == nil {
		t.Fatalf("expected invalid severity error")
	}
	if core.CodeFromError(err) != core.CodeInvalidEnvelope {
		t.Fatalf("expected INVALID_ENVELOPE, got %s", core.CodeFromError(err))
	}
}
