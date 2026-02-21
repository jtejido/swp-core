package p1events

import (
	"encoding/binary"
	"fmt"
)

const (
	wtVarint = 0
	wtBytes  = 2
)

type EventRecord struct {
	EventID   string
	EventType string
	Severity  string
	TsUnixMs  uint64
	MsgID     []byte
	TaskID    []byte
	RPCID     []byte
	Body      []byte
}

type EvtPublish struct {
	Event EventRecord
}

type EvtSubscribe struct {
	Filter string
}

type EvtUnsubscribe struct {
	SubscriptionID string
}

type EvtBatch struct {
	Events []EventRecord
}

type EvtErr struct {
	Code    string
	Message string
}

func EncodePayloadPublish(v EvtPublish) ([]byte, error) {
	return encodeWrapper(1, encodePublish(v)), nil
}

func DecodePayloadPublish(payload []byte) (EvtPublish, error) {
	inner, err := decodeWrapper(payload, 1)
	if err != nil {
		return EvtPublish{}, err
	}
	return decodePublish(inner)
}

func EncodePayloadSubscribe(v EvtSubscribe) ([]byte, error) {
	return encodeWrapper(2, encodeSubscribe(v)), nil
}

func DecodePayloadSubscribe(payload []byte) (EvtSubscribe, error) {
	inner, err := decodeWrapper(payload, 2)
	if err != nil {
		return EvtSubscribe{}, err
	}
	return decodeSubscribe(inner)
}

func EncodePayloadUnsubscribe(v EvtUnsubscribe) ([]byte, error) {
	return encodeWrapper(3, encodeUnsubscribe(v)), nil
}

func DecodePayloadUnsubscribe(payload []byte) (EvtUnsubscribe, error) {
	inner, err := decodeWrapper(payload, 3)
	if err != nil {
		return EvtUnsubscribe{}, err
	}
	return decodeUnsubscribe(inner)
}

func EncodePayloadBatch(v EvtBatch) ([]byte, error) {
	return encodeWrapper(4, encodeBatch(v)), nil
}

func DecodePayloadBatch(payload []byte) (EvtBatch, error) {
	inner, err := decodeWrapper(payload, 4)
	if err != nil {
		return EvtBatch{}, err
	}
	return decodeBatch(inner)
}

func EncodePayloadErr(v EvtErr) ([]byte, error) {
	return encodeWrapper(5, encodeErr(v)), nil
}

func DecodePayloadErr(payload []byte) (EvtErr, error) {
	inner, err := decodeWrapper(payload, 5)
	if err != nil {
		return EvtErr{}, err
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

func encodePublish(v EvtPublish) []byte {
	var out []byte
	out = appendKey(out, 1, wtBytes)
	out = appendBytes(out, encodeEventRecord(v.Event))
	return out
}

func decodePublish(b []byte) (EvtPublish, error) {
	var out EvtPublish
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return EvtPublish{}, err
		}
		if field == 1 {
			if wt != wtBytes {
				return EvtPublish{}, fmt.Errorf("evt_publish.event wrong wire type")
			}
			ev, err := decodeEventRecord(val)
			if err != nil {
				return EvtPublish{}, err
			}
			out.Event = ev
		}
		b = b[n:]
	}
	return out, nil
}

func encodeSubscribe(v EvtSubscribe) []byte {
	var out []byte
	if v.Filter != "" {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, []byte(v.Filter))
	}
	return out
}

func decodeSubscribe(b []byte) (EvtSubscribe, error) {
	var out EvtSubscribe
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return EvtSubscribe{}, err
		}
		if field == 1 {
			if wt != wtBytes {
				return EvtSubscribe{}, fmt.Errorf("evt_subscribe.filter wrong wire type")
			}
			out.Filter = string(val)
		}
		b = b[n:]
	}
	return out, nil
}

func encodeUnsubscribe(v EvtUnsubscribe) []byte {
	var out []byte
	if v.SubscriptionID != "" {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, []byte(v.SubscriptionID))
	}
	return out
}

