# ADR 0005: First-class shells are Bash and zsh

## Status

Accepted

## Context

Open question 2 in [DESIGN.md](../DESIGN.md) asked which shells receive first-class integration. PATH installation is the default and does not depend on a shell. Shell integration is optional. The project supports Linux and macOS ([ADR 0001](0001-linux-and-macos-first.md)) and is maintained by one person.

Bash and zsh cover the interactive shells on those two systems. Fish, PowerShell, and other shells are real, and each one has its own startup file and its own rules for whether a function reaches a child process.

## Decision

1. **First-class shells.** Shell integration is written for Bash and zsh. Those are the shells we implement, test, and document.

2. **POSIX where it can be.** Shared install and integration logic is POSIX `sh`, so Bash and zsh run the same script. Shell-specific code is limited to the startup-file snippet each shell must source. Linux and macOS are the platforms, using the portable path handling from ADR 0001.

3. **Other shells.** Fish, PowerShell, and the rest get PATH installation only. They do not get a startup-file integration in this phase.

## Consequences

- A Fish or PowerShell user can still put the wrapper earlier on `PATH`. They do not get the optional shell-integration mode.
- Session-off stays inside the wrapper, using the kernel session id. It does not depend on this shell integration.
- Startup-file edits stay between the marked comments from the reversibility rules, with one marked block per shell we actually edit.
