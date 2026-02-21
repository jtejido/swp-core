package p1fixture

import (
	"encoding/binary"
	"fmt"
)

const (
	wtVarint = 0
	wtBytes  = 2
)

type Decision struct {
	Reject bool
	Code   string
	Reason string
}

// Evaluate parses the synthetic fixture payload used by conformance vectors and
// applies deterministic reject rules keyed by vector_id.
func Evaluate(payload []byte, expectedProfile string, rejectRules map[string]Decision) (Decision, error) {
	profile, vectorID, err := DecodeProfileAndVectorID(payload)
	if err != nil {
		return Decision{}, err
	}
	if profile != expectedProfile {
		return Decision{
			Reject: true,
			Code:   "INVALID_PROFILE_PAYLOAD",
			Reason: fmt.Sprintf("invalid %s payload marker", expectedProfile),
		}, nil
	}
	if d, ok := rejectRules[vectorID]; ok {
		return d, nil
	}
	return Decision{Reject: false, Code: "OK"}, nil
}

func DecodeProfileAndVectorID(payload []byte) (string, string, error) {
	var profile string
	var vectorID string

	for len(payload) > 0 {
		field, wt, val, n, err := consumeField(payload)
		if err != nil {
			return "", "", err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return "", "", fmt.Errorf("fixture profile marker has wrong wire type")
			}
			profile = string(val)
		case 2:
			if wt != wtBytes {
				return "", "", fmt.Errorf("fixture vector id has wrong wire type")
			}
			vectorID = string(val)
		}
		payload = payload[n:]
	}

	if profile == "" {
		return "", "", fmt.Errorf("fixture payload missing profile marker")
	}
	if vectorID == "" {
		return "", "", fmt.Errorf("fixture payload missing vector id")
	}
	return profile, vectorID, nil
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
