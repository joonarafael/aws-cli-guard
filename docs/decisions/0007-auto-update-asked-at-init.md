# ADR 0007: Auto-update is chosen during init

## Status

Accepted

## Context

[Auto-update and update notices](../DESIGN.md#124-auto-update-and-update-notices) left the default open. The rest of that behavior is already specified: a setting turns updates on or off, an off setting still suggests a newer release on stderr, an on setting downloads and verifies a release, and the check does not change the AWS command's stdout or exit status, does not stall it, and does not send credentials or command arguments.

`aws-guard init` is the moment the user is setting the tool up. That is the right time to ask. [ADR 0006](0006-identity-verification-off-by-default.md) adds `settings.toml` for preferences that are not policy.

## Decision

1. **Init asks.** `aws-guard init` asks whether automatic updates should be on. The answer is written to `[updates] auto` in `settings.toml`.

2. **Unanswered means off.** When the key is missing, automatic updates stay off. The existing suggestion for a newer release still applies. Init asks when it creates the settings file, including when the policy file already exists and init refuses to replace that policy file.

3. **The question is explicit.** There is no silent default of on. A build that does not yet contain the updater does not ask and does not download. When the updater ships, init asks if the answer is not already stored.

## Consequences

- A machine does not start replacing its binary until the person who ran init says so.
- Someone who initialized an older build is asked the next time init runs and the answer is absent. Init still must not overwrite `config.toml`.
- `OP-4` implements the prompt, the download, and the suggestion. It reads and writes the settings file from ADR 0006.
