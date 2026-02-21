package server

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"swp-spec-kit/poc/internal/core"
	"swp-spec-kit/poc/internal/p1events"
	runtimeclock "swp-spec-kit/poc/internal/runtime/clock"
	runtimecontext "swp-spec-kit/poc/internal/runtime/context"
	runtimeerrors "swp-spec-kit/poc/internal/runtime/errors"
)

func (s *Server) emitProfileEvent(
	ctx context.Context,
	env core.Envelope,
	eventType string,
	severity string,
	body map[string]any,
	taskID []byte,
	rpcID []byte,
) {
	if s == nil {
		return
	}

	ts := runtimeclock.UnixMilli(nil)
	if body == nil {
		body = map[string]any{}
	}
	body["profile_id"] = env.ProfileID
	if code, ok := body["code"].(string); ok && code != "" {
		body["canonical_code"] = runtimeerrors.Canonical(code)
	}
	bodyBytes, _ := json.Marshal(body)

	ev := p1events.EventRecord{
		EventID:   fmt.Sprintf("%s-%d", eventType, time.Now().UnixNano()),
		EventType: eventType,
		Severity:  severity,
		TsUnixMs:  ts,
		MsgID:     append([]byte(nil), env.MsgID...),
		TaskID:    append([]byte(nil), taskID...),
		RPCID:     append([]byte(nil), rpcID...),
		Body:      bodyBytes,
	}

	if corr, ok := runtimecontext.CorrelationFromContext(ctx); ok {
		if len(ev.MsgID) == 0 && len(corr.MsgID) > 0 {
			ev.MsgID = append([]byte(nil), corr.MsgID...)
		}
		if len(ev.TaskID) == 0 && len(corr.TaskID) > 0 {
			ev.TaskID = append([]byte(nil), corr.TaskID...)
		}
		if len(ev.RPCID) == 0 && len(corr.RPCID) > 0 {
			ev.RPCID = append([]byte(nil), corr.RPCID...)
		}
	}

	if len(ev.TaskID) == 0 || len(ev.RPCID) == 0 {
		doc := s.runtime.obs.GetDoc()
		if len(ev.TaskID) == 0 && len(doc.TaskID) > 0 {
			ev.TaskID = append([]byte(nil), doc.TaskID...)
		}
		if len(ev.RPCID) == 0 && len(doc.RPCID) > 0 {
			ev.RPCID = append([]byte(nil), doc.RPCID...)
		}
	}

	if err := s.runtime.events.Publish(ev); err != nil {
		s.logger.Printf("telemetry publish failed: %v", err)
	}
}
