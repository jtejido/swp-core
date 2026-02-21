package server

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"swp-spec-kit/poc/internal/core"
	"swp-spec-kit/poc/internal/p1cred"
)

const (
	credMsgTypePresent  = 1
	credMsgTypeDelegate = 2
	credMsgTypeRevoke   = 3
	credMsgTypeErr      = 4
)

const maxDelegationDepth = 8

func handleSWPCred(ctx context.Context, env core.Envelope) ([]core.Envelope, error) {
	return handleSWPCredWithBackend(ctx, env, defaultBackends.cred)
}

func (s *Server) handleSWPCred(ctx context.Context, env core.Envelope) ([]core.Envelope, error) {
	return handleSWPCredWithBackend(ctx, env, s.runtime.cred)
}

func handleSWPCredWithBackend(_ context.Context, env core.Envelope, backend CredBackend) ([]core.Envelope, error) {
	now := uint64(time.Now().UnixMilli())

	switch env.MsgType {
	case credMsgTypePresent:
		present, err := p1cred.DecodePayloadPresent(env.Payload)
		if err != nil {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid CRED present payload: %w", err))
		}
		if strings.TrimSpace(present.CredType) == "" {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("cred_type required"))
		}
		if len(present.Credential) == 0 {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("credential required"))
		}
		if !isSupportedCredType(present.CredType) {
			return []core.Envelope{newCredErrEnvelope(env.MsgID, now, "UNSUPPORTED_CRED_TYPE", "credential type not supported")}, nil
		}
		if bytes.Contains(bytes.ToLower(present.Credential), []byte("invalid")) {
			return []core.Envelope{newCredErrEnvelope(env.MsgID, now, "INVALID_CREDENTIAL", "invalid credential")}, nil
		}
		if bytes.Contains(bytes.ToLower(present.Credential), []byte("expired")) {
			return []core.Envelope{newCredErrEnvelope(env.MsgID, now, "EXPIRED", "credential expired")}, nil
		}

		backend.EnsureChain(present.ChainID)
		if backend.IsRevoked(present.ChainID) {
			return []core.Envelope{newCredErrEnvelope(env.MsgID, now, "REVOKED", "credential chain revoked")}, nil
		}
		return nil, nil

	case credMsgTypeDelegate:
		del, err := p1cred.DecodePayloadDelegate(env.Payload)
		if err != nil {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid CRED delegate payload: %w", err))
		}
		if len(del.ChainID) == 0 {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("chain_id required"))
		}
		if len(del.Delegation) == 0 {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("delegation required"))
		}
		if del.ExpiresAtUnixMs == 0 {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("expires_at_unix_ms required"))
		}
		if del.ExpiresAtUnixMs <= now {
			return []core.Envelope{newCredErrEnvelope(env.MsgID, now, "EXPIRED", "delegation expired")}, nil
		}

		if backend.IsRevoked(del.ChainID) {
			return []core.Envelope{newCredErrEnvelope(env.MsgID, now, "REVOKED", "credential chain revoked")}, nil
		}
		if depth := backend.IncrementChainDepth(del.ChainID); depth > maxDelegationDepth {
			return []core.Envelope{newCredErrEnvelope(env.MsgID, now, "CHAIN_LIMIT", "delegation chain length exceeded")}, nil
		}
		return nil, nil

	case credMsgTypeRevoke:
		rev, err := p1cred.DecodePayloadRevoke(env.Payload)
		if err != nil {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid CRED revoke payload: %w", err))
		}
		if len(rev.ChainID) == 0 {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("chain_id required"))
		}
		backend.Revoke(rev.ChainID)
		return nil, nil

	default:
		return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid CRED msg_type %d", env.MsgType))
	}
}

func isSupportedCredType(credType string) bool {
	switch strings.ToLower(strings.TrimSpace(credType)) {
	case "jwt", "mtls", "opaque":
		return true
	default:
		return false
	}
}

func newCredErrEnvelope(msgID []byte, ts uint64, code, message string) core.Envelope {
	payload, err := p1cred.EncodePayloadErr(p1cred.CredErr{Code: code, Message: message})
	if err != nil {
		payload = nil
	}
	return core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileSWPCred,
		MsgType:   credMsgTypeErr,
		MsgID:     append([]byte(nil), msgID...),
		TsUnixMs:  ts,
		Payload:   payload,
	}
}
