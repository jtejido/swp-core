package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"swp-spec-kit/poc/internal/p1agdisc"
	"swp-spec-kit/poc/internal/p1artifact"
	"swp-spec-kit/poc/internal/p1events"
	"swp-spec-kit/poc/internal/p1obs"
	"swp-spec-kit/poc/internal/p1policyhint"
	"swp-spec-kit/poc/internal/p1rpc"
	"swp-spec-kit/poc/internal/p1state"
	"swp-spec-kit/poc/internal/p1tooldisc"
)

var (
	errA2AUnknownTask        = errors.New("a2a unknown task")
	errA2ATaskConflict       = errors.New("a2a task conflict")
	errA2ATerminalConflict   = errors.New("a2a terminal conflict")
	errArtifactChunkOrdering = errors.New("artifact chunk ordering violation")
)

type A2ATaskRecord struct {
	Kind          string
	Input         []byte
	Terminal      bool
	TerminalOK    bool
	TerminalOut   []byte
	TerminalError string
}

type A2ABackend interface {
	UpsertTask(taskID []byte, kind string, input []byte) (bool, error)
	GetTask(taskID []byte) (A2ATaskRecord, bool)
	SetTerminal(taskID []byte, ok bool, output []byte, errMsg string) error
}

type ArtifactRecord struct {
	Offer     p1artifact.ArtOffer
	Data      []byte
	NextChunk uint64
}

type ArtifactBackend interface {
	PutOffer(offer p1artifact.ArtOffer)
	GetArtifact(artifactID string) (ArtifactRecord, bool)
	AppendChunk(chunk p1artifact.ArtChunk) (ArtifactRecord, error)
}

type StateBackend interface {
	PutState(put p1state.StatePut)
	GetState(stateID []byte) (p1state.StatePut, bool)
	HasState(stateID []byte) bool
}

type AGDISCBackend interface {
	GetAgentCard(agentID string) (p1agdisc.AgdiscDoc, bool)
}

type ToolDiscBackend interface {
	ListTools() []p1tooldisc.ToolDescriptor
	GetTool(toolID, version string) (p1tooldisc.ToolDescriptor, bool)
}

type RPCBackendMessage struct {
	MsgType    uint64
	Resp       p1rpc.RpcResp
	Err        p1rpc.RpcErr
	StreamItem p1rpc.RpcStreamItem
}

type RPCBackend interface {
	HandleRequest(req p1rpc.RpcReq) ([]RPCBackendMessage, error)
	HandleCancel() (p1rpc.RpcErr, error)
}

type EventsBackend interface {
	Publish(event p1events.EventRecord) error
	Subscribe(filter string) ([]p1events.EventRecord, error)
	Unsubscribe(subscriptionID string) error
}

type CredBackend interface {
	EnsureChain(chainID []byte)
	IncrementChainDepth(chainID []byte) int
	IsRevoked(chainID []byte) bool
	Revoke(chainID []byte)
}

type PolicyHintBackend interface {
	GetConstraint(key string) (p1policyhint.Constraint, bool)
	SetConstraint(c p1policyhint.Constraint)
}

type RelayBackend interface {
	CreateDelivery(deliveryID []byte) (bool, uint32, string)
	MarkAck(deliveryID []byte)
	MarkNack(deliveryID []byte, retryable bool) (uint32, string)
	GetDelivery(deliveryID []byte) (uint32, string, bool)
}

type OBSBackend interface {
	SetDoc(doc p1obs.ObsDoc)
	GetDoc() p1obs.ObsDoc
}

type runtimeBackends struct {
	a2a        A2ABackend
	artifact   ArtifactBackend
	state      StateBackend
	agdisc     AGDISCBackend
	tooldisc   ToolDiscBackend
	rpc        RPCBackend
	events     EventsBackend
	cred       CredBackend
	policyHint PolicyHintBackend
	relay      RelayBackend
	obs        OBSBackend
}

type Option func(*runtimeBackends)

func WithA2ABackend(b A2ABackend) Option {
	return func(r *runtimeBackends) {
		if b != nil {
			r.a2a = b
		}
	}
}

func WithArtifactBackend(b ArtifactBackend) Option {
	return func(r *runtimeBackends) {
		if b != nil {
			r.artifact = b
		}
	}
}

func WithStateBackend(b StateBackend) Option {
	return func(r *runtimeBackends) {
		if b != nil {
			r.state = b
		}
	}
}

func WithAGDISCBackend(b AGDISCBackend) Option {
	return func(r *runtimeBackends) {
		if b != nil {
			r.agdisc = b
		}
	}
}

