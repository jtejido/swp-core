package p1rpc

import (
	"encoding/binary"
	"fmt"
)

// Proto schema alignment (proto/swp_rpc.proto):
// message Payload { oneof body { RpcReq req=1; RpcResp resp=2; RpcErr err=3; RpcStreamItem stream_item=4; RpcCancel cancel=5; } }
// message RpcReq { bytes rpc_id=1; string method=2; bytes params=3; string idempotency_key=4; }
// message RpcResp { bytes rpc_id=1; bytes result=2; }
// message RpcErr { bytes rpc_id=1; string error_code=2; bool retryable=3; string error_message=4; }
// message RpcStreamItem { bytes rpc_id=1; uint64 seq_no=2; bytes item=3; bool is_terminal=4; }
// message RpcCancel { bytes rpc_id=1; string reason=2; }

const (
	wtVarint = 0
	wt64Bit  = 1
	wtBytes  = 2
	wt32Bit  = 5
)

type RpcReq struct {
	RPCID          []byte
	Method         string
	Params         []byte
	IdempotencyKey string
}

type RpcResp struct {
	RPCID  []byte
	Result []byte
}

type RpcErr struct {
	RPCID        []byte
	ErrorCode    string
	Retryable    bool
	ErrorMessage string
}

type RpcStreamItem struct {
	RPCID      []byte
	SeqNo      uint64
	Item       []byte
	IsTerminal bool
}

type RpcCancel struct {
	RPCID  []byte
	Reason string
}

func EncodePayloadReq(v RpcReq) ([]byte, error) {
	inner := encodeRpcReq(v)
	return encodeWrapper(1, inner), nil
}

func DecodePayloadReq(payload []byte) (RpcReq, error) {
	inner, err := decodeWrapper(payload, 1)
	if err != nil {
		return RpcReq{}, err
	}
	return decodeRpcReq(inner)
}

func EncodePayloadResp(v RpcResp) ([]byte, error) {
	inner := encodeRpcResp(v)
	return encodeWrapper(2, inner), nil
}

func DecodePayloadResp(payload []byte) (RpcResp, error) {
	inner, err := decodeWrapper(payload, 2)
	if err != nil {
		return RpcResp{}, err
	}
	return decodeRpcResp(inner)
}

func EncodePayloadErr(v RpcErr) ([]byte, error) {
	inner := encodeRpcErr(v)
	return encodeWrapper(3, inner), nil
}

func DecodePayloadErr(payload []byte) (RpcErr, error) {
	inner, err := decodeWrapper(payload, 3)
	if err != nil {
		return RpcErr{}, err
	}
	return decodeRpcErr(inner)
}

func EncodePayloadStreamItem(v RpcStreamItem) ([]byte, error) {
	inner := encodeRpcStreamItem(v)
	return encodeWrapper(4, inner), nil
}

func DecodePayloadStreamItem(payload []byte) (RpcStreamItem, error) {
	inner, err := decodeWrapper(payload, 4)
	if err != nil {
		return RpcStreamItem{}, err
	}
	return decodeRpcStreamItem(inner)
}

func EncodePayloadCancel(v RpcCancel) ([]byte, error) {
	inner := encodeRpcCancel(v)
	return encodeWrapper(5, inner), nil
}

func DecodePayloadCancel(payload []byte) (RpcCancel, error) {
	inner, err := decodeWrapper(payload, 5)
	if err != nil {
		return RpcCancel{}, err
	}
	return decodeRpcCancel(inner)
}

func encodeWrapper(oneofField uint64, inner []byte) []byte {
	var out []byte
	out = appendKey(out, oneofField, wtBytes)
	out = appendBytes(out, inner)
	return out
}

func decodeWrapper(payload []byte, expectedField uint64) ([]byte, error) {
	for len(payload) > 0 {
		field, wt, val, n, err := consumeField(payload)
		if err != nil {
			return nil, err
		}
		if wt == wtBytes && field == expectedField {
			return val, nil
		}
		payload = payload[n:]
	}
	return nil, fmt.Errorf("missing wrapper field %d", expectedField)
}

