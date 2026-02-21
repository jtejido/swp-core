package server

import (
	"context"
	"crypto/sha256"
	"errors"
	"testing"

	"swp-spec-kit/poc/internal/core"
	"swp-spec-kit/poc/internal/p1a2a"
	"swp-spec-kit/poc/internal/p1agdisc"
	"swp-spec-kit/poc/internal/p1artifact"
	"swp-spec-kit/poc/internal/p1cred"
	"swp-spec-kit/poc/internal/p1events"
	"swp-spec-kit/poc/internal/p1obs"
	"swp-spec-kit/poc/internal/p1policyhint"
	"swp-spec-kit/poc/internal/p1relay"
	"swp-spec-kit/poc/internal/p1rpc"
	"swp-spec-kit/poc/internal/p1state"
	"swp-spec-kit/poc/internal/p1tooldisc"
)

type mockA2ABackend struct {
	upsertCalled bool
	getCalled    bool
	setCalled    bool
}

func (m *mockA2ABackend) UpsertTask(_ []byte, _ string, _ []byte) (bool, error) {
	m.upsertCalled = true
	return true, nil
}

func (m *mockA2ABackend) GetTask(_ []byte) (A2ATaskRecord, bool) {
	m.getCalled = true
	return A2ATaskRecord{}, true
}

func (m *mockA2ABackend) SetTerminal(_ []byte, _ bool, _ []byte, _ string) error {
	m.setCalled = true
	return nil
}

type mockArtifactBackend struct {
	putOfferCalled bool
}

func (m *mockArtifactBackend) PutOffer(_ p1artifact.ArtOffer) {
	m.putOfferCalled = true
}

func (m *mockArtifactBackend) GetArtifact(_ string) (ArtifactRecord, bool) {
	return ArtifactRecord{}, true
}

func (m *mockArtifactBackend) AppendChunk(_ p1artifact.ArtChunk) (ArtifactRecord, error) {
	return ArtifactRecord{}, nil
}

type mockStateBackend struct {
	putStateCalled bool
}

func (m *mockStateBackend) PutState(_ p1state.StatePut) {
	m.putStateCalled = true
}

func (m *mockStateBackend) GetState(_ []byte) (p1state.StatePut, bool) {
	return p1state.StatePut{}, false
}

func (m *mockStateBackend) HasState(_ []byte) bool {
	return true
}

type mockCredBackend struct {
	ensureCalled bool
}

type mockAGDISCBackend struct {
	getCalled bool
}

func (m *mockAGDISCBackend) GetAgentCard(_ string) (p1agdisc.AgdiscDoc, bool) {
	m.getCalled = true
	return p1agdisc.AgdiscDoc{
		AgentID:        "agent.mock",
		SchemaRevision: "v1",
		CardPayload:    []byte(`{"name":"Mock Agent"}`),
		ETag:           "etag-mock-v1",
		MaxAgeMs:       60000,
	}, true
}

type mockToolDiscBackend struct {
	listCalled bool
	getCalled  bool
}

func (m *mockToolDiscBackend) ListTools() []p1tooldisc.ToolDescriptor {
	m.listCalled = true
	return []p1tooldisc.ToolDescriptor{{
		ToolID:    "mock-tool",
		Name:      "Mock Tool",
		Version:   "1.0.0",
		SchemaRef: "swp://schemas/tools/mock/v1",
	}}
}

func (m *mockToolDiscBackend) GetTool(toolID, version string) (p1tooldisc.ToolDescriptor, bool) {
	m.getCalled = true
	if toolID == "mock-tool" && (version == "" || version == "1.0.0") {
		return p1tooldisc.ToolDescriptor{
			ToolID:    "mock-tool",
			Name:      "Mock Tool",
			Version:   "1.0.0",
			SchemaRef: "swp://schemas/tools/mock/v1",
		}, true
	}
	return p1tooldisc.ToolDescriptor{}, false
}

func (m *mockCredBackend) EnsureChain(_ []byte) {
	m.ensureCalled = true
}

func (m *mockCredBackend) IncrementChainDepth(_ []byte) int {
	return 1
}

func (m *mockCredBackend) IsRevoked(_ []byte) bool {
	return false
}

func (m *mockCredBackend) Revoke(_ []byte) {}

type mockPolicyHintBackend struct {
	setCalled bool
}

func (m *mockPolicyHintBackend) GetConstraint(_ string) (p1policyhint.Constraint, bool) {
	return p1policyhint.Constraint{}, false
}

func (m *mockPolicyHintBackend) SetConstraint(_ p1policyhint.Constraint) {
	m.setCalled = true
}

type mockRelayBackend struct {
	createCalled bool
}

