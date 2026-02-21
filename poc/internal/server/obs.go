package server

import (
	"context"
	"fmt"

	"swp-spec-kit/poc/internal/core"
	"swp-spec-kit/poc/internal/p1obs"
	runtimeclock "swp-spec-kit/poc/internal/runtime/clock"
	runtimevalidate "swp-spec-kit/poc/internal/runtime/validate"
)

const (
	obsMsgTypeSet = 1
	obsMsgTypeGet = 2
	obsMsgTypeDoc = 3
	obsMsgTypeErr = 4
)

func handleSWPOBS(ctx context.Context, env core.Envelope) ([]core.Envelope, error) {
	return handleSWPOBSWithBackend(ctx, env, defaultBackends.obs)
}

func (s *Server) handleSWPOBS(ctx context.Context, env core.Envelope) ([]core.Envelope, error) {
	return handleSWPOBSWithBackend(ctx, env, s.runtime.obs)
}

func handleSWPOBSWithBackend(_ context.Context, env core.Envelope, backend OBSBackend) ([]core.Envelope, error) {
	now := runtimeclock.UnixMilli(nil)

	switch env.MsgType {
	case obsMsgTypeSet:
		set, err := p1obs.DecodePayloadSet(env.Payload)
		if err != nil {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid OBS set payload: %w", err))
		}
		if err := validateTraceparent(set.Traceparent); err != nil {
			return nil, core.Wrap(core.CodeInvalidEnvelope, err)
		}
		backend.SetDoc(p1obs.ObsDoc{
			Traceparent: set.Traceparent,
			Tracestate:  set.Tracestate,
			MsgID:       append([]byte(nil), set.MsgID...),
			TaskID:      append([]byte(nil), set.TaskID...),
			RPCID:       append([]byte(nil), set.RPCID...),
		})
		return nil, nil

	case obsMsgTypeGet:
		if _, err := p1obs.DecodePayloadGet(env.Payload); err != nil {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid OBS get payload: %w", err))
		}
		doc := backend.GetDoc()

		payload, err := p1obs.EncodePayloadDoc(doc)
		if err != nil {
			return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("encode OBS doc payload: %w", err))
		}
		return []core.Envelope{newOBSEnvelope(env.MsgID, obsMsgTypeDoc, now, payload)}, nil

	default:
		return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid OBS msg_type %d", env.MsgType))
	}
}

func validateTraceparent(traceparent string) error {
	return runtimevalidate.Traceparent(traceparent)
}

func newOBSEnvelope(msgID []byte, msgType, ts uint64, payload []byte) core.Envelope {
	return core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileSWPOBS,
		MsgType:   msgType,
		MsgID:     append([]byte(nil), msgID...),
		TsUnixMs:  ts,
		Payload:   payload,
	}
}
