# ADR 0032: Bash and zsh shell startup integration

## Status

Accepted

## Context

[ADR 0023](0023-product-finish-before-agent-detection.md) schedules `IN-2` after install, upgrade, and the operational criteria. [ADR 0025](0025-installer-and-releases.md) ships `aws-guard install` and prints a pasteable `export PATH` line; it does not edit shell startup files. [ADR 0005](0005-bash-and-zsh.md) names Bash and zsh as first-class shells and keeps shared logic in POSIX `sh`. [DESIGN.md](../DESIGN.md) Mode B allowed a shell function; reversibility requires marked startup edits.

The practical gap is a new terminal: a one-shot `export PATH` does not persist. Enforcement stays the Go wrapper on `PATH`, not a forwarding function. Session off and persistent disable do not depend on shell integration.

## Decision

1. **Command.** `aws-guard shell bash|zsh [--prefix dir] [--remove]`. The shell name is required. Flags are only `--prefix` and `--remove`, in that order when both appear. Unknown arguments print `usage: aws-guard shell bash|zsh [--prefix dir] [--remove]` and exit 3. No terminal and no confirmation.

2. **Startup files.** `bash` edits `<home>/.bashrc` only. `zsh` edits `<home>/.zshrc` only. `<home>` uses the same resolution as the default install prefix: user database, then `HOME` ([ADR 0014](0014-guard-paths-from-user-database.md)). Fish and other shells are out of scope; they keep the pasteable line from install.

3. **Prefix.** An omitted `--prefix` is the default install directory. A relative `--prefix` is resolved once to an absolute path. The prefix must pass `PrefixUpgradeable` before any startup read or write. Persistent-disable layout in the prefix refuses with no startup write. Prefix characters must not contain newline, carriage return, or single quote.

4. **Marked block.** Identical for both shells, inserted or updated only between these exact comment lines:

```sh
# aws-cli-guard begin
export PATH='<prefix>:$PATH'
# aws-cli-guard end
```

   The block is appended when absent so it wins over earlier `PATH` assignments in the same file. A second run with the same prefix is a no-op. A second run with a different verified prefix replaces only the interior of the existing block. `--remove` deletes the begin line through the end line inclusive. A missing startup file or missing block on remove exits 0 and writes nothing.

5. **Startup file rules.** At most one begin and one end marker; malformed marker layout refuses with no write. Files containing carriage return refuse. Maximum size 1 MiB. Symlinks and non-regular files refuse without following links. New files are mode `0644`; existing files keep their mode. Atomic replace uses a same-directory partial file.

6. **What does not change.** `aws-guard install`, `aws-guard upgrade`, `install.sh`, the release archives, the seven-key manifest, and auto-update do not edit startup files. Install success adds a detail line pointing at `aws-guard shell`. Upgrade does not rewrite an existing block when the prefix path is unchanged.

7. **Stdout and stderr.** On change: detail `updated shell startup` and `startup: <absolute path>`. Unchanged: detail `shell startup left unchanged`. Errors use fixed reason lines with no paths on the error line, matching install style.

## Consequences

- [ROADMAP.md](../ROADMAP.md) marks `IN-2` done; the detection tail (`AG-6`) is next per [ADR 0023](0023-product-finish-before-agent-detection.md).
- [DEVELOPMENT.md](../DEVELOPMENT.md) next ADR number is `0033`.
- Login Bash that never sources `.bashrc` still needs the pasteable `export PATH` line or a profile that sources `.bashrc`.
- Reversal without a working wrapper is manual: delete the marked block from the startup file.
