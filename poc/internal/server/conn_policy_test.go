package server

import (
	"errors"
	"testing"
	"time"
)

func TestConnPolicyRateLimitExceeded(t *testing.T) {
	base := time.Unix(0, 0)
	p := newConnPolicy(base)
	p.maxFrames = 2
	p.rateWindow = time.Second

	if err := p.check(base, []byte("msg-1")); err != nil {
		t.Fatalf("first message should pass: %v", err)
	}
	if err := p.check(base.Add(100*time.Millisecond), []byte("msg-2")); err != nil {
		t.Fatalf("second message should pass: %v", err)
	}
	err := p.check(base.Add(200*time.Millisecond), []byte("msg-3"))
	if !errors.Is(err, errRateLimitExceeded) {
		t.Fatalf("expected rate limit exceeded, got %v", err)
	}
}

func TestConnPolicyDuplicateMsgID(t *testing.T) {
	base := time.Unix(0, 0)
	p := newConnPolicy(base)
	p.maxFrames = 100
	p.rateWindow = time.Second
	p.msgIDWindow = 2 * time.Second

	msgID := []byte("same-id")
	if err := p.check(base, msgID); err != nil {
		t.Fatalf("first message should pass: %v", err)
	}
	err := p.check(base.Add(500*time.Millisecond), msgID)
	if !errors.Is(err, errDuplicateMsgID) {
		t.Fatalf("expected duplicate msg_id error, got %v", err)
	}
	if err := p.check(base.Add(3*time.Second), msgID); err != nil {
		t.Fatalf("message after duplicate window should pass: %v", err)
	}
}