func WithToolDiscBackend(b ToolDiscBackend) Option {
	return func(r *runtimeBackends) {
		if b != nil {
			r.tooldisc = b
		}
	}
}

func WithRPCBackend(b RPCBackend) Option {
	return func(r *runtimeBackends) {
		if b != nil {
			r.rpc = b
		}
	}
}

func WithEventsBackend(b EventsBackend) Option {
	return func(r *runtimeBackends) {
		if b != nil {
			r.events = b
		}
	}
}

func WithCredBackend(b CredBackend) Option {
	return func(r *runtimeBackends) {
		if b != nil {
			r.cred = b
		}
	}
}

func WithPolicyHintBackend(b PolicyHintBackend) Option {
	return func(r *runtimeBackends) {
		if b != nil {
			r.policyHint = b
		}
	}
}

func WithRelayBackend(b RelayBackend) Option {
	return func(r *runtimeBackends) {
		if b != nil {
			r.relay = b
		}
	}
}

func WithOBSBackend(b OBSBackend) Option {
	return func(r *runtimeBackends) {
		if b != nil {
			r.obs = b
		}
	}
}

func newRuntimeBackends(opts ...Option) runtimeBackends {
	r := runtimeBackends{
		a2a:        newInMemoryA2ABackend(),
		artifact:   newInMemoryArtifactBackend(),
		state:      newInMemoryStateBackend(),
		agdisc:     newInMemoryAGDISCBackend(),
		tooldisc:   newInMemoryToolDiscBackend(),
		rpc:        newInMemoryRPCBackend(),
		events:     newInMemoryEventsBackend(),
		cred:       newInMemoryCredBackend(),
		policyHint: newInMemoryPolicyHintBackend(),
		relay:      newInMemoryRelayBackend(),
		obs:        newInMemoryOBSBackend(),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&r)
		}
	}
	return r
}

var defaultBackends = newRuntimeBackends()

type inMemoryA2ABackend struct {
	mu    sync.RWMutex
	tasks map[string]A2ATaskRecord
}

func newInMemoryA2ABackend() *inMemoryA2ABackend {
	return &inMemoryA2ABackend{tasks: map[string]A2ATaskRecord{}}
}

func (b *inMemoryA2ABackend) UpsertTask(taskID []byte, kind string, input []byte) (bool, error) {
	key := string(taskID)
	b.mu.Lock()
	defer b.mu.Unlock()
	if existing, ok := b.tasks[key]; ok {
		if existing.Kind == kind && bytes.Equal(existing.Input, input) {
			return false, nil
		}
		return false, errA2ATaskConflict
	}
	b.tasks[key] = A2ATaskRecord{Kind: kind, Input: append([]byte(nil), input...)}
	return true, nil
}

func (b *inMemoryA2ABackend) GetTask(taskID []byte) (A2ATaskRecord, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	rec, ok := b.tasks[string(taskID)]
	if !ok {
		return A2ATaskRecord{}, false
	}
	return A2ATaskRecord{
		Kind:          rec.Kind,
		Input:         append([]byte(nil), rec.Input...),
		Terminal:      rec.Terminal,
		TerminalOK:    rec.TerminalOK,
		TerminalOut:   append([]byte(nil), rec.TerminalOut...),
		TerminalError: rec.TerminalError,
	}, true
}

func (b *inMemoryA2ABackend) SetTerminal(taskID []byte, ok bool, output []byte, errMsg string) error {
	key := string(taskID)
	b.mu.Lock()
	defer b.mu.Unlock()
	rec, exists := b.tasks[key]
	if !exists {
		return errA2AUnknownTask
	}
	if rec.Terminal {
		if rec.TerminalOK == ok && bytes.Equal(rec.TerminalOut, output) && rec.TerminalError == errMsg {
			return nil
		}
		return errA2ATerminalConflict
	}
	rec.Terminal = true
	rec.TerminalOK = ok
	rec.TerminalOut = append([]byte(nil), output...)
	rec.TerminalError = errMsg
	b.tasks[key] = rec
	return nil
}

type inMemoryArtifactBackend struct {
	mu      sync.RWMutex
	records map[string]ArtifactRecord
}

func newInMemoryArtifactBackend() *inMemoryArtifactBackend {
	return &inMemoryArtifactBackend{records: map[string]ArtifactRecord{}}
}

func (b *inMemoryArtifactBackend) PutOffer(offer p1artifact.ArtOffer) {
	b.mu.Lock()
	b.records[offer.ArtifactID] = ArtifactRecord{Offer: offer, Data: nil, NextChunk: 0}
	b.mu.Unlock()
}

