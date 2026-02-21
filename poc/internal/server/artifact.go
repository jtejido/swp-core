package server

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"swp-spec-kit/poc/internal/core"
	"swp-spec-kit/poc/internal/p1artifact"
)

const (
	artifactMsgTypeOffer = 1
	artifactMsgTypeGet   = 2
	artifactMsgTypeChunk = 3
	artifactMsgTypeAck   = 4
	artifactMsgTypeErr   = 5
)

func handleSWPArtifact(ctx context.Context, env core.Envelope) ([]core.Envelope, error) {
	return handleSWPArtifactWithBackend(ctx, env, defaultBackends.artifact)
}

func (s *Server) handleSWPArtifact(ctx context.Context, env core.Envelope) ([]core.Envelope, error) {
	return handleSWPArtifactWithBackend(ctx, env, s.runtime.artifact)
}

func handleSWPArtifactWithBackend(_ context.Context, env core.Envelope, backend ArtifactBackend) ([]core.Envelope, error) {
	now := uint64(time.Now().UnixMilli())

	switch env.MsgType {
	case artifactMsgTypeOffer:
		offer, err := p1artifact.DecodePayloadOffer(env.Payload)
		if err != nil {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid ARTIFACT offer payload: %w", err))
		}
		if strings.TrimSpace(offer.ArtifactID) == "" {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("artifact_id required"))
		}
		backend.PutOffer(offer)
		return nil, nil

	case artifactMsgTypeGet:
		get, err := p1artifact.DecodePayloadGet(env.Payload)
		if err != nil {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid ARTIFACT get payload: %w", err))
		}
		if strings.TrimSpace(get.ArtifactID) == "" {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("artifact_id required"))
		}
		rec, ok := backend.GetArtifact(get.ArtifactID)
		if !ok {
			errPayload, err := p1artifact.EncodePayloadErr(p1artifact.ArtErr{Code: "NOT_FOUND", Message: "artifact not found"})
			if err != nil {
				return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("encode ARTIFACT not-found error: %w", err))
			}
			return []core.Envelope{newArtifactEnvelope(env.MsgID, artifactMsgTypeErr, now, errPayload)}, nil
		}

		start := get.Start
		end := get.End
		total := uint64(len(rec.Data))
		if start > total {
			start = total
		}
		if end == 0 || end > total {
			end = total
		}
		if start > end {
			errPayload, err := p1artifact.EncodePayloadErr(p1artifact.ArtErr{Code: "INVALID_RANGE", Message: "start greater than end"})
			if err != nil {
				return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("encode ARTIFACT invalid-range error: %w", err))
			}
			return []core.Envelope{newArtifactEnvelope(env.MsgID, artifactMsgTypeErr, now, errPayload)}, nil
		}

		chunkPayload, err := p1artifact.EncodePayloadChunk(p1artifact.ArtChunk{
			ArtifactID:  get.ArtifactID,
			ChunkIndex:  0,
			Offset:      start,
			Data:        append([]byte(nil), rec.Data[start:end]...),
			IsTerminal:  end == total,
			ResumeToken: fmt.Sprintf("%s:%d", get.ArtifactID, end),
		})
		if err != nil {
			return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("encode ARTIFACT chunk response: %w", err))
		}
		return []core.Envelope{newArtifactEnvelope(env.MsgID, artifactMsgTypeChunk, now, chunkPayload)}, nil

	case artifactMsgTypeChunk:
		chunk, err := p1artifact.DecodePayloadChunk(env.Payload)
		if err != nil {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid ARTIFACT chunk payload: %w", err))
		}
		if strings.TrimSpace(chunk.ArtifactID) == "" {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("artifact_id required"))
		}

		rec, err := backend.AppendChunk(chunk)
		if err != nil {
			if err == errArtifactChunkOrdering {
				errPayload, e := p1artifact.EncodePayloadErr(p1artifact.ArtErr{Code: "ORDERING", Message: "unexpected chunk index"})
				if e != nil {
					return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("encode ARTIFACT ordering error: %w", e))
				}
				return []core.Envelope{newArtifactEnvelope(env.MsgID, artifactMsgTypeErr, now, errPayload)}, nil
			}
			return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("append ARTIFACT chunk: %w", err))
		}

		if chunk.IsTerminal {
			if rec.Offer.TotalSize > 0 && uint64(len(rec.Data)) != rec.Offer.TotalSize {
				errPayload, err := p1artifact.EncodePayloadErr(p1artifact.ArtErr{Code: "SIZE_MISMATCH", Message: "artifact size mismatch"})
				if err != nil {
					return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("encode ARTIFACT size mismatch error: %w", err))
				}
				return []core.Envelope{newArtifactEnvelope(env.MsgID, artifactMsgTypeErr, now, errPayload)}, nil
			}
			if len(rec.Offer.Hash) > 0 && strings.EqualFold(rec.Offer.HashAlg, "sha256") {
				sum := sha256.Sum256(rec.Data)
				if !bytes.Equal(sum[:], rec.Offer.Hash) {
					errPayload, err := p1artifact.EncodePayloadErr(p1artifact.ArtErr{Code: "INTEGRITY_MISMATCH", Message: "artifact hash mismatch"})
					if err != nil {
						return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("encode ARTIFACT integrity error: %w", err))
					}
					return []core.Envelope{newArtifactEnvelope(env.MsgID, artifactMsgTypeErr, now, errPayload)}, nil
				}
			}
		}

		ackPayload, err := p1artifact.EncodePayloadAck(p1artifact.ArtAck{
			ArtifactID: chunk.ArtifactID,
			ChunkIndex: chunk.ChunkIndex,
		})
		if err != nil {
			return nil, core.Wrap(core.CodeInternalError, fmt.Errorf("encode ARTIFACT ack: %w", err))
		}
		return []core.Envelope{newArtifactEnvelope(env.MsgID, artifactMsgTypeAck, now, ackPayload)}, nil

	case artifactMsgTypeAck:
		if _, err := p1artifact.DecodePayloadAck(env.Payload); err != nil {
			return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid ARTIFACT ack payload: %w", err))
		}
		return nil, nil

	default:
		return nil, core.Wrap(core.CodeInvalidEnvelope, fmt.Errorf("invalid ARTIFACT msg_type %d", env.MsgType))
	}
}

func newArtifactEnvelope(msgID []byte, msgType, ts uint64, payload []byte) core.Envelope {
	return core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: ProfileSWPArtifact,
		MsgType:   msgType,
		MsgID:     append([]byte(nil), msgID...),
		TsUnixMs:  ts,
		Payload:   payload,
	}
}
