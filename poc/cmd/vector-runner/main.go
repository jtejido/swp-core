package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"swp-spec-kit/poc/internal/core"
)

type expectation struct {
	OK            bool   `json:"ok"`
	Code          string `json:"code"`
	Version       uint64 `json:"version,omitempty"`
	ProfileID     uint64 `json:"profile_id,omitempty"`
	MsgType       uint64 `json:"msg_type,omitempty"`
	MinPayloadLen int    `json:"min_payload_len,omitempty"`
}

type vector struct {
	Name   string      `json:"name"`
	Bin    string      `json:"bin"`
	Expect expectation `json:"expect"`
}

func main() {
	pattern := flag.String("pattern", "conformance/vectors/poc_*.json", "glob for vector JSON files")
	flag.Parse()

	files, err := filepath.Glob(*pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "glob failed: %v\n", err)
		os.Exit(2)
	}
	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "no vectors matched: %s\n", *pattern)
		os.Exit(2)
	}
	sort.Strings(files)

	failures := 0
	for _, path := range files {
		if err := runOne(path); err != nil {
			failures++
			fmt.Fprintf(os.Stderr, "FAIL %s: %v\n", filepath.Base(path), err)
		} else {
			fmt.Printf("PASS %s\n", filepath.Base(path))
		}
	}

	if failures > 0 {
		fmt.Fprintf(os.Stderr, "%d vector(s) failed\n", failures)
		os.Exit(1)
	}
	fmt.Printf("all %d vectors passed\n", len(files))
}

func runOne(jsonPath string) error {
	raw, err := os.ReadFile(jsonPath)
	if err != nil {
		return err
	}
	var v vector
	if err := json.Unmarshal(raw, &v); err != nil {
		return err
	}

	binPath := filepath.Join(filepath.Dir(jsonPath), v.Bin)
	binData, err := os.ReadFile(binPath)
	if err != nil {
		return err
	}

	validator := core.DefaultValidator()
	validator.KnownProfiles = map[uint64]struct{}{1: {}, 12: {}}
	validator.EnforceKnownProfile = true

	_, code, err := decodeAndValidate(binData, validator)
	if v.Expect.OK {
		if err != nil {
			return fmt.Errorf("expected success but got %s: %v", code, err)
		}
		if code != core.CodeOK {
			return fmt.Errorf("expected code OK but got %s", code)
		}
		if v.Expect.Code != "" && v.Expect.Code != string(core.CodeOK) {
			return fmt.Errorf("invalid vector expect.code for OK case: %q", v.Expect.Code)
		}
		if _, cErr := assertEnvelope(binData, v.Expect); cErr != nil {
			return cErr
		}
		return nil
	}

	if err == nil {
		return fmt.Errorf("expected error %s but got success", v.Expect.Code)
	}
	if v.Expect.Code != "" && string(code) != v.Expect.Code {
		return fmt.Errorf("expected code %s got %s", v.Expect.Code, code)
	}
	return nil
}

func assertEnvelope(binData []byte, expect expectation) (core.Envelope, error) {
	frame, err := core.ReadFrame(bytes.NewReader(binData), core.DefaultLimits().MaxFrameBytes)
	if err != nil {
		return core.Envelope{}, err
	}
	env, err := core.DecodeEnvelopeE1(frame, core.DefaultLimits())
	if err != nil {
		return core.Envelope{}, err
	}
	if expect.Version != 0 && env.Version != expect.Version {
		return env, fmt.Errorf("version mismatch: got %d want %d", env.Version, expect.Version)
	}
	if expect.ProfileID != 0 && env.ProfileID != expect.ProfileID {
		return env, fmt.Errorf("profile_id mismatch: got %d want %d", env.ProfileID, expect.ProfileID)
	}
	if expect.MsgType != 0 && env.MsgType != expect.MsgType {
		return env, fmt.Errorf("msg_type mismatch: got %d want %d", env.MsgType, expect.MsgType)
	}
	if expect.MinPayloadLen > 0 && len(env.Payload) < expect.MinPayloadLen {
		return env, fmt.Errorf("payload too short: got %d want at least %d", len(env.Payload), expect.MinPayloadLen)
	}
	return env, nil
}

func decodeAndValidate(data []byte, validator core.Validator) (core.Envelope, core.Code, error) {
	frame, err := core.ReadFrame(bytes.NewReader(data), validator.Limits.MaxFrameBytes)
	if err != nil {
		return core.Envelope{}, core.CodeFromError(err), err
	}
	env, err := core.DecodeEnvelopeE1(frame, validator.Limits)
	if err != nil {
		return core.Envelope{}, core.CodeFromError(err), err
	}
	if err := validator.ValidateEnvelope(env); err != nil {
		return core.Envelope{}, core.CodeFromError(err), err
	}
	return env, core.CodeOK, nil
}
