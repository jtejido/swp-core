# Spec Vector Runner Output Schema (v1)

This document defines the JSON artifact schema emitted by:

- `poc/cmd/spec-vector-runner`

## 1. Top-level fields

- `schema_version` (integer): output schema version.
- `run` (object):
  - `pattern` (string): effective vector glob pattern(s).
  - `no_fallback` (boolean): whether strict no-fallback mode was enabled.
  - `timestamp_utc` (string): UTC RFC3339 timestamp.
  - `runner_git_sha` (string): repository commit SHA, or `nogit` when unavailable.
- `total` (integer): number of vectors executed.
- `passed` (integer): number of passing results.
- `failed` (integer): number of failing results.
- `fallback_count` (integer): number of results where `used_fallback=true`.
- `results` (array): one entry per executed vector.
- `failures` (array, optional): subset of `results` entries where `pass=false`.

## 2. Result entry fields

Each `results[]` item may include:

- `vector_id` (string)
- `path` (string): vector JSON file path
- `pass` (boolean)
- `expected` (string)
- `observed` (string)
- `expected_code` (string): runtime/legacy alias code
- `observed_code` (string): runtime/legacy alias code
- `expected_error_code` (string, optional): canonical `ERR_*`
- `observed_error_code` (string, optional): canonical `ERR_*`
- `used_fallback` (boolean)
- `fallback_mode` (string, optional): `allowed` or `disallowed`
- `detail` (string, optional)

## 3. Invariants

- `total == len(results)`
- `passed + failed == total`
- `fallback_count == count(results[].used_fallback == true)`
- if `failures` is present:
  - every item in `failures` MUST also appear in `results` by `vector_id`
  - `len(failures) == failed`

## 4. Compatibility notes

- `expected_code`/`observed_code` are retained for backward compatibility.
- canonical conformance taxonomy uses `expected_error_code`/`observed_error_code` with `ERR_*` values.
- consumers SHOULD key behavior off `schema_version` for forward compatibility.

## 5. Vector discovery behavior

- When `-pattern` is provided, the runner loads vectors matching the supplied comma-separated glob list.
- When `-pattern` is omitted, the runner uses built-in spec globs for official namespaces:
  - `core_*`, `e1_*`, `s1_*`, `mcp_*`, `a2a_*`, `agdisc_*`, `tooldisc_*`, `rpc_*`, `events_*`, `artifact_*`, `cred_*`, `policyhint_*`, `state_*`, `obs_*`, `relay_*`.
- `poc_*` vectors are not part of default discovery and MUST be executed via the dedicated POC runner.
