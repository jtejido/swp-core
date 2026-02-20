# Glossary

Canonical terms used across this repository.

## Terms

- Frame: One length-delimited unit on a stream. A frame contains one encoded envelope.
- Envelope: Core message wrapper containing routing, correlation, and payload bytes.
- Profile: A separately specified payload format identified by `profile_id`.
- Binding: A separately specified transport and/or security requirements set used with Core.
- Conformance vector: A deterministic test artifact with raw bytes and expected outcomes.
- `version`: SWP Core protocol version.
- `profile_id`: Numeric identifier that selects the payload profile.
- `msg_type`: Profile-defined message type code interpreted within `profile_id`.
- `msg_id`: Opaque correlation identifier for message-level tracking.
- `flags`: Core bitfield for extensibility and behavior hints.
- `ts_unix_ms`: Sender timestamp in Unix epoch milliseconds.
- `payload`: Opaque bytes interpreted only by the addressed profile.

## Naming convention

- Use snake_case for core envelope field names in prose and examples.
- Use `msg_id`, `msg_type`, and `profile_id` consistently; do not mix with aliases.
- Use "A2A" as the active term for agent-to-agent profile language.
