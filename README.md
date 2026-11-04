# github.com/gerege-systems/pdf

```sh
go get github.com/gerege-systems/pdf
```

## Fork notice

This module is a fork of [`github.com/digitorus/pdf`](https://github.com/digitorus/pdf) **v0.1.2**
(itself forked from [`rsc.io/pdf`](https://github.com/rsc/pdf)), maintained by Gerege Systems for the
e-ID Mongolia certificate authority and its Go SDK (`github.com/gerege-systems/eid-mongolia-sdk-go`).
The original copyright and BSD-style license are kept unchanged in [LICENSE](LICENSE).

The API is identical to digitorus/pdf v0.1.2; only the import path differs. On top of v0.1.2 it adds
denial-of-service hardening for hostile input (2026-10-21):

- arrays that contain `endobj`, and arrays, hex strings and literal strings left open at the end of an
  EOF-tolerant buffer (object streams, content streams), are reported as malformed instead of looping
  forever;
- an object stream `/Extends` chain that loops back is no longer followed;
- object stream scanning stops at the end of the stream when `/N` overstates the number of objects.
- objects nested deeper than 1000 levels (arrays, dictionaries, `N G obj` definitions) are reported as
  malformed instead of exhausting the goroutine stack, which is a fatal error `recover` cannot catch
  (2026-11-04).

### Why a separate module path

The hardening was first shipped as `replace github.com/digitorus/pdf => github.com/craftzbay/pdf
v0.1.3-eid.2`. Go does not apply `replace` directives of dependencies, so programs importing the SDK
without copying that line silently built the unpatched v0.1.2. Publishing the fork under its own path
lets the CA and the SDK import it directly, so every consumer gets the fix with no `replace`.

Upstream digitorus/pdf v0.3.0 fixes the same inputs but was not adopted: its new filter-chain decoding
of object streams bypasses our zip-bomb pre-check, and its unbounded object cache raised peak heap on a
large legitimate PDF from 41 MiB to 435 MiB. The `github.com/digitorus/pdfsign` revision we test against also
still targets the v0.1.x API.

`go 1.17` and the exported API stay as in v0.1.2; parser behaviour for well-formed documents is unchanged.

---

This project is forked from rsc/pdf.
