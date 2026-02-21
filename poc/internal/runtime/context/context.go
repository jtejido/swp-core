package context

import (
	"context"
)

type MessageMeta struct {
	ProfileID uint64
	MsgID     []byte
}

type Correlation struct {
	Traceparent string
	Tracestate  string
	MsgID       []byte
	TaskID      []byte
	RPCID       []byte
}

type messageMetaKey struct{}
type correlationKey struct{}

func WithMessageMeta(ctx context.Context, m MessageMeta) context.Context {
	return context.WithValue(ctx, messageMetaKey{}, MessageMeta{
		ProfileID: m.ProfileID,
		MsgID:     append([]byte(nil), m.MsgID...),
	})
}

func MessageMetaFromContext(ctx context.Context) (MessageMeta, bool) {
	v := ctx.Value(messageMetaKey{})
	meta, ok := v.(MessageMeta)
	if !ok {
		return MessageMeta{}, false
	}
	meta.MsgID = append([]byte(nil), meta.MsgID...)
	return meta, true
}

func WithCorrelation(ctx context.Context, c Correlation) context.Context {
	return context.WithValue(ctx, correlationKey{}, Correlation{
		Traceparent: c.Traceparent,
		Tracestate:  c.Tracestate,
		MsgID:       append([]byte(nil), c.MsgID...),
		TaskID:      append([]byte(nil), c.TaskID...),
		RPCID:       append([]byte(nil), c.RPCID...),
	})
}

func CorrelationFromContext(ctx context.Context) (Correlation, bool) {
	v := ctx.Value(correlationKey{})
	c, ok := v.(Correlation)
	if !ok {
		return Correlation{}, false
	}
	c.MsgID = append([]byte(nil), c.MsgID...)
	c.TaskID = append([]byte(nil), c.TaskID...)
	c.RPCID = append([]byte(nil), c.RPCID...)
	return c, true
}
