# Changelog

## v0.1.4 — 2026-11-04

### Security — bound object nesting depth (fatal stack overflow)

- **What:** `readObject` now panics with "objects nested deeper than 1000 levels" (a normal parse error that
  callers already `recover`) once arrays, dictionaries or nested `N G obj` definitions exceed `maxObjectDepth`
  = 1000 levels. `lex.go`, tests in `depth_test.go`.
- **Why:** the recursion was unbounded. A few megabytes of `[[[[…` or `1 0 obj 1 0 obj …` — compressed to a few
  kilobytes inside a FlateDecode object stream — overflowed the goroutine stack (1 GB by default): `fatal error:
  stack overflow`, which kills the whole process and cannot be recovered. The e-ID CA and Go SDK pre-scan for
  deep nesting, but that scan missed objects reached through a non-integer object stream offset (read as 0 here)
  and nested `N G obj` chains.
- **Verified:** `go test ./...` (new tests fail with `fatal error: stack overflow` under a 32 MiB stack without
  the limit), `go vet`, `GOOS=linux GOARCH=amd64 go build ./...`; e-ID CA corpus replay (683 PDFs × 2
  profiles) byte-identical.

## v0.1.3 — 2026-11-04

### Changed — module path `github.com/gerege-systems/pdf`

- **What:** the digitorus/pdf v0.1.2 fork with the 2026-10-21 DoS hardening (`6255353`, `f931f72`, formerly
  `github.com/craftzbay/pdf v0.1.3-eid.2`) is published under its own module path. Code is byte-identical to
  `v0.1.3-eid.2` (commit `f931f72`) apart from `go.mod`, the `pdfpasswd` import, documentation and
  `scripts/changelog.sh`.
- **Why:** Go ignores `replace` directives of dependencies, so consumers of the e-ID Go SDK that did not copy the
  `replace` line built the unpatched upstream v0.1.2. A direct import removes the need for `replace`.
- **Migrate:** import `github.com/gerege-systems/pdf` instead of `github.com/digitorus/pdf` and delete
  `replace github.com/digitorus/pdf => github.com/craftzbay/pdf …`.
- **Verified:** `go build`, `go vet`, `go test ./...`; `git diff f931f72 -- '*.go'` limited to the import path.

## v0.1.3-eid.2 — 2026-10-21 (as github.com/craftzbay/pdf)

- Stop scanning an object stream at its end when `/N` overstates its objects.

## v0.1.3-eid.1 — 2026-10-21 (as github.com/craftzbay/pdf)

- Reject unterminated arrays and strings instead of looping forever; stop following a looping object stream
  `/Extends` chain.
