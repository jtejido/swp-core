package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"

	"swp-spec-kit/poc/internal/core"
)

const (
	ProfileMCPMap uint64 = 1
	ProfileSWPRPC uint64 = 12
)

type Server struct {
	logger    *log.Logger
	limits    core.Limits
	validator core.Validator
	router    *core.Router
}

func New(logger *log.Logger) *Server {
	if logger == nil {
		logger = log.Default()
	}
	limits := core.DefaultLimits()
	validator := core.DefaultValidator()
	validator.Limits = limits
	validator.KnownProfiles = map[uint64]struct{}{
		ProfileMCPMap: {},
		ProfileSWPRPC: {},
	}

	router := core.NewRouter()
	router.Register(ProfileMCPMap, handleMCP)
	router.Register(ProfileSWPRPC, handleSWPRPC)

	return &Server{
		logger:    logger,
		limits:    limits,
		validator: validator,
		router:    router,
	}
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

		responses, err := s.router.Dispatch(ctx, env)
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