func encodeRpcReq(v RpcReq) []byte {
	var out []byte
	if len(v.RPCID) > 0 {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, v.RPCID)
	}
	if v.Method != "" {
		out = appendKey(out, 2, wtBytes)
		out = appendBytes(out, []byte(v.Method))
	}
	if len(v.Params) > 0 {
		out = appendKey(out, 3, wtBytes)
		out = appendBytes(out, v.Params)
	}
	if v.IdempotencyKey != "" {
		out = appendKey(out, 4, wtBytes)
		out = appendBytes(out, []byte(v.IdempotencyKey))
	}
	return out
}

func decodeRpcReq(b []byte) (RpcReq, error) {
	var out RpcReq
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return RpcReq{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return RpcReq{}, fmt.Errorf("rpc_req.rpc_id wrong wire type")
			}
			out.RPCID = append([]byte(nil), val...)
		case 2:
			if wt != wtBytes {
				return RpcReq{}, fmt.Errorf("rpc_req.method wrong wire type")
			}
			out.Method = string(val)
		case 3:
			if wt != wtBytes {
				return RpcReq{}, fmt.Errorf("rpc_req.params wrong wire type")
			}
			out.Params = append([]byte(nil), val...)
		case 4:
			if wt != wtBytes {
				return RpcReq{}, fmt.Errorf("rpc_req.idempotency_key wrong wire type")
			}
			out.IdempotencyKey = string(val)
		}
		b = b[n:]
	}
	if out.Method == "" {
		return RpcReq{}, fmt.Errorf("rpc_req.method required")
	}
	return out, nil
}

func encodeRpcResp(v RpcResp) []byte {
	var out []byte
	if len(v.RPCID) > 0 {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, v.RPCID)
	}
	out = appendKey(out, 2, wtBytes)
	out = appendBytes(out, v.Result)
	return out
}

func decodeRpcResp(b []byte) (RpcResp, error) {
	var out RpcResp
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return RpcResp{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return RpcResp{}, fmt.Errorf("rpc_resp.rpc_id wrong wire type")
			}
			out.RPCID = append([]byte(nil), val...)
		case 2:
			if wt != wtBytes {
				return RpcResp{}, fmt.Errorf("rpc_resp.result wrong wire type")
			}
			out.Result = append([]byte(nil), val...)
		}
		b = b[n:]
	}
	return out, nil
}

func encodeRpcErr(v RpcErr) []byte {
	var out []byte
	if len(v.RPCID) > 0 {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, v.RPCID)
	}
	if v.ErrorCode != "" {
		out = appendKey(out, 2, wtBytes)
		out = appendBytes(out, []byte(v.ErrorCode))
	}
	out = appendKey(out, 3, wtVarint)
	out = binary.AppendUvarint(out, boolToU64(v.Retryable))
	if v.ErrorMessage != "" {
		out = appendKey(out, 4, wtBytes)
		out = appendBytes(out, []byte(v.ErrorMessage))
	}
	return out
}

func decodeRpcErr(b []byte) (RpcErr, error) {
	var out RpcErr
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return RpcErr{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return RpcErr{}, fmt.Errorf("rpc_err.rpc_id wrong wire type")
			}
			out.RPCID = append([]byte(nil), val...)
		case 2:
			if wt != wtBytes {
				return RpcErr{}, fmt.Errorf("rpc_err.error_code wrong wire type")
			}
			out.ErrorCode = string(val)
		case 3:
			if wt != wtVarint {
				return RpcErr{}, fmt.Errorf("rpc_err.retryable wrong wire type")
			}
			vv, _, err := consumeVarintValue(val)
			if err != nil {
				return RpcErr{}, err
			}
			out.Retryable = vv != 0
		case 4:
			if wt != wtBytes {
				return RpcErr{}, fmt.Errorf("rpc_err.error_message wrong wire type")
			}
			out.ErrorMessage = string(val)
		}
		b = b[n:]
	}
	return out, nil
}

func encodeRpcStreamItem(v RpcStreamItem) []byte {
	var out []byte
	if len(v.RPCID) > 0 {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, v.RPCID)
	}
	out = appendKey(out, 2, wtVarint)
	out = binary.AppendUvarint(out, v.SeqNo)
	out = appendKey(out, 3, wtBytes)
	out = appendBytes(out, v.Item)
	out = appendKey(out, 4, wtVarint)
	out = binary.AppendUvarint(out, boolToU64(v.IsTerminal))
	return out
}

