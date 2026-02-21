package server

import (
	"context"
	"fmt"
	"strings"
	"time"

	"swp-spec-kit/poc/internal/core"
	"swp-spec-kit/poc/internal/p1policyhint"
)

const (
	policyHintMsgTypeSet       = 1
	policyHintMsgTypeAck       = 2
	policyHintMsgTypeViolation = 3
	policyHintMsgTypeErr       = 4
)

var knownPolicyHintKeys = map[string]struct{}{
	"no_external_network": {},
	"no_pii":              {},
	"cost_limit":          {},
	"region":              {},
}

func handleSWPPolicyHint(ctx context.Context, env core.Envelope) ([]core.Envelope, error) {
	return handleSWPPolicyHintWithBackend(ctx, env, defaultBackends.policyHint)
}

func (s *Server) handleSWPPolicyHint(ctx context.Context, env core.Envelope) ([]core.Envelope, error) {
	return handleSWPPolicyHintWithBackend(ctx, env, s.runtime.policyHint)
}

func handleSWPPolicyHintWithBackend(_ context.Context, env core.Envelope, backend PolicyHintBackend) ([]core.Envelope, error) {
	now := uint64(time.Now().UnixMilli())

	switch env.MsgType {
	case policyHintMsgTypeSet:
		set, err := p1policyhint.DecodePayloadSet(env.Payload)
		if err != nil {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid POLICYHINT set payload: %w", err))
		}
		for _, c := range set.Constraints {
			if strings.TrimSpace(c.Key) == "" {
				return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("constraint key required"))
			}
			mode := strings.ToUpper(strings.TrimSpace(c.Mode))
			if mode == "" {
				mode = "MAY"
			}
			if mode != "MUST" && mode != "SHOULD" && mode != "MAY" {
				return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid constraint mode"))
			}

			_, known := knownPolicyHintKeys[c.Key]
			if !known && mode == "MUST" {
				payload, err := p1policyhint.EncodePayloadViolation(p1policyhint.PolicyViolation{
					Key:        c.Key,
					ScopeRef:   c.ScopeRef,
					ReasonCode: "UNKNOWN_KEY",
				})
				if err != nil {
					return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("encode POLICYHINT unknown-key violation: %w", err))
				}
				return []core.Envelope{newPolicyHintEnvelope(env.MsgID, policyHintMsgTypeViolation, now, payload)}, nil
			}

			existing, ok := backend.GetConstraint(c.Key)
			if ok && strings.ToUpper(strings.TrimSpace(existing.Mode)) == "MUST" && mode == "MUST" && existing.Value != c.Value {
				payload, err := p1policyhint.EncodePayloadViolation(p1policyhint.PolicyViolation{
					Key:        c.Key,
					ScopeRef:   c.ScopeRef,
					ReasonCode: "CONFLICT",
				})
				if err != nil {
					return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("encode POLICYHINT conflict violation: %w", err))
				}
				return []core.Envelope{newPolicyHintEnvelope(env.MsgID, policyHintMsgTypeViolation, now, payload)}, nil
			}
			backend.SetConstraint(p1policyhint.Constraint{Key: c.Key, Value: c.Value, Mode: mode, ScopeRef: c.ScopeRef})
		}

		ackPayload, err := p1policyhint.EncodePayloadAck(p1policyhint.PolicyHintAck{AckID: string(env.MsgID)})
		if err != nil {
			return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("encode POLICYHINT ack: %w", err))
		}
		return []core.Envelope{newPolicyHintEnvelope(env.MsgID, policyHintMsgTypeAck, now, ackPayload)}, nil

	case policyHintMsgTypeAck:
		if _, err := p1policyhint.DecodePayloadAck(env.Payload); err != nil {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid POLICYHINT ack payload: %w", err))
		}
		return nil, nil

	case policyHintMsgTypeViolation:
		if _, err := p1policyhint.DecodePayloadViolation(env.Payload); err != nil {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid POLICYHINT violation payload: %w", err))
		}
		return nil, nil

	default:
		return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid POLICYHINT msg_type %d", env.MsgType))
	}
}

func newPolicyHintEnvelope(msgID []byte, msgType, ts uint64, payload []byte) core.Envelope {
	return core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileSWPPolicyHint,
		MsgType:   msgType,
		MsgID:     append([]byte(nil), msgID...),
		TsUnixMs:  ts,
		Payload:   payload,
	}
}
