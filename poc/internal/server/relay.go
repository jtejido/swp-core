package server

import (
	"context"
	"fmt"
	"time"

	"swp-spec-kit/poc/internal/core"
	"swp-spec-kit/poc/internal/p1relay"
)

const (
	relayMsgTypePublish = 1
	relayMsgTypeAck     = 2
	relayMsgTypeNack    = 3
	relayMsgTypeStatus  = 4
	relayMsgTypeErr     = 5
)

type relayDelivery struct {
	attemptCount uint32
	state        string
}

func handleSWPRelay(ctx context.Context, env core.Envelope) ([]core.Envelope, error) {
	return handleSWPRelayWithBackend(ctx, env, defaultBackends.relay)
}

func (s *Server) handleSWPRelay(ctx context.Context, env core.Envelope) ([]core.Envelope, error) {
	return handleSWPRelayWithBackend(ctx, env, s.runtime.relay)
}

func handleSWPRelayWithBackend(_ context.Context, env core.Envelope, backend RelayBackend) ([]core.Envelope, error) {
	now := uint64(time.Now().UnixMilli())

	switch env.MsgType {
	case relayMsgTypePublish:
		pub, err := p1relay.DecodePayloadPublish(env.Payload)
		if err != nil {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid RELAY publish payload: %w", err))
		}
		if len(pub.DeliveryID) == 0 {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("delivery_id required"))
		}

		created, attempts, state := backend.CreateDelivery(pub.DeliveryID)
		if !created {
			statusPayload, err := p1relay.EncodePayloadStatus(p1relay.RelayStatus{
				DeliveryID:   pub.DeliveryID,
				State:        "duplicate",
				AttemptCount: attempts,
			})
			if err != nil {
				return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("encode RELAY duplicate status: %w", err))
			}
			_ = state
			return []core.Envelope{newRelayEnvelope(env.MsgID, relayMsgTypeStatus, now, statusPayload)}, nil
		}

		ackPayload, err := p1relay.EncodePayloadAck(p1relay.RelayAck{DeliveryID: pub.DeliveryID})
		if err != nil {
			return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("encode RELAY ack: %w", err))
		}
		return []core.Envelope{newRelayEnvelope(env.MsgID, relayMsgTypeAck, now, ackPayload)}, nil

	case relayMsgTypeAck:
		ack, err := p1relay.DecodePayloadAck(env.Payload)
		if err != nil {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid RELAY ack payload: %w", err))
		}
		if len(ack.DeliveryID) == 0 {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("delivery_id required"))
		}
		backend.MarkAck(ack.DeliveryID)
		return nil, nil

	case relayMsgTypeNack:
		nack, err := p1relay.DecodePayloadNack(env.Payload)
		if err != nil {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid RELAY nack payload: %w", err))
		}
		if len(nack.DeliveryID) == 0 {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("delivery_id required"))
		}

		attempts, state := backend.MarkNack(nack.DeliveryID, nack.Retryable)
		statusPayload, err := p1relay.EncodePayloadStatus(p1relay.RelayStatus{
			DeliveryID:   nack.DeliveryID,
			State:        state,
			AttemptCount: attempts,
		})
		if err != nil {
			return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("encode RELAY status: %w", err))
		}
		return []core.Envelope{newRelayEnvelope(env.MsgID, relayMsgTypeStatus, now, statusPayload)}, nil

	case relayMsgTypeStatus:
		statusReq, err := p1relay.DecodePayloadStatus(env.Payload)
		if err != nil {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid RELAY status payload: %w", err))
		}
		if len(statusReq.DeliveryID) == 0 {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("delivery_id required"))
		}

		attempts, state, ok := backend.GetDelivery(statusReq.DeliveryID)
		if !ok {
			errPayload, err := p1relay.EncodePayloadErr(p1relay.RelayErr{Code: "NOT_FOUND", Message: "delivery not found"})
			if err != nil {
				return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("encode RELAY not-found error: %w", err))
			}
			return []core.Envelope{newRelayEnvelope(env.MsgID, relayMsgTypeErr, now, errPayload)}, nil
		}
		statusPayload, err := p1relay.EncodePayloadStatus(p1relay.RelayStatus{
			DeliveryID:   statusReq.DeliveryID,
			State:        state,
			AttemptCount: attempts,
		})
		if err != nil {
			return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("encode RELAY status response: %w", err))
		}
		return []core.Envelope{newRelayEnvelope(env.MsgID, relayMsgTypeStatus, now, statusPayload)}, nil

	case relayMsgTypeErr:
		if _, err := p1relay.DecodePayloadErr(env.Payload); err != nil {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid RELAY err payload: %w", err))
		}
		return nil, nil

	default:
		return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid RELAY msg_type %d", env.MsgType))
	}
}

func newRelayEnvelope(msgID []byte, msgType, ts uint64, payload []byte) core.Envelope {
	return core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileSWPRelay,
		MsgType:   msgType,
		MsgID:     append([]byte(nil), msgID...),
		TsUnixMs:  ts,
		Payload:   payload,
	}
}
