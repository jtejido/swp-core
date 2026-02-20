package core

import (
	"bytes"
	"testing"
)

func TestWriteReadFrame(t *testing.T) {
	var buf bytes.Buffer
	payload := []byte("abc")
	if err := WriteFrame(&buf, payload, DefaultLimits().MaxFrameBytes); err != nil {
		t.Fatalf("WriteFrame failed: %v", err)
	}
	got, err := ReadFrame(&buf, DefaultLimits().MaxFrameBytes)
	if err != nil {
		t.Fatalf("ReadFrame failed: %v", err)
	}
	if string(got) != string(payload) {
		t.Fatalf("frame payload mismatch: got %q want %q", got, payload)
	}
}

func TestReadFrameZeroLengthInvalid(t *testing.T) {
	buf := bytes.NewBuffer([]byte{0, 0, 0, 0})
	_, err := ReadFrame(buf, DefaultLimits().MaxFrameBytes)
	if err == nil {
		t.Fatalf("expected error")
	}
	if CodeFromError(err) != CodeInvalidFrame {
		t.Fatalf("unexpected code: %s", CodeFromError(err))
	}
}
