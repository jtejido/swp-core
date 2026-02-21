package server

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"swp-spec-kit/poc/internal/core"
	"swp-spec-kit/poc/internal/p1a2a"
)

const (
	a2aMsgTypeHandshake = 1
	a2aMsgTypeTask      = 2
	a2aMsgTypeEvent     = 3
	a2aMsgTypeResult    = 4
)

func handleA2A(ctx context.Context, env core.Envelope) ([]core.Envelope, error) {
	return handleA2AWithBackend(ctx, env, defaultBackends.a2a)
}

func (s *Server) handleA2A(ctx context.Context, env core.Envelope) ([]core.Envelope, error) {
	return handleA2AWithBackend(ctx, env, s.runtime.a2a)
}

func handleA2AWithBackend(_ context.Context, env core.Envelope, backend A2ABackend) ([]core.Envelope, error) {
	now := uint64(time.Now().UnixMilli())

	switch env.MsgType {
	case a2aMsgTypeHandshake:
		hs, err := p1a2a.DecodePayloadHandshake(env.Payload)
		if err != nil {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid A2A handshake payload: %w", err))
		}
		if strings.TrimSpace(hs.AgentID) == "" {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("agent_id required"))
		}
		return nil, nil

	case a2aMsgTypeTask:
		task, err := p1a2a.DecodePayloadTask(env.Payload)
		if err != nil {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid A2A task payload: %w", err))
		}
		if len(task.TaskID) == 0 {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("task_id required"))
		}
		if strings.TrimSpace(task.Kind) == "" {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("task kind required"))
		}

		created, err := backend.UpsertTask(task.TaskID, task.Kind, task.Input)
		if err != nil {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("conflicting duplicate task_id"))
		}
		if !created {
			return nil, nil
		}

		if strings.HasPrefix(strings.ToLower(task.Kind), "unsupported") {
			payload, err := p1a2a.EncodePayloadResult(p1a2a.Result{
				TaskID:       task.TaskID,
				OK:           false,
				ErrorMessage: "unsupported capability",
			})
			if err != nil {
				return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("encode A2A unsupported-capability result: %w", err))
			}
			return []core.Envelope{newA2AEnvelope(env.MsgID, a2aMsgTypeResult, now, payload)}, nil
		}

		if bytes.Contains(bytes.ToLower(task.Input), []byte("malformed")) {
			payload, err := p1a2a.EncodePayloadResult(p1a2a.Result{
				TaskID:       task.TaskID,
				OK:           false,
				ErrorMessage: "malformed task input",
			})
			if err != nil {
				return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("encode A2A malformed-input result: %w", err))
			}
			return []core.Envelope{newA2AEnvelope(env.MsgID, a2aMsgTypeResult, now, payload)}, nil
		}
		return nil, nil

	case a2aMsgTypeEvent:
		ev, err := p1a2a.DecodePayloadEvent(env.Payload)
		if err != nil {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid A2A event payload: %w", err))
		}
		if len(ev.TaskID) == 0 {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("task_id required"))
		}
		if strings.TrimSpace(ev.Message) == "" && len(ev.EventPayload) == 0 {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("event content required"))
		}
		state, ok := backend.GetTask(ev.TaskID)
		if !ok {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("unknown task_id"))
		}
		if state.Terminal {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("event after terminal result"))
		}
		return nil, nil

	case a2aMsgTypeResult:
		res, err := p1a2a.DecodePayloadResult(env.Payload)
		if err != nil {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid A2A result payload: %w", err))
		}
		if len(res.TaskID) == 0 {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("task_id required"))
		}

		err = backend.SetTerminal(res.TaskID, res.OK, res.Output, res.ErrorMessage)
		if err != nil {
			switch err {
			case errA2AUnknownTask:
				return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("unknown task_id"))
			case errA2ATerminalConflict:
				return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("conflicting duplicate terminal result"))
			default:
				return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("set A2A terminal result: %w", err))
			}
		}
		return nil, nil

	default:
		return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid A2A msg_type %d", env.MsgType))
	}
}

func newA2AEnvelope(msgID []byte, msgType, ts uint64, payload []byte) core.Envelope {
	return core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileA2A,
		MsgType:   msgType,
		MsgID:     append([]byte(nil), msgID...),
		TsUnixMs:  ts,
		Payload:   payload,
	}
}
