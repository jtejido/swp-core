package server

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"swp-spec-kit/poc/internal/core"
	"swp-spec-kit/poc/internal/p1tooldisc"
)

const (
	tooldiscMsgTypeListReq  = 1
	tooldiscMsgTypeListResp = 2
	tooldiscMsgTypeGetReq   = 3
	tooldiscMsgTypeGetResp  = 4
	tooldiscMsgTypeErr      = 5
)

func handleSWPToolDisc(ctx context.Context, env core.Envelope) ([]core.Envelope, error) {
	return handleSWPToolDiscWithBackend(ctx, env, defaultBackends.tooldisc)
}

func (s *Server) handleSWPToolDisc(ctx context.Context, env core.Envelope) ([]core.Envelope, error) {
	return handleSWPToolDiscWithBackend(ctx, env, s.runtime.tooldisc)
}

func handleSWPToolDiscWithBackend(_ context.Context, env core.Envelope, backend ToolDiscBackend) ([]core.Envelope, error) {
	now := uint64(time.Now().UnixMilli())

	switch env.MsgType {
	case tooldiscMsgTypeListReq:
		req, err := p1tooldisc.DecodePayloadListReq(env.Payload)
		if err != nil {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid TOOLDISC list request payload: %w", err))
		}
		resp := listTools(req, backend.ListTools())
		payload, err := p1tooldisc.EncodePayloadListResp(resp)
		if err != nil {
			return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("encode TOOLDISC list response: %w", err))
		}
		return []core.Envelope{newToolDiscEnvelope(env.MsgID, tooldiscMsgTypeListResp, now, payload)}, nil

	case tooldiscMsgTypeGetReq:
		req, err := p1tooldisc.DecodePayloadGetReq(env.Payload)
		if err != nil {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid TOOLDISC get request payload: %w", err))
		}
		tool, ok := backend.GetTool(req.ToolID, req.Version)
		if !ok {
			errPayload, err := p1tooldisc.EncodePayloadErr(p1tooldisc.TooldiscErr{
				Code:    "NOT_FOUND",
				Message: "tool not found",
			})
			if err != nil {
				return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("encode TOOLDISC not-found error payload: %w", err))
			}
			return []core.Envelope{newToolDiscEnvelope(env.MsgID, tooldiscMsgTypeErr, now, errPayload)}, nil
		}
		payload, err := p1tooldisc.EncodePayloadGetResp(p1tooldisc.TooldiscGetResp{Tool: tool})
		if err != nil {
			return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("encode TOOLDISC get response: %w", err))
		}
		return []core.Envelope{newToolDiscEnvelope(env.MsgID, tooldiscMsgTypeGetResp, now, payload)}, nil

	default:
		return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid TOOLDISC msg_type %d", env.MsgType))
	}
}

func listTools(req p1tooldisc.TooldiscListReq, tools []p1tooldisc.ToolDescriptor) p1tooldisc.TooldiscListResp {
	filtered := filterTools(req.Filter, tools)
	start := 0
	if req.PageToken != "" {
		if n, err := strconv.Atoi(req.PageToken); err == nil && n >= 0 {
			start = n
		}
	}
	if start > len(filtered) {
		start = len(filtered)
	}

	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = len(filtered)
	}
	if pageSize > 100 {
		pageSize = 100
	}

	end := start + pageSize
	if end > len(filtered) {
		end = len(filtered)
	}

	resp := p1tooldisc.TooldiscListResp{
		Tools: filtered[start:end],
	}
	if end < len(filtered) {
		resp.NextPageToken = strconv.Itoa(end)
	}
	return resp
}

func filterTools(filter string, tools []p1tooldisc.ToolDescriptor) []p1tooldisc.ToolDescriptor {
	if strings.TrimSpace(filter) == "" {
		out := make([]p1tooldisc.ToolDescriptor, 0, len(tools))
		for _, t := range tools {
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
	f := strings.ToLower(strings.TrimSpace(filter))
	out := make([]p1tooldisc.ToolDescriptor, 0, len(tools))
	for _, t := range tools {
		if strings.Contains(strings.ToLower(t.ToolID), f) || strings.Contains(strings.ToLower(t.Name), f) {
			out = append(out, p1tooldisc.ToolDescriptor{
				ToolID:            t.ToolID,
				Name:              t.Name,
				Version:           t.Version,
				SchemaRef:         t.SchemaRef,
				DescriptorPayload: append([]byte(nil), t.DescriptorPayload...),
			})
		}
	}
	return out
}

func newToolDiscEnvelope(msgID []byte, msgType, ts uint64, payload []byte) core.Envelope {
	return core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileSWPToolDisc,
		MsgType:   msgType,
		MsgID:     append([]byte(nil), msgID...),
		TsUnixMs:  ts,
		Payload:   payload,
	}
}
