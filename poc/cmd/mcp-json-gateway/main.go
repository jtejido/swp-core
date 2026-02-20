package main

import (
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"swp-spec-kit/poc/internal/core"
)

const profileMCPMap = 1

func main() {
	listen := flag.String("listen", ":8080", "HTTP listen address")
	swpAddr := flag.String("swp", "127.0.0.1:7777", "SWP TCP address")
	flag.Parse()

	h := &handler{swpAddr: *swpAddr}
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", h.handleMCP)

	log.Printf("mcp-json-gateway listening on %s (swp=%s)", *listen, *swpAddr)
	if err := http.ListenAndServe(*listen, mux); err != nil {
		log.Fatalf("http serve: %v", err)
	}
}

type handler struct {
	swpAddr string
	mu      sync.Mutex
	conn    net.Conn
}

func (h *handler) handleMCP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(r.Body, 2*1024*1024))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	msgType := uint64(1)
	if isNotification(payload) {
		msgType = 3
	}

	msgID, err := newMsgID(16)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	env := core.Envelope{
		Version:   core.CoreVersion,
		ProfileID: profileMCPMap,
		MsgType:   msgType,
		MsgID:     msgID,
		TsUnixMs:  uint64(time.Now().UnixMilli()),
		Payload:   payload,
	}

	respPayload, err := h.send(env, msgType != 3)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	if msgType == 3 {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"ok":true}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(respPayload)
}

func (h *handler) send(env core.Envelope, wantResponse bool) ([]byte, error) {
	body, err := core.EncodeEnvelopeE1(env)
	if err != nil {
		return nil, fmt.Errorf("encode envelope: %w", err)
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		conn, err := h.ensureConnLocked()
		if err != nil {
			return nil, fmt.Errorf("dial swp: %w", err)
		}

		if err := conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
			lastErr = fmt.Errorf("set write deadline: %w", err)
			h.resetConnLocked()
			continue
		}
		if err := core.WriteFrame(conn, body, core.DefaultLimits().MaxFrameBytes); err != nil {
			lastErr = fmt.Errorf("write frame: %w", err)
			h.resetConnLocked()
			continue
		}
		if !wantResponse {
			return nil, nil
		}

		if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
			lastErr = fmt.Errorf("set read deadline: %w", err)
			h.resetConnLocked()
			continue
		}
		respFrame, err := core.ReadFrame(conn, core.DefaultLimits().MaxFrameBytes)
		if err != nil {
			lastErr = fmt.Errorf("read frame: %w", err)
			h.resetConnLocked()
			continue
		}
		resp, err := core.DecodeEnvelopeE1(respFrame, core.DefaultLimits())
		if err != nil {
			lastErr = fmt.Errorf("decode response: %w", err)
			h.resetConnLocked()
			continue
		}
		return resp.Payload, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("request failed")
	}
	return nil, lastErr
}

func (h *handler) ensureConnLocked() (net.Conn, error) {
	if h.conn != nil {
		return h.conn, nil
	}
	conn, err := net.Dial("tcp", h.swpAddr)
	if err != nil {
		return nil, err
	}
	h.conn = conn
	return h.conn, nil
}

func (h *handler) resetConnLocked() {
	if h.conn != nil {
		_ = h.conn.Close()
		h.conn = nil
	}
}

func (h *handler) close() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.resetConnLocked()
}

func isNotification(payload []byte) bool {
	var msg map[string]any
	if err := json.Unmarshal(payload, &msg); err != nil {
		return false
	}
	_, hasID := msg["id"]
	return !hasID
}

func newMsgID(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}
