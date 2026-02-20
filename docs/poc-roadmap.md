# POC Roadmap (2-3 Weeks)

## Week 1

- Implement Core E1 framing/codec/validator/router in Go.
- Implement MCP Mapping profile (`profile_id=1`) request/response.
- Add unit tests for framing, codec, and validation.

## Week 2

- Implement SWP-RPC profile (`profile_id=12`) request/response and streaming items.
- Add integration test for MCP flow + SWP-RPC streaming flow.
- Add vector generator and vector runner using `.json` + real `.bin` fixtures.

## Week 3 (Hardening)

- Add mcp JSON gateway endpoint for external clients.
- Add podman compose workflow and make targets.
- Add retry/timeout tuning and richer negative vectors.
