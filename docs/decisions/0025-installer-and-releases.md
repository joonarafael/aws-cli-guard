# ADR 0025: Installer layout and GitHub Releases

## Status

Accepted

## Context

[ADR 0023](0023-product-finish-before-agent-detection.md) schedules `IN-1` and `OP-3` together, after migrate and before upgrade, session off, and persistent disable. [ADR 0017](0017-persistent-disable-swaps-aws.md) already fixes the manifest's three paths and the sibling disable script. [Publish pipeline](../DESIGN.md#123-publish-pipeline) fixes the channel and leaves asset names, the install directory, and the release workflow open.

The default install directory is a private directory under the user's home. The user can choose a different directory, including a shared bin directory. That choice must not replace an existing real AWS CLI. The real binary and the wrapper have to remain two files, which is also how discovery finds the real one ([Reversibility](../DESIGN.md#reversibility)).

## Decision

1. **One layout for both epics.** `IN-1` and `OP-3` use this layout. The implementation ships them together. This ADR does not implement either epic. `IN-3` is upgrade. `OP-2` is the swap. `IN-2` is the Bash and zsh startup edit. Those stay later.

2. **Install directory.** The default prefix is `<home>/.local/share/aws-cli-guard`, using the home directory from [ADR 0014](0014-guard-paths-from-user-database.md). This tree is not the policy, grants, or cache directories. `aws-guard install` accepts `--prefix <dir>`. A relative prefix is resolved against the current directory once, and every recorded path is absolute. When home resolution fails, install refuses unless `--prefix` is set. No environment variable selects the prefix.

3. **A chosen directory never replaces an existing file.** This holds for the default prefix and for any `--prefix`, including `/usr/local/bin`. The installer creates files only inside the prefix, and only under these four names: `aws`, `aws-guard`, `aws-cli-guard-disable`, and `aws-cli-guard-manifest`. It also reserves `aws.aside` and `aws-cli-guard-disabled` for [ADR 0017](0017-persistent-disable-swaps-aws.md). It does not create those two reserved names.

   Before it creates the prefix or any file, it `lstat`s all six paths. If any entry exists — a regular file, a directory, a symlink, or a broken symlink — the command exits 3 and writes nothing. It does not open that entry for writing, and it does not unlink, rename, or truncate it.

   A free name is created with `O_CREAT|O_EXCL`. The installer does not `rename` onto a destination, because rename replaces an existing entry. If exclusive create fails because the name appeared, the command stops and unlinks only files this invocation created. A later run refuses while the name exists, including a partial file left by a crash, names that path, and does not delete it.

   The path recorded as the real AWS CLI is not one of those six paths. Discovery may open it read-only to see whether it is a wrapper. The installer never opens it for writing and never unlinks or renames it. After install, the user puts the prefix earlier on `PATH`. That shadows the real binary. The original file stays where it was found.

4. **Real CLI discovery.** `aws-guard install` walks `PATH` and skips wrappers, using the same file-identity and build-info checks as the wrapper. It does not read `AWS_CLI_GUARD_REAL_AWS` and does not copy that variable into the manifest, so a caller cannot persist an override as the stand-in's target. If no real CLI is found, or the path it finds is one of the six prefix paths, install exits 3 and writes nothing. It does not run the real CLI.

5. **Manifest.** `aws-cli-guard-manifest` is `key=value`, one key per line, value everything after the first `=`. The keys, in this order, are `version`, `aws`, `aws_guard`, `real_aws`, `disable_script`, `aside`, and `marker`. Paths are absolute. The file is not evaluated as shell. A prefix containing a newline or a single quote is refused, so the success line `export PATH='<prefix>:$PATH'` is safe to paste. The `version` value is the version stamped into the release binary, or `dev` when the binary is unstamped.

6. **`aws-guard install`.** It takes an optional `--prefix` and no other arguments. Unknown arguments print `usage: aws-guard install [--prefix dir]` and exit 3. It needs no terminal. It copies three siblings of the running executable: that executable as `aws-guard`, the sibling `aws`, and the sibling `aws-cli-guard-disable`. A missing sibling refuses with nothing written. The command does not download, does not edit shell startup files, and does not read or write policy, settings, grants, or cache. On success it prints the prefix and the `export PATH` line. `install.sh` is the only downloader, and it never writes the prefix itself.