func (b *inMemoryArtifactBackend) GetArtifact(artifactID string) (ArtifactRecord, bool) {
	b.mu.RLock()
	rec, ok := b.records[artifactID]
	b.mu.RUnlock()
	if !ok {
		return ArtifactRecord{}, false
	}
	return ArtifactRecord{
		Offer:     cloneOffer(rec.Offer),
		Data:      append([]byte(nil), rec.Data...),
		NextChunk: rec.NextChunk,
	}, true
}

func (b *inMemoryArtifactBackend) AppendChunk(chunk p1artifact.ArtChunk) (ArtifactRecord, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	rec, ok := b.records[chunk.ArtifactID]
	if !ok {
		rec = ArtifactRecord{}
	}
	if chunk.ChunkIndex != rec.NextChunk {
		return ArtifactRecord{}, errArtifactChunkOrdering
	}
	rec.Data = append(rec.Data, chunk.Data...)
	rec.NextChunk++
	b.records[chunk.ArtifactID] = rec
	return ArtifactRecord{Offer: cloneOffer(rec.Offer), Data: append([]byte(nil), rec.Data...), NextChunk: rec.NextChunk}, nil
}

func cloneOffer(in p1artifact.ArtOffer) p1artifact.ArtOffer {
	return p1artifact.ArtOffer{
		ArtifactID: in.ArtifactID,
		TotalSize:  in.TotalSize,
		HashAlg:    in.HashAlg,
		Hash:       append([]byte(nil), in.Hash...),
		Metadata:   append([]byte(nil), in.Metadata...),
	}
}

type inMemoryStateBackend struct {
	mu     sync.RWMutex
	states map[string]p1state.StatePut
}

func newInMemoryStateBackend() *inMemoryStateBackend {
	return &inMemoryStateBackend{states: map[string]p1state.StatePut{}}
}

func (b *inMemoryStateBackend) PutState(put p1state.StatePut) {
	b.mu.Lock()
	b.states[string(put.StateID)] = cloneStatePut(put)
	b.mu.Unlock()
}

func (b *inMemoryStateBackend) GetState(stateID []byte) (p1state.StatePut, bool) {
	b.mu.RLock()
	put, ok := b.states[string(stateID)]
	b.mu.RUnlock()
	if !ok {
		return p1state.StatePut{}, false
	}
	return cloneStatePut(put), true
}

func (b *inMemoryStateBackend) HasState(stateID []byte) bool {
	b.mu.RLock()
	_, ok := b.states[string(stateID)]
	b.mu.RUnlock()
	return ok
}

type inMemoryAGDISCBackend struct {
	mu    sync.RWMutex
	cards map[string]p1agdisc.AgdiscDoc
}

func newInMemoryAGDISCBackend() *inMemoryAGDISCBackend {
	return &inMemoryAGDISCBackend{
		cards: map[string]p1agdisc.AgdiscDoc{
			"agent.demo": {
				AgentID:        "agent.demo",
				SchemaRevision: "v1",
				CardPayload:    []byte(`{"name":"Demo Agent","capabilities":["echo","count"]}`),
				ETag:           "etag-agent-demo-v1",
				MaxAgeMs:       60000,
			},
		},
	}
}

func (b *inMemoryAGDISCBackend) GetAgentCard(agentID string) (p1agdisc.AgdiscDoc, bool) {
	b.mu.RLock()
	card, ok := b.cards[agentID]
	b.mu.RUnlock()
	if !ok {
		return p1agdisc.AgdiscDoc{}, false
	}
	return p1agdisc.AgdiscDoc{
		AgentID:        card.AgentID,
		SchemaRevision: card.SchemaRevision,
		CardPayload:    append([]byte(nil), card.CardPayload...),
		ETag:           card.ETag,
		MaxAgeMs:       card.MaxAgeMs,
	}, true
}

type inMemoryToolDiscBackend struct {
	mu    sync.RWMutex
	tools []p1tooldisc.ToolDescriptor
}

func newInMemoryToolDiscBackend() *inMemoryToolDiscBackend {
	return &inMemoryToolDiscBackend{
		tools: []p1tooldisc.ToolDescriptor{
			{
				ToolID:    "echo",
				Name:      "Echo",
				Version:   "1.0.0",
				SchemaRef: "swp://schemas/tools/echo/v1",
			},
			{
				ToolID:    "count",
				Name:      "Counter",
				Version:   "1.0.0",
				SchemaRef: "swp://schemas/tools/count/v1",
			},
		},
	}
}

