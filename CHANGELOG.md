# Changelog

## v0.1.3 — 2026-11-04

### Changed — module path `github.com/gerege-systems/pdf`

- **What:** the digitorus/pdf v0.1.2 fork with the 2026-10-21 DoS hardening (`6255353`, `f931f72`, formerly
  `github.com/craftzbay/pdf v0.1.3-eid.2`) is published under its own module path. Code is byte-identical to
  `v0.1.3-eid.2` apart from `go.mod`, the `pdfpasswd` import and this documentation.
- **Why:** Go ignores `replace` directives of dependencies, so consumers of the e-ID Go SDK that did not copy the
  `replace` line built the unpatched upstream v0.1.2. A direct import removes the need for `replace`.
- **Migrate:** import `github.com/gerege-systems/pdf` instead of `github.com/digitorus/pdf` and delete
  `replace github.com/digitorus/pdf => github.com/craftzbay/pdf …`.
- **Verified:** `go build`, `go vet`, `go test ./...`; `git diff v0.1.3-eid.2 -- '*.go'` limited to the import path.

## v0.1.3-eid.2 — 2026-10-21 (as github.com/craftzbay/pdf)

- Stop scanning an object stream at its end when `/N` overstates its objects.

## v0.1.3-eid.1 — 2026-10-21 (as github.com/craftzbay/pdf)

- Reject unterminated arrays and strings instead of looping forever; stop following a looping object stream
  `/Extends` chain.
