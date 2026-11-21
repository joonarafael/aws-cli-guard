# Design

## The problem

AWS IAM Identity Center lets one login reach many CLI profiles. That convenience makes it easy to run a powerful profile simply because its credentials are already on disk.

AI coding agents increasingly run terminal commands with the developer's identity. They can invoke the AWS CLI with the same profiles the developer uses, including production or administrative access.

AWS CLI Guard adds a separate, local policy layer between a process and the AWS CLI. The goal is to reduce accidental misuse, not to replace IAM or guarantee that every AWS call on the machine is authorized.

## What it is, and what it is not

AWS CLI Guard is a lightweight wrapper around AWS CLI v2. It sits earlier on `PATH`, inspects each invocation, and either refuses it, allows it, or hands off login flows that are intentionally out of band.

It is not:

- An IAM, SSO, or organizational security boundary
- A way to enforce permissions inside AWS accounts
- Interception for SDKs, boto3, or arbitrary HTTP clients
- Guaranteed detection of every AI agent or every bypass

Stronger guarantees require controls outside this tool. How the wrapper behaves in practice is documented in [ARCHITECTURE.md](ARCHITECTURE.md).

## Goals and non-goals

### Goals

- Intercept normal `aws` invocations that reach the wrapper on `PATH`.
- Resolve the effective profile and credential subject as accurately as practical.
- Apply local policy before starting the real AWS CLI.
- Support `allow`, `prompt`, and `deny` for named profiles.
- Support explicit, time-limited, machine-wide grants approved on a terminal outside the refused command.
- Optionally apply stricter rules when a caller is positively recognized as an AI agent.
- Handle common CLI configuration patterns without executing untrusted credential commands for side effects.
- Preserve the real CLI's arguments, environment, stdout, stderr, and exit codes when a command runs.
- Fail closed when policy or grants cannot be loaded or trusted.
- Provide session-scoped and installation-scoped ways to turn the wrapper off without a hidden environment bypass.
- Stay small, local-first, and explainable.

### Non-goals

- Replacing Identity Center, SSO, or credential management.
- Inspecting every API operation for read-only versus destructive semantics.
- Preventing a determined local user from bypassing the wrapper.
- Requiring a central service or shared organization policy store.
- Proving that a caller is human when no agent signal matched.

## Core ideas

### PATH wrapper

The user installs an executable named `aws` that appears before the real AWS CLI on `PATH`. The wrapper parses enough of the invocation to resolve profile and global options, loads local policy and grants, and either refuses or replaces itself with the real binary.

Discovery must never call another wrapper copy in a loop. The real binary is located by skipping wrapper identities and by honoring an explicit real-CLI override only when it points at a non-wrapper.

Arguments are forwarded unchanged. The `aws` entrypoint exposes no wrapper flags; control lives in `aws-guard` and in documented operator commands.

### Policy actions

Each named profile has an action:

- **allow** — run the command when other rules permit.
- **prompt** — refuse until the user creates a grant on this machine.
- **deny** — refuse for that profile.

Separate rules can apply to credential-export commands and to recognized agents. Details of evaluation order are in [ARCHITECTURE.md](ARCHITECTURE.md).

### Grants

A grant is a time-limited approval stored on disk. It is machine-wide for the installation, not tied to a shell session. Creation requires the control plane, a controlling terminal, and typed confirmation of the profile name. Headless automation must pre-create grants or allow profiles in policy; the wrapper does not read stdin to approve a running `aws` process.

### Agent awareness

Agent rules are opt-in. Detection uses available signals such as environment markers configured in policy. Only a positive match treats the caller as an agent. A missing marker does not prove a human and does not trigger agent denials.

Built-in marker names ship in the binary catalog; extending detection is deliberate configuration work, not an implicit default.

### Operator off switches

Two mechanisms exist and are intentionally visible:

- **Session off** — for one terminal session only, via `aws-guard off` and `aws-guard on`, with confirmation and binding to the controlling terminal device.
- **Persistent disable** — swaps the installed `aws` for a stand-in that execs the real CLI without policy, recoverable with `aws-guard enable`.

There is no supported environment variable that turns policy off for a single `aws` invocation or redirects guard-owned files in production builds.

### Credential export

Commands that print long-term access keys are a separate policy class. Default policy is restrictive because those commands export secrets to stdout.

### Login handoff

Recognized `aws sso login` invocations are handed to the real CLI without policy so users can refresh SSO sessions. Other login flows follow normal guarded evaluation unless documented otherwise in [ARCHITECTURE.md](ARCHITECTURE.md).

## Principles

### Fail closed

Missing or corrupt policy, settings, or grants refuse the command. Cache corruption may warn and continue; authorization data does not.

### Forward arguments unchanged

Do not rewrite the user's argv. Do not wrap execution in a shell.

### Preserve CLI exit codes

On Linux and macOS the wrapper process becomes the real AWS CLI. Wrapper cancellations use exit status **3**, reserved for "command was not executed."

### Never print secrets

Messages, logs, and panic paths must not echo keys, tokens, environment values, or raw argv. Go panics print argument values, so secrets stay out of parameters.

### Read files strictly

Unknown fields, duplicate JSON keys, and unsupported schema versions are errors. Migrations are explicit `aws-guard migrate`, not silent rewrites on every command.

### Reversible changes

Recovery moves broken files aside instead of deleting history. Install and disable operations use atomic writes and exclusive create where collisions would be dangerous.

### Local-first

Policy, grants, settings, and session-off state live under the user's home directory layout described in [ARCHITECTURE.md](ARCHITECTURE.md). No phone-home policy server.

### Explainable

Refusals name the rule that fired. The control plane can list policy and grants without running AWS.

### Honest about bypasses

Calling the real CLI by absolute path, another `PATH`, or an SDK bypasses the wrapper. Editing local policy or grants is allowed by design for the same user. Documentation states these limits plainly.

## User experience

Wrapper lines on stderr start with `aws-cli-guard:`. Cancellations use `aws-cli-guard: error:` and state that the command was not executed. Warnings use `aws-cli-guard: warning:`.

When a command is allowed, stdout and stderr are the real CLI's. Guidance must not tell users to disable the wrapper through a secret environment variable or to invoke an undocumented real-CLI path.

Grants are created only through `aws-guard allow` with terminal confirmation, or by editing policy to `allow`.

## Security and reliability themes

- Execute the real CLI with an argument vector, never a shell string.
- Create private directories at `0700` and files at `0600`; refuse symlink traps on guard-owned paths.
- Each invocation reloads policy from disk; a failed load does not leave a previous allow in effect.
- Tests should prove that a refused command never starts the real AWS CLI.

Mechanisms for paths, decoding, grants fingerprints, install layout, logging, and updates are specified in [ARCHITECTURE.md](ARCHITECTURE.md). Surprises discovered during implementation are recorded in [FINDINGS.md](FINDINGS.md).

## Accepted limitations

The wrapper is PATH-based and same-user. Documented bypasses and operator controls are listed in the [README Limitations](../README.md#limitations) section. They include absolute-path invocation, SDK use, spoofed agent markers, `HOME` fallback for guard paths, and CLI versions below the supported floor.

AWS-side IAM and account policies remain authoritative for what a profile can do in AWS.

## Summary

AWS CLI Guard stays small, local-first, and conservative when it cannot tell whether an invocation is permitted. It adds deliberate friction to risky profile use without blocking routine development. Profile-aware approval, durable local policy, and optional agent detection work together so you keep control over which AWS identities your shell and tools may use; stronger guarantees still require controls outside the wrapper.
