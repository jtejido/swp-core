package server

import (
	"context"
	"testing"

	"swp-spec-kit/poc/internal/core"
	"swp-spec-kit/poc/internal/p1a2a"
)

func resetA2ATasks(t *testing.T) {
	t.Helper()
	defaultBackends.a2a = newInMemoryA2ABackend()
}

func TestHandleA2ALifecycleSuccess(t *testing.T) {
	resetA2ATasks(t)

	hsPayload, err := p1a2a.EncodePayloadHandshake(p1a2a.Handshake{AgentID: "agent.demo", Capabilities: []string{"demo.run"}})
	if err != nil {
		t.Fatalf("encode handshake payload: %v", err)
	}
	_, err = handleA2A(context.Background(), core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileA2A,
		MsgType:   a2aMsgTypeHandshake,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   hsPayload,
	})
	if err != nil {
		t.Fatalf("handshake failed: %v", err)
	}

	taskPayload, err := p1a2a.EncodePayloadTask(p1a2a.Task{TaskID: []byte("task-1"), Kind: "demo.run", Input: []byte(`{"x":1}`)})
	if err != nil {
		t.Fatalf("encode task payload: %v", err)
	}
	out, err := handleA2A(context.Background(), core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileA2A,
		MsgType:   a2aMsgTypeTask,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   taskPayload,
	})
	if err != nil {
		t.Fatalf("task failed: %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("expected no task response, got %d", len(out))
	}

	eventPayload, err := p1a2a.EncodePayloadEvent(p1a2a.Event{TaskID: []byte("task-1"), Message: "running"})
	if err != nil {
		t.Fatalf("encode event payload: %v", err)
	}
	_, err = handleA2A(context.Background(), core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileA2A,
		MsgType:   a2aMsgTypeEvent,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   eventPayload,
	})
	if err != nil {
		t.Fatalf("event failed: %v", err)
	}

	resultPayload, err := p1a2a.EncodePayloadResult(p1a2a.Result{TaskID: []byte("task-1"), OK: true, Output: []byte("done")})
	if err != nil {
		t.Fatalf("encode result payload: %v", err)
	}
	_, err = handleA2A(context.Background(), core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileA2A,
		MsgType:   a2aMsgTypeResult,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   resultPayload,
	})
	if err != nil {
		t.Fatalf("result failed: %v", err)
	}
}

func TestHandleA2AEventBeforeTaskRejected(t *testing.T) {
	resetA2ATasks(t)

	eventPayload, _ := p1a2a.EncodePayloadEvent(p1a2a.Event{TaskID: []byte("missing-task"), Message: "running"})
	_, err := handleA2A(context.Background(), core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileA2A,
		MsgType:   a2aMsgTypeEvent,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   eventPayload,
	})
	if err == nil {
		t.Fatalf("expected event-before-task rejection")
	}
	if core.CodeFromError(err) != core.CodeInvalidEnvelope {
		t.Fatalf("expected INVALID_ENVELOPE, got %s", core.CodeFromError(err))
	}
}

func TestHandleA2ADuplicateTaskConflictRejected(t *testing.T) {
	resetA2ATasks(t)

	payload1, _ := p1a2a.EncodePayloadTask(p1a2a.Task{TaskID: []byte("task-1"), Kind: "demo.run", Input: []byte("a")})
	payload2, _ := p1a2a.EncodePayloadTask(p1a2a.Task{TaskID: []byte("task-1"), Kind: "demo.run", Input: []byte("b")})

	_, err := handleA2A(context.Background(), core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileA2A,
		MsgType:   a2aMsgTypeTask,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   payload1,
	})
	if err != nil {
		t.Fatalf("first task failed: %v", err)
	}

	_, err = handleA2A(context.Background(), core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileA2A,
		MsgType:   a2aMsgTypeTask,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   payload2,
	})
	if err == nil {
		t.Fatalf("expected conflicting duplicate task rejection")
	}
}

func TestHandleA2APostTerminalEventRejected(t *testing.T) {
	resetA2ATasks(t)

	taskPayload, _ := p1a2a.EncodePayloadTask(p1a2a.Task{TaskID: []byte("task-1"), Kind: "demo.run", Input: []byte("a")})
	resultPayload, _ := p1a2a.EncodePayloadResult(p1a2a.Result{TaskID: []byte("task-1"), OK: true, Output: []byte("ok")})
	eventPayload, _ := p1a2a.EncodePayloadEvent(p1a2a.Event{TaskID: []byte("task-1"), Message: "late"})

	_, _ = handleA2A(context.Background(), core.Envelope{Version: core.CoreVersion, ProfileID: ProfileA2A, MsgType: a2aMsgTypeTask, MsgID: []byte("12345678abcdefgh"), Payload: taskPayload})
	_, _ = handleA2A(context.Background(), core.Envelope{Version: core.CoreVersion, ProfileID: ProfileA2A, MsgType: a2aMsgTypeResult, MsgID: []byte("12345678abcdefgh"), Payload: resultPayload})

	_, err := handleA2A(context.Background(), core.Envelope{Version: core.CoreVersion, ProfileID: ProfileA2A, MsgType: a2aMsgTypeEvent, MsgID: []byte("12345678abcdefgh"), Payload: eventPayload})
	if err == nil {
		t.Fatalf("expected post-terminal event rejection")
	}
}

func TestHandleA2AUnsupportedCapabilityReturnsResultFailure(t *testing.T) {
	resetA2ATasks(t)

	taskPayload, _ := p1a2a.EncodePayloadTask(p1a2a.Task{TaskID: []byte("task-unsupported"), Kind: "unsupported.capability", Input: []byte("a")})
	out, err := handleA2A(context.Background(), core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileA2A,
		MsgType:   a2aMsgTypeTask,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   taskPayload,
	})
	if err != nil {
		t.Fatalf("task failed unexpectedly: %v", err)
	}
	if len(out) != 1 || out[0].MsgType != a2aMsgTypeResult {
		t.Fatalf("expected terminal result for unsupported capability, got %+v", out)
	}
	res, err := p1a2a.DecodePayloadResult(out[0].Payload)
	if err != nil {
		t.Fatalf("decode result payload failed: %v", err)
	}
	if res.OK {
		t.Fatalf("expected failing result")
	}
}
