package p1rpc

import "testing"

func TestPayloadReqRoundTrip(t *testing.T) {
	in := RpcReq{RPCID: []byte("r1"), Method: "demo.echo", Params: []byte(`{"x":1}`), IdempotencyKey: "k1"}
	b, err := EncodePayloadReq(in)
	if err != nil {
		t.Fatalf("EncodePayloadReq failed: %v", err)
	}
	out, err := DecodePayloadReq(b)
	if err != nil {
		t.Fatalf("DecodePayloadReq failed: %v", err)
	}
	if string(out.RPCID) != string(in.RPCID) || out.Method != in.Method || string(out.Params) != string(in.Params) || out.IdempotencyKey != in.IdempotencyKey {
		t.Fatalf("roundtrip mismatch: got %+v want %+v", out, in)
	}
}

func TestPayloadRespRoundTrip(t *testing.T) {
	in := RpcResp{RPCID: []byte("r1"), Result: []byte("ok")}
	b, err := EncodePayloadResp(in)
	if err != nil {
		t.Fatalf("EncodePayloadResp failed: %v", err)
	}
	out, err := DecodePayloadResp(b)
	if err != nil {
		t.Fatalf("DecodePayloadResp failed: %v", err)
	}
	if string(out.RPCID) != string(in.RPCID) || string(out.Result) != string(in.Result) {
		t.Fatalf("roundtrip mismatch")
	}
}

func TestPayloadErrRoundTrip(t *testing.T) {
	in := RpcErr{RPCID: []byte("r1"), ErrorCode: "internal", Retryable: true, ErrorMessage: "boom"}
	b, err := EncodePayloadErr(in)
	if err != nil {
		t.Fatalf("EncodePayloadErr failed: %v", err)
	}
	out, err := DecodePayloadErr(b)
	if err != nil {
		t.Fatalf("DecodePayloadErr failed: %v", err)
	}
	if string(out.RPCID) != string(in.RPCID) || out.ErrorCode != in.ErrorCode || out.Retryable != in.Retryable || out.ErrorMessage != in.ErrorMessage {
		t.Fatalf("roundtrip mismatch")
	}
}

func TestPayloadStreamItemRoundTrip(t *testing.T) {
	in := RpcStreamItem{RPCID: []byte("r1"), SeqNo: 2, Item: []byte("2"), IsTerminal: false}
	b, err := EncodePayloadStreamItem(in)
	if err != nil {
		t.Fatalf("EncodePayloadStreamItem failed: %v", err)
	}
	out, err := DecodePayloadStreamItem(b)
	if err != nil {
		t.Fatalf("DecodePayloadStreamItem failed: %v", err)
	}
	if string(out.RPCID) != string(in.RPCID) || out.SeqNo != in.SeqNo || string(out.Item) != string(in.Item) || out.IsTerminal != in.IsTerminal {
		t.Fatalf("roundtrip mismatch")
	}
}

func TestPayloadCancelRoundTrip(t *testing.T) {
	in := RpcCancel{RPCID: []byte("r1"), Reason: "user"}
	b, err := EncodePayloadCancel(in)
	if err != nil {
		t.Fatalf("EncodePayloadCancel failed: %v", err)
	}
	out, err := DecodePayloadCancel(b)
	if err != nil {
		t.Fatalf("DecodePayloadCancel failed: %v", err)
	}
	if string(out.RPCID) != string(in.RPCID) || out.Reason != in.Reason {
		t.Fatalf("roundtrip mismatch")
	}
}

func TestDecodePayloadReqInvalid(t *testing.T) {
	if _, err := DecodePayloadReq([]byte{0x08, 0x01}); err == nil {
		t.Fatalf("expected error")
	}
}
