package p1state

import (
	"encoding/binary"
	"fmt"
)

const (
	wtVarint = 0
	wtBytes  = 2
)

type StatePut struct {
	StateID   []byte
	Blob      []byte
	ParentIDs [][]byte
	Metadata  []byte
}

type StateGet struct {
	StateID []byte
}

type StateDelta struct {
	StateID   []byte
	Delta     []byte
	ParentIDs [][]byte
}

type StateErr struct {
	Code    string
	Message string
}

func EncodePayloadPut(v StatePut) ([]byte, error) {
	return encodeWrapper(1, encodePut(v)), nil
}

func DecodePayloadPut(payload []byte) (StatePut, error) {
	inner, err := decodeWrapper(payload, 1)
	if err != nil {
		return StatePut{}, err
	}
	return decodePut(inner)
}

func EncodePayloadGet(v StateGet) ([]byte, error) {
	return encodeWrapper(2, encodeGet(v)), nil
}

func DecodePayloadGet(payload []byte) (StateGet, error) {
	inner, err := decodeWrapper(payload, 2)
	if err != nil {
		return StateGet{}, err
	}
	return decodeGet(inner)
}

func EncodePayloadDelta(v StateDelta) ([]byte, error) {
	return encodeWrapper(3, encodeDelta(v)), nil
}

func DecodePayloadDelta(payload []byte) (StateDelta, error) {
	inner, err := decodeWrapper(payload, 3)
	if err != nil {
		return StateDelta{}, err
	}
	return decodeDelta(inner)
}

func EncodePayloadErr(v StateErr) ([]byte, error) {
	return encodeWrapper(4, encodeErr(v)), nil
}

func DecodePayloadErr(payload []byte) (StateErr, error) {
	inner, err := decodeWrapper(payload, 4)
	if err != nil {
		return StateErr{}, err
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

func encodePut(v StatePut) []byte {
	var out []byte
	if len(v.StateID) > 0 {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, v.StateID)
	}
	if len(v.Blob) > 0 {
		out = appendKey(out, 2, wtBytes)
		out = appendBytes(out, v.Blob)
	}
	for _, pid := range v.ParentIDs {
		out = appendKey(out, 3, wtBytes)
		out = appendBytes(out, pid)
	}
	if len(v.Metadata) > 0 {
		out = appendKey(out, 4, wtBytes)
		out = appendBytes(out, v.Metadata)
	}
	return out
}

func decodePut(b []byte) (StatePut, error) {
	var out StatePut
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return StatePut{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return StatePut{}, fmt.Errorf("state_put.state_id wrong wire type")
			}
			out.StateID = append([]byte(nil), val...)
		case 2:
			if wt != wtBytes {
				return StatePut{}, fmt.Errorf("state_put.blob wrong wire type")
			}
			out.Blob = append([]byte(nil), val...)
		case 3:
			if wt != wtBytes {
				return StatePut{}, fmt.Errorf("state_put.parent_ids wrong wire type")
			}
			out.ParentIDs = append(out.ParentIDs, append([]byte(nil), val...))
		case 4:
			if wt != wtBytes {
				return StatePut{}, fmt.Errorf("state_put.metadata wrong wire type")
			}
			out.Metadata = append([]byte(nil), val...)
		}
		b = b[n:]
	}
	return out, nil
}

func encodeGet(v StateGet) []byte {
	var out []byte
	if len(v.StateID) > 0 {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, v.StateID)
	}
	return out
}

func decodeGet(b []byte) (StateGet, error) {
	var out StateGet
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return StateGet{}, err
		}
		if field == 1 {
			if wt != wtBytes {
				return StateGet{}, fmt.Errorf("state_get.state_id wrong wire type")
			}
			out.StateID = append([]byte(nil), val...)
		}
		b = b[n:]
	}
	return out, nil
}

func encodeDelta(v StateDelta) []byte {
	var out []byte
	if len(v.StateID) > 0 {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, v.StateID)
	}
	if len(v.Delta) > 0 {
		out = appendKey(out, 2, wtBytes)
		out = appendBytes(out, v.Delta)
	}
	for _, pid := range v.ParentIDs {
		out = appendKey(out, 3, wtBytes)
		out = appendBytes(out, pid)
	}
	return out
}

func decodeDelta(b []byte) (StateDelta, error) {
	var out StateDelta
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return StateDelta{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return StateDelta{}, fmt.Errorf("state_delta.state_id wrong wire type")
			}
			out.StateID = append([]byte(nil), val...)
		case 2:
			if wt != wtBytes {
				return StateDelta{}, fmt.Errorf("state_delta.delta wrong wire type")
			}
			out.Delta = append([]byte(nil), val...)
		case 3:
			if wt != wtBytes {
				return StateDelta{}, fmt.Errorf("state_delta.parent_ids wrong wire type")
			}
			out.ParentIDs = append(out.ParentIDs, append([]byte(nil), val...))
		}
		b = b[n:]
	}
	return out, nil
}

func encodeErr(v StateErr) []byte {
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

func decodeErr(b []byte) (StateErr, error) {
	var out StateErr
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return StateErr{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return StateErr{}, fmt.Errorf("state_err.code wrong wire type")
			}
			out.Code = string(val)
		case 2:
			if wt != wtBytes {
				return StateErr{}, fmt.Errorf("state_err.message wrong wire type")
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
