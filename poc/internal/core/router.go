package core

import (
	"context"
	"fmt"
)

type Handler func(context.Context, Envelope) ([]Envelope, error)

type Router struct {
	handlers map[uint64]Handler
}

func NewRouter() *Router {
	return &Router{handlers: make(map[uint64]Handler)}
}

func (r *Router) Register(profileID uint64, h Handler) {
	r.handlers[profileID] = h
}

func (r *Router) Dispatch(ctx context.Context, env Envelope) ([]Envelope, error) {
	h, ok := r.handlers[env.ProfileID]
	if !ok {
		return nil, Wrap(CodeUnknownProfile, fmt.Errorf("unknown profile_id %d", env.ProfileID))
	}
	return h(ctx, env)
}
