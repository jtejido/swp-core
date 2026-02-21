package p1a2a

import (
	"encoding/binary"
	"fmt"
)

const (
	wtVarint = 0
	wtBytes  = 2
)

type Handshake struct {
	AgentID      string
	Capabilities []string
}

type Task struct {
	TaskID []byte
	Kind   string
	Input  []byte
}

type Event struct {
	TaskID       []byte
	Message      string
	EventPayload []byte
}

type Result struct {
	TaskID       []byte
	OK           bool
	Output       []byte
	ErrorMessage string
}

func EncodePayloadHandshake(v Handshake) ([]byte, error) {
	return encodeWrapper(1, encodeHandshake(v)), nil
}

func DecodePayloadHandshake(payload []byte) (Handshake, error) {
	inner, err := decodeWrapper(payload, 1)
	if err != nil {
		return Handshake{}, err
	}
	return decodeHandshake(inner)
}

func EncodePayloadTask(v Task) ([]byte, error) {
	return encodeWrapper(2, encodeTask(v)), nil
}

func DecodePayloadTask(payload []byte) (Task, error) {
	inner, err := decodeWrapper(payload, 2)
	if err != nil {
		return Task{}, err
	}
	return decodeTask(inner)
}

func EncodePayloadEvent(v Event) ([]byte, error) {
	return encodeWrapper(3, encodeEvent(v)), nil
}

func DecodePayloadEvent(payload []byte) (Event, error) {
	inner, err := decodeWrapper(payload, 3)
	if err != nil {
		return Event{}, err
	}
	return decodeEvent(inner)
}

func EncodePayloadResult(v Result) ([]byte, error) {
	return encodeWrapper(4, encodeResult(v)), nil
}

func DecodePayloadResult(payload []byte) (Result, error) {
	inner, err := decodeWrapper(payload, 4)
	if err != nil {
		return Result{}, err
	}
	return decodeResult(inner)
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

func encodeHandshake(v Handshake) []byte {
	var out []byte
	if v.AgentID != "" {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, []byte(v.AgentID))
	}
	for _, cap := range v.Capabilities {
		out = appendKey(out, 2, wtBytes)
		out = appendBytes(out, []byte(cap))
	}
	return out
}

func decodeHandshake(b []byte) (Handshake, error) {
	var out Handshake
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return Handshake{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return Handshake{}, fmt.Errorf("a2a_handshake.agent_id wrong wire type")
			}
			out.AgentID = string(val)
		case 2:
			if wt != wtBytes {
				return Handshake{}, fmt.Errorf("a2a_handshake.capabilities wrong wire type")
			}
			out.Capabilities = append(out.Capabilities, string(val))
		}
		b = b[n:]
	}
	return out, nil
}

func encodeTask(v Task) []byte {
	var out []byte
	if len(v.TaskID) > 0 {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, v.TaskID)
	}
	if v.Kind != "" {
		out = appendKey(out, 2, wtBytes)
		out = appendBytes(out, []byte(v.Kind))
	}
	if len(v.Input) > 0 {
		out = appendKey(out, 3, wtBytes)
		out = appendBytes(out, v.Input)
	}
	return out
}

func decodeTask(b []byte) (Task, error) {
	var out Task
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return Task{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return Task{}, fmt.Errorf("a2a_task.task_id wrong wire type")
			}
			out.TaskID = append([]byte(nil), val...)
		case 2:
			if wt != wtBytes {
				return Task{}, fmt.Errorf("a2a_task.kind wrong wire type")
			}
			out.Kind = string(val)
		case 3:
			if wt != wtBytes {
				return Task{}, fmt.Errorf("a2a_task.input wrong wire type")
			}
			out.Input = append([]byte(nil), val...)
		}
		b = b[n:]
	}
	return out, nil
}

func encodeEvent(v Event) []byte {
	var out []byte
	if len(v.TaskID) > 0 {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, v.TaskID)
	}
	if v.Message != "" {
		out = appendKey(out, 2, wtBytes)
		out = appendBytes(out, []byte(v.Message))
	}
	if len(v.EventPayload) > 0 {
		out = appendKey(out, 3, wtBytes)
		out = appendBytes(out, v.EventPayload)
	}
	return out
}

func decodeEvent(b []byte) (Event, error) {
	var out Event
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return Event{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return Event{}, fmt.Errorf("a2a_event.task_id wrong wire type")
			}
			out.TaskID = append([]byte(nil), val...)
		case 2:
			if wt != wtBytes {
				return Event{}, fmt.Errorf("a2a_event.message wrong wire type")
			}
			out.Message = string(val)
		case 3:
			if wt != wtBytes {
				return Event{}, fmt.Errorf("a2a_event.event_payload wrong wire type")
			}
			out.EventPayload = append([]byte(nil), val...)
		}
		b = b[n:]
	}
	return out, nil
}

func encodeResult(v Result) []byte {
	var out []byte
	if len(v.TaskID) > 0 {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, v.TaskID)
	}
	if v.OK {
		out = appendKey(out, 2, wtVarint)
		out = binary.AppendUvarint(out, 1)
	}
	if len(v.Output) > 0 {
		out = appendKey(out, 3, wtBytes)
		out = appendBytes(out, v.Output)
	}
	if v.ErrorMessage != "" {
		out = appendKey(out, 4, wtBytes)
		out = appendBytes(out, []byte(v.ErrorMessage))
	}
	return out
}

func decodeResult(b []byte) (Result, error) {
	var out Result
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return Result{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return Result{}, fmt.Errorf("a2a_result.task_id wrong wire type")
			}
			out.TaskID = append([]byte(nil), val...)
		case 2:
			if wt != wtVarint {
				return Result{}, fmt.Errorf("a2a_result.ok wrong wire type")
			}
			vv, _, err := consumeVarintValue(val)
			if err != nil {
				return Result{}, err
			}
			out.OK = vv != 0
		case 3:
			if wt != wtBytes {
				return Result{}, fmt.Errorf("a2a_result.output wrong wire type")
			}
			out.Output = append([]byte(nil), val...)
		case 4:
			if wt != wtBytes {
				return Result{}, fmt.Errorf("a2a_result.error_message wrong wire type")
			}
			out.ErrorMessage = string(val)
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
