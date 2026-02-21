package server

import (
	"context"
	"crypto/sha256"
	"testing"
	"time"

	"swp-spec-kit/poc/internal/core"
	"swp-spec-kit/poc/internal/p1artifact"
	"swp-spec-kit/poc/internal/p1cred"
	"swp-spec-kit/poc/internal/p1state"
)

func resetArtifactState(t *testing.T) {
	t.Helper()
	defaultBackends.artifact = newInMemoryArtifactBackend()
}

func resetStateStore(t *testing.T) {
	t.Helper()
	defaultBackends.state = newInMemoryStateBackend()
}

func resetCredStore(t *testing.T) {
	t.Helper()
	defaultBackends.cred = newInMemoryCredBackend()
}

func TestHandleSWPArtifactOfferChunkGet(t *testing.T) {
	resetArtifactState(t)

	data := []byte("hello")
	h := sha256.Sum256(data)
	offerPayload, err := p1artifact.EncodePayloadOffer(p1artifact.ArtOffer{
		ArtifactID: "artifact-1",
		TotalSize:  uint64(len(data)),
		HashAlg:    "sha256",
		Hash:       h[:],
	})
	if err != nil {
		t.Fatalf("encode offer payload: %v", err)
	}
	offerEnv := core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileSWPArtifact,
		MsgType:   artifactMsgTypeOffer,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   offerPayload,
	}
	out, err := handleSWPArtifact(context.Background(), offerEnv)
	if err != nil {
		t.Fatalf("offer failed: %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("expected no offer response, got %d", len(out))
	}

	chunkPayload, err := p1artifact.EncodePayloadChunk(p1artifact.ArtChunk{
		ArtifactID: "artifact-1",
		ChunkIndex: 0,
		Offset:     0,
		Data:       data,
		IsTerminal: true,
	})
	if err != nil {
		t.Fatalf("encode chunk payload: %v", err)
	}
	chunkEnv := core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileSWPArtifact,
		MsgType:   artifactMsgTypeChunk,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   chunkPayload,
	}
	out, err = handleSWPArtifact(context.Background(), chunkEnv)
	if err != nil {
		t.Fatalf("chunk failed: %v", err)
	}
	if len(out) != 1 || out[0].MsgType != artifactMsgTypeAck {
		t.Fatalf("expected one ack response, got %+v", out)
	}

	getPayload, err := p1artifact.EncodePayloadGet(p1artifact.ArtGet{ArtifactID: "artifact-1"})
	if err != nil {
		t.Fatalf("encode get payload: %v", err)
	}
	getEnv := core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileSWPArtifact,
		MsgType:   artifactMsgTypeGet,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   getPayload,
	}
	out, err = handleSWPArtifact(context.Background(), getEnv)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if len(out) != 1 || out[0].MsgType != artifactMsgTypeChunk {
		t.Fatalf("expected one chunk response, got %+v", out)
	}
	chunk, err := p1artifact.DecodePayloadChunk(out[0].Payload)
	if err != nil {
		t.Fatalf("decode chunk response: %v", err)
	}
	if string(chunk.Data) != "hello" {
		t.Fatalf("expected chunk data hello, got %q", string(chunk.Data))
	}
}

func TestHandleSWPArtifactChunkOrderingError(t *testing.T) {
	resetArtifactState(t)
	offerPayload, _ := p1artifact.EncodePayloadOffer(p1artifact.ArtOffer{ArtifactID: "artifact-2", TotalSize: 1})
	_, _ = handleSWPArtifact(context.Background(), core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileSWPArtifact,
		MsgType:   artifactMsgTypeOffer,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   offerPayload,
	})

	chunkPayload, _ := p1artifact.EncodePayloadChunk(p1artifact.ArtChunk{ArtifactID: "artifact-2", ChunkIndex: 1, Data: []byte("x")})
	out, err := handleSWPArtifact(context.Background(), core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileSWPArtifact,
		MsgType:   artifactMsgTypeChunk,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   chunkPayload,
	})
	if err != nil {
		t.Fatalf("chunk ordering failed unexpectedly: %v", err)
	}
	if len(out) != 1 || out[0].MsgType != artifactMsgTypeErr {
		t.Fatalf("expected one ARTIFACT err response, got %+v", out)
	}
	aerr, err := p1artifact.DecodePayloadErr(out[0].Payload)
	if err != nil {
		t.Fatalf("decode artifact error payload: %v", err)
	}
	if aerr.Code != "ORDERING" {
		t.Fatalf("expected ORDERING, got %q", aerr.Code)
	}
}

