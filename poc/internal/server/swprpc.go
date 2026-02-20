package server

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"swp-spec-kit/poc/internal/core"
	"swp-spec-kit/poc/internal/p1rpc"
)

const (
	rpcMsgTypeReq        = 1
	rpcMsgTypeResp       = 2
	rpcMsgTypeErr        = 3
	rpcMsgTypeStreamItem = 4
	rpcMsgTypeCancel     = 5
)

func handleSWPRPC(_ context.Context, env core.Envelope) ([]core.Envelope, error) {
	now := uint64(time.Now().UnixMilli())

	switch env.MsgType {
	case rpcMsgTypeCancel:
		resp, err := p1rpc.EncodePayloadErr(p1rpc.RpcErr{
			RPCID:        []byte{},
			ErrorCode:    "cancelled",
			Retryable:    false,
			ErrorMessage: "cancel received",
		})
		if err != nil {
			return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("encode RPC error payload: %w", err))
		}
		return []core.Envelope{newRPCEnvelope(env.MsgID, rpcMsgTypeErr, now, resp)}, nil
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

	switch req.Method {
	case "demo.echo":
		resp, err := p1rpc.EncodePayloadResp(p1rpc.RpcResp{
			RPCID:  req.RPCID,
			Result: req.Params,
		})
		if err != nil {
			return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("encode RPC response payload: %w", err))
		}
		return []core.Envelope{newRPCEnvelope(env.MsgID, rpcMsgTypeResp, now, resp)}, nil

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

		out := make([]core.Envelope, 0, count+1)
		for i := 1; i <= count; i++ {
			item, err := p1rpc.EncodePayloadStreamItem(p1rpc.RpcStreamItem{
				RPCID:      req.RPCID,
				SeqNo:      uint64(i),
				Item:       []byte(fmt.Sprintf("%d", i)),
				IsTerminal: false,
			})
			if err != nil {
				return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("encode RPC stream item payload: %w", err))
			}
			out = append(out, newRPCEnvelope(env.MsgID, rpcMsgTypeStreamItem, now, item))
		}
		terminalResult, _ := json.Marshal(map[string]any{"count": count, "done": true})
		terminal, err := p1rpc.EncodePayloadResp(p1rpc.RpcResp{
			RPCID:  req.RPCID,
			Result: terminalResult,
		})
		if err != nil {
			return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("encode RPC terminal response payload: %w", err))
		}
		out = append(out, newRPCEnvelope(env.MsgID, rpcMsgTypeResp, now, terminal))
		return out, nil

	case "demo.fail":
		errPayload, err := p1rpc.EncodePayloadErr(p1rpc.RpcErr{
			RPCID:        req.RPCID,
			ErrorCode:    "internal",
			Retryable:    false,
			ErrorMessage: "forced failure",
		})
		if err != nil {
			return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("encode RPC error payload: %w", err))
		}
		return []core.Envelope{newRPCEnvelope(env.MsgID, rpcMsgTypeErr, now, errPayload)}, nil

	default:
		errPayload, err := p1rpc.EncodePayloadErr(p1rpc.RpcErr{
			RPCID:        req.RPCID,
			ErrorCode:    "unknown_method",
			Retryable:    false,
			ErrorMessage: "unknown method",
		})
		if err != nil {
			return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("encode RPC unknown-method payload: %w", err))
		}
		return []core.Envelope{newRPCEnvelope(env.MsgID, rpcMsgTypeErr, now, errPayload)}, nil
	}
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
