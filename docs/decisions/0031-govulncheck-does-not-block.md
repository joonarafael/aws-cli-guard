# ADR 0031: govulncheck does not block pull requests

## Status

Accepted

## Context

CI runs [govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck) on the Go version pinned in `go.mod` and in `.github/workflows/ci.yml`. That scan reports vulnerabilities in the standard library of the pinned toolchain, and the pinned scanner can throw when that toolchain is newer than the scanner's analysis libraries.

Those results are worth seeing. They are not a reason to hold a pull request. The findings move when the Go pin or the scanner pin moves, and a crash is a scanner failure rather than a defect in the change under review.

## Decision

1. **Still run the scan.** The Source Code Hygiene job runs the pinned govulncheck against `./...` on the pinned Go version.

2. **Do not fail the job.** A non-zero exit, including vulnerabilities found (govulncheck status 3, which `go run` surfaces as exit 1) and a scanner crash, does not fail Source Code Hygiene. A pull request can merge while the scan is red or the scanner throws.

3. **Report both places.** The job summary and a sticky pull-request comment carry the same report. The comment is replaced in place on the next run, including a clean run. The report names the pinned Go version and the pinned govulncheck version, lists called vulnerabilities without their example traces, and says that the check does not block merging. A crash quotes the start of the scanner error instead of a finding list.

4. **Leave the log.** The step output keeps the full scanner log. The comment links to that Actions run.

## Consequences

- Known standard-library findings on the pinned toolchain do not stop a merge. The sticky comment is what reviewers see.
- A scanner crash on the pinned Go version does not stop a merge either. The comment says the scanner threw.
- Clearing the report still means bumping the Go pin, the govulncheck pin, or both. This record does not choose either bump.
