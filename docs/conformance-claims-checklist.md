# Conformance Claims Checklist

Use this checklist when asserting protocol conformance.

Conformance class taxonomy is defined in `docs/conformance-classes.md`.

## 1. SWP Core Claim

Required:
- Pass 100% of required Core vectors in `conformance/vectors/catalog.md`.
- Pass 100% of required E1 vectors (`e1_*`) in `conformance/vectors/catalog.md`.
- Implement framing and envelope validation rules from `docs/core-spec.md`.
- Implement deterministic unknown-profile handling for both:
  - response path available
  - no response path available

Evidence to provide:
- test report with vector IDs and pass/fail status
- canonical reject mappings using `expected_error_code` (`ERR_*`) per `docs/error-codes.md`
- parser/runtime configuration (`MAX_FRAME_BYTES`, `MAX_PAYLOAD_BYTES`, `MIN_MSG_ID_BYTES`, `MAX_MSG_ID_BYTES`)
- implementation version and build metadata

Recommended artifact bundle for publication:
- run:
  - `make conformance-core`
  - `make conformance-pack`
- generated artifacts:
  - `artifacts/conformance/core.default.json`
  - `artifacts/conformance/core.strict.json`
  - `artifacts/conformance/swp-conformance-bundle.tar.gz`
- include:
  - command lines used
  - one-line summary output (default + strict)
  - generated JSON artifacts

## 2. Profile Claims

Each claimed profile requires 100% pass on its vector namespace and documented deterministic behavior.

### MCP Mapping (`mcp_*`)

Required:
- JSON-RPC byte preservation in relay mode
- request/response/notification semantics
- deterministic `msg_id` reuse for response correlation

### A2A (`a2a_*`)

Required:
- task lifecycle closure semantics
- unknown-task handling
- duplicate/conflict deterministic behavior

### SWP Foundation

- SWP-RPC (`rpc_*`): lifecycle, retryability, idempotency, streaming ordering
- SWP-EVENTS (`events_*`): required fields, ordering, correlation propagation
- SWP-OBS (`obs_*`): trace context validity/preservation

### SWP Discovery

- SWP-AGDISC (`agdisc_*`): retrieval, invalid-card handling, caching semantics
- SWP-TOOLDISC (`tooldisc_*`): list/get semantics, descriptor/schema invariants

### SWP Data and Governance

- SWP-ARTIFACT (`artifact_*`): chunk integrity, resume/range behavior
- SWP-STATE (`state_*`): parent DAG validation and hash consistency
- SWP-POLICYHINT (`policyhint_*`): unknown-key/conflict/violation behavior
- SWP-CRED (`cred_*`): expiry, chain limits, invalid-credential handling
- SWP-RELAY (`relay_*`): at-least-once, dedupe, retry, dead-letter behavior

Evidence to provide (for each claimed profile):
- vector report for that profile namespace
- lifecycle/error examples from implementation logs
- policy statement for deterministic conflict/error handling

## 3. S1 Binding Claim

Required:
- Pass 100% of required S1 vectors (`s1_*`).
- Enforce channel-auth before frame processing.
- Enforce fail-closed behavior on auth/integrity/downgrade failure.

Evidence to provide:
- S1 vector report
- channel policy configuration and identity surfacing method
- security event samples for auth/integrity failures

## 4. Federation-Ready Claim

Required:
- Valid Core claim
- Valid claim for at least one profile
- Valid S1 claim

Evidence to provide:
- aggregated report containing all required vector sets
- deployment trust model summary (identity namespace + tenant boundaries)
- change-control record for profile/core/binding versions

## 5. Class Claim (`SWP Cx`)

Required:
- Declare claimed class (`C0`..`C5`) from `docs/conformance-classes.md`.
- Pass 100% of required vector namespaces for that class.
- Declare exact Core/profile/binding versions used for the claim.

Evidence to provide:
- aggregated vector report for all required namespaces in class
- class claim statement (for example `SWP C3`)
- implementation/version metadata
