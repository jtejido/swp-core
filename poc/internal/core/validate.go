package core

import (
	"fmt"
	"time"
)

type Validator struct {
	Limits Limits

	KnownProfiles       map[uint64]struct{}
	EnforceKnownProfile bool

	EnforceTimestamp bool
	AllowZeroTS      bool
	MaxClockSkew     time.Duration
	Now              func() time.Time
}

func DefaultValidator() Validator {
	return Validator{
		Limits:              DefaultLimits(),
		EnforceKnownProfile: true,
		EnforceTimestamp:    false,
		AllowZeroTS:         true,
		MaxClockSkew:        5 * time.Minute,
		Now:                 time.Now,
	}
}

func (v Validator) ValidateEnvelope(env Envelope) error {
	if env.Version != CoreVersion {
		return Wrap(CodeUnsupportedVersion, fmt.Errorf("unsupported version %d", env.Version))
	}

	if v.EnforceKnownProfile {
		if _, ok := v.KnownProfiles[env.ProfileID]; !ok {
			return Wrap(CodeUnknownProfile, fmt.Errorf("unknown profile_id %d", env.ProfileID))
		}
	}

	if env.MsgType == 0 {
		return Wrap(CodeInvalidEnvelope, fmt.Errorf("msg_type must be non-zero"))
	}

	if len(env.MsgID) < v.Limits.MinMsgIDBytes || len(env.MsgID) > v.Limits.MaxMsgIDBytes {
		return Wrap(CodeInvalidEnvelope, fmt.Errorf("msg_id length %d not in [%d,%d]", len(env.MsgID), v.Limits.MinMsgIDBytes, v.Limits.MaxMsgIDBytes))
	}

	if uint32(len(env.Payload)) > v.Limits.MaxPayloadBytes {
		return Wrap(CodeInvalidEnvelope, fmt.Errorf("payload length %d exceeds max %d", len(env.Payload), v.Limits.MaxPayloadBytes))
	}

	if v.EnforceTimestamp {
		if env.TsUnixMs == 0 {
			if !v.AllowZeroTS {
				return Wrap(CodeInvalidEnvelope, fmt.Errorf("zero timestamp is not allowed"))
			}
		} else {
			now := v.Now()
			t := time.UnixMilli(int64(env.TsUnixMs))
			delta := now.Sub(t)
			if delta < 0 {
				delta = -delta
			}
			if delta > v.MaxClockSkew {
				return Wrap(CodeInvalidEnvelope, fmt.Errorf("timestamp skew %s exceeds max %s", delta, v.MaxClockSkew))
			}
		}
	}

	return nil
}
