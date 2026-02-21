package server

import (
	"context"
	"fmt"

	"swp-spec-kit/poc/internal/core"
	"swp-spec-kit/poc/internal/p1rpc"
	runtimeclock "swp-spec-kit/poc/internal/runtime/clock"
)

const (
	rpcMsgTypeReq        = 1
	rpcMsgTypeResp       = 2
	rpcMsgTypeErr        = 3
	rpcMsgTypeStreamItem = 4
	rpcMsgTypeCancel     = 5
)

func handleSWPRPC(ctx context.Context, env core.Envelope) ([]core.Envelope, error) {
	return handleSWPRPCWithBackend(ctx, env, defaultBackends.rpc, nil)
}

func (s *Server) handleSWPRPC(ctx context.Context, env core.Envelope) ([]core.Envelope, error) {
	return handleSWPRPCWithBackend(ctx, env, s.runtime.rpc, func(eventType, severity string, body map[string]any, rpcID []byte) {
		s.emitProfileEvent(ctx, env, eventType, severity, body, nil, rpcID)
	})
}

func handleSWPRPCWithBackend(
	_ context.Context,
	env core.Envelope,
	backend RPCBackend,
	emit func(eventType, severity string, body map[string]any, rpcID []byte),
) ([]core.Envelope, error) {
	now := runtimeclock.UnixMilli(nil)

	switch env.MsgType {
	case rpcMsgTypeCancel:
		cerr, err := backend.HandleCancel()
		if err != nil {
			return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("backend RPC cancel: %w", err))
		}
		if emit != nil {
			emit("swp.rpc.cancel", "info", map[string]any{
				"code": cerr.ErrorCode,
			}, cerr.RPCID)
		}
		payload, err := p1rpc.EncodePayloadErr(cerr)
		if err != nil {
			return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("encode RPC error payload: %w", err))
		}
		return []core.Envelope{newRPCEnvelope(env.MsgID, rpcMsgTypeErr, now, payload)}, nil
	case rpcMsgTypeReq:
	default:
		return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid SWP-RPC msg_type %d", env.MsgType))
	}

	req, err := p1rpc.DecodePayloadReq(env.Payload)
	if err != nil {
		return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid RPC request protobuf payload: %w", err))
	}
	if req.Method == "" {
		return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("missing method"))
	}
	if emit != nil {
		emit("swp.rpc.request", "info", map[string]any{
			"method": req.Method,
		}, req.RPCID)
	}

	msgs, err := backend.HandleRequest(req)
	if err != nil {
		if emit != nil {
			emit("swp.rpc.error", "error", map[string]any{
				"method": req.Method,
				"code":   "INTERNAL_ERROR",
			}, req.RPCID)
		}
		return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("backend RPC request: %w", err))
	}

	out := make([]core.Envelope, 0, len(msgs))
	for _, msg := range msgs {
		switch msg.MsgType {
		case rpcMsgTypeResp:
			payload, err := p1rpc.EncodePayloadResp(msg.Resp)
			if err != nil {
				return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("encode RPC response payload: %w", err))
			}
			if emit != nil {
				emit("swp.rpc.response", "info", map[string]any{
					"method": req.Method,
				}, msg.Resp.RPCID)
			}
			out = append(out, newRPCEnvelope(env.MsgID, rpcMsgTypeResp, now, payload))
		case rpcMsgTypeErr:
			payload, err := p1rpc.EncodePayloadErr(msg.Err)
			if err != nil {
				return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("encode RPC error payload: %w", err))
			}
			if emit != nil {
				emit("swp.rpc.response", "warn", map[string]any{
					"method": req.Method,
					"code":   msg.Err.ErrorCode,
				}, msg.Err.RPCID)
			}
			out = append(out, newRPCEnvelope(env.MsgID, rpcMsgTypeErr, now, payload))
		case rpcMsgTypeStreamItem:
			payload, err := p1rpc.EncodePayloadStreamItem(msg.StreamItem)
			if err != nil {
				return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("encode RPC stream item payload: %w", err))
			}
			if emit != nil {
				emit("swp.rpc.stream", "debug", map[string]any{
					"method": req.Method,
					"seq_no": msg.StreamItem.SeqNo,
				}, msg.StreamItem.RPCID)
			}
			out = append(out, newRPCEnvelope(env.MsgID, rpcMsgTypeStreamItem, now, payload))
		default:
			return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("backend returned unsupported RPC message type %d", msg.MsgType))
		}
	}

	return out, nil
}

func newRPCEnvelope(msgID []byte, msgType, ts uint64, payload []byte) core.Envelope {
	return core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileSWPRPC,
		MsgType:   msgType,
		MsgID:     append([]byte(nil), msgID...),
		TsUnixMs:  ts,
		Payload:   payload,
	}
}
