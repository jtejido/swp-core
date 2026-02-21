package p1agdisc

import (
	"encoding/binary"
	"fmt"
)

const (
	wtVarint = 0
	wtBytes  = 2
)

type AgdiscGet struct {
	AgentID     string
	IfNoneMatch string
}

type AgdiscDoc struct {
	AgentID        string
	SchemaRevision string
	CardPayload    []byte
	ETag           string
	MaxAgeMs       uint64
}

type AgdiscNotModified struct {
	AgentID string
	ETag    string
}

type AgdiscErr struct {
	Code    string
	Message string
}

func EncodePayloadGet(v AgdiscGet) ([]byte, error) {
	return encodeWrapper(1, encodeGet(v)), nil
}

func DecodePayloadGet(payload []byte) (AgdiscGet, error) {
	inner, err := decodeWrapper(payload, 1)
	if err != nil {
		return AgdiscGet{}, err
	}
	return decodeGet(inner)
}

func EncodePayloadDoc(v AgdiscDoc) ([]byte, error) {
	return encodeWrapper(2, encodeDoc(v)), nil
}

func DecodePayloadDoc(payload []byte) (AgdiscDoc, error) {
	inner, err := decodeWrapper(payload, 2)
	if err != nil {
		return AgdiscDoc{}, err
	}
	return decodeDoc(inner)
}

func EncodePayloadNotModified(v AgdiscNotModified) ([]byte, error) {
	return encodeWrapper(3, encodeNotModified(v)), nil
}

func DecodePayloadNotModified(payload []byte) (AgdiscNotModified, error) {
	inner, err := decodeWrapper(payload, 3)
	if err != nil {
		return AgdiscNotModified{}, err
	}
	return decodeNotModified(inner)
}

func EncodePayloadErr(v AgdiscErr) ([]byte, error) {
	return encodeWrapper(4, encodeErr(v)), nil
}

func DecodePayloadErr(payload []byte) (AgdiscErr, error) {
	inner, err := decodeWrapper(payload, 4)
	if err != nil {
		return AgdiscErr{}, err
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

func encodeGet(v AgdiscGet) []byte {
	var out []byte
	if v.AgentID != "" {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, []byte(v.AgentID))
	}
	if v.IfNoneMatch != "" {
		out = appendKey(out, 2, wtBytes)
		out = appendBytes(out, []byte(v.IfNoneMatch))
	}
	return out
}

func decodeGet(b []byte) (AgdiscGet, error) {
	var out AgdiscGet
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return AgdiscGet{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return AgdiscGet{}, fmt.Errorf("agdisc_get.agent_id wrong wire type")
			}
			out.AgentID = string(val)
		case 2:
			if wt != wtBytes {
				return AgdiscGet{}, fmt.Errorf("agdisc_get.if_none_match wrong wire type")
			}
			out.IfNoneMatch = string(val)
		}
		b = b[n:]
	}
	if out.AgentID == "" {
		return AgdiscGet{}, fmt.Errorf("agdisc_get.agent_id required")
	}
	return out, nil
}

func encodeDoc(v AgdiscDoc) []byte {
	var out []byte
	if v.AgentID != "" {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, []byte(v.AgentID))
	}
	if v.SchemaRevision != "" {
		out = appendKey(out, 2, wtBytes)
		out = appendBytes(out, []byte(v.SchemaRevision))
	}
	if len(v.CardPayload) > 0 {
		out = appendKey(out, 3, wtBytes)
		out = appendBytes(out, v.CardPayload)
	}
	if v.ETag != "" {
		out = appendKey(out, 4, wtBytes)
		out = appendBytes(out, []byte(v.ETag))
	}
	if v.MaxAgeMs != 0 {
		out = appendKey(out, 5, wtVarint)
		out = binary.AppendUvarint(out, v.MaxAgeMs)
	}
	return out
}

func decodeDoc(b []byte) (AgdiscDoc, error) {
	var out AgdiscDoc
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return AgdiscDoc{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return AgdiscDoc{}, fmt.Errorf("agdisc_doc.agent_id wrong wire type")
			}
			out.AgentID = string(val)
		case 2:
			if wt != wtBytes {
				return AgdiscDoc{}, fmt.Errorf("agdisc_doc.schema_revision wrong wire type")
			}
			out.SchemaRevision = string(val)
		case 3:
			if wt != wtBytes {
				return AgdiscDoc{}, fmt.Errorf("agdisc_doc.card_payload wrong wire type")
			}
			out.CardPayload = append([]byte(nil), val...)
		case 4:
			if wt != wtBytes {
				return AgdiscDoc{}, fmt.Errorf("agdisc_doc.etag wrong wire type")
			}
			out.ETag = string(val)
		case 5:
			if wt != wtVarint {
				return AgdiscDoc{}, fmt.Errorf("agdisc_doc.max_age_ms wrong wire type")
			}
			vv, _, err := consumeVarintValue(val)
			if err != nil {
				return AgdiscDoc{}, err
			}
			out.MaxAgeMs = vv
		}
		b = b[n:]
	}
	return out, nil
}

func encodeNotModified(v AgdiscNotModified) []byte {
	var out []byte
	if v.AgentID != "" {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, []byte(v.AgentID))
	}
	if v.ETag != "" {
		out = appendKey(out, 2, wtBytes)
		out = appendBytes(out, []byte(v.ETag))
	}
	return out
}

func decodeNotModified(b []byte) (AgdiscNotModified, error) {
	var out AgdiscNotModified
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return AgdiscNotModified{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return AgdiscNotModified{}, fmt.Errorf("agdisc_not_modified.agent_id wrong wire type")
			}
			out.AgentID = string(val)
		case 2:
			if wt != wtBytes {
				return AgdiscNotModified{}, fmt.Errorf("agdisc_not_modified.etag wrong wire type")
			}
			out.ETag = string(val)
		}
		b = b[n:]
	}
	return out, nil
}

func encodeErr(v AgdiscErr) []byte {
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

func decodeErr(b []byte) (AgdiscErr, error) {
	var out AgdiscErr
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return AgdiscErr{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return AgdiscErr{}, fmt.Errorf("agdisc_err.code wrong wire type")
			}
			out.Code = string(val)
		case 2:
			if wt != wtBytes {
				return AgdiscErr{}, fmt.Errorf("agdisc_err.message wrong wire type")
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
