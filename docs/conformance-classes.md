# Conformance Classes (Draft)

This document defines shorthand conformance classes so implementers can make scoped, comparable claims.

## Class definitions

- `C0` (Core Baseline): Core + E1 + S1
- `C1` (MCP Bridge): C0 + MCP Mapping profile
- `C2` (A2A Baseline): C0 + A2A profile
- `C3` (SWP Runtime): C0 + SWP-RPC + SWP-EVENTS + SWP-OBS
- `C4` (SWP Data Plane): C3 + SWP-ARTIFACT + SWP-STATE
- `C5` (SWP Federation): C4 + SWP-AGDISC + SWP-TOOLDISC + SWP-CRED + SWP-POLICYHINT + SWP-RELAY

## Required vector namespaces per class

- `C0`: `core_*`, `e1_*`, `s1_*`
- `C1`: C0 + `mcp_*`
- `C2`: C0 + `a2a_*`
- `C3`: C0 + `rpc_*`, `events_*`, `obs_*`
- `C4`: C3 + `artifact_*`, `state_*`
- `C5`: C4 + `agdisc_*`, `tooldisc_*`, `cred_*`, `policyhint_*`, `relay_*`

## Claim format

Implementations SHOULD publish claims in this form:

- `SWP Cx` (example: `SWP C1`)
- optional additive profile claims beyond class baseline
- version references for Core, bindings, and profiles

## Notes

- Classes are cumulative.
- A valid class claim requires 100% pass for all required vector namespaces in that class.
- Classes do not prevent publishing narrower claims outside class taxonomy.
