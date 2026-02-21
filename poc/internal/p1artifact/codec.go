package p1artifact

import (
	"encoding/binary"
	"fmt"
)

const (
	wtVarint = 0
	wtBytes  = 2
)

type ArtOffer struct {
	ArtifactID string
	TotalSize  uint64
	HashAlg    string
	Hash       []byte
	Metadata   []byte
}

type ArtGet struct {
	ArtifactID  string
	Start       uint64
	End         uint64
	ResumeToken string
}

type ArtChunk struct {
	ArtifactID  string
	ChunkIndex  uint64
	Offset      uint64
	Data        []byte
	IsTerminal  bool
	ResumeToken string
}

type ArtAck struct {
	ArtifactID string
	ChunkIndex uint64
}

type ArtErr struct {
	Code      string
	Message   string
	Retryable bool
}

func EncodePayloadOffer(v ArtOffer) ([]byte, error) {
	return encodeWrapper(1, encodeOffer(v)), nil
}

func DecodePayloadOffer(payload []byte) (ArtOffer, error) {
	inner, err := decodeWrapper(payload, 1)
	if err != nil {
		return ArtOffer{}, err
	}
	return decodeOffer(inner)
}

func EncodePayloadGet(v ArtGet) ([]byte, error) {
	return encodeWrapper(2, encodeGet(v)), nil
}

func DecodePayloadGet(payload []byte) (ArtGet, error) {
	inner, err := decodeWrapper(payload, 2)
	if err != nil {
		return ArtGet{}, err
	}
	return decodeGet(inner)
}

func EncodePayloadChunk(v ArtChunk) ([]byte, error) {
	return encodeWrapper(3, encodeChunk(v)), nil
}

func DecodePayloadChunk(payload []byte) (ArtChunk, error) {
	inner, err := decodeWrapper(payload, 3)
	if err != nil {
		return ArtChunk{}, err
	}
	return decodeChunk(inner)
}

func EncodePayloadAck(v ArtAck) ([]byte, error) {
	return encodeWrapper(4, encodeAck(v)), nil
}

func DecodePayloadAck(payload []byte) (ArtAck, error) {
	inner, err := decodeWrapper(payload, 4)
	if err != nil {
		return ArtAck{}, err
	}
	return decodeAck(inner)
}

func EncodePayloadErr(v ArtErr) ([]byte, error) {
	return encodeWrapper(5, encodeErr(v)), nil
}

func DecodePayloadErr(payload []byte) (ArtErr, error) {
	inner, err := decodeWrapper(payload, 5)
	if err != nil {
		return ArtErr{}, err
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

func encodeOffer(v ArtOffer) []byte {
	var out []byte
	if v.ArtifactID != "" {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, []byte(v.ArtifactID))
	}
	if v.TotalSize != 0 {
		out = appendKey(out, 2, wtVarint)
		out = binary.AppendUvarint(out, v.TotalSize)
	}
	if v.HashAlg != "" {
		out = appendKey(out, 3, wtBytes)
		out = appendBytes(out, []byte(v.HashAlg))
	}
	if len(v.Hash) > 0 {
		out = appendKey(out, 4, wtBytes)
		out = appendBytes(out, v.Hash)
	}
	if len(v.Metadata) > 0 {
		out = appendKey(out, 5, wtBytes)
		out = appendBytes(out, v.Metadata)
	}
	return out
}

func decodeOffer(b []byte) (ArtOffer, error) {
	var out ArtOffer
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return ArtOffer{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return ArtOffer{}, fmt.Errorf("art_offer.artifact_id wrong wire type")
			}
			out.ArtifactID = string(val)
		case 2:
			if wt != wtVarint {
				return ArtOffer{}, fmt.Errorf("art_offer.total_size wrong wire type")
			}
			vv, _, err := consumeVarintValue(val)
			if err != nil {
				return ArtOffer{}, err
			}
			out.TotalSize = vv
		case 3:
			if wt != wtBytes {
				return ArtOffer{}, fmt.Errorf("art_offer.hash_alg wrong wire type")
			}
			out.HashAlg = string(val)
		case 4:
			if wt != wtBytes {
				return ArtOffer{}, fmt.Errorf("art_offer.hash wrong wire type")
			}
			out.Hash = append([]byte(nil), val...)
		case 5:
			if wt != wtBytes {
				return ArtOffer{}, fmt.Errorf("art_offer.metadata wrong wire type")
			}
			out.Metadata = append([]byte(nil), val...)
		}
		b = b[n:]
	}
	return out, nil
}

func encodeGet(v ArtGet) []byte {
	var out []byte
	if v.ArtifactID != "" {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, []byte(v.ArtifactID))
	}
	if v.Start != 0 {
		out = appendKey(out, 2, wtVarint)
		out = binary.AppendUvarint(out, v.Start)
	}
	if v.End != 0 {
		out = appendKey(out, 3, wtVarint)
		out = binary.AppendUvarint(out, v.End)
	}
	if v.ResumeToken != "" {
		out = appendKey(out, 4, wtBytes)
		out = appendBytes(out, []byte(v.ResumeToken))
	}
	return out
}

