# ADR 0017: Persistent disable swaps the installed `aws`

## Status

Accepted

## Context

[Persistent system-wide disable and enable](../DESIGN.md#122-persistent-system-wide-disable-and-enable) left the mechanism open. Disable and enable have to work when the installed Go binary does not run. The disabled state has to persist across terminals, logins, and reboots, cover every session that uses this installation, and be visible. Recovery must not execute that binary. No environment variable means off.

The enabled `aws` is a Go binary earlier on `PATH`. Shell integration is optional, and only Bash and zsh get it. A child process does not inherit a shell function. Two copies of the wrapper on `PATH` must not hand off to each other.

Three candidates were on the table.

A sentinel that only shell integration reads misses the default `PATH` install, child processes, and every other shell. A shell script as the permanent `aws` survives a broken Go binary and adds a shell process to every healthy command. Invoking the real AWS CLI by an absolute path recovers one command and leaves the broken `aws` on `PATH` for the next script. Deleting the `aws` name and letting `PATH` fall through can find another copy of the wrapper.

The requirement is a way to enter the disabled state without the Go binary. After that, the `aws` name has to be something other than that binary.

## Decision

1. **Swap the install's `aws` name.** Disable and enable change the file at the `aws` path in the installation. They do not edit shell startup files, and shell integration does not participate. While the installation is enabled, that path is the Go wrapper, and a command does not pass through a shell. While it is disabled, that path is a short script.

2. **What the installer records.** The manifest records three absolute paths: the Go wrapper, the `aws` path it occupies, and the real AWS CLI found at install time. The installer also places one POSIX `sh` script beside the installation. That script is not named `aws`. The manifest and the install documentation name it. `IN-1` lays this down.

3. **One implementation.** `aws-guard disable` and `aws-guard enable` call that script when the Go binary runs. When the Go binary does not run, the user runs the script by the documented path. Recovery does not execute `aws` or `aws-guard`.

4. **Confirmation.** Disable and enable require a controlling terminal and an explicit confirmation, the same rule as `aws-guard off`. A headless caller cannot complete them. A refused confirmation leaves the installation unchanged and exits 3.

5. **Disable.** The script writes the stand-in `aws` to a temporary name in the install directory, renames the Go binary to a stable aside name recorded in the manifest, then renames the stand-in onto the `aws` path. It writes a marker file in that same installation. The marker, the aside binary, and the stand-in live with the installation. They do not live under the user config, state, or cache directories.

6. **The stand-in.** It prints one stderr line, `aws-cli-guard: warning:`, that this installation is disabled and policy is not applied. The line does not include the real AWS CLI path. It then executes the real AWS CLI at the absolute path from the manifest, with the original arguments. The caller's exit code is that program's exit code. If the recorded file is missing, the stand-in prints `aws-cli-guard: error:`, does not search `PATH`, and exits 3.

7. **Enable.** The script removes the stand-in and the marker, and moves the aside binary back to the `aws` path. The aside name stays in the manifest, so a swap that stopped halfway can be finished from the manifest without guessing which file is the wrapper.

8. **Scope.** A per-user install covers that user's `aws` path. A machine-wide install covers the machine, and writing that directory is an admin action. Persistent disable keeps every session of this installation off until enable, including a session that `aws-guard on` had turned back on. Session-off applies only while the installation is enabled.

9. **Enabled discovery is unchanged.** The manifest path is what the stand-in executes. While the installation is enabled, the Go wrapper keeps the discovery rules it already has.

10. **No environment variable.** Nothing in the environment selects the disabled state. A caller that can write the install directory can swap the same files. That change is persistent and visible.

11. **Last resort.** When the install directory itself cannot be used, the documented recovery is to invoke the real AWS CLI by absolute path. That recovers one command. It is not the disable mechanism.

## Consequences

- `IN-1` records the three paths and installs the sibling script. `OP-2` is the swap, and it waits for that layout.
- A moved or deleted AWS CLI makes the disabled `aws` stop, with status 3, until a later install refreshes the manifest path. The stand-in does not walk `PATH`.
- An agent that can already type on the user's terminal can confirm disable. An agent that can write the install directory can swap the files without the script. Both are accepted limits, of the same kind as editing the policy file. The confirmation stops a headless run of the script.
- [Reversibility](../DESIGN.md#reversibility) keeps policy and grants under the user directories. The disable marker and the aside binary are install files, recorded in the manifest.
