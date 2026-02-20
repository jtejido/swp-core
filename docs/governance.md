# Governance and Registry Operations (Informative)

## 1. Goals

Define lightweight process for:

- profile ID allocation
- extension-type allocation for E1 TLVs
- versioning and deprecation workflow

## 2. Process model

### 2.1 Maintainer-led stage

- maintainers approve and merge spec changes
- each change SHOULD include conformance impact notes

### 2.2 Steering-stage model

Recommended evolution:

- 3-7 member steering group
- supermajority approval for breaking changes and new mandatory requirements

## 3. Registry rules

- `profile_id` values MUST NOT be reused
- E1 extension types `0-15` are reserved
- new allocations SHOULD include:
  - rationale
  - conformance vector additions
  - compatibility impact statement

## 4. Deprecation policy

- announce deprecation prior to removal
- provide replacement path and migration timeline

