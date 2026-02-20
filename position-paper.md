# SWP (SWP Protocol): A Minimal Envelope + Profile Substrate for Federated Agent/Tool Systems

## Executive summary

Modern AI platform teams are converging on **agents** (long-running, stateful task executors) and **tool fabrics** (databases, search, code execution, business APIs). In practice, these systems are assembled from heterogeneous components: multiple agent runtimes, multiple tool protocols, multiple hosting models, and cross-organization federation constraints. The result is an integration surface dominated by **bespoke adapters**, inconsistent security/observability, and hard-to-test behaviors.

**SWP (SWP Protocol)** is a *minimal*, standards-oriented substrate that standardizes **only what must be common** to make interoperability and federation tractable:
- a compact **binary frame + envelope** for message boundaries, correlation, dispatch, and limits
- a registry of **profiles** (semantics) that sit on top of the envelope
- **conformance classes and vectors** to make “compatible” verifiable
- a baseline **security binding** expectation (authenticated confidential channel)

SWP deliberately does **not** attempt to replace existing ecosystems (e.g., MCP for tool invocation or A2A for agent communication). Instead, it provides a stable envelope and dispatch layer that can **bridge** them, while enabling new profiles for common infrastructure primitives (RPC, events, artifact transfer, relay, etc.).

The goal is to solve the coordination problem: **how do independently-developed agent/tool systems communicate across boundaries with enforceable compatibility, consistent governance hooks, and predictable operational behavior?**

---

## The problem: interoperability is failing in the field

### 1) Message boundaries and semantics are being reinvented
Across agent stacks, message transport frequently devolves into ad hoc combinations of:
- JSON over HTTP
- custom WebSocket/SSE streaming
- framework-specific RPC and event schemas
- one-off “artifact upload” APIs with inconsistent resumability and integrity guarantees

Teams repeatedly re-implement the same “plumbing”:
- message framing
- correlation IDs and lifecycle tracking
- streaming partial results
- cancellation
- retries and at-least-once delivery patterns
- size and rate enforcement
- audit/trace propagation

### 2) Federation multiplies complexity
Once an agent/tool system crosses *organizational* or *tenant* boundaries, requirements harden:
- stable peer identity surface for authorization decisions
- predictable limits to prevent abuse (frame/payload caps)
- consistent observability hooks (trace correlation)
- policy constraints (data residency, PII restrictions, budgets)
- artifact transfer with integrity and resumability

Without a shared substrate, these requirements are repeatedly “bolted on,” producing brittle, divergent integrations.

### 3) Lack of verifiable compatibility
Even when two parties claim “we implement protocol X,” practical interop failures occur because:
- edge cases are unspecified (oversized frames, unknown fields, timing skew)
- versioning expectations are unclear
- error semantics vary
- payload encoding differs (JSON vs protobuf vs bespoke)

Absent **golden vectors** and conformance claims, “compatible” is a guess.

---

## Why existing options don’t solve this alone

### MCP and A2A are necessary, but they address different layers
- **MCP** focuses on tool invocation semantics and tool discovery for model-to-tool interaction.
- **A2A** focuses on agent-to-agent interaction semantics (tasks, streaming, artifacts).

Both are early and evolving, and neither is primarily a general-purpose *federation substrate* with a profile registry and conformance class model spanning multiple ecosystems.

### gRPC/HTTP are transports, not interoperability contracts
gRPC and HTTP/2 are excellent transports and already provide framing at their layer. However, “use HTTP/2” does not standardize:
- a common envelope for correlation, dispatch, and limits across agent/tool semantics
- profile identity and registries
- consistent cross-ecosystem mapping (e.g., MCP↔A2A bridging)
- conformance vectors and conformance classes for verifiable claims

SWP is designed to sit above transports: it can run on top of HTTP/2, QUIC, raw TCP streams, etc., while preserving a stable message envelope and profile dispatch model.

---

## The SWP bet: standardize the minimum that must be common

SWP is explicitly a **protocol family**:
- **Core**: framing + envelope + dispatch + limits
- **Payload binding(s)**: how profile payload bytes are encoded (e.g., P1 protobuf)
- **Profiles**: semantics for specific problem domains
- **Bindings**: transport/security expectations

### SWP Core (SWP Core)
SWP Core answers: *where does one message end and the next begin, and how do we route it?*
- **Framing**: length-prefixed frames on a byte stream (message boundaries)
- **Envelope**: minimal common header fields:
  - `version`
  - `profile_id` (dispatch key)
  - `msg_type` (profile-local type)
  - `msg_id` (correlation)
  - `flags`
  - `ts_unix_ms`
  - `payload` (opaque to core)

Core defines mandatory parsing limits and rejection behavior to make implementations robust by default.

### E1 (mandatory envelope encoding)
To avoid “compatible but not interoperable,” SWP defines at least one mandatory-to-implement envelope encoding:
- compact varint + bytes encoding
- fixed field order for deterministic parsing
- extension TLV block for forward compatibility

