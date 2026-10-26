# ADR 0003: Wrapper cancellations exit 3

## Status

Accepted

## Context

[User Experience](../DESIGN.md#9-user-experience) requires one exit status for wrapper refusals, a missing policy, a corrupt policy, and corrupt grants. That status must not be one of the AWS CLI's documented codes: 0, 1, 2, 130, 252, 253, 254, or 255. Shells also use 126 ("found but not executable"), 127 ("not found"), and 128 plus a signal number.

Allowed commands replace the wrapper with the real AWS CLI, so those exit codes stay the CLI's codes. The gap is which single code to use when the wrapper itself cancels the command, and how a caller tells that cancellation from a CLI error that happens to be non-zero.

Status 3 carries no standard name that means "permission denied" or "bad configuration." One code has to cover every cancellation, so a name that fits only one of those cases would mislead. The stderr line is what makes the source clear.

## Decision

1. **Status.** When AWS CLI Guard cancels a command, it exits 3. This includes policy refusals, a missing policy, corrupt policy, corrupt grants, corrupt settings, and a failure to find or start the real AWS CLI.

2. **Error line.** Every such cancellation writes the reason to stderr. Each line starts with `aws-cli-guard: error:`. The reason says the command was not executed and that AWS CLI Guard is what stopped it. The line is single-line, with control characters stripped, and it follows the secret-handling rules.

3. **Other wrapper lines.** A wrapper message that does not cancel the command uses `aws-cli-guard: warning:` when it is a warning, and `aws-cli-guard:` when it is other wrapper text. The `default` profile warning uses the warning prefix.

4. **After handoff.** Once the wrapper has been replaced by the real AWS CLI, that program's stdout and stderr are unchanged. The wrapper cannot prefix them. The same is true of a crash the wrapper did not catch.

5. **Warnings are not status 3.** A version warning from [ADR 0002](0002-aws-cli-version-floor.md) does not by itself exit 3. An older CLI is handed off, and the caller sees the CLI's status. A newer CLI continues into policy, which may still cancel with status 3.

## Consequences

- Scripts can treat exit 3 as "AWS CLI Guard cancelled this" and treat every other non-zero status as the AWS CLI, a signal, or the shell.
- Exit 3 will still collide with any other program that picked 3. The prefixed error line is the disambiguator, and it is part of the contract, not an optional flourish.
- The missing-CLI path in `cmd/aws` exits 1 today. The exit-status epic changes that path to 3 and adds the prefix.
- Longer guidance, such as how to allow a profile, is more wrapper lines with the same `aws-cli-guard:` prefix. It stays free of secrets, raw argument vectors, a disabling environment variable, and the path of the real AWS CLI.
