package p1obs

import (
	"encoding/binary"
	"fmt"
)

const (
	wtVarint = 0
	wtBytes  = 2
)

type ObsSet struct {
	Traceparent string
	Tracestate  string
	MsgID       []byte
	TaskID      []byte
	RPCID       []byte
}

type ObsGet struct {
	IncludeCurrent bool
}

type ObsDoc struct {
	Traceparent string
	Tracestate  string
	MsgID       []byte
	TaskID      []byte
	RPCID       []byte
}

type ObsErr struct {
	Code    string
	Message string
}

func EncodePayloadSet(v ObsSet) ([]byte, error) {
	return encodeWrapper(1, encodeSet(v)), nil
}

func DecodePayloadSet(payload []byte) (ObsSet, error) {
	inner, err := decodeWrapper(payload, 1)
	if err != nil {
		return ObsSet{}, err
	}
	return decodeSet(inner)
}

func EncodePayloadGet(v ObsGet) ([]byte, error) {
	return encodeWrapper(2, encodeGet(v)), nil
}

func DecodePayloadGet(payload []byte) (ObsGet, error) {
	inner, err := decodeWrapper(payload, 2)
	if err != nil {
		return ObsGet{}, err
	}
	return decodeGet(inner)
}

func EncodePayloadDoc(v ObsDoc) ([]byte, error) {
	return encodeWrapper(3, encodeDoc(v)), nil
}

func DecodePayloadDoc(payload []byte) (ObsDoc, error) {
	inner, err := decodeWrapper(payload, 3)
	if err != nil {
		return ObsDoc{}, err
	}
	return decodeDoc(inner)
}

func EncodePayloadErr(v ObsErr) ([]byte, error) {
	return encodeWrapper(4, encodeErr(v)), nil
}

func DecodePayloadErr(payload []byte) (ObsErr, error) {
	inner, err := decodeWrapper(payload, 4)
	if err != nil {
		return ObsErr{}, err
	}
	return decodeErr(inner)
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

func encodeSet(v ObsSet) []byte {
	var out []byte
	if v.Traceparent != "" {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, []byte(v.Traceparent))
	}
	if v.Tracestate != "" {
		out = appendKey(out, 2, wtBytes)
		out = appendBytes(out, []byte(v.Tracestate))
	}
	if len(v.MsgID) > 0 {
		out = appendKey(out, 3, wtBytes)
		out = appendBytes(out, v.MsgID)
	}
	if len(v.TaskID) > 0 {
		out = appendKey(out, 4, wtBytes)
		out = appendBytes(out, v.TaskID)
	}
	if len(v.RPCID) > 0 {
		out = appendKey(out, 5, wtBytes)
		out = appendBytes(out, v.RPCID)
	}
	return out
}

func decodeSet(b []byte) (ObsSet, error) {
	var out ObsSet
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return ObsSet{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return ObsSet{}, fmt.Errorf("obs_set.traceparent wrong wire type")
			}
			out.Traceparent = string(val)
		case 2:
			if wt != wtBytes {
				return ObsSet{}, fmt.Errorf("obs_set.tracestate wrong wire type")
			}
			out.Tracestate = string(val)
		case 3:
			if wt != wtBytes {
				return ObsSet{}, fmt.Errorf("obs_set.msg_id wrong wire type")
			}
			out.MsgID = append([]byte(nil), val...)
		case 4:
			if wt != wtBytes {
				return ObsSet{}, fmt.Errorf("obs_set.task_id wrong wire type")
			}
			out.TaskID = append([]byte(nil), val...)
		case 5:
			if wt != wtBytes {
				return ObsSet{}, fmt.Errorf("obs_set.rpc_id wrong wire type")
			}
			out.RPCID = append([]byte(nil), val...)
		}
		b = b[n:]
	}
	return out, nil
}

func encodeGet(v ObsGet) []byte {
	var out []byte
	out = appendKey(out, 1, wtVarint)
	if v.IncludeCurrent {
		out = binary.AppendUvarint(out, 1)
	} else {
		out = binary.AppendUvarint(out, 0)
	}
	return out
}

func decodeGet(b []byte) (ObsGet, error) {
	var out ObsGet
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return ObsGet{}, err
		}
		if field == 1 {
			if wt != wtVarint {
				return ObsGet{}, fmt.Errorf("obs_get.include_current wrong wire type")
			}
			vv, _, err := consumeVarintValue(val)
			if err != nil {
				return ObsGet{}, err
			}
			out.IncludeCurrent = vv != 0
		}
		b = b[n:]
	}
	return out, nil
}

