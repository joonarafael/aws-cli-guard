# ADR 0030: Auto-update and update notices

## Status

Accepted

## Context

[ADR 0023](0023-product-finish-before-agent-detection.md) schedules `OP-4` after verbose logging and before Bash and zsh integration. [ADR 0007](0007-auto-update-asked-at-init.md) decides that `aws-guard init` asks, that a missing answer means off, and that a build without the updater does not ask and does not download. [Auto-update and update notices](../DESIGN.md#124-auto-update-and-update-notices) require a suggestion when updates are off, a checksum-verified download and replace when they are on, an unchanged AWS stdout and exit status, no stall, and a check that contacts only the release channel.

[ADR 0025](0025-installer-and-releases.md) fixes the asset names, `checksums.txt`, and the rule that a floating latest URL is not the install path. Published GitHub Releases omit drafts. [ADR 0026](0026-upgrade-behavior.md) replaces the four install files and then migrates, and it never writes `real_aws`. The wrapper then replaces its process with the real AWS CLI, so it cannot finish a download after the handoff. A download before the handoff would stall the command.

`[updates]` is still an unknown settings field ([ADR 0029](0029-verbose-logging.md)). Init does not ask.

## Decision

1. **This record does not implement `OP-4`.** Until the implementation lands, the settings decoder keeps rejecting `[updates]`, and init does not ask and does not download. There is no public `aws-guard update` command.

2. **Settings.** Schema version stays 1. Optional `[updates]` in `settings.toml`:

```toml
[updates]
auto = false
```

A present table requires boolean `auto`. `true` and `false` are both valid. A missing `auto` is `settings.KindInvalid` on key `updates.auto`. An absent table, an absent key, and a missing file all mean off. Unknown fields still fail closed. `[identity]` and `[logging]` are unchanged. `settings.Recover` keeps writing the current default body and does not write `[updates]`. The load result distinguishes absent, `true`, and `false`. Absent and `false` both leave the updater off. Only an absent answer makes init ask.

3. **Init asks before it writes.** When the answer is already stored, init keeps today's result: both files present exits 3 with `the policy file already exists`, and a missing settings file is created without a second question. When the file is present and unusable, init exits 3 with the existing settings diagnosis, writes nothing, and does not ask.

When the answer is absent and a controlling terminal can be opened, init reads one line from that device, the same device as `terminal.Confirm`, not from stdin. The prompt is `automatic updates? type on or off`. The line `on` stores `true`. The line `off` stores `false`. Any other line exits 3 with `the update choice was not confirmed` and writes nothing, including no new policy file.

When the answer is absent and no controlling terminal can be opened, init does not exit 3 and does not write the key. On the success path it prints `automatic updates stay off` on stdout.

A new settings file after an answer is the current default body plus `[updates]`. An existing usable file is rewritten in place, preserving `version`, `[identity]`, and `[logging]`, and adding `[updates]`. Comments are not kept. The write is atomic, mode `0600`, and refuses a symlink. Policy bytes stay unchanged when the policy file already exists. Success text uses `updated settings file` when an existing settings file gained the key, and the existing created or left-unchanged lines otherwise.

4. **Which `aws` invocations participate.** Only a handoff that has already loaded settings successfully and is about to exec the real CLI: an allow, including a prompt with a current grant, and session off. Login, the version handoff, discovery failure, and every cancellation skip the check. Those paths do not load settings for this feature and do not spawn a helper. The verbose decision block gains no update field. Update lines are ordinary tight stderr, written after the existing tight lines and before the verbose block.

5. **The `aws` path does not contact the network and does not wait.** It may read `updates.json`, print one cached notice, and start a helper. It does not wait for the helper, the network, or a lock. It does not call the blocking `files.OpenLock`. To clear a notice it takes a non-blocking exclusive flock (`LOCK_EX|LOCK_NB`) on `updates.json.lock`. If that lock is held or cannot be opened, it skips the notice and still execs. Spawn does not hold the lock. A spawn that fails to start is ignored. Stdout and the exit status stay the AWS command's. The command that starts a check is not the command that reports the result.

6. **When the wrapper skips the check.** Skip, and still exec, when any of these hold:

- `version.Version` is not `X.Y.Z` (`^[0-9]+\.[0-9]+\.[0-9]+$`). A `dev` stamp never checks, suggests, or replaces.
- The directory of `os.Executable` is not a verified installation under the [ADR 0026](0026-upgrade-behavior.md) rules, or the running file is not the same file as the manifest `aws` path. No other prefix is searched. `--prefix` is not accepted.
- The cache path cannot be resolved.
- `updates.json` is present but unusable. The wrapper prints the same warning as an unusable `cli-version.json`, naming `updates.json`, and does not rewrite the file. `aws-guard recover cache` remains the way to replace it.

7. **Cache.** One file, `updates.json`, an immediate child of the cache directory ([On-disk files](../DESIGN.md#on-disk-files)). Mode `0600`, directory `0700`, atomic write, symlink refusal, and a 1 MiB cap, matching `cli-version.json`. Strict JSON, version 1. Unknown fields and duplicate keys are corruption. The `aws` path does not create the cache directory. The helper may.

```json
{
  "version": 1,
  "checked_at": "2026-11-12T12:00:00Z",
  "current": "1.2.3",
  "latest": "1.3.0",
  "notice": "available"
}
```

`checked_at` is RFC3339 UTC. `current` and a non-empty `latest` match `X.Y.Z`. `notice` is `""`, `available`, `updated`, or `migrate-failed`. `available` requires `latest` to be numerically newer than `current`. `updated` and `migrate-failed` require `latest` equal to `current`. Components compare as decimal integers, so `03` equals `3`.

The wrapper treats the cache as fresh when `checked_at` is within the past 24 hours and `current` equals the running stamp. A missing file, a `checked_at` in the future, a `checked_at` 24 hours old or older, or a `current` that differs from the running stamp is stale. The wrapper starts a helper only when the cache is usable and stale. The helper's own re-read uses the narrower rule in decision 9.

8. **Notices.** The wrapper prints a notice only when the file is usable, `current` equals the running stamp, and the lock was acquired. It then clears `notice` and leaves `checked_at`, `current`, and `latest` unchanged.

- `available` prints detail `release X.Y.Z is available`, using `latest`. The line names that version and contains no URL.
- `updated` prints detail `updated this installation to X.Y.Z`.
- `migrate-failed` prints warning `update replaced the programs; policy and grants were not updated` and detail `fix the file and run aws-guard migrate`.

One helper run produces one notice. After it is cleared, the line stays quiet until a later helper run sets it again.

9. **Helper.** The wrapper starts its own executable with `Setsid`, stdin, stdout, and stderr on `/dev/null`, and argv of exactly one argument, `aws-cli-guard-update-check`. The wrapper entrypoint recognizes that sole argument before `gate.Decide` and does not exec the real CLI. Any other argv is an ordinary invocation. The word is not in usage text. The helper exits 0, including when the check fails. A `dev` stamp, an unverified prefix, or an unusable cache makes it exit 0 without a write, and it does not rewrite an unusable cache. Before any request it unsets `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN`, `AWS_SECURITY_TOKEN`, `AWS_PROFILE`, and `AWS_DEFAULT_PROFILE`. It keeps the rest of the environment, including proxy and CA variables. It takes the blocking lock and re-reads the cache. On that re-read, fresh means `checked_at` is within the past 24 hours and not in the future. A `current` value other than the helper's own stamp is still fresh. A fresh record is left unchanged. It does not read stdin, open `/dev/tty`, or write `settings.toml`.

10. **Release lookup.** The only production URLs are compiled in. No environment variable, flag, or argument selects them. Tests pass a base URL and an HTTP client into the library in process. The entrypoint does not read an override.

- Latest: `GET https://api.github.com/repos/joonarafael/aws-cli-guard/releases/latest` with `User-Agent: aws-cli-guard/X.Y.Z`, `Accept: application/vnd.github+json`, and `X-GitHub-Api-Version: 2022-11-28`. No `Authorization` header and no body.
- The body is capped at 1 MiB. The helper reads `tag_name`, `draft`, and `prerelease` only. A `draft` or `prerelease` of true, or a `tag_name` other than `v` plus `X.Y.Z`, means no usable release.
- Download base: `https://github.com/joonarafael/aws-cli-guard/releases/download/vX.Y.Z`. Asset names are [ADR 0025](0025-installer-and-releases.md). The platform is `runtime.GOOS` and `runtime.GOARCH`, and only `linux`/`darwin` with `amd64`/`arm64` downloads. Anything else records the check and does not download.
- Each request times out after 30 seconds. There are no retries. Redirects may be followed. The archive and `checksums.txt` both come from that pinned base URL.

A newer release is a numerically greater `X.Y.Z`. An equal or smaller tag is not installed. The helper does not downgrade.

11. **What the helper writes back.** It refreshes `checked_at` after a finished attempt. On a network, HTTP, or parse failure it does not replace. When the file already exists it keeps `current`, `latest`, and `notice`. When no file exists yet it writes the running stamp as `current`, an empty `latest`, and an empty `notice`.

When the remote version is not newer, `notice` is empty and `latest` is the remote version. When it is newer and `auto` is absent or false, `notice` is `available` and nothing is downloaded. When it is newer and `auto` is true, the helper checks the prefix and the upgrade preflight before it downloads. A disabled prefix (`aws.aside` or `aws-cli-guard-disabled` present), an unsupported platform, or an unusable policy or grants file sets `notice` to `available` and does not download. A missing policy file and a missing grants file are normal. A present file that does not load is unusable. A failed download or a checksum mismatch sets `notice` to `available` and does not replace.

12. **Download, replace, migrate.** The helper downloads into a new temporary directory and removes that directory before it exits. `checksums.txt` is capped at 1 MiB. The archive is capped at 64 MiB. The checksum line is `sha256sum` form: 64 lowercase hex digits, two spaces, the archive basename. A missing line or a mismatched digest does not unpack and does not replace. The archive must contain exactly `aws`, `aws-guard`, and `aws-cli-guard-disable`: no extra members, no directories, no absolute paths, and no `..`.

Replacement uses the [ADR 0026](0026-upgrade-behavior.md) replace of those three programs and `aws-cli-guard-manifest`. The manifest keeps every path. `version` becomes the new `X.Y.Z`, which is the stamp in the downloaded binaries, not the running process's stamp. `real_aws` is not rediscovered and not opened for writing. A prefix that upgrade would refuse is not modified. There is no rollback and no second repair path. The helper does not migrate unless the replace succeeded.

After a successful replace it runs `policy.Migrate` and `grants.Migrate` in process, as upgrade does. A missing policy file and a missing grants file stay normal. `settings.toml` and the cache are not migrated. The live `aws` path still never migrates. Success sets `current` and `latest` to the new version and `notice` to `updated`. A migrate failure leaves the programs in place, sets `current` and `latest` to the new version, and sets `notice` to `migrate-failed`.

13. **Out of scope.** Shell startup edits stay `IN-2`. Signing stays out, as in [ADR 0025](0025-installer-and-releases.md). Identity verification stays `AD-1`. The helper does not exec the real AWS CLI and does not apply policy to an AWS command.

## Consequences

- [ROADMAP.md](../ROADMAP.md) keeps `OP-4` as to do. The next product change is the implementation of this record. `IN-2` stays after that. Init in this tree still does not ask.
- [DESIGN.md](../DESIGN.md#124-auto-update-and-update-notices) points here.
- [DEVELOPMENT.md](../DEVELOPMENT.md) next ADR number is `0031`.
- A settings file that contains `[updates]` is unreadable to this binary and to older binaries, which reject unknown fields and cancel. That starts when the implementation writes the key.
- The same-command AWS process is unchanged by a check. The user sees a suggestion or an update line on a later `aws` invocation.
- [ADR 0026](0026-upgrade-behavior.md) replace is the only write inside the prefix. `real_aws` stays the path recorded at install time.