func (b *inMemoryToolDiscBackend) ListTools() []p1tooldisc.ToolDescriptor {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make([]p1tooldisc.ToolDescriptor, 0, len(b.tools))
	for _, t := range b.tools {
		out = append(out, p1tooldisc.ToolDescriptor{
			ToolID:            t.ToolID,
			Name:              t.Name,
			Version:           t.Version,
			SchemaRef:         t.SchemaRef,
			DescriptorPayload: append([]byte(nil), t.DescriptorPayload...),
		})
	}
	return out
}

func (b *inMemoryToolDiscBackend) GetTool(toolID, version string) (p1tooldisc.ToolDescriptor, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, t := range b.tools {
		if t.ToolID == toolID && (version == "" || t.Version == version) {
			return p1tooldisc.ToolDescriptor{
				ToolID:            t.ToolID,
				Name:              t.Name,
				Version:           t.Version,
				SchemaRef:         t.SchemaRef,
				DescriptorPayload: append([]byte(nil), t.DescriptorPayload...),
			}, true
		}
	}
	return p1tooldisc.ToolDescriptor{}, false
}

type inMemoryRPCBackend struct{}

func newInMemoryRPCBackend() *inMemoryRPCBackend {
	return &inMemoryRPCBackend{}
}

func (b *inMemoryRPCBackend) HandleCancel() (p1rpc.RpcErr, error) {
	return p1rpc.RpcErr{
		RPCID:        []byte{},
		ErrorCode:    "cancelled",
		Retryable:    false,
		ErrorMessage: "cancel received",
	}, nil
}

func (b *inMemoryRPCBackend) HandleRequest(req p1rpc.RpcReq) ([]RPCBackendMessage, error) {
	switch req.Method {
	case "demo.echo":
		return []RPCBackendMessage{{
			MsgType: rpcMsgTypeResp,
			Resp: p1rpc.RpcResp{
				RPCID:  req.RPCID,
				Result: req.Params,
			},
		}}, nil

	case "demo.stream.count":
		count := 5
		var p struct {
			Count int `json:"count"`
		}
		if len(req.Params) > 0 {
			_ = json.Unmarshal(req.Params, &p)
			if p.Count > 0 {
				count = p.Count
			}
		}
		if count > 100 {
			count = 100
		}

		out := make([]RPCBackendMessage, 0, count+1)
		for i := 1; i <= count; i++ {
			out = append(out, RPCBackendMessage{
				MsgType: rpcMsgTypeStreamItem,
				StreamItem: p1rpc.RpcStreamItem{
					RPCID:      req.RPCID,
					SeqNo:      uint64(i),
					Item:       []byte(fmt.Sprintf("%d", i)),
					IsTerminal: false,
				},
			})
		}
		terminalResult, _ := json.Marshal(map[string]any{"count": count, "done": true})
		out = append(out, RPCBackendMessage{
			MsgType: rpcMsgTypeResp,
			Resp: p1rpc.RpcResp{
				RPCID:  req.RPCID,
				Result: terminalResult,
			},
		})
		return out, nil

	case "demo.fail":
		return []RPCBackendMessage{{
			MsgType: rpcMsgTypeErr,
			Err: p1rpc.RpcErr{
				RPCID:        req.RPCID,
				ErrorCode:    "internal",
				Retryable:    false,
				ErrorMessage: "forced failure",
			},
		}}, nil

	default:
		return []RPCBackendMessage{{
			MsgType: rpcMsgTypeErr,
			Err: p1rpc.RpcErr{
				RPCID:        req.RPCID,
				ErrorCode:    "unknown_method",
				Retryable:    false,
				ErrorMessage: "unknown method",
			},
		}}, nil
	}
}

type inMemoryEventsBackend struct{}

func newInMemoryEventsBackend() *inMemoryEventsBackend {
	return &inMemoryEventsBackend{}
}

func (b *inMemoryEventsBackend) Publish(_ p1events.EventRecord) error {
	return nil
}

func (b *inMemoryEventsBackend) Subscribe(_ string) ([]p1events.EventRecord, error) {
	return nil, nil
}

func (b *inMemoryEventsBackend) Unsubscribe(_ string) error {
	return nil
}

type inMemoryCredBackend struct {
	mu       sync.RWMutex
	revoked  map[string]bool
	chainLen map[string]int
}

func newInMemoryCredBackend() *inMemoryCredBackend {
	return &inMemoryCredBackend{revoked: map[string]bool{}, chainLen: map[string]int{}}
}

func (b *inMemoryCredBackend) EnsureChain(chainID []byte) {
	key := string(chainID)
	if key == "" {
		return
	}
	b.mu.Lock()
	if b.chainLen[key] == 0 {
		b.chainLen[key] = 1
	}
	b.mu.Unlock()
}