func encodeDoc(v ObsDoc) []byte {
	var out []byte
	if v.Traceparent != "" {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, []byte(v.Traceparent))
	}
	if v.Tracestate != "" {
		out = appendKey(out, 2, wtBytes)
		out = appendBytes(out, []byte(v.Tracestate))
	}
	if len(v.MsgID) > 0 {
		out = appendKey(out, 3, wtBytes)
		out = appendBytes(out, v.MsgID)
	}
	if len(v.TaskID) > 0 {
		out = appendKey(out, 4, wtBytes)
		out = appendBytes(out, v.TaskID)
	}
	if len(v.RPCID) > 0 {
		out = appendKey(out, 5, wtBytes)
		out = appendBytes(out, v.RPCID)
	}
	return out
}

func decodeDoc(b []byte) (ObsDoc, error) {
	var out ObsDoc
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return ObsDoc{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return ObsDoc{}, fmt.Errorf("obs_doc.traceparent wrong wire type")
			}
			out.Traceparent = string(val)
		case 2:
			if wt != wtBytes {
				return ObsDoc{}, fmt.Errorf("obs_doc.tracestate wrong wire type")
			}
			out.Tracestate = string(val)
		case 3:
			if wt != wtBytes {
				return ObsDoc{}, fmt.Errorf("obs_doc.msg_id wrong wire type")
			}
			out.MsgID = append([]byte(nil), val...)
		case 4:
			if wt != wtBytes {
				return ObsDoc{}, fmt.Errorf("obs_doc.task_id wrong wire type")
			}
			out.TaskID = append([]byte(nil), val...)
		case 5:
			if wt != wtBytes {
				return ObsDoc{}, fmt.Errorf("obs_doc.rpc_id wrong wire type")
			}
			out.RPCID = append([]byte(nil), val...)
		}
		b = b[n:]
	}
	return out, nil
}

func encodeErr(v ObsErr) []byte {
	var out []byte
	if v.Code != "" {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, []byte(v.Code))
	}
	if v.Message != "" {
		out = appendKey(out, 2, wtBytes)
		out = appendBytes(out, []byte(v.Message))
	}
	return out
}

func decodeErr(b []byte) (ObsErr, error) {
	var out ObsErr
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return ObsErr{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return ObsErr{}, fmt.Errorf("obs_err.code wrong wire type")
			}
			out.Code = string(val)
		case 2:
			if wt != wtBytes {
				return ObsErr{}, fmt.Errorf("obs_err.message wrong wire type")
			}
			out.Message = string(val)
		}
		b = b[n:]
	}
	return out, nil
}

func appendKey(dst []byte, field, wt uint64) []byte {
	return binary.AppendUvarint(dst, (field<<3)|wt)
}

func appendBytes(dst []byte, b []byte) []byte {
	dst = binary.AppendUvarint(dst, uint64(len(b)))
	return append(dst, b...)
}

func consumeField(b []byte) (field uint64, wt uint64, val []byte, consumed int, err error) {
	key, n := binary.Uvarint(b)
	if n <= 0 {
		return 0, 0, nil, 0, fmt.Errorf("invalid protobuf key")
	}
	field = key >> 3
	wt = key & 0x7
	idx := n

	switch wt {
	case wtVarint:
		_, vn := binary.Uvarint(b[idx:])
		if vn <= 0 {
			return 0, 0, nil, 0, fmt.Errorf("invalid protobuf varint value")
		}
		return field, wt, b[idx : idx+vn], idx + vn, nil
	case wtBytes:
		l, ln := binary.Uvarint(b[idx:])
		if ln <= 0 {
			return 0, 0, nil, 0, fmt.Errorf("invalid protobuf bytes length")
		}
		idx += ln
		if idx+int(l) > len(b) {
			return 0, 0, nil, 0, fmt.Errorf("truncated protobuf bytes value")
		}
		return field, wt, b[idx : idx+int(l)], idx + int(l), nil
	default:
		return 0, 0, nil, 0, fmt.Errorf("unsupported protobuf wire type %d", wt)
	}
}

func consumeVarintValue(b []byte) (uint64, int, error) {
	v, n := binary.Uvarint(b)
	if n <= 0 {
		return 0, 0, fmt.Errorf("invalid varint value")
	}
	return v, n, nil
}
