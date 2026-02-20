# SWP-ARTIFACT (Draft v0.1)

## 1. Scope

SWP-ARTIFACT defines resumable artifact transfer with integrity validation.

**Profile ID:** `14`

## 2. Message model

Supported message classes:

- `ART_OFFER`
- `ART_GET`
- `ART_CHUNK`
- `ART_ACK` (optional)
- `ART_ERR`

## 3. msg_type assignments

- `1`: `ART_OFFER`
- `2`: `ART_GET`
- `3`: `ART_CHUNK`
- `4`: `ART_ACK`
- `5`: `ART_ERR`

Any other `msg_type` is invalid for this profile version.

Payload encoding for this profile MUST support P1:

- `docs/profile-payload-encoding-p1.md`
- Normative schema annex: `proto/swp_artifact.proto`

P1 opaque-bytes fields for this profile:

- `ArtOffer.hash`
- `ArtOffer.metadata`
- `ArtChunk.data`


## 4. Artifact invariants

- content-addressable `artifact_id` is recommended.
- transfer metadata MUST include total size and hash metadata.
- chunking MUST define either byte range or index+size deterministically.
- resume MUST be possible via token or range restart rule.
- integrity MUST be verifiable at chunk and/or full-artifact level.

## 5. Behavior

- `ART_OFFER` advertises artifact metadata.
- `ART_GET` requests transfer or range.
- `ART_CHUNK` carries data segment.
- `ART_ACK` acknowledges progress when used.

## 6. Conformance requirements

A conforming implementation MUST:

- reject unsupported `msg_type` values.
- enforce chunk ordering and range validity rules.
- detect and reject integrity mismatches.
- support deterministic resume behavior.

See vectors in `conformance/vectors/catalog.md` under `artifact_*`.
