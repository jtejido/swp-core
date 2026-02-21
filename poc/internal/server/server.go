package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"swp-spec-kit/poc/internal/core"
	runtimecontext "swp-spec-kit/poc/internal/runtime/context"
)

const (
	ProfileMCPMap        uint64 = 1
	ProfileA2A           uint64 = 2
	ProfileSWPAGDISC     uint64 = 10
	ProfileSWPToolDisc   uint64 = 11
	ProfileSWPRPC        uint64 = 12
	ProfileSWPEvents     uint64 = 13
	ProfileSWPArtifact   uint64 = 14
	ProfileSWPCred       uint64 = 15
	ProfileSWPPolicyHint uint64 = 16
	ProfileSWPState      uint64 = 17
	ProfileSWPOBS        uint64 = 18
	ProfileSWPRelay      uint64 = 19
)

type Server struct {
	logger    *log.Logger
	limits    core.Limits
	validator core.Validator
	router    *core.Router
	runtime   runtimeBackends
}

const (
	defaultConnRateWindow         = time.Second
	defaultConnMaxFramesPerWindow = 128
	defaultDuplicateMsgIDWindow   = 5 * time.Second
)

var (
	errRateLimitExceeded = errors.New("rate limit exceeded")
	errDuplicateMsgID    = errors.New("duplicate in-flight msg_id")
)

type connPolicy struct {
	windowStart time.Time
	frameCount  int
	rateWindow  time.Duration
	maxFrames   int
	msgIDWindow time.Duration
	seenMsgID   map[string]time.Time
}

func newConnPolicy(now time.Time) *connPolicy {
	return &connPolicy{
		windowStart: now,
		rateWindow:  defaultConnRateWindow,
		maxFrames:   defaultConnMaxFramesPerWindow,
		msgIDWindow: defaultDuplicateMsgIDWindow,
		seenMsgID:   make(map[string]time.Time),
	}
}

func (p *connPolicy) check(now time.Time, msgID []byte) error {
	if now.Sub(p.windowStart) >= p.rateWindow {
		p.windowStart = now
		p.frameCount = 0
	}
	p.frameCount++
	if p.frameCount > p.maxFrames {
		return errRateLimitExceeded
	}

	threshold := now.Add(-p.msgIDWindow)
	for k, ts := range p.seenMsgID {
		if ts.Before(threshold) {
			delete(p.seenMsgID, k)
		}
	}

	key := string(msgID)
	if last, ok := p.seenMsgID[key]; ok && now.Sub(last) < p.msgIDWindow {
		return errDuplicateMsgID
	}
	p.seenMsgID[key] = now
	return nil
}

func New(logger *log.Logger, opts ...Option) *Server {
	if logger == nil {
		logger = log.Default()
	}
	runtime := newRuntimeBackends(opts...)
	limits := core.DefaultLimits()
	validator := core.DefaultValidator()
	validator.Limits = limits
	validator.KnownProfiles = map[uint64]struct{}{
		ProfileMCPMap:        {},
		ProfileA2A:           {},
		ProfileSWPAGDISC:     {},
		ProfileSWPToolDisc:   {},
		ProfileSWPRPC:        {},
		ProfileSWPEvents:     {},
		ProfileSWPArtifact:   {},
		ProfileSWPCred:       {},
		ProfileSWPPolicyHint: {},
		ProfileSWPState:      {},
		ProfileSWPOBS:        {},
		ProfileSWPRelay:      {},
	}

	s := &Server{
		logger:    logger,
		limits:    limits,
		validator: validator,
		runtime:   runtime,
	}
	router := core.NewRouter()
	router.Register(ProfileMCPMap, s.handleMCP)
	router.Register(ProfileA2A, s.handleA2A)
	router.Register(ProfileSWPAGDISC, s.handleSWPAGDISC)
	router.Register(ProfileSWPToolDisc, s.handleSWPToolDisc)
	router.Register(ProfileSWPRPC, s.handleSWPRPC)
	router.Register(ProfileSWPEvents, s.handleSWPEvents)
	router.Register(ProfileSWPArtifact, s.handleSWPArtifact)
	router.Register(ProfileSWPCred, s.handleSWPCred)
	router.Register(ProfileSWPPolicyHint, s.handleSWPPolicyHint)
	router.Register(ProfileSWPState, s.handleSWPState)
	router.Register(ProfileSWPOBS, s.handleSWPOBS)
	router.Register(ProfileSWPRelay, s.handleSWPRelay)
	s.router = router
	return s
}

func (s *Server) Serve(ctx context.Context, ln net.Listener) error {
	for {
		conn, err := ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			select {
			case <-ctx.Done():
				return nil
			default:
			}
			return fmt.Errorf("accept: %w", err)
		}
		go s.handleConn(ctx, conn)
	}
}

func (s *Server) handleConn(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	policy := newConnPolicy(time.Now())
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		frame, err := core.ReadFrame(conn, s.limits.MaxFrameBytes)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			s.logger.Printf("read frame error: %v", err)
			return
		}

		env, err := core.DecodeEnvelopeE1(frame, s.limits)
		if err != nil {
			s.logger.Printf("decode envelope error: %v", err)
			return
		}

		if err := s.validator.ValidateEnvelope(env); err != nil {
			s.logger.Printf("validate envelope error: %v", err)
			return
		}
		if err := policy.check(time.Now(), env.MsgID); err != nil {
			if errors.Is(err, errRateLimitExceeded) || errors.Is(err, errDuplicateMsgID) {
				s.logger.Printf("connection policy violation: %v", err)
			} else {
				s.logger.Printf("connection policy error: %v", err)
			}
			return
		}

		reqCtx := runtimecontext.WithMessageMeta(ctx, runtimecontext.MessageMeta{
			ProfileID: env.ProfileID,
			MsgID:     env.MsgID,
		})
		obsDoc := s.runtime.obs.GetDoc()
		reqCtx = runtimecontext.WithCorrelation(reqCtx, runtimecontext.Correlation{
			Traceparent: obsDoc.Traceparent,
			Tracestate:  obsDoc.Tracestate,
			MsgID:       obsDoc.MsgID,
			TaskID:      obsDoc.TaskID,
			RPCID:       obsDoc.RPCID,
		})

		responses, err := s.router.Dispatch(reqCtx, env)
		if err != nil {
			s.logger.Printf("dispatch error: %v", err)
			return
		}

		for _, resp := range responses {
			encoded, err := core.EncodeEnvelopeE1(resp)
			if err != nil {
				s.logger.Printf("encode response error: %v", err)
				return
			}
			if err := core.WriteFrame(conn, encoded, s.limits.MaxFrameBytes); err != nil {
				s.logger.Printf("write response error: %v", err)
				return
			}
		}
	}
}
