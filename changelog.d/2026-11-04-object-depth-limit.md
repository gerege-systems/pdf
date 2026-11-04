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
