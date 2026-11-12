# ADR 0026: Upgrade behavior

## Status

Accepted

## Context

[ADR 0025](0025-installer-and-releases.md) ships `aws-guard install`, `install.sh`, and a tag-driven release workflow. Its consequences say a second install refuses and does not run migrate until upgrade exists. [FS-5](0024-migrate-command.md) is done: `aws-guard migrate` rewrites older supported policy and grants files and is non-interactive so an upgrade path can call it. [ADR 0023](0023-product-finish-before-agent-detection.md) schedules `IN-3` after the installer and release pipeline.

Policy, settings, grants, and cache live outside the install prefix ([DESIGN.md](../DESIGN.md#on-disk-files)). Upgrade must refresh the wrapper binaries without touching the real AWS CLI path recorded at install time.

## Decision

1. **Entry points.** `aws-guard upgrade [--prefix dir]` upgrades an existing installation and refuses when the prefix is not a verified installation. `aws-guard install [--prefix dir]` keeps exclusive create when all six reserved names under the prefix are absent, and takes the same upgrade path when the prefix is already a verified installation. [install.sh](../../install.sh) still execs `aws-guard install`; a second `sh install.sh --version X.Y.Z` with `--prefix` therefore upgrades. No new shell flag. The upgrade-versus-fresh decision stays in Go because the shell cannot resolve the user-database default prefix.

2. **Files that may change.** Only four names inside the prefix: `aws`, `aws-guard`, `aws-cli-guard-disable`, and `aws-cli-guard-manifest`. The manifest keeps every path except `version`, which becomes the running binary's release stamp (`dev` or `X.Y.Z`). There is no version comparison; an older pinned release may replace a newer one. `real_aws` is never rediscovered and never opened for writing. If `aws.aside` or `aws-cli-guard-disabled` exists under the prefix, the command exits 3 before any write ([ADR 0017](0017-persistent-disable-swaps-aws.md)).

3. **Prefix classification.** Six reserved paths: the four install files, plus `aws.aside` and `aws-cli-guard-disabled`. All six absent means fresh install with exclusive create. All six present with a valid manifest and wrapper identity means upgrade. Any other combination exits 3 with no write.

4. **Manifest and wrapper identity.** The manifest is `key=value`, exactly seven keys in fixed order: `version`, `aws`, `aws_guard`, `real_aws`, `disable_script`, `aside`, `marker`. No extra keys, no empty values. `aws`, `aws_guard`, `disable_script`, `aside`, and `marker` must be the canonical basenames joined onto the resolved prefix. Each of `aws` and `aws-guard` must be a regular non-symlink file whose Go build-info main package path matches the sibling about to be copied: installed `aws` matches the bundle `aws` (`cmd/aws`), installed `aws-guard` matches the running `aws-guard` (`cmd/aws-guard`). The disable script must be a regular non-symlink file at the recorded path; its bytes are not inspected. Failure exits 3 with a fixed reason and does not name paths on the error line.

5. **Replacement.** Copy each of the three programs from the running bundle when source and destination are not the same file (so `aws-guard upgrade` from the installed binary does not truncate itself). Use a same-directory partial file with `O_NOFOLLOW`, fsync, lstat again, then rename. Mode `0755` for the three programs and `0644` for the manifest. Write the manifest last. A partial failure after one successful rename does not roll back; stderr tells the user to re-run the same command.

6. **Migrate.** After a successful prefix replace, run `policy.Migrate` and `grants.Migrate` in process. Load policy and grants before any prefix write. A missing policy file is normal on upgrade and does not create a file. A missing grants file stays normal. `settings.toml` and the cache are not read or written. The standalone `aws-guard migrate` command keeps [ADR 0024](0024-migrate-command.md): a missing policy still cancels there. The live `aws` path still never migrates.

7. **Migration failure.** No new command, no binary rollback, no automatic delete. Exit 3. Print existing corruption diagnosis when applicable, then a short stderr guide. Two openings: if the prefix was not touched, say nothing in it changed; if the programs were already replaced, say so and that policy and grants were not deleted. Then: a `*.migrated.<timestamp>` sibling holds the previous copy when a rewrite started; fix the file and run `aws-guard migrate`; corrupt files may be reset with `aws-guard recover policy` or `aws-guard recover grants` on a terminal; to start over, move the config, state, and cache directories aside and run `aws-guard init` (moving config also moves `settings.toml`, which can be moved back). Print resolved directory paths when home resolution works. Do not print `real_aws` or file contents. A broken prefix is separate: move the prefix aside and run install again; that does not delete policy.

8. **Needs no terminal.** Upgrade and install on an existing prefix do not read stdin and do not open `/dev/tty`.

## Consequences

- [ADR 0025](0025-installer-and-releases.md) consequence on second install is replaced: a verified prefix upgrades instead of refusing.
- [ROADMAP.md](../ROADMAP.md) marks `IN-3` done; `OP-1` and `OP-2` are next in the product-finish sequence.
- Auto-update reuses this replace path ([ADR 0030](0030-auto-update.md)). It still must not write `real_aws`.
