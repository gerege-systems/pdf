# changelog.d — one change, one file

Do not edit `CHANGELOG.md` in a branch. Each change adds `changelog.d/<YYYY-MM-DD>-<topic>.md` with one
`### <title>` section: what · why · where (files/PR) · how it was verified.

Release (the commit before the tag): `scripts/changelog.sh assemble "v<version> — <YYYY-MM-DD>"` moves the
fragments under a new heading in `CHANGELOG.md` and deletes them.