7. **`install.sh`.** POSIX `sh`, `set -eu`. Usage is `install.sh --version X.Y.Z [--prefix dir] [--base-url url]`. The version matches `^[0-9]+\.[0-9]+\.[0-9]+$`. The default base URL is `https://github.com/joonarafael/aws-cli-guard/releases/download/vX.Y.Z`. No environment variable selects the URL. The script detects OS and architecture, downloads that version's archive and `checksums.txt` over the base URL, verifies SHA-256, and refuses on a mismatch without unpacking. It unpacks into a fresh temporary directory, never into the prefix. The archive must contain exactly `aws`, `aws-guard`, and `aws-cli-guard-disable`: no extra members, no directories, no absolute paths, and no `..`. It then execs `"$tmp/aws-guard" install`, passing `--prefix` only when the user passed one, and removes the temporary directory. OS and architecture are Linux and macOS, amd64 and arm64 ([ADR 0001](0001-linux-and-macos-first.md)). `uname` `Linux` or `Darwin` with `x86_64` selects amd64. `Linux` with `aarch64` or `arm64`, and `Darwin` with `arm64`, select arm64. Anything else exits 3. A refusal prints `aws-cli-guard: error:` and exits 3. The script does not invoke `aws`.

8. **Disable script until `OP-2`.** The installed `aws-cli-guard-disable` is the sibling script ADR 0017 names. Any invocation, including `disable` and `enable`, prints `aws-cli-guard: error:` that persistent disable is not available in this installation yet, and exits 3. It does not read the manifest, move files, or execute the real CLI. `OP-2` replaces that body with the swap and keeps the path and the reserved aside and marker names.

9. **Release workflow.** A tag `vX.Y.Z` whose `X.Y.Z` matches the version pattern runs a workflow separate from the pull-request gate. The workflow builds `CGO_ENABLED=0` binaries for the four platforms, stages them with `aws-cli-guard-disable`, and uploads a draft GitHub Release. Publishing the draft is a manual step. The assets are:

   ```text
   aws-cli-guard_X.Y.Z_linux_amd64.tar.gz
   aws-cli-guard_X.Y.Z_linux_arm64.tar.gz
   aws-cli-guard_X.Y.Z_darwin_amd64.tar.gz
   aws-cli-guard_X.Y.Z_darwin_arm64.tar.gz
   checksums.txt
   install.sh
   ```

   `checksums.txt` is SHA-256 of the four archives, in `sha256sum` form (`hex`, two spaces, basename). `install.sh` is the repository file, unchanged. The workflow stamps `X.Y.Z` into the binary. It does not sign artifacts. Install documentation names one version. A floating latest URL is not the install path.

## Consequences

- A custom prefix is safe because the installer only adds new names, and an occupied name aborts before any write. Choosing `/usr/local/bin` when `aws` is already there refuses. Choosing it when `aws` is absent adds a new file and leaves every other entry alone.
- `IN-3` may replace bytes inside a verified prefix ([ADR 0026](0026-upgrade-behavior.md)). A second `aws-guard install` or `install.sh` run on that prefix upgrades instead of refusing. It still must not write `real_aws`.
- `OP-2` may rename the installed wrapper to the recorded aside path and write the stand-in at the recorded `aws` path. It must not rename or replace `real_aws`.
- Shell startup edits stay in `IN-2`. Until the user changes `PATH`, the new directory is not selected.
- Dogfooding from `bin/` stays the development path until the implementation lands. A local `aws-guard install` works when the three siblings are staged together, which is what the release archive does.
- Logging is [ADR 0029](0029-verbose-logging.md). Auto-update is [ADR 0030](0030-auto-update.md). It verifies the same checksums and still must not write `real_aws`. This record does not implement it.
- [ROADMAP.md](../ROADMAP.md) keeps `IN-1` and `OP-3` as to do until the implementation lands.