func decodeGet(b []byte) (ArtGet, error) {
	var out ArtGet
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return ArtGet{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return ArtGet{}, fmt.Errorf("art_get.artifact_id wrong wire type")
			}
			out.ArtifactID = string(val)
		case 2:
			if wt != wtVarint {
				return ArtGet{}, fmt.Errorf("art_get.start wrong wire type")
			}
			vv, _, err := consumeVarintValue(val)
			if err != nil {
				return ArtGet{}, err
			}
			out.Start = vv
		case 3:
			if wt != wtVarint {
				return ArtGet{}, fmt.Errorf("art_get.end wrong wire type")
			}
			vv, _, err := consumeVarintValue(val)
			if err != nil {
				return ArtGet{}, err
			}
			out.End = vv
		case 4:
			if wt != wtBytes {
				return ArtGet{}, fmt.Errorf("art_get.resume_token wrong wire type")
			}
			out.ResumeToken = string(val)
		}
		b = b[n:]
	}
	return out, nil
}

func encodeChunk(v ArtChunk) []byte {
	var out []byte
	if v.ArtifactID != "" {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, []byte(v.ArtifactID))
	}
	if v.ChunkIndex != 0 {
		out = appendKey(out, 2, wtVarint)
		out = binary.AppendUvarint(out, v.ChunkIndex)
	}
	if v.Offset != 0 {
		out = appendKey(out, 3, wtVarint)
		out = binary.AppendUvarint(out, v.Offset)
	}
	if len(v.Data) > 0 {
		out = appendKey(out, 4, wtBytes)
		out = appendBytes(out, v.Data)
	}
	if v.IsTerminal {
		out = appendKey(out, 5, wtVarint)
		out = binary.AppendUvarint(out, 1)
	}
	if v.ResumeToken != "" {
		out = appendKey(out, 6, wtBytes)
		out = appendBytes(out, []byte(v.ResumeToken))
	}
	return out
}

func decodeChunk(b []byte) (ArtChunk, error) {
	var out ArtChunk
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return ArtChunk{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return ArtChunk{}, fmt.Errorf("art_chunk.artifact_id wrong wire type")
			}
			out.ArtifactID = string(val)
		case 2:
			if wt != wtVarint {
				return ArtChunk{}, fmt.Errorf("art_chunk.chunk_index wrong wire type")
			}
			vv, _, err := consumeVarintValue(val)
			if err != nil {
				return ArtChunk{}, err
			}
			out.ChunkIndex = vv
		case 3:
			if wt != wtVarint {
				return ArtChunk{}, fmt.Errorf("art_chunk.offset wrong wire type")
			}
			vv, _, err := consumeVarintValue(val)
			if err != nil {
				return ArtChunk{}, err
			}
			out.Offset = vv
		case 4:
			if wt != wtBytes {
				return ArtChunk{}, fmt.Errorf("art_chunk.data wrong wire type")
			}
			out.Data = append([]byte(nil), val...)
		case 5:
			if wt != wtVarint {
				return ArtChunk{}, fmt.Errorf("art_chunk.is_terminal wrong wire type")
			}
			vv, _, err := consumeVarintValue(val)
			if err != nil {
				return ArtChunk{}, err
			}
			out.IsTerminal = vv != 0
		case 6:
			if wt != wtBytes {
				return ArtChunk{}, fmt.Errorf("art_chunk.resume_token wrong wire type")
			}
			out.ResumeToken = string(val)
		}
		b = b[n:]
	}
	return out, nil
}

func encodeAck(v ArtAck) []byte {
	var out []byte
	if v.ArtifactID != "" {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, []byte(v.ArtifactID))
	}
	if v.ChunkIndex != 0 {
		out = appendKey(out, 2, wtVarint)
		out = binary.AppendUvarint(out, v.ChunkIndex)
	}
	return out
}

func decodeAck(b []byte) (ArtAck, error) {
	var out ArtAck
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return ArtAck{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return ArtAck{}, fmt.Errorf("art_ack.artifact_id wrong wire type")
			}
			out.ArtifactID = string(val)
		case 2:
			if wt != wtVarint {
				return ArtAck{}, fmt.Errorf("art_ack.chunk_index wrong wire type")
			}
			vv, _, err := consumeVarintValue(val)
			if err != nil {
				return ArtAck{}, err
			}
			out.ChunkIndex = vv
		}
		b = b[n:]
	}
	return out, nil
}

func encodeErr(v ArtErr) []byte {
	var out []byte
	if v.Code != "" {
		out = appendKey(out, 1, wtBytes)
		out = appendBytes(out, []byte(v.Code))
	}
	if v.Message != "" {
		out = appendKey(out, 2, wtBytes)
		out = appendBytes(out, []byte(v.Message))
	}
	if v.Retryable {
		out = appendKey(out, 3, wtVarint)
		out = binary.AppendUvarint(out, 1)
	}
	return out
}

func decodeErr(b []byte) (ArtErr, error) {
	var out ArtErr
	for len(b) > 0 {
		field, wt, val, n, err := consumeField(b)
		if err != nil {
			return ArtErr{}, err
		}
		switch field {
		case 1:
			if wt != wtBytes {
				return ArtErr{}, fmt.Errorf("art_err.code wrong wire type")
			}
			out.Code = string(val)
		case 2:
			if wt != wtBytes {
				return ArtErr{}, fmt.Errorf("art_err.message wrong wire type")
			}
			out.Message = string(val)
		case 3:
			if wt != wtVarint {
				return ArtErr{}, fmt.Errorf("art_err.retryable wrong wire type")
			}
			vv, _, err := consumeVarintValue(val)
			if err != nil {
				return ArtErr{}, err
			}
			out.Retryable = vv != 0
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
