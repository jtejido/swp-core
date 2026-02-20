package server

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"swp-spec-kit/poc/internal/core"
)

const (
	mcpMsgTypeRequest      = 1
	mcpMsgTypeResponse     = 2
	mcpMsgTypeNotification = 3
)

type mcpRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type mcpResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id,omitempty"`
	Result  any    `json:"result,omitempty"`
	Error   any    `json:"error,omitempty"`
}

func handleMCP(_ context.Context, env core.Envelope) ([]core.Envelope, error) {
	if env.MsgType != mcpMsgTypeRequest && env.MsgType != mcpMsgTypeNotification {
		return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid MCP msg_type %d", env.MsgType))
	}

	var req mcpRequest
	if err := json.Unmarshal(env.Payload, &req); err != nil {
		return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid JSON-RPC payload: %w", err))
	}

	if env.MsgType == mcpMsgTypeNotification {
		return nil, nil
	}

	resp := mcpResponse{JSONRPC: "2.0", ID: req.ID}
	switch req.Method {
	case "tools/list":
		resp.Result = map[string]any{
			"tools": []map[string]any{
				{
					"name":        "echo",
					"description": "Echo tool for SWP POC",
					"inputSchema": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"text": map[string]any{"type": "string"},
						},
						"required": []string{"text"},
					},
				},
			},
		}
	case "tools/call":
		var p struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		}
		_ = json.Unmarshal(req.Params, &p)
		if p.Name == "echo" {
			resp.Result = map[string]any{
				"content": []map[string]any{{
					"type": "text",
					"text": fmt.Sprintf("echo: %v", p.Arguments["text"]),
				}},
			}
		} else {
			resp.Error = map[string]any{"code": -32601, "message": "tool not found"}
		}
	default:
		resp.Error = map[string]any{"code": -32601, "message": "method not found"}
	}

	payload, err := json.Marshal(resp)
	if err != nil {
		return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("marshal response: %w", err))
	}

	return []core.Envelope{{
		Version:   core.CoreVersion,
		ProfileID: ProfileMCPMap,
		MsgType:   mcpMsgTypeResponse,
		MsgID:     append([]byte(nil), env.MsgID...),
		TsUnixMs:  uint64(time.Now().UnixMilli()),
		Payload:   payload,
	}}, nil
}
