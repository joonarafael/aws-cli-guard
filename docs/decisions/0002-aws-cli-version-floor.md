# ADR 0002: AWS CLI 2.9.0 is the oldest supported release

## Status

Accepted

## Context

Open question 12 in [DESIGN.md](../DESIGN.md) asked which AWS CLI versions and configuration features the wrapper must support. Profile parsing and config reading need a fixed grammar. AWS CLI 2.9.0, released in November 2022, is the first release that understands `[sso-session]` sections and `aws sso login --sso-session`. That is the configuration shape current Identity Center docs describe.

The project is maintained by one person. Matching CLIs from before that release is not a goal.

## Decision

1. **Floor.** The supported range starts at AWS CLI 2.9.0. Version 1, and version 2 older than 2.9.0, are outside that range.

2. **Older than the floor.** The wrapper logs a warning and hands the original arguments to the real AWS CLI. It does not load policy, evaluate grants, or otherwise guard the command. The exit status is the CLI's status.

3. **Known-supported ceiling.** The binary carries the newest AWS CLI version this project has verified. That constant starts at 2.9.0 and is raised only in a change that records the newer release as checked.

4. **Newer than the ceiling.** The wrapper logs a warning that the installed CLI is newer than the release this build knows, then continues the normal guarded path. The command is still parsed, still subject to policy, and still refused when policy says so.

5. **Version not readable.** If the wrapper cannot tell that the real CLI is 2.9.0 or newer, it uses the older-than-floor path: warn, then hand off with no policy. A comparable version at or above the floor, including one above the ceiling, is the only case that enters the guarded path.

Compare versions as numeric components. The warning is a wrapper warning, as specified in [ADR 0003](0003-wrapper-exit-status.md), and it is written before process replacement.

## Consequences

- Users on a CLI older than 2.9.0 get the real CLI with no guardrails from this tool, after a warning. That limit belongs in the bypass documentation.
- A current CLI will warn on every command until the known-supported constant is raised past it. The guarded path still runs.
- Config and global-option behavior are the behavior of the 2.9.0 line through the known-supported release. Features this build does not understand follow the fail-closed rule already in the design: an unresolved subject is not treated as an allow.
- The version is read from the real AWS CLI the wrapper already found, so the check does not invoke the wrapper again.
