# ADR 0014: Guard-owned paths come from the user database

## Status

Accepted

## Context

[On-disk files](../DESIGN.md#on-disk-files) placed policy, settings, grants, and cache under directories chosen by environment variables: `AWS_CLI_GUARD_CONFIG_DIR`, `AWS_CLI_GUARD_STATE_DIR`, and `AWS_CLI_GUARD_CACHE_DIR` first, then `XDG_CONFIG_HOME`, `XDG_STATE_HOME`, and `XDG_CACHE_HOME`, then `HOME`. `internal/paths` implements that order.

Any process can set those variables on the same command that runs `aws`. `AWS_CLI_GUARD_CONFIG_DIR=./x aws ...` makes the wrapper read a permissive `config.toml` the caller wrote, and `AWS_CLI_GUARD_STATE_DIR` does the same for grants. Dropping only the guard variables is not enough. `HOME=/tmp/h`, with `/tmp/h/.aws` symlinked to the real `~/.aws`, still selects a forged policy while the real CLI keeps using the real config and SSO cache.

The design already rejects an environment variable whose presence disables policy ([Turn the wrapper off for one terminal session](../DESIGN.md#121-turn-the-wrapper-off-for-one-terminal-session)). Redirecting the policy directory has the same effect and is cheaper than editing the policy file, which is persistent and visible.

Some users cannot be found in the local user database. With `CGO_ENABLED=0`, Go's `os/user` reads `/etc/passwd` on Linux, so LDAP and SSSD accounts are missing there. The wrapper must keep working for them.

## Decision

1. **Home directory.** Guard-owned paths are built from the home directory of the process's user in the user database, through `os/user.Current()`. The lookup is by user id. It does not read `HOME`, `USER`, or `LOGNAME`.

2. **Fallback.** When the lookup fails, or returns an empty or non-absolute home directory, the wrapper uses `HOME`. This is the supported path for accounts the database lookup cannot see, and it is not an error. It does not cancel the command and it does not print a warning on the ordinary path. Verbose logging ([Logging](../DESIGN.md#125-logging)) records which source was used once it exists.

3. **No home.** When the lookup fails and `HOME` is unset, empty, or not absolute, the wrapper stops, as before. It does not fall back to the current directory or to `/tmp`.

4. **Fixed layout.** Under that home directory the paths are fixed: `.config/aws-cli-guard/` for `config.toml` and `settings.toml`, `.local/state/aws-cli-guard/grants.json`, and `.cache/aws-cli-guard/`. `XDG_CONFIG_HOME`, `XDG_STATE_HOME`, and `XDG_CACHE_HOME` are not read. One rule covers all three directories, including the cache.

5. **Test-only overrides.** `AWS_CLI_GUARD_CONFIG_DIR`, `AWS_CLI_GUARD_STATE_DIR`, and `AWS_CLI_GUARD_CACHE_DIR` are compiled only into builds with the `awsguard_testpaths` build tag. `make build` does not set that tag, so a production binary has no code that reads them. `make test` sets the tag, and the integration-test helper passes it when it builds the wrapper subprocess. The user-database lookup is injectable inside `internal/paths`, so unit tests can cover both the lookup and the fallback.

6. **Both binaries.** `aws` and `aws-guard` use the same resolution, so `aws-guard init`, `allow`, and `recover` write where the wrapper reads.

7. **AWS files are unchanged.** Reading `~/.aws/config` and `~/.aws/credentials` for the grant fingerprint still follows `AWS_CONFIG_FILE`, `AWS_SHARED_CREDENTIALS_FILE`, and `HOME` ([ADR 0009](0009-aws-config-read.md)). Those paths must match what the real AWS CLI reads, and the real CLI follows the environment.

8. **macOS.** Before merging, run `make check` on the macOS laptop to confirm that `os/user.Current()` resolves the home directory in a `CGO_ENABLED=0` build ([ADR 0001](0001-linux-and-macos-first.md)).

## Consequences

- An environment variable on the `aws` command line can no longer select a different policy or grants file for users found in the user database.
- Users whose lookup falls back to `HOME` keep the old exposure to a redirected `HOME`. That is listed under [Bypass Limitations](../DESIGN.md#10-bypass-limitations).
- A process that can write `~/.config/aws-cli-guard/config.toml` or the grants file can still change policy. That stays a documented bypass. Unlike an environment variable, it is persistent and visible.
- Users who set `XDG_*` directories elsewhere get the fixed layout instead. Dotfiles that sync `~/.config` still sync policy and still do not sync grants.
- Dogfooding a `make build` binary uses the real home directory. To dogfood against a scratch policy, build with the `awsguard_testpaths` tag.
- `FD-5` in [ROADMAP.md](../ROADMAP.md) implements this ADR and updates [DEVELOPMENT.md](../DEVELOPMENT.md) when it lands.
