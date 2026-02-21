package server

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"swp-spec-kit/poc/internal/core"
	"swp-spec-kit/poc/internal/p1state"
)

const (
	stateMsgTypePut   = 1
	stateMsgTypeGet   = 2
	stateMsgTypeDelta = 3
	stateMsgTypeErr   = 4
)

func handleSWPState(ctx context.Context, env core.Envelope) ([]core.Envelope, error) {
	return handleSWPStateWithBackend(ctx, env, defaultBackends.state)
}

func (s *Server) handleSWPState(ctx context.Context, env core.Envelope) ([]core.Envelope, error) {
	return handleSWPStateWithBackend(ctx, env, s.runtime.state)
}

func handleSWPStateWithBackend(_ context.Context, env core.Envelope, backend StateBackend) ([]core.Envelope, error) {
	now := uint64(time.Now().UnixMilli())

	switch env.MsgType {
	case stateMsgTypePut:
		put, err := p1state.DecodePayloadPut(env.Payload)
		if err != nil {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid STATE put payload: %w", err))
		}
		if len(put.StateID) == 0 {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("state_id required"))
		}
		if len(put.Blob) == 0 {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("blob required"))
		}

		sum := sha256.Sum256(put.Blob)
		if !bytes.Equal(sum[:], put.StateID) {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("state_id/hash mismatch"))
		}

		for _, parentID := range put.ParentIDs {
			if !backend.HasState(parentID) {
				return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("parent state missing"))
			}
		}
		backend.PutState(put)
		return nil, nil

	case stateMsgTypeGet:
		get, err := p1state.DecodePayloadGet(env.Payload)
		if err != nil {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid STATE get payload: %w", err))
		}
		if len(get.StateID) == 0 {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("state_id required"))
		}

		put, ok := backend.GetState(get.StateID)
		if !ok {
			errPayload, err := p1state.EncodePayloadErr(p1state.StateErr{Code: "NOT_FOUND", Message: "state not found"})
			if err != nil {
				return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("encode STATE not-found error: %w", err))
			}
			return []core.Envelope{newStateEnvelope(env.MsgID, stateMsgTypeErr, now, errPayload)}, nil
		}

		payload, err := p1state.EncodePayloadPut(put)
		if err != nil {
			return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("encode STATE get response: %w", err))
		}
		return []core.Envelope{newStateEnvelope(env.MsgID, stateMsgTypePut, now, payload)}, nil

	case stateMsgTypeDelta:
		delta, err := p1state.DecodePayloadDelta(env.Payload)
		if err != nil {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid STATE delta payload: %w", err))
		}
		if len(delta.StateID) == 0 {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("state_id required"))
		}

		if !backend.HasState(delta.StateID) {
			errPayload, err := p1state.EncodePayloadErr(p1state.StateErr{Code: "NOT_FOUND", Message: "state not found"})
			if err != nil {
				return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("encode STATE delta not-found error: %w", err))
			}
			return []core.Envelope{newStateEnvelope(env.MsgID, stateMsgTypeErr, now, errPayload)}, nil
		}
		return nil, nil

	default:
		return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid STATE msg_type %d", env.MsgType))
	}
}

func cloneStatePut(in p1state.StatePut) p1state.StatePut {
	out := p1state.StatePut{
		StateID:  append([]byte(nil), in.StateID...),
		Blob:     append([]byte(nil), in.Blob...),
		Metadata: append([]byte(nil), in.Metadata...),
	}
	for _, parentID := range in.ParentIDs {
		out.ParentIDs = append(out.ParentIDs, append([]byte(nil), parentID...))
	}
	return out
}

func newStateEnvelope(msgID []byte, msgType, ts uint64, payload []byte) core.Envelope {
	return core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileSWPState,
		MsgType:   msgType,
		MsgID:     append([]byte(nil), msgID...),
		TsUnixMs:  ts,
		Payload:   payload,
	}
}