### P1 (profile payload encoding binding)
SWP profiles must not be “prose-only.” SWP therefore defines a mandatory payload encoding for profiles:
- **P1: protobuf (proto3)** for profile payload bytes
- normative `.proto` annexes per profile
- forward-compatibility via unknown field preservation/ignoring rules

### Security binding (S1 baseline)
SWP does not invent cryptography. It assumes deployments choose a secure channel binding, with **S1** as the baseline expectation:
- authenticated confidential channel
- stable peer identity surfaced to profiles (authorization remains out of scope)

---

## Profiles: infrastructure primitives that agent systems actually need

SWP profiles are intentionally “boring” in the best way: common primitives that appear repeatedly in real deployments.

Examples include:
- **AGDISC**: agent discovery (agent card, capabilities, endpoints)
- **TOOLDISC**: tool discovery (tool descriptors + schema refs + hints)
- **RPC**: generic request/response with streaming and idempotency
- **EVENTS**: event stream records (progress/log/audit) with correlation
- **ARTIFACT**: artifact transfer with hashing, chunking, resumability
- **STATE**: immutable, content-addressed state blobs (optional DAG)
- **OBS**: observability context propagation (trace correlation)
- **RELAY**: store-and-forward delivery patterns (at-least-once + dedupe)
- **POLICYHINT**: portable constraints and violation signaling
- **CRED**: credential/delegation carrier (federation-heavy; optional)

Crucially, SWP can also define compatibility profiles that bridge existing ecosystems:
- **MCPMAP**: carry MCP JSON-RPC bytes opaquely without semantic reinterpretation
- **A2A**: align to A2A semantics while using SWP’s envelope, encoding, and conformance framework

---

## What SWP deliberately excludes

To remain standardizable and adoptable, SWP explicitly avoids:
- **A new IAM system** (authentication/authorization stays in channel bindings and platform policy)
- **A new policy language** (POLICYHINT is key/value constraints, not a full DSL)
- **A mandated transport** (HTTP/2, QUIC, TCP, WebSocket can be used via transport bindings)
- **Framework-specific agent semantics** (profiles define interop primitives, not “your agent framework”)
- **Replacing MCP or A2A** (SWP is a substrate that can bridge them)

---

## Conformance: make compatibility verifiable, not aspirational

SWP treats conformance as a first-class deliverable:
- **Conformance classes** define what a claim means (e.g., “Core+E1+S1”, “Core+MCPMAP”, “Core+RPC+EVENTS”).
- **Golden vectors** provide byte-level fixtures and expected outcomes for accept/reject and decoding behavior.
- Conformance claims become auditable (“passes C1 vectors”) rather than interpretive.

This is the critical difference between a “spec people quote” and a “standard people rely on.”

---

## Adoption path: what happens next

SWP succeeds through a wedge that reduces real integration cost today.

### Phase 1: Gateway and bridges (practical wedge)
- Deploy an **SWP Gateway** that speaks:
  - MCP on one side (JSON-RPC tool calls)
  - SWP (Core+E1) internally, dispatching MCPMAP payloads
- Demonstrate federated policy enforcement points (limits, identity surface, correlation) at the gateway.

### Phase 2: Second independent implementation + conformance badge
- Produce a second implementation (different language/team)
- Both implementations pass the same conformance classes (e.g., C0 + MCPMAP)
- Publish a “Conformant” badge tied to vector runs

### Phase 3: Expand to foundational SWP profiles
- Add SWP-RPC, SWP-EVENTS, SWP-ARTIFACT as the practical “infrastructure trio”
- Integrate OBS context propagation for end-to-end traceability

### Phase 4: Federation-heavy extensions (as demanded)
- POLICYHINT and CRED only when federation requirements are concrete
- RELAY where offline/intermittent agents require store-and-forward

---

## Success criteria

SWP is successful when:
1. **Two independent implementations** interoperate using Core+E1 and pass a published conformance class.
2. A real organization deploys an **SWP Gateway** to bridge at least two heterogeneous agent/tool ecosystems.
3. A profile registry process exists with clear rules (allocation, deprecation, compatibility).
4. The conformance suite is used as a gating artifact (CI, release criteria, vendor claims).

---

## Governance recommendation (lightweight, credible)

Start with a maintainer-led model (fast iteration), then evolve:
- Stage 1: maintainers manage changes, requiring spec diffs + conformance impact notes
- Stage 2: small steering group (users + implementers) governs:
  - breaking changes
  - new mandatory requirements
  - profile_id allocations

Registries (profile IDs, extension types) should be explicit, stable, and non-reused.

---

## Appendix: where the technical artifacts live

This position paper intentionally does not restate the full spec. The technical artifacts are:
- SWP Core specification and E1 encoding binding
- P1 payload encoding binding (protobuf) and normative proto annexes
- Profile specifications (AGDISC, TOOLDISC, RPC, EVENTS, ARTIFACT, …)
- Conformance classes and golden vectors
- Architecture and sequence diagrams

These artifacts are the authoritative source for implementation and verification.
