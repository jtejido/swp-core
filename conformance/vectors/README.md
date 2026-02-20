
# Golden vectors

Put raw vectors under this folder.

Suggested structure:
- `vector_0001_valid_frame.bin`
- `vector_0001_expected.json`

expected.json may include:
- expected parse outcome (OK / error code)
- canonical expected reject code (`expected_error_code`, using `ERR_*` taxonomy from `docs/error-codes.md`)
- expected envelope fields (version, profile_id, msg_type, msg_id length, flags)
- expected binding behavior (error response, stream close, or accept)

Naming guidance:
- `core_####_<case>.bin`
- `core_####_<case>.json`
- `e1_####_<case>.bin`
- `e1_####_<case>.json`
- `mcp_####_<case>.bin`
- `mcp_####_<case>.json`
- `a2a_####_<case>.bin`
- `a2a_####_<case>.json`
- `rpc_####_<case>.bin`
- `events_####_<case>.bin`
- `obs_####_<case>.bin`
- `agdisc_####_<case>.bin`
- `tooldisc_####_<case>.bin`
- `artifact_####_<case>.bin`
- `state_####_<case>.bin`
- `policyhint_####_<case>.bin`
- `cred_####_<case>.bin`
- `relay_####_<case>.bin`

Start from `conformance/vectors/catalog.md` for required initial vector cases.
Template expected metadata files are pre-generated as `conformance/vectors/<vector_id>.json`.

Note:
- Most vectors are runtime wire vectors (`.bin` + expected metadata).
- Some vectors are process/spec vectors (for example compatibility policy checks) and may use metadata-only evidence.
- Runtime and mixed vectors in this repository are concrete fixtures.
- Process vectors use `.evidence.md` artifacts instead of `.bin` fixtures.
