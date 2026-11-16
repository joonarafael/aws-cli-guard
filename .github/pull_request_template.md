# Pull request

## What and why

<!-- What changes and why. Link the issue: "Closes #123". -->

## User-visible behavior

<!-- What a user sees differently. "None" for refactors, tests, and docs. -->

## Checklist

<!-- Tick what applies; delete lines that don't. -->

- [ ] `make check` passes
- [ ] Tests cover the change, including that a refused command never runs the real AWS CLI
- [ ] Ambiguous cases still cancel with exit status 3
- [ ] No secrets, raw arguments, environment values, or real AWS CLI path in output or errors
- [ ] No new flag on `aws`, and no environment variable that turns policy off or moves guard files
- [ ] A changed on-disk format bumps its schema version and `aws-guard migrate` handles it
- [ ] Platform-sensitive code: `make check` also passed on macOS
- [ ] `README.md`, `docs/ARCHITECTURE.md`, or `docs/FINDINGS.md` updated where affected
