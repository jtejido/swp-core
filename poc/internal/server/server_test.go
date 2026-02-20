package server

import (
	"context"
	"encoding/json"
	"testing"

	"swp-spec-kit/poc/internal/core"
	"swp-spec-kit/poc/internal/p1rpc"
)

func TestHandleMCPToolsList(t *testing.T) {
	payload, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      "x",
		"method":  "tools/list",
		"params":  map[string]any{},
	})
	env := core.Envelope{
		Version:   1,
		ProfileID: ProfileMCPMap,
		MsgType:   1,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   payload,
	}

	out, err := handleMCP(context.Background(), env)
	if err != nil {
		t.Fatalf("handleMCP failed: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 response, got %d", len(out))
	}
	if out[0].MsgType != 2 {
		t.Fatalf("expected response msg_type=2, got %d", out[0].MsgType)
	}
	if !json.Valid(out[0].Payload) {
		t.Fatalf("expected valid JSON payload")
	}
}

func TestHandleSWPRPCStreaming(t *testing.T) {
	params, _ := json.Marshal(map[string]any{
		"count": 2,
	})
	payload, err := p1rpc.EncodePayloadReq(p1rpc.RpcReq{
		RPCID:  []byte("r1"),
		Method: "demo.stream.count",
		Params: params,
	})
	if err != nil {
		t.Fatalf("encode request payload failed: %v", err)
	}
	env := core.Envelope{
		Version:   1,
		ProfileID: ProfileSWPRPC,
		MsgType:   1,
		MsgID:     []byte("abcdef1234567890"),
		Payload:   payload,
	}

	out, err := handleSWPRPC(context.Background(), env)
	if err != nil {
		t.Fatalf("handleSWPRPC failed: %v", err)
	}
	if len(out) != 3 {
		t.Fatalf("expected 3 envelopes (2 stream + 1 terminal), got %d", len(out))
	}
	if out[0].MsgType != 4 || out[1].MsgType != 4 || out[2].MsgType != 2 {
		t.Fatalf("unexpected message type sequence: %d, %d, %d", out[0].MsgType, out[1].MsgType, out[2].MsgType)
	}

	item1, err := p1rpc.DecodePayloadStreamItem(out[0].Payload)
	if err != nil {
		t.Fatalf("decode stream item failed: %v", err)
	}
	if item1.SeqNo != 1 {
		t.Fatalf("expected seq_no 1, got %d", item1.SeqNo)
	}

	resp, err := p1rpc.DecodePayloadResp(out[2].Payload)
	if err != nil {
		t.Fatalf("decode terminal response failed: %v", err)
	}
	if string(resp.RPCID) != "r1" {
		t.Fatalf("expected rpc_id r1, got %q", string(resp.RPCID))
	}
}
