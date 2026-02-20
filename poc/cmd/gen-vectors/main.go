package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"swp-spec-kit/poc/internal/core"
	"swp-spec-kit/poc/internal/p1rpc"
)

type expectation struct {
	OK            bool   `json:"ok"`
	Code          string `json:"code"`
	Version       uint64 `json:"version,omitempty"`
	ProfileID     uint64 `json:"profile_id,omitempty"`
	MsgType       uint64 `json:"msg_type,omitempty"`
	MinPayloadLen int    `json:"min_payload_len,omitempty"`
}

type vector struct {
	Name   string      `json:"name"`
	Bin    string      `json:"bin"`
	Expect expectation `json:"expect"`
}

func main() {
	base := filepath.Join("conformance", "vectors")
	if err := os.MkdirAll(base, 0o755); err != nil {
		panic(err)
	}

	mustWriteVector(base, "poc_0001_valid_mcp_request", validMCP(), expectation{
		OK:            true,
		Code:          string(core.CodeOK),
		Version:       1,
		ProfileID:     1,
		MsgType:       1,
		MinPayloadLen: 10,
	})

	mustWriteVector(base, "poc_0002_valid_swprpc_request", validRPC(), expectation{
		OK:            true,
		Code:          string(core.CodeOK),
		Version:       1,
		ProfileID:     12,
		MsgType:       1,
		MinPayloadLen: 10,
	})

	mustWriteVector(base, "poc_0003_invalid_version", invalidVersion(), expectation{
		OK:   false,
		Code: string(core.CodeUnsupportedVersion),
	})

	mustWriteVector(base, "poc_0004_invalid_empty_msg_id", emptyMsgID(), expectation{
		OK:   false,
		Code: string(core.CodeInvalidEnvelope),
	})

	mustWriteVector(base, "poc_0005_invalid_truncated_frame", truncatedFrame(), expectation{
		OK:   false,
		Code: string(core.CodeInvalidFrame),
	})

	mustWriteVector(base, "poc_0006_invalid_varint_overflow", varintOverflow(), expectation{
		OK:   false,
		Code: string(core.CodeInvalidFrame),
	})

	fmt.Println("generated POC vectors")
}

func mustWriteVector(base, name string, framed []byte, expect expectation) {
	binName := name + ".bin"
	jsonName := name + ".json"
	if err := os.WriteFile(filepath.Join(base, binName), framed, 0o644); err != nil {
		panic(err)
	}
	v := vector{Name: name, Bin: binName, Expect: expect}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(filepath.Join(base, jsonName), append(b, '\n'), 0o644); err != nil {
		panic(err)
	}
}

func frameFromEnvelope(env core.Envelope) []byte {
	body, err := core.EncodeEnvelopeE1(env)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	if err := core.WriteFrame(&buf, body, core.DefaultLimits().MaxFrameBytes); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func validMCP() []byte {
	env := core.Envelope{
		Version:   1,
		ProfileID: 1,
		MsgType:   1,
		Flags:     0,
		TsUnixMs:  uint64(time.Now().UnixMilli()),
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   []byte(`{"jsonrpc":"2.0","id":"1","method":"tools/list","params":{}}`),
	}
	return frameFromEnvelope(env)
}

func validRPC() []byte {
	payload, err := p1rpc.EncodePayloadReq(p1rpc.RpcReq{
		RPCID:  []byte("r1"),
		Method: "demo.echo",
		Params: []byte(`{"hello":"world"}`),
	})
	if err != nil {
		panic(err)
	}
	env := core.Envelope{
		Version:   1,
		ProfileID: 12,
		MsgType:   1,
		Flags:     0,
		TsUnixMs:  uint64(time.Now().UnixMilli()),
		MsgID:     []byte("abcdef1234567890"),
		Payload:   payload,
	}
	return frameFromEnvelope(env)
}

func invalidVersion() []byte {
	env := core.Envelope{
		Version:   2,
		ProfileID: 1,
		MsgType:   1,
		Flags:     0,
		TsUnixMs:  uint64(time.Now().UnixMilli()),
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   []byte(`{"jsonrpc":"2.0"}`),
	}
	return frameFromEnvelope(env)
}

func emptyMsgID() []byte {
	env := core.Envelope{
		Version:   1,
		ProfileID: 1,
		MsgType:   1,
		Flags:     0,
		TsUnixMs:  uint64(time.Now().UnixMilli()),
		MsgID:     []byte{},
		Payload:   []byte(`{"jsonrpc":"2.0"}`),
	}
	return frameFromEnvelope(env)
}

func truncatedFrame() []byte {
	full := validMCP()
	return full[:len(full)-3]
}

func varintOverflow() []byte {
	// Build a frame whose first uvarint (version) overflows: 11 continuation bytes.
	body := []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80}
	out := make([]byte, 4+len(body))
	binary.BigEndian.PutUint32(out[:4], uint32(len(body)))
	copy(out[4:], body)
	return out
}
