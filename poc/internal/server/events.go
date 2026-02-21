package server

import (
	"context"
	"fmt"

	"swp-spec-kit/poc/internal/core"
	"swp-spec-kit/poc/internal/p1events"
	runtimeclock "swp-spec-kit/poc/internal/runtime/clock"
	runtimecontext "swp-spec-kit/poc/internal/runtime/context"
	runtimevalidate "swp-spec-kit/poc/internal/runtime/validate"
)

const (
	eventsMsgTypePublish     = 1
	eventsMsgTypeSubscribe   = 2
	eventsMsgTypeUnsubscribe = 3
	eventsMsgTypeBatch       = 4
	eventsMsgTypeErr         = 5
)

func handleSWPEvents(ctx context.Context, env core.Envelope) ([]core.Envelope, error) {
	return handleSWPEventsWithBackend(ctx, env, defaultBackends.events, defaultBackends.obs)
}

func (s *Server) handleSWPEvents(ctx context.Context, env core.Envelope) ([]core.Envelope, error) {
	return handleSWPEventsWithBackend(ctx, env, s.runtime.events, s.runtime.obs)
}

func handleSWPEventsWithBackend(ctx context.Context, env core.Envelope, backend EventsBackend, obsBackend OBSBackend) ([]core.Envelope, error) {
	now := runtimeclock.UnixMilli(nil)

	switch env.MsgType {
	case eventsMsgTypePublish:
		pub, err := p1events.DecodePayloadPublish(env.Payload)
		if err != nil {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid EVENTS publish payload: %w", err))
		}
		pub.Event = enrichEventRecord(ctx, env, pub.Event, obsBackend, now)
		if err := validateEventRecord(pub.Event); err != nil {
			return nil, core.Wrap(core.CodeInvalidEnvelope, err)
		}
		if err := backend.Publish(pub.Event); err != nil {
			return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("backend EVENTS publish: %w", err))
		}
		return nil, nil

	case eventsMsgTypeSubscribe:
		sub, err := p1events.DecodePayloadSubscribe(env.Payload)
		if err != nil {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid EVENTS subscribe payload: %w", err))
		}
		events, err := backend.Subscribe(sub.Filter)
		if err != nil {
			return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("backend EVENTS subscribe: %w", err))
		}
		payload, err := p1events.EncodePayloadBatch(p1events.EvtBatch{Events: events})
		if err != nil {
			return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("encode EVENTS batch response: %w", err))
		}
		return []core.Envelope{newEventsEnvelope(env.MsgID, eventsMsgTypeBatch, now, payload)}, nil

	case eventsMsgTypeUnsubscribe:
		unsub, err := p1events.DecodePayloadUnsubscribe(env.Payload)
		if err != nil {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid EVENTS unsubscribe payload: %w", err))
		}
		if err := backend.Unsubscribe(unsub.SubscriptionID); err != nil {
			return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("backend EVENTS unsubscribe: %w", err))
		}
		return nil, nil

	default:
		return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid EVENTS msg_type %d", env.MsgType))
	}
}

func enrichEventRecord(
	ctx context.Context,
	env core.Envelope,
	ev p1events.EventRecord,
	obsBackend OBSBackend,
	now uint64,
) p1events.EventRecord {
	if ev.TsUnixMs == 0 {
		ev.TsUnixMs = now
	}

	if len(ev.MsgID) == 0 {
		if meta, ok := runtimecontext.MessageMetaFromContext(ctx); ok && len(meta.MsgID) > 0 {
			ev.MsgID = append([]byte(nil), meta.MsgID...)
		} else {
			ev.MsgID = append([]byte(nil), env.MsgID...)
		}
	}

	if corr, ok := runtimecontext.CorrelationFromContext(ctx); ok {
		if len(ev.TaskID) == 0 && len(corr.TaskID) > 0 {
			ev.TaskID = append([]byte(nil), corr.TaskID...)
		}
		if len(ev.RPCID) == 0 && len(corr.RPCID) > 0 {
			ev.RPCID = append([]byte(nil), corr.RPCID...)
		}
	}

	if obsBackend != nil && (!runtimevalidate.HasCorrelation(ev.MsgID, ev.TaskID, ev.RPCID) || len(ev.TaskID) == 0 || len(ev.RPCID) == 0) {
		doc := obsBackend.GetDoc()
		if len(ev.MsgID) == 0 && len(doc.MsgID) > 0 {
			ev.MsgID = append([]byte(nil), doc.MsgID...)
		}
		if len(ev.TaskID) == 0 && len(doc.TaskID) > 0 {
			ev.TaskID = append([]byte(nil), doc.TaskID...)
		}
		if len(ev.RPCID) == 0 && len(doc.RPCID) > 0 {
			ev.RPCID = append([]byte(nil), doc.RPCID...)
		}
	}

	return ev
}

func validateEventRecord(ev p1events.EventRecord) error {
	if err := runtimevalidate.RequireNonEmpty("event_id", ev.EventID); err != nil {
		return err
	}
	if err := runtimevalidate.RequireNonEmpty("event_type", ev.EventType); err != nil {
		return err
	}
	if err := runtimevalidate.Severity(ev.Severity); err != nil {
		return err
	}
	if !runtimevalidate.HasCorrelation(ev.MsgID, ev.TaskID, ev.RPCID) {
		return fmt.Errorf("at least one correlation key required")
	}
	return nil
}

func newEventsEnvelope(msgID []byte, msgType, ts uint64, payload []byte) core.Envelope {
	return core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileSWPEvents,
		MsgType:   msgType,
		MsgID:     append([]byte(nil), msgID...),
		TsUnixMs:  ts,
		Payload:   payload,
	}
}