func (b *inMemoryCredBackend) IncrementChainDepth(chainID []byte) int {
	key := string(chainID)
	b.mu.Lock()
	defer b.mu.Unlock()
	b.chainLen[key] = b.chainLen[key] + 1
	return b.chainLen[key]
}

func (b *inMemoryCredBackend) IsRevoked(chainID []byte) bool {
	b.mu.RLock()
	revoked := b.revoked[string(chainID)]
	b.mu.RUnlock()
	return revoked
}

func (b *inMemoryCredBackend) Revoke(chainID []byte) {
	b.mu.Lock()
	b.revoked[string(chainID)] = true
	b.mu.Unlock()
}

type inMemoryPolicyHintBackend struct {
	mu          sync.RWMutex
	constraints map[string]p1policyhint.Constraint
}

func newInMemoryPolicyHintBackend() *inMemoryPolicyHintBackend {
	return &inMemoryPolicyHintBackend{constraints: map[string]p1policyhint.Constraint{}}
}

func (b *inMemoryPolicyHintBackend) GetConstraint(key string) (p1policyhint.Constraint, bool) {
	b.mu.RLock()
	c, ok := b.constraints[key]
	b.mu.RUnlock()
	return c, ok
}

func (b *inMemoryPolicyHintBackend) SetConstraint(c p1policyhint.Constraint) {
	b.mu.Lock()
	b.constraints[c.Key] = c
	b.mu.Unlock()
}

type inMemoryRelayBackend struct {
	mu         sync.RWMutex
	deliveries map[string]relayDelivery
}

func newInMemoryRelayBackend() *inMemoryRelayBackend {
	return &inMemoryRelayBackend{deliveries: map[string]relayDelivery{}}
}

func (b *inMemoryRelayBackend) CreateDelivery(deliveryID []byte) (bool, uint32, string) {
	key := string(deliveryID)
	b.mu.Lock()
	defer b.mu.Unlock()
	if d, ok := b.deliveries[key]; ok {
		return false, d.attemptCount, d.state
	}
	b.deliveries[key] = relayDelivery{attemptCount: 1, state: "queued"}
	return true, 1, "queued"
}

func (b *inMemoryRelayBackend) MarkAck(deliveryID []byte) {
	key := string(deliveryID)
	b.mu.Lock()
	if d, ok := b.deliveries[key]; ok {
		d.state = "acked"
		b.deliveries[key] = d
	}
	b.mu.Unlock()
}

func (b *inMemoryRelayBackend) MarkNack(deliveryID []byte, retryable bool) (uint32, string) {
	key := string(deliveryID)
	b.mu.Lock()
	defer b.mu.Unlock()
	d := b.deliveries[key]
	d.attemptCount++
	if retryable {
		d.state = "retry"
	} else {
		d.state = "dead-letter"
	}
	b.deliveries[key] = d
	return d.attemptCount, d.state
}

func (b *inMemoryRelayBackend) GetDelivery(deliveryID []byte) (uint32, string, bool) {
	key := string(deliveryID)
	b.mu.RLock()
	d, ok := b.deliveries[key]
	b.mu.RUnlock()
	if !ok {
		return 0, "", false
	}
	return d.attemptCount, d.state, true
}

type inMemoryOBSBackend struct {
	mu  sync.RWMutex
	doc p1obs.ObsDoc
}

func newInMemoryOBSBackend() *inMemoryOBSBackend {
	return &inMemoryOBSBackend{}
}

func (b *inMemoryOBSBackend) SetDoc(doc p1obs.ObsDoc) {
	b.mu.Lock()
	b.doc = p1obs.ObsDoc{
		Traceparent: doc.Traceparent,
		Tracestate:  doc.Tracestate,
		MsgID:       append([]byte(nil), doc.MsgID...),
		TaskID:      append([]byte(nil), doc.TaskID...),
		RPCID:       append([]byte(nil), doc.RPCID...),
	}
	b.mu.Unlock()
}

func (b *inMemoryOBSBackend) GetDoc() p1obs.ObsDoc {
	b.mu.RLock()
	doc := b.doc
	b.mu.RUnlock()
	return p1obs.ObsDoc{
		Traceparent: doc.Traceparent,
		Tracestate:  doc.Tracestate,
		MsgID:       append([]byte(nil), doc.MsgID...),
		TaskID:      append([]byte(nil), doc.TaskID...),
		RPCID:       append([]byte(nil), doc.RPCID...),
	}
}
