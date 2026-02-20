package core

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

func EncodeEnvelopeE1(env Envelope) ([]byte, error) {
	var out []byte
	out = binary.AppendUvarint(out, env.Version)
	out = binary.AppendUvarint(out, env.ProfileID)
	out = binary.AppendUvarint(out, env.MsgType)
	out = binary.AppendUvarint(out, env.Flags)
	out = binary.AppendUvarint(out, env.TsUnixMs)

	out = appendBytes(out, env.MsgID)

	extBytes, err := encodeExtensions(env.Extensions)
	if err != nil {
		return nil, Wrap(CodeInvalidEnvelope, err)
	}
	out = appendBytes(out, extBytes)

	out = appendBytes(out, env.Payload)
	return out, nil
}

func DecodeEnvelopeE1(body []byte, limits Limits) (Envelope, error) {
	r := bytes.NewReader(body)

	version, err := readUvarint(r)
	if err != nil {
		return Envelope{}, Wrap(CodeInvalidFrame, fmt.Errorf("decode version: %w", err))
	}
	profileID, err := readUvarint(r)
	if err != nil {
		return Envelope{}, Wrap(CodeInvalidFrame, fmt.Errorf("decode profile_id: %w", err))
	}
	msgType, err := readUvarint(r)
	if err != nil {
		return Envelope{}, Wrap(CodeInvalidFrame, fmt.Errorf("decode msg_type: %w", err))
	}
	flags, err := readUvarint(r)
	if err != nil {
		return Envelope{}, Wrap(CodeInvalidFrame, fmt.Errorf("decode flags: %w", err))
	}
	ts, err := readUvarint(r)
	if err != nil {
		return Envelope{}, Wrap(CodeInvalidFrame, fmt.Errorf("decode ts_unix_ms: %w", err))
	}

	msgID, err := readBytes(r)
	if err != nil {
		return Envelope{}, Wrap(CodeInvalidFrame, fmt.Errorf("decode msg_id: %w", err))
	}
	extRaw, err := readBytes(r)
	if err != nil {
		return Envelope{}, Wrap(CodeInvalidFrame, fmt.Errorf("decode extensions: %w", err))
	}
	if len(extRaw) > limits.MaxExtBytes {
		return Envelope{}, Wrap(CodeInvalidEnvelope, fmt.Errorf("extensions length %d exceeds max %d", len(extRaw), limits.MaxExtBytes))
	}
	ext, err := decodeExtensions(extRaw)
	if err != nil {
		return Envelope{}, Wrap(CodeInvalidFrame, fmt.Errorf("decode extension TLV: %w", err))
	}

	payload, err := readBytes(r)
	if err != nil {
		return Envelope{}, Wrap(CodeInvalidFrame, fmt.Errorf("decode payload: %w", err))
	}
	if uint32(len(payload)) > limits.MaxPayloadBytes {
		return Envelope{}, Wrap(CodeInvalidEnvelope, fmt.Errorf("payload length %d exceeds max %d", len(payload), limits.MaxPayloadBytes))
	}

	if r.Len() != 0 {
		return Envelope{}, Wrap(CodeInvalidFrame, fmt.Errorf("trailing bytes: %d", r.Len()))
	}

	return Envelope{
		Version:    version,
		ProfileID:  profileID,
		MsgType:    msgType,
		Flags:      flags,
		TsUnixMs:   ts,
		MsgID:      msgID,
		Extensions: ext,
		Payload:    payload,
	}, nil
}

func appendBytes(dst, v []byte) []byte {
	dst = binary.AppendUvarint(dst, uint64(len(v)))
	return append(dst, v...)
}

func readUvarint(r io.ByteReader) (uint64, error) {
	v, err := binary.ReadUvarint(r)
	if err != nil {
		return 0, err
	}
	return v, nil
}

func readBytes(r *bytes.Reader) ([]byte, error) {
	n, err := readUvarint(r)
	if err != nil {
		return nil, err
	}
	if n > uint64(r.Len()) {
		return nil, io.ErrUnexpectedEOF
	}
	out := make([]byte, int(n))
	if _, err := io.ReadFull(r, out); err != nil {
		return nil, err
	}
	return out, nil
}

func encodeExtensions(ext []Extension) ([]byte, error) {
	var out []byte
	for _, e := range ext {
		out = binary.AppendUvarint(out, e.Type)
		out = appendBytes(out, e.Value)
	}
	return out, nil
}

func decodeExtensions(raw []byte) ([]Extension, error) {
	r := bytes.NewReader(raw)
	ext := make([]Extension, 0)
	for r.Len() > 0 {
		t, err := readUvarint(r)
		if err != nil {
			return nil, err
		}
		v, err := readBytes(r)
		if err != nil {
			return nil, err
		}
		ext = append(ext, Extension{Type: t, Value: v})
	}
	return ext, nil
}
