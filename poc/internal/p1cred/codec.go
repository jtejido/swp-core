package p1cred

import (
	"encoding/binary"
	"fmt"
)

const (
	wtVarint = 0
	wtBytes  = 2
)

type CredPresent struct {
	CredType   string
	Credential []byte
	ChainID    []byte
}

type CredDelegate struct {
	ChainID         []byte
	Delegation      []byte
	ExpiresAtUnixMs uint64
}

type CredRevoke struct {
	ChainID []byte
	Reason  string
}

type CredErr struct {
	Code    string
	Message string
}

func EncodePayloadPresent(v CredPresent) ([]byte, error) {
	return encodeWrapper(1, encodePresent(v)), nil
}

func DecodePayloadPresent(payload []byte) (CredPresent, error) {
	inner, err := decodeWrapper(payload, 1)
	if err != nil {
		return CredPresent{}, err
	}
	return decodePresent(inner)
}

func EncodePayloadDelegate(v CredDelegate) ([]byte, error) {
	return encodeWrapper(2, encodeDelegate(v)), nil
}

func DecodePayloadDelegate(payload []byte) (CredDelegate, error) {
	inner, err := decodeWrapper(payload, 2)
	if err != nil {
		return CredDelegate{}, err
	}
	return decodeDelegate(inner)
}

func EncodePayloadRevoke(v CredRevoke) ([]byte, error) {
	return encodeWrapper(3, encodeRevoke(v)), nil
}

func DecodePayloadRevoke(payload []byte) (CredRevoke, error) {
	inner, err := decodeWrapper(payload, 3)
	if err != nil {
		return CredRevoke{}, err
	}
	return decodeRevoke(inner)
}

func EncodePayloadErr(v CredErr) ([]byte, error) {
	return encodeWrapper(4, encodeErr(v)), nil
}

func DecodePayloadErr(payload []byte) (CredErr, error) {
	inner, err := decodeWrapper(payload, 4)
	if err != nil {
		return CredErr{}, err
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

func encodePresent(v CredPresent) []byte {
	var out []byte
	if v.CredType != "" {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, []byte(v.CredType))
	}
	if len(v.Credential) > 0 {
		out = appendKey(out, 2, wtBytes)
		out = appendBytes(out, v.Credential)
	}
	if len(v.ChainID) > 0 {
		out = appendKey(out, 3, wtBytes)
		out = appendBytes(out, v.ChainID)
	}
	return out
}

func decodePresent(b []byte) (CredPresent, error) {
	var out CredPresent
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return CredPresent{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return CredPresent{}, fmt.Errorf("cred_present.cred_type wrong wire type")
			}
			out.CredType = string(val)
		case 2:
			if wt != wtBytes {
				return CredPresent{}, fmt.Errorf("cred_present.credential wrong wire type")
			}
			out.Credential = append([]byte(nil), val...)
		case 3:
			if wt != wtBytes {
				return CredPresent{}, fmt.Errorf("cred_present.chain_id wrong wire type")
			}
			out.ChainID = append([]byte(nil), val...)
		}
		b = b[n:]
	}
	return out, nil
}

func encodeDelegate(v CredDelegate) []byte {
	var out []byte
	if len(v.ChainID) > 0 {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, v.ChainID)
	}
	if len(v.Delegation) > 0 {
		out = appendKey(out, 2, wtBytes)
		out = appendBytes(out, v.Delegation)
	}
	if v.ExpiresAtUnixMs != 0 {
		out = appendKey(out, 3, wtVarint)
		out = binary.AppendUvarint(out, v.ExpiresAtUnixMs)
	}
	return out
}

func decodeDelegate(b []byte) (CredDelegate, error) {
	var out CredDelegate
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return CredDelegate{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return CredDelegate{}, fmt.Errorf("cred_delegate.chain_id wrong wire type")
			}
			out.ChainID = append([]byte(nil), val...)
		case 2:
			if wt != wtBytes {
				return CredDelegate{}, fmt.Errorf("cred_delegate.delegation wrong wire type")
			}
			out.Delegation = append([]byte(nil), val...)
		case 3:
			if wt != wtVarint {
				return CredDelegate{}, fmt.Errorf("cred_delegate.expires_at_unix_ms wrong wire type")
			}
			vv, _, err := consumeVarintValue(val)
			if err != nil {
				return CredDelegate{}, err
			}
			out.ExpiresAtUnixMs = vv
		}
		b = b[n:]
	}
	return out, nil
}

func encodeRevoke(v CredRevoke) []byte {
	var out []byte
	if len(v.ChainID) > 0 {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, v.ChainID)
	}
	if v.Reason != "" {
		out = appendKey(out, 2, wtBytes)
		out = appendBytes(out, []byte(v.Reason))
	}
	return out
}

func decodeRevoke(b []byte) (CredRevoke, error) {
	var out CredRevoke
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return CredRevoke{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return CredRevoke{}, fmt.Errorf("cred_revoke.chain_id wrong wire type")
			}
			out.ChainID = append([]byte(nil), val...)
		case 2:
			if wt != wtBytes {
				return CredRevoke{}, fmt.Errorf("cred_revoke.reason wrong wire type")
			}
			out.Reason = string(val)
		}
		b = b[n:]
	}
	return out, nil
}

func encodeErr(v CredErr) []byte {
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

func decodeErr(b []byte) (CredErr, error) {
	var out CredErr
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return CredErr{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return CredErr{}, fmt.Errorf("cred_err.code wrong wire type")
			}
			out.Code = string(val)
		case 2:
			if wt != wtBytes {
				return CredErr{}, fmt.Errorf("cred_err.message wrong wire type")
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