func (m *mockRelayBackend) CreateDelivery(_ []byte) (bool, uint32, string) {
	m.createCalled = true
	return true, 1, "queued"
}

func (m *mockRelayBackend) MarkAck(_ []byte) {}

func (m *mockRelayBackend) MarkNack(_ []byte, _ bool) (uint32, string) {
	return 2, "retry"
}

func (m *mockRelayBackend) GetDelivery(_ []byte) (uint32, string, bool) {
	return 1, "queued", true
}

type mockOBSBackend struct {
	setCalled bool
	getCalled bool
	doc       p1obs.ObsDoc
}

func (m *mockOBSBackend) SetDoc(doc p1obs.ObsDoc) {
	m.setCalled = true
	m.doc = doc
}

func (m *mockOBSBackend) GetDoc() p1obs.ObsDoc {
	m.getCalled = true
	if m.doc.Traceparent == "" {
		return p1obs.ObsDoc{Traceparent: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"}
	}
	return m.doc
}

type mockRPCBackend struct {
	requestCalled bool
	cancelCalled  bool
}

func (m *mockRPCBackend) HandleRequest(req p1rpc.RpcReq) ([]RPCBackendMessage, error) {
	m.requestCalled = true
	return []RPCBackendMessage{{
		MsgType: rpcMsgTypeResp,
		Resp: p1rpc.RpcResp{
			RPCID:  req.RPCID,
			Result: []byte(`{"ok":true}`),
		},
	}}, nil
}

func (m *mockRPCBackend) HandleCancel() (p1rpc.RpcErr, error) {
	m.cancelCalled = true
	return p1rpc.RpcErr{
		RPCID:        []byte("mock"),
		ErrorCode:    "cancelled",
		Retryable:    false,
		ErrorMessage: "cancelled by test",
	}, nil
}

type mockEventsBackend struct {
	publishCalled     bool
	subscribeCalled   bool
	unsubscribeCalled bool
	published         []p1events.EventRecord
}

func (m *mockEventsBackend) Publish(ev p1events.EventRecord) error {
	m.publishCalled = true
	m.published = append(m.published, ev)
	return nil
}

func (m *mockEventsBackend) Subscribe(_ string) ([]p1events.EventRecord, error) {
	m.subscribeCalled = true
	return []p1events.EventRecord{{
		EventID:   "mock-event-1",
		EventType: "mock.event",
		Severity:  "info",
		TsUnixMs:  1,
	}}, nil
}

func (m *mockEventsBackend) Unsubscribe(_ string) error {
	m.unsubscribeCalled = true
	return nil
}

func TestServerUsesInjectedA2ABackend(t *testing.T) {
	m := &mockA2ABackend{}
	s := New(nil, WithA2ABackend(m))
	payload, err := p1a2a.EncodePayloadTask(p1a2a.Task{TaskID: []byte("task-1"), Kind: "demo.run", Input: []byte("x")})
	if err != nil {
		t.Fatalf("encode A2A task payload: %v", err)
	}
	_, err = s.router.Dispatch(context.Background(), core.Envelope{ProfileID: ProfileA2A, MsgType: a2aMsgTypeTask, MsgID: []byte("12345678abcdefgh"), Payload: payload})
	if err != nil {
		t.Fatalf("dispatch A2A task: %v", err)
	}
	if !m.upsertCalled {
		t.Fatalf("expected injected A2A backend UpsertTask to be called")
	}
}

func TestServerUsesInjectedArtifactBackend(t *testing.T) {
	m := &mockArtifactBackend{}
	s := New(nil, WithArtifactBackend(m))
	payload, err := p1artifact.EncodePayloadOffer(p1artifact.ArtOffer{ArtifactID: "artifact-1"})
	if err != nil {
		t.Fatalf("encode ARTIFACT offer payload: %v", err)
	}
	_, err = s.router.Dispatch(context.Background(), core.Envelope{ProfileID: ProfileSWPArtifact, MsgType: artifactMsgTypeOffer, MsgID: []byte("12345678abcdefgh"), Payload: payload})
	if err != nil {
		t.Fatalf("dispatch ARTIFACT offer: %v", err)
	}
	if !m.putOfferCalled {
		t.Fatalf("expected injected Artifact backend PutOffer to be called")
	}
}

func TestServerUsesInjectedAGDISCBackend(t *testing.T) {
	m := &mockAGDISCBackend{}
	s := New(nil, WithAGDISCBackend(m))
	payload, err := p1agdisc.EncodePayloadGet(p1agdisc.AgdiscGet{AgentID: "agent.mock"})
	if err != nil {
		t.Fatalf("encode AGDISC get payload: %v", err)
	}
	out, err := s.router.Dispatch(context.Background(), core.Envelope{
		ProfileID: ProfileSWPAGDISC,
		MsgType:   agdiscMsgTypeGet,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   payload,
	})
	if err != nil {
		t.Fatalf("dispatch AGDISC get: %v", err)
	}
	if !m.getCalled {
		t.Fatalf("expected injected AGDISC backend GetAgentCard to be called")
	}
	if len(out) != 1 || out[0].MsgType != agdiscMsgTypeDoc {
		t.Fatalf("expected AGDISC doc response, got %+v", out)
	}
}

func TestServerUsesInjectedToolDiscBackend(t *testing.T) {
	m := &mockToolDiscBackend{}
	s := New(nil, WithToolDiscBackend(m))

	listPayload, err := p1tooldisc.EncodePayloadListReq(p1tooldisc.TooldiscListReq{})
	if err != nil {
		t.Fatalf("encode TOOLDISC list payload: %v", err)
	}
	out, err := s.router.Dispatch(context.Background(), core.Envelope{
		ProfileID: ProfileSWPToolDisc,
		MsgType:   tooldiscMsgTypeListReq,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   listPayload,
	})
	if err != nil {
		t.Fatalf("dispatch TOOLDISC list: %v", err)
	}
	if !m.listCalled {
		t.Fatalf("expected injected ToolDisc backend ListTools to be called")
	}
	if len(out) != 1 || out[0].MsgType != tooldiscMsgTypeListResp {
		t.Fatalf("expected TOOLDISC list response, got %+v", out)
	}

	getPayload, err := p1tooldisc.EncodePayloadGetReq(p1tooldisc.TooldiscGetReq{ToolID: "mock-tool"})
	if err != nil {
		t.Fatalf("encode TOOLDISC get payload: %v", err)
	}
	out, err = s.router.Dispatch(context.Background(), core.Envelope{
		ProfileID: ProfileSWPToolDisc,
		MsgType:   tooldiscMsgTypeGetReq,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   getPayload,
	})
	if err != nil {
		t.Fatalf("dispatch TOOLDISC get: %v", err)
	}
	if !m.getCalled {
		t.Fatalf("expected injected ToolDisc backend GetTool to be called")
	}
	if len(out) != 1 || out[0].MsgType != tooldiscMsgTypeGetResp {
		t.Fatalf("expected TOOLDISC get response, got %+v", out)
	}
}

func TestServerUsesInjectedStateBackend(t *testing.T) {
	m := &mockStateBackend{}
	s := New(nil, WithStateBackend(m))
	blob := []byte("state")
	h := sha256.Sum256(blob)
	payload, err := p1state.EncodePayloadPut(p1state.StatePut{StateID: h[:], Blob: blob})
	if err != nil {
		t.Fatalf("encode STATE put payload: %v", err)
	}
	_, err = s.router.Dispatch(context.Background(), core.Envelope{ProfileID: ProfileSWPState, MsgType: stateMsgTypePut, MsgID: []byte("12345678abcdefgh"), Payload: payload})
	if err != nil {
		t.Fatalf("dispatch STATE put: %v", err)
	}
	if !m.putStateCalled {
		t.Fatalf("expected injected State backend PutState to be called")
	}
}

func TestServerUsesInjectedCredBackend(t *testing.T) {
	m := &mockCredBackend{}
	s := New(nil, WithCredBackend(m))
	payload, err := p1cred.EncodePayloadPresent(p1cred.CredPresent{CredType: "jwt", Credential: []byte("token"), ChainID: []byte("chain-1")})
	if err != nil {
		t.Fatalf("encode CRED present payload: %v", err)
	}
	_, err = s.router.Dispatch(context.Background(), core.Envelope{ProfileID: ProfileSWPCred, MsgType: credMsgTypePresent, MsgID: []byte("12345678abcdefgh"), Payload: payload})
	if err != nil {
		t.Fatalf("dispatch CRED present: %v", err)
	}
	if !m.ensureCalled {
		t.Fatalf("expected injected Cred backend EnsureChain to be called")
	}
}

func TestServerUsesInjectedPolicyHintBackend(t *testing.T) {
	m := &mockPolicyHintBackend{}
	s := New(nil, WithPolicyHintBackend(m))
	payload, err := p1policyhint.EncodePayloadSet(p1policyhint.PolicyHintSet{Constraints: []p1policyhint.Constraint{{Key: "no_external_network", Value: "true", Mode: "MUST"}}})
	if err != nil {
		t.Fatalf("encode POLICYHINT set payload: %v", err)
	}
	_, err = s.router.Dispatch(context.Background(), core.Envelope{ProfileID: ProfileSWPPolicyHint, MsgType: policyHintMsgTypeSet, MsgID: []byte("12345678abcdefgh"), Payload: payload})
	if err != nil {
		t.Fatalf("dispatch POLICYHINT set: %v", err)
	}
	if !m.setCalled {
		t.Fatalf("expected injected PolicyHint backend SetConstraint to be called")
	}
}

func TestServerUsesInjectedRelayBackend(t *testing.T) {
	m := &mockRelayBackend{}
	s := New(nil, WithRelayBackend(m))
	payload, err := p1relay.EncodePayloadPublish(p1relay.RelayPublish{DeliveryID: []byte("delivery-1"), Topic: "demo", Payload: []byte("x")})
	if err != nil {
		t.Fatalf("encode RELAY publish payload: %v", err)
	}
	_, err = s.router.Dispatch(context.Background(), core.Envelope{ProfileID: ProfileSWPRelay, MsgType: relayMsgTypePublish, MsgID: []byte("12345678abcdefgh"), Payload: payload})
	if err != nil {
		t.Fatalf("dispatch RELAY publish: %v", err)
	}
	if !m.createCalled {
		t.Fatalf("expected injected Relay backend CreateDelivery to be called")
	}
}

func TestServerUsesInjectedOBSBackend(t *testing.T) {
	m := &mockOBSBackend{}
	s := New(nil, WithOBSBackend(m))
	setPayload, err := p1obs.EncodePayloadSet(p1obs.ObsSet{Traceparent: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"})
	if err != nil {
		t.Fatalf("encode OBS set payload: %v", err)
	}
	_, err = s.router.Dispatch(context.Background(), core.Envelope{ProfileID: ProfileSWPOBS, MsgType: obsMsgTypeSet, MsgID: []byte("12345678abcdefgh"), Payload: setPayload})
	if err != nil {
		t.Fatalf("dispatch OBS set: %v", err)
	}
	if !m.setCalled {
		t.Fatalf("expected injected OBS backend SetDoc to be called")
	}

	getPayload, err := p1obs.EncodePayloadGet(p1obs.ObsGet{IncludeCurrent: true})
	if err != nil {
		t.Fatalf("encode OBS get payload: %v", err)
	}
	_, err = s.router.Dispatch(context.Background(), core.Envelope{ProfileID: ProfileSWPOBS, MsgType: obsMsgTypeGet, MsgID: []byte("12345678abcdefgh"), Payload: getPayload})
	if err != nil {
		t.Fatalf("dispatch OBS get: %v", err)
	}
	if !m.getCalled {
		t.Fatalf("expected injected OBS backend GetDoc to be called")
	}
}

func TestServerUsesInjectedRPCBackend(t *testing.T) {
	m := &mockRPCBackend{}
	s := New(nil, WithRPCBackend(m))

	reqPayload, err := p1rpc.EncodePayloadReq(p1rpc.RpcReq{
		RPCID:  []byte("rpc-mock"),
		Method: "demo.any",
		Params: []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("encode RPC request payload: %v", err)
	}
	out, err := s.router.Dispatch(context.Background(), core.Envelope{
		ProfileID: ProfileSWPRPC,
		MsgType:   rpcMsgTypeReq,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   reqPayload,
	})
	if err != nil {
		t.Fatalf("dispatch RPC request: %v", err)
	}
	if !m.requestCalled {
		t.Fatalf("expected injected RPC backend HandleRequest to be called")
	}
	if len(out) != 1 || out[0].MsgType != rpcMsgTypeResp {
		t.Fatalf("expected one RPC response envelope, got %+v", out)
	}

	out, err = s.router.Dispatch(context.Background(), core.Envelope{
		ProfileID: ProfileSWPRPC,
		MsgType:   rpcMsgTypeCancel,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   []byte{0x0a, 0x00},
	})
	if err != nil {
		t.Fatalf("dispatch RPC cancel: %v", err)
	}
	if !m.cancelCalled {
		t.Fatalf("expected injected RPC backend HandleCancel to be called")
	}
	if len(out) != 1 || out[0].MsgType != rpcMsgTypeErr {
		t.Fatalf("expected one RPC error envelope for cancel, got %+v", out)
	}
}

func TestServerUsesInjectedEventsBackend(t *testing.T) {
	m := &mockEventsBackend{}
	s := New(nil, WithEventsBackend(m))

	pubPayload, err := p1events.EncodePayloadPublish(p1events.EvtPublish{
		Event: p1events.EventRecord{
			EventID:   "event-1",
			EventType: "mock.event",
			Severity:  "info",
		},
	})
	if err != nil {
		t.Fatalf("encode EVENTS publish payload: %v", err)
	}
	_, err = s.router.Dispatch(context.Background(), core.Envelope{
		ProfileID: ProfileSWPEvents,
		MsgType:   eventsMsgTypePublish,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   pubPayload,
	})
	if err != nil {
		t.Fatalf("dispatch EVENTS publish: %v", err)
	}
	if !m.publishCalled {
		t.Fatalf("expected injected Events backend Publish to be called")
	}

	subPayload, err := p1events.EncodePayloadSubscribe(p1events.EvtSubscribe{Filter: "mock"})
	if err != nil {
		t.Fatalf("encode EVENTS subscribe payload: %v", err)
	}
	out, err := s.router.Dispatch(context.Background(), core.Envelope{
		ProfileID: ProfileSWPEvents,
		MsgType:   eventsMsgTypeSubscribe,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   subPayload,
	})
	if err != nil {
		t.Fatalf("dispatch EVENTS subscribe: %v", err)
	}
	if !m.subscribeCalled {
		t.Fatalf("expected injected Events backend Subscribe to be called")
	}
	if len(out) != 1 || out[0].MsgType != eventsMsgTypeBatch {
		t.Fatalf("expected one EVENTS batch response, got %+v", out)
	}
}

func TestServerEmitsEventsForMCPAndRPCWithOBSCorrelation(t *testing.T) {
	evBackend := &mockEventsBackend{}
	obsBackend := &mockOBSBackend{
		doc: p1obs.ObsDoc{
			Traceparent: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
			TaskID:      []byte("task-from-obs"),
		},
	}
	s := New(nil, WithEventsBackend(evBackend), WithOBSBackend(obsBackend))

	mcpPayload := []byte(`{"jsonrpc":"2.0","id":"x","method":"tools/list","params":{}}`)
	mcpResp, err := s.router.Dispatch(context.Background(), core.Envelope{
		ProfileID: ProfileMCPMap,
		MsgType:   mcpMsgTypeRequest,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   mcpPayload,
	})
	if err != nil {
		t.Fatalf("dispatch MCP request: %v", err)
	}
	if len(mcpResp) != 1 || mcpResp[0].MsgType != mcpMsgTypeResponse {
		t.Fatalf("expected MCP response envelope, got %+v", mcpResp)
	}

	rpcPayload, err := p1rpc.EncodePayloadReq(p1rpc.RpcReq{
		RPCID:  []byte("rpc-correlation"),
		Method: "demo.echo",
		Params: []byte(`{"ok":true}`),
	})
	if err != nil {
		t.Fatalf("encode RPC request payload: %v", err)
	}
	_, err = s.router.Dispatch(context.Background(), core.Envelope{
		ProfileID: ProfileSWPRPC,
		MsgType:   rpcMsgTypeReq,
		MsgID:     []byte("abcdef1234567890"),
		Payload:   rpcPayload,
	})
	if err != nil {
		t.Fatalf("dispatch RPC request: %v", err)
	}

	if len(evBackend.published) == 0 {
		t.Fatalf("expected emitted events from MCP/RPC handlers")
	}

	var sawMCP, sawRPC bool
	for _, ev := range evBackend.published {
		if ev.EventType == "swp.mcp.request" {
			sawMCP = true
			if string(ev.TaskID) != "task-from-obs" {
				t.Fatalf("expected MCP event task_id from OBS, got %q", string(ev.TaskID))
			}
		}
		if ev.EventType == "swp.rpc.request" {
			sawRPC = true
			if string(ev.RPCID) != "rpc-correlation" {
				t.Fatalf("expected RPC event rpc_id from request, got %q", string(ev.RPCID))
			}
			if string(ev.TaskID) != "task-from-obs" {
				t.Fatalf("expected RPC event task_id fallback from OBS, got %q", string(ev.TaskID))
			}
		}
	}
	if !sawMCP {
		t.Fatalf("expected swp.mcp.request event")
	}
	if !sawRPC {
		t.Fatalf("expected swp.rpc.request event")
	}
}

func TestServerDefaultBackendsArePerInstance(t *testing.T) {
	s1 := New(nil)
	s2 := New(nil)

	b1, ok1 := s1.runtime.a2a.(*inMemoryA2ABackend)
	b2, ok2 := s2.runtime.a2a.(*inMemoryA2ABackend)
	if !ok1 || !ok2 {
		t.Fatalf("expected in-memory A2A backends")
	}
	if b1 == b2 {
		t.Fatalf("expected distinct per-server default backends")
	}
}

type faultA2ABackend struct {
	upsertCreated bool
	upsertErr     error
	getTask       A2ATaskRecord
	getTaskOK     bool
	setErr        error
}

func (f *faultA2ABackend) UpsertTask(_ []byte, _ string, _ []byte) (bool, error) {
	return f.upsertCreated, f.upsertErr
}

func (f *faultA2ABackend) GetTask(_ []byte) (A2ATaskRecord, bool) {
	return f.getTask, f.getTaskOK
}

func (f *faultA2ABackend) SetTerminal(_ []byte, _ bool, _ []byte, _ string) error {
	return f.setErr
}

func TestA2ABackendFaultInjection(t *testing.T) {
	taskPayload, err := p1a2a.EncodePayloadTask(p1a2a.Task{TaskID: []byte("task-fault"), Kind: "demo.run", Input: []byte("x")})
	if err != nil {
		t.Fatalf("encode A2A task payload: %v", err)
	}
	resultPayload, err := p1a2a.EncodePayloadResult(p1a2a.Result{TaskID: []byte("task-fault"), OK: true, Output: []byte("ok")})
	if err != nil {
		t.Fatalf("encode A2A result payload: %v", err)
	}

	tests := []struct {
		name     string
		env      core.Envelope
		backend  A2ABackend
		wantCode core.Code
	}{
		{
			name: "task-upsert-conflict",
			env: core.Envelope{
				ProfileID: ProfileA2A, MsgType: a2aMsgTypeTask, MsgID: []byte("12345678abcdefgh"), Payload: taskPayload,
			},
			backend:  &faultA2ABackend{upsertErr: errA2ATaskConflict},
			wantCode: core.CodeInvalidEnvelope,
		},
		{
			name: "result-unknown-task",
			env: core.Envelope{
				ProfileID: ProfileA2A, MsgType: a2aMsgTypeResult, MsgID: []byte("12345678abcdefgh"), Payload: resultPayload,
			},
			backend:  &faultA2ABackend{setErr: errA2AUnknownTask},
			wantCode: core.CodeInvalidEnvelope,
		},
		{
			name: "result-terminal-conflict",
			env: core.Envelope{
				ProfileID: ProfileA2A, MsgType: a2aMsgTypeResult, MsgID: []byte("12345678abcdefgh"), Payload: resultPayload,
			},
			backend:  &faultA2ABackend{setErr: errA2ATerminalConflict},
			wantCode: core.CodeInvalidEnvelope,
		},
		{
			name: "result-unexpected-backend-error",
			env: core.Envelope{
				ProfileID: ProfileA2A, MsgType: a2aMsgTypeResult, MsgID: []byte("12345678abcdefgh"), Payload: resultPayload,
			},
			backend:  &faultA2ABackend{setErr: errors.New("backend exploded")},
			wantCode: core.CodeInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handleA2AWithBackend(context.Background(), tt.env, tt.backend)
			if err == nil {
				t.Fatalf("expected error")
			}
			if got := core.CodeFromError(err); got != tt.wantCode {
				t.Fatalf("unexpected code: got=%s want=%s err=%v", got, tt.wantCode, err)
			}
		})
	}
}

type faultArtifactBackend struct {
	appendErr error
}

func (f *faultArtifactBackend) PutOffer(_ p1artifact.ArtOffer) {}

func (f *faultArtifactBackend) GetArtifact(_ string) (ArtifactRecord, bool) {
	return ArtifactRecord{}, true
}

func (f *faultArtifactBackend) AppendChunk(_ p1artifact.ArtChunk) (ArtifactRecord, error) {
	return ArtifactRecord{}, f.appendErr
}

func TestArtifactBackendFaultInjection(t *testing.T) {
	chunkPayload, err := p1artifact.EncodePayloadChunk(p1artifact.ArtChunk{
		ArtifactID: "artifact-fault",
		ChunkIndex: 0,
		Data:       []byte("x"),
	})
	if err != nil {
		t.Fatalf("encode ARTIFACT chunk payload: %v", err)
	}
	env := core.Envelope{
		ProfileID: ProfileSWPArtifact,
		MsgType:   artifactMsgTypeChunk,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   chunkPayload,
	}

	t.Run("ordering-error-maps-to-artifact-err-envelope", func(t *testing.T) {
		out, err := handleSWPArtifactWithBackend(context.Background(), env, &faultArtifactBackend{appendErr: errArtifactChunkOrdering})
		if err != nil {
			t.Fatalf("expected artifact error envelope, got dispatch err: %v", err)
		}
		if len(out) != 1 || out[0].MsgType != artifactMsgTypeErr {
			t.Fatalf("expected one ARTIFACT_ERR envelope, got %+v", out)
		}
		aerr, derr := p1artifact.DecodePayloadErr(out[0].Payload)
		if derr != nil {
			t.Fatalf("decode ARTIFACT err payload: %v", derr)
		}
		if aerr.Code != "ORDERING" {
			t.Fatalf("expected ORDERING, got %q", aerr.Code)
		}
	})

	t.Run("unexpected-error-maps-to-internal-error", func(t *testing.T) {
		_, err := handleSWPArtifactWithBackend(context.Background(), env, &faultArtifactBackend{appendErr: errors.New("disk failure")})
		if err == nil {
			t.Fatalf("expected error")
		}
		if got := core.CodeFromError(err); got != core.CodeInternalError {
			t.Fatalf("unexpected code: got=%s want=%s err=%v", got, core.CodeInternalError, err)
		}
	})
}

type conflictPolicyHintBackend struct{}

func (c *conflictPolicyHintBackend) GetConstraint(_ string) (p1policyhint.Constraint, bool) {
	return p1policyhint.Constraint{Key: "region", Value: "us-east-1", Mode: "MUST"}, true
}

func (c *conflictPolicyHintBackend) SetConstraint(_ p1policyhint.Constraint) {}

func TestPolicyHintConflictFromInjectedBackend(t *testing.T) {
	payload, err := p1policyhint.EncodePayloadSet(p1policyhint.PolicyHintSet{
		Constraints: []p1policyhint.Constraint{{
			Key:   "region",
			Value: "eu-west-1",
			Mode:  "MUST",
		}},
	})
	if err != nil {
		t.Fatalf("encode POLICYHINT set payload: %v", err)
	}

	out, err := handleSWPPolicyHintWithBackend(context.Background(), core.Envelope{
		ProfileID: ProfileSWPPolicyHint,
		MsgType:   policyHintMsgTypeSet,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   payload,
	}, &conflictPolicyHintBackend{})
	if err != nil {
		t.Fatalf("expected POLICYHINT violation envelope, got err: %v", err)
	}
	if len(out) != 1 || out[0].MsgType != policyHintMsgTypeViolation {
		t.Fatalf("expected one POLICY_VIOLATION envelope, got %+v", out)
	}
	viol, derr := p1policyhint.DecodePayloadViolation(out[0].Payload)
	if derr != nil {
		t.Fatalf("decode POLICY_VIOLATION payload: %v", derr)
	}
	if viol.ReasonCode != "CONFLICT" {
		t.Fatalf("expected CONFLICT reason, got %q", viol.ReasonCode)
	}
}

type duplicateRelayBackend struct{}

func (d *duplicateRelayBackend) CreateDelivery(_ []byte) (bool, uint32, string) {
	return false, 7, "queued"
}

func (d *duplicateRelayBackend) MarkAck(_ []byte) {}

func (d *duplicateRelayBackend) MarkNack(_ []byte, _ bool) (uint32, string) {
	return 8, "retry"
}

func (d *duplicateRelayBackend) GetDelivery(_ []byte) (uint32, string, bool) {
	return 7, "queued", true
}

func TestRelayDuplicatePublishFromInjectedBackend(t *testing.T) {
	payload, err := p1relay.EncodePayloadPublish(p1relay.RelayPublish{
		DeliveryID: []byte("delivery-dup"),
		Topic:      "demo",
		Payload:    []byte("x"),
	})
	if err != nil {
		t.Fatalf("encode RELAY publish payload: %v", err)
	}

	out, err := handleSWPRelayWithBackend(context.Background(), core.Envelope{
		ProfileID: ProfileSWPRelay,
		MsgType:   relayMsgTypePublish,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   payload,
	}, &duplicateRelayBackend{})
	if err != nil {
		t.Fatalf("expected duplicate status envelope, got err: %v", err)
	}
	if len(out) != 1 || out[0].MsgType != relayMsgTypeStatus {
		t.Fatalf("expected one RELAY_STATUS envelope, got %+v", out)
	}
	status, derr := p1relay.DecodePayloadStatus(out[0].Payload)
	if derr != nil {
		t.Fatalf("decode RELAY status payload: %v", derr)
	}
	if status.AttemptCount != 7 {
		t.Fatalf("expected attempt_count=7, got %d", status.AttemptCount)
	}
}

type missingParentStateBackend struct{}

func (m *missingParentStateBackend) PutState(_ p1state.StatePut) {}

func (m *missingParentStateBackend) GetState(_ []byte) (p1state.StatePut, bool) {
	return p1state.StatePut{}, false
}

func (m *missingParentStateBackend) HasState(_ []byte) bool {
	return false
}

func TestStateMissingParentFromInjectedBackend(t *testing.T) {
	parent := []byte("parent")
	blob := []byte("state-with-parent")
	hash := sha256.Sum256(blob)
	payload, err := p1state.EncodePayloadPut(p1state.StatePut{
		StateID:   hash[:],
		Blob:      blob,
		ParentIDs: [][]byte{parent},
	})
	if err != nil {
		t.Fatalf("encode STATE put payload: %v", err)
	}

	_, err = handleSWPStateWithBackend(context.Background(), core.Envelope{
		ProfileID: ProfileSWPState,
		MsgType:   stateMsgTypePut,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   payload,
	}, &missingParentStateBackend{})
	if err == nil {
		t.Fatalf("expected missing parent error")
	}
	if got := core.CodeFromError(err); got != core.CodeInvalidEnvelope {
		t.Fatalf("unexpected code: got=%s want=%s err=%v", got, core.CodeInvalidEnvelope, err)
	}
}

type faultRPCBackend struct {
	requestErr    error
	cancelErr     error
	responseTypes []RPCBackendMessage
}

func (f *faultRPCBackend) HandleRequest(_ p1rpc.RpcReq) ([]RPCBackendMessage, error) {
	if f.requestErr != nil {
		return nil, f.requestErr
	}
	return f.responseTypes, nil
}

func (f *faultRPCBackend) HandleCancel() (p1rpc.RpcErr, error) {
	if f.cancelErr != nil {
		return p1rpc.RpcErr{}, f.cancelErr
	}
	return p1rpc.RpcErr{ErrorCode: "cancelled", ErrorMessage: "ok"}, nil
}

func TestRPCBackendFaultInjection(t *testing.T) {
	reqPayload, err := p1rpc.EncodePayloadReq(p1rpc.RpcReq{
		RPCID:  []byte("rpc-fault"),
		Method: "demo.fault",
		Params: []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("encode RPC request payload: %v", err)
	}

	t.Run("request-error-maps-to-internal", func(t *testing.T) {
		_, err := handleSWPRPCWithBackend(context.Background(), core.Envelope{
			ProfileID: ProfileSWPRPC,
			MsgType:   rpcMsgTypeReq,
			MsgID:     []byte("12345678abcdefgh"),
			Payload:   reqPayload,
		}, &faultRPCBackend{requestErr: errors.New("rpc backend failed")}, nil)
		if err == nil {
			t.Fatalf("expected error")
		}
		if got := core.CodeFromError(err); got != core.CodeInternalError {
			t.Fatalf("unexpected code: got=%s want=%s err=%v", got, core.CodeInternalError, err)
		}
	})

	t.Run("unsupported-backend-message-type-maps-to-internal", func(t *testing.T) {
		_, err := handleSWPRPCWithBackend(context.Background(), core.Envelope{
			ProfileID: ProfileSWPRPC,
			MsgType:   rpcMsgTypeReq,
			MsgID:     []byte("12345678abcdefgh"),
			Payload:   reqPayload,
		}, &faultRPCBackend{
			responseTypes: []RPCBackendMessage{{MsgType: 99}},
		}, nil)
		if err == nil {
			t.Fatalf("expected error")
		}
		if got := core.CodeFromError(err); got != core.CodeInternalError {
			t.Fatalf("unexpected code: got=%s want=%s err=%v", got, core.CodeInternalError, err)
		}
	})
}

type faultEventsBackend struct {
	publishErr     error
	subscribeErr   error
	unsubscribeErr error
}

func (f *faultEventsBackend) Publish(_ p1events.EventRecord) error {
	return f.publishErr
}

func (f *faultEventsBackend) Subscribe(_ string) ([]p1events.EventRecord, error) {
	return nil, f.subscribeErr
}

func (f *faultEventsBackend) Unsubscribe(_ string) error {
	return f.unsubscribeErr
}

func TestEventsBackendFaultInjection(t *testing.T) {
	publishPayload, err := p1events.EncodePayloadPublish(p1events.EvtPublish{
		Event: p1events.EventRecord{EventID: "event-fault", EventType: "fault.event", Severity: "info"},
	})
	if err != nil {
		t.Fatalf("encode EVENTS publish payload: %v", err)
	}

	_, err = handleSWPEventsWithBackend(context.Background(), core.Envelope{
		ProfileID: ProfileSWPEvents,
		MsgType:   eventsMsgTypePublish,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   publishPayload,
	}, &faultEventsBackend{publishErr: errors.New("events backend failed")}, nil)
	if err == nil {
		t.Fatalf("expected error")
	}
	if got := core.CodeFromError(err); got != core.CodeInternalError {
		t.Fatalf("unexpected code: got=%s want=%s err=%v", got, core.CodeInternalError, err)
	}
}