func TestHandleSWPStatePutAndGet(t *testing.T) {
	resetStateStore(t)

	blob := []byte("state-one")
	h := sha256.Sum256(blob)
	putPayload, err := p1state.EncodePayloadPut(p1state.StatePut{StateID: h[:], Blob: blob})
	if err != nil {
		t.Fatalf("encode put payload: %v", err)
	}
	putEnv := core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileSWPState,
		MsgType:   stateMsgTypePut,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   putPayload,
	}
	out, err := handleSWPState(context.Background(), putEnv)
	if err != nil {
		t.Fatalf("state put failed: %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("expected no put response, got %d", len(out))
	}

	getPayload, err := p1state.EncodePayloadGet(p1state.StateGet{StateID: h[:]})
	if err != nil {
		t.Fatalf("encode get payload: %v", err)
	}
	getEnv := core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileSWPState,
		MsgType:   stateMsgTypeGet,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   getPayload,
	}
	out, err = handleSWPState(context.Background(), getEnv)
	if err != nil {
		t.Fatalf("state get failed: %v", err)
	}
	if len(out) != 1 || out[0].MsgType != stateMsgTypePut {
		t.Fatalf("expected one STATE_PUT response, got %+v", out)
	}
	resp, err := p1state.DecodePayloadPut(out[0].Payload)
	if err != nil {
		t.Fatalf("decode get response payload: %v", err)
	}
	if string(resp.Blob) != "state-one" {
		t.Fatalf("unexpected state blob %q", string(resp.Blob))
	}
}

func TestHandleSWPStateHashMismatchRejected(t *testing.T) {
	resetStateStore(t)

	putPayload, _ := p1state.EncodePayloadPut(p1state.StatePut{StateID: []byte("bad"), Blob: []byte("state")})
	_, err := handleSWPState(context.Background(), core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileSWPState,
		MsgType:   stateMsgTypePut,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   putPayload,
	})
	if err == nil {
		t.Fatalf("expected hash mismatch error")
	}
	if core.CodeFromError(err) != core.CodeInvalidEnvelope {
		t.Fatalf("expected INVALID_ENVELOPE, got %s", core.CodeFromError(err))
	}
}

func TestHandleSWPCredLifecycle(t *testing.T) {
	resetCredStore(t)

	presentPayload, err := p1cred.EncodePayloadPresent(p1cred.CredPresent{
		CredType:   "jwt",
		Credential: []byte("token-123"),
		ChainID:    []byte("chain-1"),
	})
	if err != nil {
		t.Fatalf("encode present payload: %v", err)
	}
	presentEnv := core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileSWPCred,
		MsgType:   credMsgTypePresent,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   presentPayload,
	}
	out, err := handleSWPCred(context.Background(), presentEnv)
	if err != nil {
		t.Fatalf("present failed: %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("expected no present response, got %d", len(out))
	}

	delegatePayload, err := p1cred.EncodePayloadDelegate(p1cred.CredDelegate{
		ChainID:         []byte("chain-1"),
		Delegation:      []byte("delegate-a"),
		ExpiresAtUnixMs: uint64(time.Now().Add(time.Hour).UnixMilli()),
	})
	if err != nil {
		t.Fatalf("encode delegate payload: %v", err)
	}
	delegateEnv := core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileSWPCred,
		MsgType:   credMsgTypeDelegate,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   delegatePayload,
	}
	out, err = handleSWPCred(context.Background(), delegateEnv)
	if err != nil {
		t.Fatalf("delegate failed: %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("expected no delegate response, got %d", len(out))
	}

	revokePayload, err := p1cred.EncodePayloadRevoke(p1cred.CredRevoke{ChainID: []byte("chain-1"), Reason: "key compromised"})
	if err != nil {
		t.Fatalf("encode revoke payload: %v", err)
	}
	_, err = handleSWPCred(context.Background(), core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileSWPCred,
		MsgType:   credMsgTypeRevoke,
		MsgID:     []byte("12345678abcdefgh"),
		Payload:   revokePayload,
	})
	if err != nil {
		t.Fatalf("revoke failed: %v", err)
	}

	out, err = handleSWPCred(context.Background(), presentEnv)
	if err != nil {
		t.Fatalf("present after revoke failed unexpectedly: %v", err)
	}
	if len(out) != 1 || out[0].MsgType != credMsgTypeErr {
		t.Fatalf("expected one CRED_ERR response, got %+v", out)
	}
	cerr, err := p1cred.DecodePayloadErr(out[0].Payload)
	if err != nil {
		t.Fatalf("decode cred error payload: %v", err)
	}
	if cerr.Code != "REVOKED" {
		t.Fatalf("expected REVOKED, got %q", cerr.Code)
	}
}

func TestHandleSWPCredUnsupportedMsgType(t *testing.T) {
	_, err := handleSWPCred(context.Background(), core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileSWPCred,
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
