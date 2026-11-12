# ADR 0028: Persistent disable swap script

## Status

Accepted

## Context

[ADR 0017](0017-persistent-disable-swaps-aws.md) fixes the swap mechanism. [IN-1](0025-installer-and-releases.md) laid down the manifest, reserved `aws.aside` and `aws-cli-guard-disabled`, and shipped a stub `aws-cli-guard-disable` until `OP-2`. [ADR 0023](0023-product-finish-before-agent-detection.md) schedules `OP-2` after session off. ADR 0017 leaves confirmation, temp file names, stand-in generation, partial-swap recovery, and exact user-facing lines to the implementation.

## Decision

1. **Script interface.** `aws-cli-guard-disable` takes one argument: `disable` or `enable`. Anything else prints `usage: aws-cli-guard-disable disable|enable` on the error line and exits 3. It operates on the directory that contains itself (`pwd -P` of the script path). No `--prefix` and no environment variable selects state or target.

2. **Go entry points.** `aws-guard disable` and `aws-guard enable` take no arguments. Extra arguments exit 3 with a usage line and do not start the script. Each execs the sibling `aws-cli-guard-disable` with `disable` or `enable` and returns that exit code. They do not confirm again. On non-unix platforms they exit 3 with a fixed reason and do not exec.

3. **Confirmation.** The script opens `/dev/tty`, prints one prompt on stdout, and reads one line. The required phrase is `disable` or `enable`, matching the argument. Prompts:
   - `aws-cli-guard: type disable to disable this installation`
   - `aws-cli-guard: type enable to enable this installation`
   A missing controlling terminal or a mismatched phrase prints `the installation change was not confirmed` and exits 3 without changing files. Confirmation runs before manifest validation and before any rename.

4. **Names the script may touch.** Under the install directory only: `aws`, `aws.aside`, `aws-cli-guard-disabled`, `aws.incoming`, and `aws.incoming.partial`. `aws.incoming` and `aws.incoming.partial` are not manifest keys and are not reserved for install or upgrade. Leftover incoming files do not block upgrade; the next disable or enable removes them. A symlink at any of the five names is `the installation layout cannot be changed` with no rename.

5. **Manifest.** Read `aws-cli-guard-manifest` as data; do not evaluate it as shell. Require seven keys in order: `version`, `aws`, `aws_guard`, `real_aws`, `disable_script`, `aside`, `marker`. `aws`, `aws_guard`, `disable_script`, `aside`, and `marker` must equal the canonical basenames under the script directory. `real_aws` must be non-empty, absolute, and must not equal any install path under the prefix or the five swap names. Disable does not require `real_aws` to exist; the stand-in checks at exec time.

6. **Stand-in.** Generated into `aws.incoming.partial`, mode `0755`, then renamed to `aws.incoming`. POSIX `sh` with `set -eu`. It embeds only the install directory path, single-quoted. It does not embed `real_aws`. On each invocation it re-reads `real_aws` from the manifest with the same validation, prints exactly `aws-cli-guard: warning: this installation is disabled and policy is not applied`, then `exec`s the recorded path with original arguments. If the file is missing or not executable, it prints `the recorded AWS CLI is missing` and exits 3 without searching `PATH`.

7. **Disable (enabled state).** `aws.aside` and the marker absent, `aws` a regular file: write stand-in to incoming, rename `aws` to `aws.aside`, rename incoming to `aws`, create empty marker mode `0644`. Success stdout: `this installation is disabled and policy is not applied`, then `aws-guard enable turns this installation back on`. Already disabled (aside present, marker present, `aws` is stand-in): exit 0 with `this installation is already disabled` after removing leftover incoming files.

8. **Enable (disabled state).** `aws.aside` present: remove marker and incoming files if present, rename `aws.aside` onto `aws`. Success stdout: `this installation is enabled and policy applies`. Already enabled (aside and marker absent, `aws` is wrapper): exit 0 with `this installation is already enabled`.

9. **Partial states** (after confirmation): aside present, `aws` present, marker absent — disable writes marker; enable runs the enable sequence. Aside present, `aws` absent, incoming present — disable renames incoming onto `aws` and writes marker; enable renames aside onto `aws`. Aside present, `aws` absent, incoming absent — disable exits 3 with `run enable to restore this installation` and deletes nothing; enable renames aside onto `aws`. Marker present, aside absent — both exit 3 with `the installation layout cannot be changed` and delete nothing.

10. **Out of scope.** No change to policy, grants, session off, or upgrade classification (`ErrDisableLayout` when aside or marker exists). Session off applies only while the installation is enabled.

## Consequences

- [ROADMAP.md](../ROADMAP.md) marks `OP-2` done; `OP-5` is next in the product-finish sequence.
- [DEVELOPMENT.md](../DEVELOPMENT.md) next ADR number is `0029`.
- CI continues to run `shellcheck -S warning` on [scripts/aws-cli-guard-disable](../../scripts/aws-cli-guard-disable).
