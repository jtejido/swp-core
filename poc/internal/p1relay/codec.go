package p1relay

import (
	"encoding/binary"
	"fmt"
)

const (
	wtVarint = 0
	wtBytes  = 2
)

type RelayPublish struct {
	DeliveryID []byte
	Topic      string
	Payload    []byte
	TTLMs      uint64
}

type RelayAck struct {
	DeliveryID []byte
}

type RelayNack struct {
	DeliveryID []byte
	Retryable  bool
	ReasonCode string
}

type RelayStatus struct {
	DeliveryID   []byte
	State        string
	AttemptCount uint32
}

type RelayErr struct {
	Code    string
	Message string
}

func EncodePayloadPublish(v RelayPublish) ([]byte, error) {
	return encodeWrapper(1, encodePublish(v)), nil
}

func DecodePayloadPublish(payload []byte) (RelayPublish, error) {
	inner, err := decodeWrapper(payload, 1)
	if err != nil {
		return RelayPublish{}, err
	}
	return decodePublish(inner)
}

func EncodePayloadAck(v RelayAck) ([]byte, error) {
	return encodeWrapper(2, encodeAck(v)), nil
}

func DecodePayloadAck(payload []byte) (RelayAck, error) {
	inner, err := decodeWrapper(payload, 2)
	if err != nil {
		return RelayAck{}, err
	}
	return decodeAck(inner)
}

func EncodePayloadNack(v RelayNack) ([]byte, error) {
	return encodeWrapper(3, encodeNack(v)), nil
}

func DecodePayloadNack(payload []byte) (RelayNack, error) {
	inner, err := decodeWrapper(payload, 3)
	if err != nil {
		return RelayNack{}, err
	}
	return decodeNack(inner)
}

func EncodePayloadStatus(v RelayStatus) ([]byte, error) {
	return encodeWrapper(4, encodeStatus(v)), nil
}

func DecodePayloadStatus(payload []byte) (RelayStatus, error) {
	inner, err := decodeWrapper(payload, 4)
	if err != nil {
		return RelayStatus{}, err
	}
	return decodeStatus(inner)
}

func EncodePayloadErr(v RelayErr) ([]byte, error) {
	return encodeWrapper(5, encodeErr(v)), nil
}

func DecodePayloadErr(payload []byte) (RelayErr, error) {
	inner, err := decodeWrapper(payload, 5)
	if err != nil {
		return RelayErr{}, err
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

func encodePublish(v RelayPublish) []byte {
	var out []byte
	if len(v.DeliveryID) > 0 {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, v.DeliveryID)
	}
	if v.Topic != "" {
		out = appendKey(out, 2, wtBytes)
		out = appendBytes(out, []byte(v.Topic))
	}
	if len(v.Payload) > 0 {
		out = appendKey(out, 3, wtBytes)
		out = appendBytes(out, v.Payload)
	}
	if v.TTLMs != 0 {
		out = appendKey(out, 4, wtVarint)
		out = binary.AppendUvarint(out, v.TTLMs)
	}
	return out
}

func decodePublish(b []byte) (RelayPublish, error) {
	var out RelayPublish
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return RelayPublish{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return RelayPublish{}, fmt.Errorf("relay_publish.delivery_id wrong wire type")
			}
			out.DeliveryID = append([]byte(nil), val...)
		case 2:
			if wt != wtBytes {
				return RelayPublish{}, fmt.Errorf("relay_publish.topic wrong wire type")
			}
			out.Topic = string(val)
		case 3:
			if wt != wtBytes {
				return RelayPublish{}, fmt.Errorf("relay_publish.payload wrong wire type")
			}
			out.Payload = append([]byte(nil), val...)
		case 4:
			if wt != wtVarint {
				return RelayPublish{}, fmt.Errorf("relay_publish.ttl_ms wrong wire type")
			}
			vv, _, err := consumeVarintValue(val)
			if err != nil {
				return RelayPublish{}, err
			}
			out.TTLMs = vv
		}
		b = b[n:]
	}
	return out, nil
}

func encodeAck(v RelayAck) []byte {
	var out []byte
	if len(v.DeliveryID) > 0 {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, v.DeliveryID)
	}
	return out
}

func decodeAck(b []byte) (RelayAck, error) {
	var out RelayAck
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return RelayAck{}, err
		}
		if field == 1 {
			if wt != wtBytes {
				return RelayAck{}, fmt.Errorf("relay_ack.delivery_id wrong wire type")
			}
			out.DeliveryID = append([]byte(nil), val...)
		}
		b = b[n:]
	}
	return out, nil
}

func encodeNack(v RelayNack) []byte {
	var out []byte
	if len(v.DeliveryID) > 0 {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, v.DeliveryID)
	}
	if v.Retryable {
		out = appendKey(out, 2, wtVarint)
		out = binary.AppendUvarint(out, 1)
	}
	if v.ReasonCode != "" {
		out = appendKey(out, 3, wtBytes)
		out = appendBytes(out, []byte(v.ReasonCode))
	}
	return out
}

func decodeNack(b []byte) (RelayNack, error) {
	var out RelayNack
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return RelayNack{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return RelayNack{}, fmt.Errorf("relay_nack.delivery_id wrong wire type")
			}
			out.DeliveryID = append([]byte(nil), val...)
		case 2:
			if wt != wtVarint {
				return RelayNack{}, fmt.Errorf("relay_nack.retryable wrong wire type")
			}
			vv, _, err := consumeVarintValue(val)
			if err != nil {
				return RelayNack{}, err
			}
			out.Retryable = vv != 0
		case 3:
			if wt != wtBytes {
				return RelayNack{}, fmt.Errorf("relay_nack.reason_code wrong wire type")
			}
			out.ReasonCode = string(val)
		}
		b = b[n:]
	}
	return out, nil
}

func encodeStatus(v RelayStatus) []byte {
	var out []byte
	if len(v.DeliveryID) > 0 {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, v.DeliveryID)
	}
	if v.State != "" {
		out = appendKey(out, 2, wtBytes)
		out = appendBytes(out, []byte(v.State))
	}
	if v.AttemptCount != 0 {
		out = appendKey(out, 3, wtVarint)
		out = binary.AppendUvarint(out, uint64(v.AttemptCount))
	}
	return out
}

func decodeStatus(b []byte) (RelayStatus, error) {
	var out RelayStatus
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return RelayStatus{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return RelayStatus{}, fmt.Errorf("relay_status.delivery_id wrong wire type")
			}
			out.DeliveryID = append([]byte(nil), val...)
		case 2:
			if wt != wtBytes {
				return RelayStatus{}, fmt.Errorf("relay_status.state wrong wire type")
			}
			out.State = string(val)
		case 3:
			if wt != wtVarint {
				return RelayStatus{}, fmt.Errorf("relay_status.attempt_count wrong wire type")
			}
			vv, _, err := consumeVarintValue(val)
			if err != nil {
				return RelayStatus{}, err
			}
			out.AttemptCount = uint32(vv)
		}
		b = b[n:]
	}
	return out, nil
}

func encodeErr(v RelayErr) []byte {
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

func decodeErr(b []byte) (RelayErr, error) {
	var out RelayErr
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return RelayErr{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return RelayErr{}, fmt.Errorf("relay_err.code wrong wire type")
			}
			out.Code = string(val)
		case 2:
			if wt != wtBytes {
				return RelayErr{}, fmt.Errorf("relay_err.message wrong wire type")
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
