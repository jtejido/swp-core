package main

import (
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"swp-spec-kit/poc/internal/core"
	"swp-spec-kit/poc/internal/p1rpc"
)

const (
	profileMCPMap = 1
	profileSWPRPC = 12
)

func main() {
	addr := flag.String("addr", "127.0.0.1:7777", "server TCP address")
	flag.Parse()

	conn, err := net.Dial("tcp", *addr)
	if err != nil {
		log.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	if err := demoMCP(conn); err != nil {
		log.Fatalf("mcp demo failed: %v", err)
	}
	if err := demoSWPRPCStream(conn); err != nil {
		log.Fatalf("swp-rpc stream demo failed: %v", err)
	}

	log.Println("demo complete")
}

func demoMCP(conn net.Conn) error {
	msgID, err := newMsgID(16)
	if err != nil {
		return err
	}
	payload := map[string]any{
		"jsonrpc": "2.0",
		"id":      "req-tools-list",
		"method":  "tools/list",
		"params":  map[string]any{},
	}
	p, _ := json.Marshal(payload)

	req := core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: profileMCPMap,
		MsgType:   1,
		MsgID:     msgID,
		TsUnixMs:  uint64(time.Now().UnixMilli()),
		Payload:   p,
	}
	if err := writeEnvelope(conn, req); err != nil {
		return err
	}

	resp, err := readEnvelope(conn)
	if err != nil {
		return err
	}
	if resp.ProfileID != profileMCPMap || resp.MsgType != 2 {
		return fmt.Errorf("unexpected MCP response envelope profile=%d msg_type=%d", resp.ProfileID, resp.MsgType)
	}

	var pretty map[string]any
	_ = json.Unmarshal(resp.Payload, &pretty)
	b, _ := json.MarshalIndent(pretty, "", "  ")
	log.Printf("MCP tools/list response:\n%s", string(b))
	return nil
}

func demoSWPRPCStream(conn net.Conn) error {
	msgID, err := newMsgID(16)
	if err != nil {
		return err
	}
	params, _ := json.Marshal(map[string]any{"count": 3})
	p, err := p1rpc.EncodePayloadReq(p1rpc.RpcReq{
		RPCID:  []byte("rpc-stream-1"),
		Method: "demo.stream.count",
		Params: params,
	})
	if err != nil {
		return fmt.Errorf("encode rpc req payload: %w", err)
	}

	req := core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: profileSWPRPC,
		MsgType:   1,
		MsgID:     msgID,
		TsUnixMs:  uint64(time.Now().UnixMilli()),
		Payload:   p,
	}
	if err := writeEnvelope(conn, req); err != nil {
		return err
	}

	for {
		resp, err := readEnvelope(conn)
		if err != nil {
			return err
		}
		if resp.ProfileID != profileSWPRPC {
			continue
		}
		if string(resp.MsgID) != string(msgID) {
			continue
		}

		switch resp.MsgType {
		case 4:
			item, err := p1rpc.DecodePayloadStreamItem(resp.Payload)
			if err != nil {
				return fmt.Errorf("decode stream item: %w", err)
			}
			log.Printf("SWP-RPC stream item: rpc_id=%s seq=%d item=%s", string(item.RPCID), item.SeqNo, string(item.Item))
		case 2:
			out, err := p1rpc.DecodePayloadResp(resp.Payload)
			if err != nil {
				return fmt.Errorf("decode terminal response: %w", err)
			}
			log.Printf("SWP-RPC terminal response: rpc_id=%s result=%s", string(out.RPCID), string(out.Result))
			return nil
		case 3:
			e, err := p1rpc.DecodePayloadErr(resp.Payload)
			if err != nil {
				return fmt.Errorf("decode terminal error: %w", err)
			}
			return fmt.Errorf("SWP-RPC terminal error: code=%s retryable=%t message=%s", e.ErrorCode, e.Retryable, e.ErrorMessage)
		default:
			log.Printf("SWP-RPC ignored msg_type=%d payload=%s", resp.MsgType, string(resp.Payload))
		}
	}
}

func writeEnvelope(conn net.Conn, env core.Envelope) error {
	body, err := core.EncodeEnvelopeE1(env)
	if err != nil {
		return err
	}
	return core.WriteFrame(conn, body, core.DefaultLimits().MaxFrameBytes)
}

func readEnvelope(conn net.Conn) (core.Envelope, error) {
	if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return core.Envelope{}, err
	}
	frame, err := core.ReadFrame(conn, core.DefaultLimits().MaxFrameBytes)
	if err != nil {
		return core.Envelope{}, err
	}
	return core.DecodeEnvelopeE1(frame, core.DefaultLimits())
}

func newMsgID(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}
