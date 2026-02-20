# Versioning and Compatibility Policy

This policy applies to SWP Core, profiles, and bindings.

## 1. Version domains

Independent version domains:
- Core protocol version (`version` in envelope)
- Profile version (profile specification release)
- Binding version (security/transport binding specification release)

A change in one domain does not require version increment in another unless interoperability semantics are impacted.

## 2. Change classification

### 2.1 Backward-compatible changes

Examples:
- additive optional fields
- additive enum/message type values with safe unknown handling
- additional non-normative guidance and examples

Compatibility impact:
- existing compliant implementations continue operating without behavior breakage.

### 2.2 Backward-incompatible changes

Examples:
- semantic reinterpretation of existing field values
- removal of required fields or required message classes
- changed correlation or lifecycle semantics

Compatibility impact:
- requires major profile/core version change and coordinated rollout plan.

## 3. Core compatibility rules

- Core `version` MUST be incremented when backward-incompatible Core behavior is introduced.
- Core `version` SHOULD remain unchanged for editorial and non-semantic updates.
- Receivers MUST reject unsupported Core versions deterministically.
- Implementations MUST NOT perform implicit downgrade.
- If version negotiation is implemented by a binding or transport profile, negotiation MUST be explicit and authenticated by the active security binding.
- Unknown extension elements defined as ignorable by a binding/profile MUST be ignored without reinterpretation of known fields.
- Unknown extension elements still MUST respect declared size/parse limits.

Unknown profile/message handling:
- Unknown `profile_id` MUST be rejected deterministically when no explicit ignore behavior is defined by the active binding/profile.
- Unknown `msg_type` values MUST be rejected by the owning profile unless that profile explicitly defines ignore/forward-compat handling.

## 4. Profile compatibility rules

- Profiles SHOULD prefer additive evolution where possible.
- New required profile semantics that break prior behavior MUST trigger a major profile revision.
- If profile evolution cannot be represented safely with existing `profile_id`, allocate a new `profile_id`.

## 5. Binding compatibility rules

- Binding changes that weaken required guarantees are breaking changes.
- Binding changes that tighten optional operational guidance without semantic breaks are non-breaking.

## 6. Deprecation and removal

- Mark deprecated behavior in one published release before removal.
- Deprecation notice SHOULD include:
  - affected feature
  - replacement behavior
  - removal timeline

## 7. Compatibility matrix guidance

Producer/consumer compatibility should be documented per release:
- Core version support set
- Profile version support set
- Binding version support set

Recommended matrix fields:
- implementation version
- supported core versions
- supported profile IDs and versions
- supported bindings and options

## 8. Registry governance

`docs/profile-registry.md` remains authoritative for `profile_id` allocation.

Registry operations:
- new profile allocation
- deprecation status updates
- provisional to standards-track promotion
