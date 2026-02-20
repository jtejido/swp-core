package core

import (
	"encoding/binary"
	"fmt"
	"io"
)

func ReadFrame(r io.Reader, maxFrame uint32) ([]byte, error) {
	var prefix [4]byte
	if _, err := io.ReadFull(r, prefix[:]); err != nil {
		return nil, Wrap(CodeInvalidFrame, fmt.Errorf("read prefix: %w", err))
	}

	n := binary.BigEndian.Uint32(prefix[:])
	if n == 0 {
		return nil, Wrap(CodeInvalidFrame, fmt.Errorf("zero-length frame"))
	}
	if n > maxFrame {
		return nil, Wrap(CodeInvalidFrame, fmt.Errorf("frame length %d exceeds max %d", n, maxFrame))
	}

	body := make([]byte, n)
	if _, err := io.ReadFull(r, body); err != nil {
		return nil, Wrap(CodeInvalidFrame, fmt.Errorf("read frame body: %w", err))
	}
	return body, nil
}

func WriteFrame(w io.Writer, body []byte, maxFrame uint32) error {
	n := uint32(len(body))
	if n == 0 {
		return Wrap(CodeInvalidFrame, fmt.Errorf("zero-length frame"))
	}
	if n > maxFrame {
		return Wrap(CodeInvalidFrame, fmt.Errorf("frame length %d exceeds max %d", n, maxFrame))
	}

	var prefix [4]byte
	binary.BigEndian.PutUint32(prefix[:], n)
	if _, err := w.Write(prefix[:]); err != nil {
		return Wrap(CodeInternalError, fmt.Errorf("write prefix: %w", err))
	}
	if _, err := w.Write(body); err != nil {
		return Wrap(CodeInternalError, fmt.Errorf("write body: %w", err))
	}
	return nil
}
