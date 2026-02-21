package server

import (
	"context"
	"testing"

	"swp-spec-kit/poc/internal/core"
	"swp-spec-kit/poc/internal/p1policyhint"
	"swp-spec-kit/poc/internal/p1relay"
)

func resetPolicyHints(t *testing.T) {
	t.Helper()
	defaultBackends.policyHint = newInMemoryPolicyHintBackend()
}

func resetRelayStore(t *testing.T) {
	t.Helper()
	defaultBackends.relay = newInMemoryRelayBackend()
}

func TestHandleSWPPolicyHintSetAck(t *testing.T) {
	resetPolicyHints(t)

	payload, err := p1policyhint.EncodePayloadSet(p1policyhint.PolicyHintSet{Constraints: []p1policyhint.Constraint{{
		Key:   "no_external_network",
		Value: "true",
		Mode:  "MUST",
	}}})
	if err != nil {
		t.Fatalf("encode policyhint set payload: %v", err)
	}
	env := core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileSWPPolicyHint,
		MsgType:   policyHintMsgTypeSet,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   payload,
	}
	out, err := handleSWPPolicyHint(context.Background(), env)
	if err != nil {
		t.Fatalf("policyhint set failed: %v", err)
	}
	if len(out) != 1 || out[0].MsgType != policyHintMsgTypeAck {
		t.Fatalf("expected one ack response, got %+v", out)
	}
	ack, err := p1policyhint.DecodePayloadAck(out[0].Payload)
	if err != nil {
		t.Fatalf("decode policyhint ack payload: %v", err)
	}
	if ack.AckID == "" {
		t.Fatalf("expected non-empty ack_id")
	}
}

func TestHandleSWPPolicyHintUnknownMustViolation(t *testing.T) {
	resetPolicyHints(t)

	payload, err := p1policyhint.EncodePayloadSet(p1policyhint.PolicyHintSet{Constraints: []p1policyhint.Constraint{{
		Key:   "unknown_policy_key",
		Value: "x",
		Mode:  "MUST",
	}}})
	if err != nil {
		t.Fatalf("encode policyhint set payload: %v", err)
	}
	env := core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileSWPPolicyHint,
		MsgType:   policyHintMsgTypeSet,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   payload,
	}
	out, err := handleSWPPolicyHint(context.Background(), env)
	if err != nil {
		t.Fatalf("policyhint set failed: %v", err)
	}
	if len(out) != 1 || out[0].MsgType != policyHintMsgTypeViolation {
		t.Fatalf("expected one violation response, got %+v", out)
	}
	viol, err := p1policyhint.DecodePayloadViolation(out[0].Payload)
	if err != nil {
		t.Fatalf("decode policyhint violation payload: %v", err)
	}
	if viol.ReasonCode != "UNKNOWN_KEY" {
		t.Fatalf("expected UNKNOWN_KEY, got %q", viol.ReasonCode)
	}
}

func TestHandleSWPRelayPublishDuplicateAndNack(t *testing.T) {
	resetRelayStore(t)

	publishPayload, err := p1relay.EncodePayloadPublish(p1relay.RelayPublish{
		DeliveryID: []byte("delivery-1"),
		Topic:      "demo",
		Payload:    []byte("hello"),
	})
	if err != nil {
		t.Fatalf("encode relay publish payload: %v", err)
	}
	publishEnv := core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileSWPRelay,
		MsgType:   relayMsgTypePublish,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   publishPayload,
	}
	out, err := handleSWPRelay(context.Background(), publishEnv)
	if err != nil {
		t.Fatalf("relay publish failed: %v", err)
	}
	if len(out) != 1 || out[0].MsgType != relayMsgTypeAck {
		t.Fatalf("expected one ack response, got %+v", out)
	}

	out, err = handleSWPRelay(context.Background(), publishEnv)
	if err != nil {
		t.Fatalf("relay duplicate publish failed: %v", err)
	}
	if len(out) != 1 || out[0].MsgType != relayMsgTypeStatus {
		t.Fatalf("expected one status response for duplicate, got %+v", out)
	}
	status, err := p1relay.DecodePayloadStatus(out[0].Payload)
	if err != nil {
		t.Fatalf("decode relay status payload: %v", err)
	}
	if status.State != "duplicate" {
		t.Fatalf("expected duplicate state, got %q", status.State)
	}

	nackPayload, err := p1relay.EncodePayloadNack(p1relay.RelayNack{DeliveryID: []byte("delivery-1"), Retryable: false, ReasonCode: "downstream_failed"})
	if err != nil {
		t.Fatalf("encode relay nack payload: %v", err)
	}
	nackEnv := core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileSWPRelay,
		MsgType:   relayMsgTypeNack,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   nackPayload,
	}
	out, err = handleSWPRelay(context.Background(), nackEnv)
	if err != nil {
		t.Fatalf("relay nack failed: %v", err)
	}
	if len(out) != 1 || out[0].MsgType != relayMsgTypeStatus {
		t.Fatalf("expected one status response for nack, got %+v", out)
	}
	status, err = p1relay.DecodePayloadStatus(out[0].Payload)
	if err != nil {
		t.Fatalf("decode relay status payload: %v", err)
	}
	if status.State != "dead-letter" {
		t.Fatalf("expected dead-letter state, got %q", status.State)
	}
}

func TestHandleSWPRelayUnsupportedMsgType(t *testing.T) {
	_, err := handleSWPRelay(context.Background(), core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileSWPRelay,
		MsgType:   99,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   []byte{0x0a, 0x00},
	})
	if err == nil {
		t.Fatalf("expected unsupported msg_type error")
	}
	if core.CodeFromError(err) != core.CodeInvalidEnvelope {
		t.Fatalf("expected INVALID_ENVELOPE, got %s", core.CodeFromError(err))
	}
}