func decodeRpcStreamItem(b []byte) (RpcStreamItem, error) {
	var out RpcStreamItem
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return RpcStreamItem{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return RpcStreamItem{}, fmt.Errorf("rpc_stream_item.rpc_id wrong wire type")
			}
			out.RPCID = append([]byte(nil), val...)
		case 2:
			if wt != wtVarint {
				return RpcStreamItem{}, fmt.Errorf("rpc_stream_item.seq_no wrong wire type")
			}
			vv, _, err := consumeVarintValue(val)
			if err != nil {
				return RpcStreamItem{}, err
			}
			out.SeqNo = vv
		case 3:
			if wt != wtBytes {
				return RpcStreamItem{}, fmt.Errorf("rpc_stream_item.item wrong wire type")
			}
			out.Item = append([]byte(nil), val...)
		case 4:
			if wt != wtVarint {
				return RpcStreamItem{}, fmt.Errorf("rpc_stream_item.is_terminal wrong wire type")
			}
			vv, _, err := consumeVarintValue(val)
			if err != nil {
				return RpcStreamItem{}, err
			}
			out.IsTerminal = vv != 0
		}
		b = b[n:]
	}
	return out, nil
}

func encodeRpcCancel(v RpcCancel) []byte {
	var out []byte
	if len(v.RPCID) > 0 {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, v.RPCID)
	}
	if v.Reason != "" {
		out = appendKey(out, 2, wtBytes)
		out = appendBytes(out, []byte(v.Reason))
	}
	return out
}

func decodeRpcCancel(b []byte) (RpcCancel, error) {
	var out RpcCancel
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return RpcCancel{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return RpcCancel{}, fmt.Errorf("rpc_cancel.rpc_id wrong wire type")
			}
			out.RPCID = append([]byte(nil), val...)
		case 2:
			if wt != wtBytes {
				return RpcCancel{}, fmt.Errorf("rpc_cancel.reason wrong wire type")
			}
			out.Reason = string(val)
		}
		b = b[n:]
	}
	return out, nil
}

func appendKey(out []byte, fieldNum uint64, wireType uint64) []byte {
	return binary.AppendUvarint(out, (fieldNum<<3)|wireType)
}

func appendBytes(out []byte, b []byte) []byte {
	out = binary.AppendUvarint(out, uint64(len(b)))
	return append(out, b...)
}

func consumeField(b []byte) (fieldNum uint64, wireType uint64, val []byte, consumed int, err error) {
	key, n := binary.Uvarint(b)
	if n <= 0 {
		return 0, 0, nil, 0, fmt.Errorf("invalid field key")
	}
	fieldNum = key >> 3
	wireType = key & 0x7
	pos := n

	switch wireType {
	case wtVarint:
		_, vn := binary.Uvarint(b[pos:])
		if vn <= 0 {
			return 0, 0, nil, 0, fmt.Errorf("invalid varint field %d", fieldNum)
		}
		val = b[pos : pos+vn]
		pos += vn
	case wtBytes:
		l, ln := binary.Uvarint(b[pos:])
		if ln <= 0 {
			return 0, 0, nil, 0, fmt.Errorf("invalid bytes length field %d", fieldNum)
		}
		pos += ln
		if uint64(len(b[pos:])) < l {
			return 0, 0, nil, 0, fmt.Errorf("truncated bytes field %d", fieldNum)
		}
		val = b[pos : pos+int(l)]
		pos += int(l)
	case wt64Bit:
		if len(b[pos:]) < 8 {
			return 0, 0, nil, 0, fmt.Errorf("truncated 64-bit field %d", fieldNum)
		}
		val = b[pos : pos+8]
		pos += 8
	case wt32Bit:
		if len(b[pos:]) < 4 {
			return 0, 0, nil, 0, fmt.Errorf("truncated 32-bit field %d", fieldNum)
		}
		val = b[pos : pos+4]
		pos += 4
	default:
		return 0, 0, nil, 0, fmt.Errorf("unsupported wire type %d", wireType)
	}

	return fieldNum, wireType, val, pos, nil
}

func consumeVarintValue(raw []byte) (uint64, int, error) {
	v, n := binary.Uvarint(raw)
	if n <= 0 {
		return 0, 0, fmt.Errorf("invalid varint")
	}
	return v, n, nil
}

func boolToU64(v bool) uint64 {
	if v {
		return 1
	}
	return 0
}
