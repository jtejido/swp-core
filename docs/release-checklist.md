# SWP Release Checklist

Use this checklist before publishing a GitHub release and Zenodo DOI snapshot.

## 1. Pre-release validation

Run local gates from repository root:

```bash
make ci
make conformance-pack
```

Expected artifacts:

- `artifacts/conformance/core.default.json`
- `artifacts/conformance/core.strict.json`
- `artifacts/conformance/all.default.json`
- `artifacts/conformance/all.strict.json`
- `artifacts/conformance/swp-conformance-bundle.tar.gz`

## 2. Verify repository metadata

- `README.md` title, scope, and links are current.
- `.zenodo.json` metadata matches the release title and maintainers.
- `docs/publication-artifact-index.md` references current canonical docs.

## 3. Create GitHub release

1. Tag and push release:

```bash
git tag vX.Y.Z
git push origin vX.Y.Z
```

2. Create GitHub Release for `vX.Y.Z`.
3. Upload `artifacts/conformance/swp-conformance-bundle.tar.gz` as a release asset.
4. Include in release notes:
   - Core default summary line
   - Core strict summary line
   - Any known fallback-dependent vectors (if any)

## 4. Publish Zenodo DOI

If Zenodo GitHub integration is enabled:

1. Wait for Zenodo to ingest the GitHub release.
2. Open the new Zenodo draft record.
3. Verify title, authors, keywords, and description.
4. Publish the record and capture the DOI.

If uploading manually to Zenodo:

1. Create a new software upload.
2. Upload source archive (or release tarball) and conformance bundle.
3. Apply metadata from `.zenodo.json`.
4. Publish and capture DOI.

## 5. Post-release updates

- Update DOI badge/link in `README.md` if changed.
- Add DOI and release tag to publication notes/appendix text.
- Archive generated conformance JSON/log artifacts for reproducibility.
