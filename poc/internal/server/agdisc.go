package server

import (
	"context"
	"fmt"
	"time"

	"swp-spec-kit/poc/internal/core"
	"swp-spec-kit/poc/internal/p1agdisc"
)

const (
	agdiscMsgTypeGet         = 1
	agdiscMsgTypeDoc         = 2
	agdiscMsgTypeNotModified = 3
	agdiscMsgTypeErr         = 4
)

func handleSWPAGDISC(ctx context.Context, env core.Envelope) ([]core.Envelope, error) {
	return handleSWPAGDISCWithBackend(ctx, env, defaultBackends.agdisc)
}

func (s *Server) handleSWPAGDISC(ctx context.Context, env core.Envelope) ([]core.Envelope, error) {
	return handleSWPAGDISCWithBackend(ctx, env, s.runtime.agdisc)
}

func handleSWPAGDISCWithBackend(_ context.Context, env core.Envelope, backend AGDISCBackend) ([]core.Envelope, error) {
	now := uint64(time.Now().UnixMilli())

	if env.MsgType != agdiscMsgTypeGet {
		return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid AGDISC msg_type %d", env.MsgType))
	}

	req, err := p1agdisc.DecodePayloadGet(env.Payload)
	if err != nil {
		return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid AGDISC get payload: %w", err))
	}

	card, ok := backend.GetAgentCard(req.AgentID)
	if !ok {
		payload, err := p1agdisc.EncodePayloadErr(p1agdisc.AgdiscErr{
			Code:    "NOT_FOUND",
			Message: "agent card not found",
		})
		if err != nil {
			return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("encode AGDISC error payload: %w", err))
		}
		return []core.Envelope{newAGDISCEnvelope(env.MsgID, agdiscMsgTypeErr, now, payload)}, nil
	}

	if req.IfNoneMatch != "" && req.IfNoneMatch == card.ETag {
		payload, err := p1agdisc.EncodePayloadNotModified(p1agdisc.AgdiscNotModified{
			AgentID: card.AgentID,
			ETag:    card.ETag,
		})
		if err != nil {
			return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("encode AGDISC not-modified payload: %w", err))
		}
		return []core.Envelope{newAGDISCEnvelope(env.MsgID, agdiscMsgTypeNotModified, now, payload)}, nil
	}

	payload, err := p1agdisc.EncodePayloadDoc(card)
	if err != nil {
		return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("encode AGDISC doc payload: %w", err))
	}
	return []core.Envelope{newAGDISCEnvelope(env.MsgID, agdiscMsgTypeDoc, now, payload)}, nil
}

func newAGDISCEnvelope(msgID []byte, msgType, ts uint64, payload []byte) core.Envelope {
	return core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileSWPAGDISC,
		MsgType:   msgType,
		MsgID:     append([]byte(nil), msgID...),
		TsUnixMs:  ts,
		Payload:   payload,
	}
}
