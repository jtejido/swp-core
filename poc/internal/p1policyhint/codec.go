package p1policyhint

import (
	"encoding/binary"
	"fmt"
)

const (
	wtVarint = 0
	wtBytes  = 2
)

type Constraint struct {
	Key      string
	Value    string
	Mode     string
	ScopeRef string
}

type PolicyHintSet struct {
	Constraints []Constraint
}

type PolicyHintAck struct {
	AckID string
}

type PolicyViolation struct {
	Key        string
	ScopeRef   string
	ReasonCode string
}

type PolicyErr struct {
	Code    string
	Message string
}

func EncodePayloadSet(v PolicyHintSet) ([]byte, error) {
	return encodeWrapper(1, encodeSet(v)), nil
}

func DecodePayloadSet(payload []byte) (PolicyHintSet, error) {
	inner, err := decodeWrapper(payload, 1)
	if err != nil {
		return PolicyHintSet{}, err
	}
	return decodeSet(inner)
}

func EncodePayloadAck(v PolicyHintAck) ([]byte, error) {
	return encodeWrapper(2, encodeAck(v)), nil
}

func DecodePayloadAck(payload []byte) (PolicyHintAck, error) {
	inner, err := decodeWrapper(payload, 2)
	if err != nil {
		return PolicyHintAck{}, err
	}
	return decodeAck(inner)
}

func EncodePayloadViolation(v PolicyViolation) ([]byte, error) {
	return encodeWrapper(3, encodeViolation(v)), nil
}

func DecodePayloadViolation(payload []byte) (PolicyViolation, error) {
	inner, err := decodeWrapper(payload, 3)
	if err != nil {
		return PolicyViolation{}, err
	}
	return decodeViolation(inner)
}

func EncodePayloadErr(v PolicyErr) ([]byte, error) {
	return encodeWrapper(4, encodeErr(v)), nil
}

func DecodePayloadErr(payload []byte) (PolicyErr, error) {
	inner, err := decodeWrapper(payload, 4)
	if err != nil {
		return PolicyErr{}, err
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

func encodeSet(v PolicyHintSet) []byte {
	var out []byte
	for _, c := range v.Constraints {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, encodeConstraint(c))
	}
	return out
}

func decodeSet(b []byte) (PolicyHintSet, error) {
	var out PolicyHintSet
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return PolicyHintSet{}, err
		}
		if field == 1 {
			if wt != wtBytes {
				return PolicyHintSet{}, fmt.Errorf("policy_hint_set.constraints wrong wire type")
			}
			c, err := decodeConstraint(val)
			if err != nil {
				return PolicyHintSet{}, err
			}
			out.Constraints = append(out.Constraints, c)
		}
		b = b[n:]
	}
	return out, nil
}

func encodeAck(v PolicyHintAck) []byte {
	var out []byte
	if v.AckID != "" {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, []byte(v.AckID))
	}
	return out
}

func decodeAck(b []byte) (PolicyHintAck, error) {
	var out PolicyHintAck
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return PolicyHintAck{}, err
		}
		if field == 1 {
			if wt != wtBytes {
				return PolicyHintAck{}, fmt.Errorf("policy_hint_ack.ack_id wrong wire type")
			}
			out.AckID = string(val)
		}
		b = b[n:]
	}
	return out, nil
}

func encodeViolation(v PolicyViolation) []byte {
	var out []byte
	if v.Key != "" {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, []byte(v.Key))
	}
	if v.ScopeRef != "" {
		out = appendKey(out, 2, wtBytes)
		out = appendBytes(out, []byte(v.ScopeRef))
	}
	if v.ReasonCode != "" {
		out = appendKey(out, 3, wtBytes)
		out = appendBytes(out, []byte(v.ReasonCode))
	}
	return out
}

func decodeViolation(b []byte) (PolicyViolation, error) {
	var out PolicyViolation
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return PolicyViolation{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return PolicyViolation{}, fmt.Errorf("policy_violation.key wrong wire type")
			}
			out.Key = string(val)
		case 2:
			if wt != wtBytes {
				return PolicyViolation{}, fmt.Errorf("policy_violation.scope_ref wrong wire type")
			}
			out.ScopeRef = string(val)
		case 3:
			if wt != wtBytes {
				return PolicyViolation{}, fmt.Errorf("policy_violation.reason_code wrong wire type")
			}
			out.ReasonCode = string(val)
		}
		b = b[n:]
	}
	return out, nil
}

func encodeErr(v PolicyErr) []byte {
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

func decodeErr(b []byte) (PolicyErr, error) {
	var out PolicyErr
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return PolicyErr{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return PolicyErr{}, fmt.Errorf("policy_err.code wrong wire type")
			}
			out.Code = string(val)
		case 2:
			if wt != wtBytes {
				return PolicyErr{}, fmt.Errorf("policy_err.message wrong wire type")
			}
			out.Message = string(val)
		}
		b = b[n:]
	}
	return out, nil
}

func encodeConstraint(v Constraint) []byte {
	var out []byte
	if v.Key != "" {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, []byte(v.Key))
	}
	if v.Value != "" {
		out = appendKey(out, 2, wtBytes)
		out = appendBytes(out, []byte(v.Value))
	}
	if v.Mode != "" {
		out = appendKey(out, 3, wtBytes)
		out = appendBytes(out, []byte(v.Mode))
	}
	if v.ScopeRef != "" {
		out = appendKey(out, 4, wtBytes)
		out = appendBytes(out, []byte(v.ScopeRef))
	}
	return out
}

func decodeConstraint(b []byte) (Constraint, error) {
	var out Constraint
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return Constraint{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return Constraint{}, fmt.Errorf("constraint.key wrong wire type")
			}
			out.Key = string(val)
		case 2:
			if wt != wtBytes {
				return Constraint{}, fmt.Errorf("constraint.value wrong wire type")
			}
			out.Value = string(val)
		case 3:
			if wt != wtBytes {
				return Constraint{}, fmt.Errorf("constraint.mode wrong wire type")
			}
			out.Mode = string(val)
		case 4:
			if wt != wtBytes {
				return Constraint{}, fmt.Errorf("constraint.scope_ref wrong wire type")
			}
			out.ScopeRef = string(val)
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