func decodeUnsubscribe(b []byte) (EvtUnsubscribe, error) {
	var out EvtUnsubscribe
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return EvtUnsubscribe{}, err
		}
		if field == 1 {
			if wt != wtBytes {
				return EvtUnsubscribe{}, fmt.Errorf("evt_unsubscribe.subscription_id wrong wire type")
			}
			out.SubscriptionID = string(val)
		}
		b = b[n:]
	}
	return out, nil
}

func encodeBatch(v EvtBatch) []byte {
	var out []byte
	for _, e := range v.Events {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, encodeEventRecord(e))
	}
	return out
}

func decodeBatch(b []byte) (EvtBatch, error) {
	var out EvtBatch
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return EvtBatch{}, err
		}
		if field == 1 {
			if wt != wtBytes {
				return EvtBatch{}, fmt.Errorf("evt_batch.events wrong wire type")
			}
			ev, err := decodeEventRecord(val)
			if err != nil {
				return EvtBatch{}, err
			}
			out.Events = append(out.Events, ev)
		}
		b = b[n:]
	}
	return out, nil
}

func encodeErr(v EvtErr) []byte {
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

func decodeErr(b []byte) (EvtErr, error) {
	var out EvtErr
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return EvtErr{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return EvtErr{}, fmt.Errorf("evt_err.code wrong wire type")
			}
			out.Code = string(val)
		case 2:
			if wt != wtBytes {
				return EvtErr{}, fmt.Errorf("evt_err.message wrong wire type")
			}
			out.Message = string(val)
		}
		b = b[n:]
	}
	return out, nil
}

func encodeEventRecord(v EventRecord) []byte {
	var out []byte
	if v.EventID != "" {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, []byte(v.EventID))
	}
	if v.EventType != "" {
		out = appendKey(out, 2, wtBytes)
		out = appendBytes(out, []byte(v.EventType))
	}
	if v.Severity != "" {
		out = appendKey(out, 3, wtBytes)
		out = appendBytes(out, []byte(v.Severity))
	}
	if v.TsUnixMs != 0 {
		out = appendKey(out, 4, wtVarint)
		out = binary.AppendUvarint(out, v.TsUnixMs)
	}
	if len(v.MsgID) > 0 {
		out = appendKey(out, 5, wtBytes)
		out = appendBytes(out, v.MsgID)
	}
	if len(v.TaskID) > 0 {
		out = appendKey(out, 6, wtBytes)
		out = appendBytes(out, v.TaskID)
	}
	if len(v.RPCID) > 0 {
		out = appendKey(out, 7, wtBytes)
		out = appendBytes(out, v.RPCID)
	}
	if len(v.Body) > 0 {
		out = appendKey(out, 8, wtBytes)
		out = appendBytes(out, v.Body)
	}
	return out
}

func decodeEventRecord(b []byte) (EventRecord, error) {
	var out EventRecord
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return EventRecord{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return EventRecord{}, fmt.Errorf("event_record.event_id wrong wire type")
			}
			out.EventID = string(val)
		case 2:
			if wt != wtBytes {
				return EventRecord{}, fmt.Errorf("event_record.event_type wrong wire type")
			}
			out.EventType = string(val)
		case 3:
			if wt != wtBytes {
				return EventRecord{}, fmt.Errorf("event_record.severity wrong wire type")
			}
			out.Severity = string(val)
		case 4:
			if wt != wtVarint {
				return EventRecord{}, fmt.Errorf("event_record.ts_unix_ms wrong wire type")
			}
			vv, _, err := consumeVarintValue(val)
			if err != nil {
				return EventRecord{}, err
			}
			out.TsUnixMs = vv
		case 5:
			if wt != wtBytes {
				return EventRecord{}, fmt.Errorf("event_record.msg_id wrong wire type")
			}
			out.MsgID = append([]byte(nil), val...)
		case 6:
			if wt != wtBytes {
				return EventRecord{}, fmt.Errorf("event_record.task_id wrong wire type")
			}
			out.TaskID = append([]byte(nil), val...)
		case 7:
			if wt != wtBytes {
				return EventRecord{}, fmt.Errorf("event_record.rpc_id wrong wire type")
			}
			out.RPCID = append([]byte(nil), val...)
		case 8:
			if wt != wtBytes {
				return EventRecord{}, fmt.Errorf("event_record.body wrong wire type")
			}
			out.Body = append([]byte(nil), val...)
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
