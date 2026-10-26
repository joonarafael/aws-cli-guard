# Initial Design Document

## 1. Overview

AWS CLI Guard is a lightweight, local-first wrapper around AWS CLI v2 that adds user-controlled authorization guardrails to AWS profile usage.

AWS IAM Identity Center (formerly AWS SSO) allows users to authenticate a session and obtain access to multiple AWS profiles associated with that session. While convenient, this behavior makes it easy to use a powerful profile simply because its credentials are already available.

This becomes more concerning as AI coding agents gain the ability to execute terminal commands autonomously. An agent operating in a developer's environment may be able to invoke the AWS CLI using the same profiles available to the developer, including profiles with production or administrative privileges.

AWS CLI Guard aims to reduce the risk of accidental access by introducing a separate, local policy layer between a process and the AWS CLI.

The tool is intended as a defense-in-depth measure against accidental misuse, not as a strict organizational security boundary or a watertight security mechanism.

The primary goals are:

- Require explicit user approval before a profile can be used.
- Allow users to permanently allow trusted profiles, such as explicitly designated read-only profiles.
- Allow users to deny particular profiles entirely.
- Support explicit, time-limited grants so a profile can be used on this machine without asking again on each command.
- Apply stricter policies when a caller appears to be an AI agent.
- Provide a simple, user-controlled configuration format.
- Preserve normal AWS CLI behavior whenever a command is allowed.
- Make bypasses and limitations understandable rather than claiming comprehensive enforcement.
- Turn the wrapper fully off for the current terminal session only, with a dedicated command.
- Disable and re-enable the wrapper persistently for the whole installation, including when a published build is broken.
- Distribute builds through a cross-platform GitHub Releases pipeline, with an installer and direct downloads.
- Offer auto-update that can be turned off, and suggest an update when a newer version is available.
- Log tightly by default, and support verbose logging for troubleshooting.

Session off, persistent disable, GitHub Releases, auto-update, and logging are recorded operational criteria. [Section 12](#12-recorded-operational-criteria) specifies them so they are not forgotten. They are outside the current roadmap.

The implementation language will be Go.

## 2. Goals and Non-Goals

### Goals

- Intercept normal AWS CLI invocations.
- Identify the effective AWS profile as accurately as practical.
- Apply local policy before executing the real AWS CLI.
- Support allow, prompt, and deny decisions.
- Support user-defined profile policies.
- Support explicit system-wide grants that expire after a TTL.
- Detect known AI coding agents using available signals.
- Apply agent-specific policies, including unconditional profile denials.
- Handle common AWS CLI configuration and credential-provider mechanisms safely.
- Support both PATH-based installation and an optional shell-integration mode.
- When a profile is not approved yet, fail the invocation and explain how to allow it on this machine before retrying, including headless use.
- Fail closed when policy evaluation is unavailable or ambiguous in ways that could bypass a configured restriction.
- Meet the recorded operational criteria in [Recorded operational criteria](#12-recorded-operational-criteria): session-scoped off, persistent system-wide disable and enable, GitHub Releases with direct download, optional auto-update, and tight plus verbose logging. Those criteria are specified so they are not forgotten. They are not scheduled work yet.

### Non-Goals

- Replacing AWS IAM Identity Center or AWS SSO.
- Managing AWS credentials or implementing an alternative authentication system.
- Replacing IAM policies, permission boundaries, service control policies, or other AWS authorization mechanisms.
- Providing guaranteed detection of all AI agents.
- Preventing a determined local user from bypassing the wrapper.
- Intercepting every possible AWS API request from arbitrary applications and SDKs.
- Requiring a centralized service, account, or organization-wide configuration.
- Inspecting every AWS API operation to determine whether it is read-only or destructive.

The tool is a local safety mechanism designed to make accidental misuse less likely.

## 3. Core Architecture

The tool will be installed as an executable named `aws` placed before the actual AWS CLI in the user's PATH.

Example:

```text
    PATH
      ~/.local/bin/aws       -> AWS CLI Guard wrapper
      /usr/local/bin/aws     -> Real AWS CLI v2
```

When a process executes `aws`, the wrapper:

1. Parses the invocation sufficiently to identify the effective profile and relevant options.
2. Determines the available caller-context signals.
3. Loads the local configuration and applicable policy.
4. Evaluates profile restrictions and caller-specific rules.
5. Allows the invocation, refuses it with guidance when a required grant is missing, or denies it.
6. If allowed, executes the real AWS CLI with the original arguments and appropriate environment.
7. On Linux and macOS, replaces this process with the real AWS CLI so the caller observes that program's exit code.

The wrapper must locate the real AWS CLI without accidentally invoking itself recursively.

The original arguments are forwarded unchanged. The wrapper does not add, remove, or reorder them, and it does not construct a shell command from them.

### Wrapper control plane

The `aws` entrypoint has no flags of its own. Guard behavior is controlled by the `aws-guard` binary and by `AWS_CLI_GUARD_*` environment variables. A flag on `aws` would collide with AWS CLI global options and would break the rule that arguments are forwarded unchanged.

`AWS_CLI_GUARD_REAL_AWS` overrides discovery of the real binary. Other `AWS_CLI_GUARD_*` variables are defined with the feature that reads them. The wrapper does not invent a flag form of the same control.

### Installation modes

Two modes should be considered:

#### Mode A: PATH-based wrapper

The installer places the wrapper in a directory earlier in PATH than the real AWS CLI.

Advantages:

- Works across shells and with many child processes.
- Requires little or no shell-specific code.
- Is the preferred default for the initial implementation.

Limitations:

- Can be bypassed by invoking the real AWS executable through an absolute path or another PATH.
- Does not intercept direct SDK usage or alternative credential access.
- Does not guarantee that every subprocess inherits the intended PATH.

#### Mode B: Shell integration

An optional shell function or integration script forwards AWS invocations to AWS CLI Guard.

Advantages:

- Can provide additional shell-session context.
- Can be useful for shell-specific UX or debugging.

Limitations:

- Functions are not universally inherited by subprocesses.
- Behavior differs between Bash, Zsh, Fish, PowerShell, and other shells.
- A child process can bypass the function.

The installation process may let users select either mode. PATH-based installation should be the initial implementation priority.

Neither mode should be described as complete enforcement.

### Reversibility

Uninstall is out of scope. Anything the software changes on the machine still has to be reversible later, without guessing and without a working wrapper binary.

- The installer adds the wrapper as a separate `aws` earlier on `PATH`. It never replaces or overwrites the real AWS CLI. Both files remain, which is also how the wrapper finds the real one.
- Every path an installer creates or edits is recorded in a manifest next to the installation. A later uninstall reads that manifest.
- Edits to shell startup files sit between marked begin and end comments, so only those lines need to be removed.
- Policy, grants, cache, session-off records, and a disable sentinel live only under the directories in [On-disk files](#on-disk-files). The tool does not write into `~/.aws` or into another program's files.
- User policy is kept unless a future uninstall is explicitly asked to purge it. Grants, cache, and session records can be removed without removing policy.
- The manifest and the shell markers are plain files. Recovery does not depend on running the installed `aws` or `aws-guard` binary.

## 4. Policy Model

Policies are configured locally by the user.

Each profile has one of three basic actions:

- `allow`: execute without a grant.
- `prompt`: execute only when a valid system-wide grant exists for that profile. Otherwise fail the invocation and tell the user how to allow it. The `aws` process does not ask an interactive question.
- `deny`: refuse execution. A grant cannot override this.

The default policy should be `prompt`.

A profile-specific `deny` takes precedence over a normal allow rule or a grant. Agent-specific restrictions are evaluated before a grant can authorize execution.

The initial policy evaluation order should be:

1. Resolve the effective invocation and the credential subject, where possible. A named profile, the profile literally named `default`, environment credentials, and other ambient credentials are different subjects. See [Environment-based credentials](#64-environment-based-credentials).
2. Load configuration. A missing policy file is not a corrupt file. See [Missing policy](#missing-policy) and [Corrupted files](#corrupted-files).
3. Apply explicit unconditional deny rules.
4. Apply applicable caller-context restrictions. An unknown caller uses the normal profile policy. See [Agent-specific policy](#85-agent-specific-policy).
5. If the command would print access keys, apply the credential-export rule. A profile allow or a grant does not override it. See [Commands that print access keys](#commands-that-print-access-keys).
6. Apply profile-specific policy.
7. If the action is `prompt`, require a valid system-wide grant. See [Grants](#5-grants).
8. Execute the real AWS CLI if permitted.

The exact precedence between global and profile-specific rules should be documented and deterministic. Explicit denials should never be silently overridden by broader allow rules.

### Example configuration

User policy is TOML. Illustrative, not a finalized schema:

```toml
version = 1

[defaults]
action = "prompt"

[profiles.dev-readonly]
action = "allow"

[profiles.staging-admin]
action = "prompt"

[profiles.production-admin]
action = "prompt"
# Default duration for `aws-guard allow` when --ttl is omitted.
# This field does not grant anything by itself.
approval_ttl = "5m"

[profiles.production-breakglass]
action = "deny"

# Recognized agents only. Unknown callers use the profile policy above.
[agents]
default = "deny"

[agents.profiles.dev-readonly]
action = "allow"

# Commands that print access keys. Default is prompt: fail until set to "allow".
[credential_export]
action = "prompt"
```

A build that does not enforce a section rejects a file that contains it. Ignoring `agents` or `credential_export` would let the file claim a guarantee the binary does not provide. Adding an enforced section is a schema change, as described in [Versions and schema changes](#versions-and-schema-changes).

`allow_readonly` is not a policy key. A profile is not read-only because its name contains a word such as `readonly`. If the tool later supports a read-only classification, the user must set it explicitly, or the tool must establish it through a documented, sufficiently reliable method. A `classification` field would be a user assertion unless independently verified. It is not an AWS authorization guarantee.

### Commands that print access keys

Some AWS CLI commands print credential material. The initial set is explicit:

- `aws configure export-credentials`
- `aws configure get` when the requested setting is `aws_access_key_id`, `aws_secret_access_key`, or `aws_session_token`

If `configure get` is invoked and the wrapper cannot tell whether the setting is one of those, it treats the command as secret-printing. The set grows only by an intentional change. Ordinary API calls are not in it. `aws configure list-profiles`, `aws configure set`, and SSO login are not in it.

The default action is `prompt`, including when the `[credential_export]` section is omitted. The invocation fails and the message points at that setting in the policy file. Setting it to `allow` is the override. Setting it to `deny` refuses the commands even for an otherwise allowed profile. A profile `allow` and a system-wide grant do not permit these commands.

### Configuration behavior

The tool should:

- Support a default policy for unspecified profiles.
- Support profile-specific overrides.
- Validate configuration before each policy evaluation. Each `aws` invocation loads policy from disk. There is no long-lived daemon and no cached allow decision.
- Reject malformed policy rather than treating it as allow.
- Explain configuration errors clearly, without printing file bodies or credential values.
- Keep configuration local by default.
- Create grants only through the explicit allow command.
- Provide a command to inspect the effective policy.
- Provide a command to validate the configuration.
- Create policy, grant, and cache files with mode `0600`, and their directories with mode `0700`, using an explicit chmod rather than the process umask.
- Never store AWS credentials in its own configuration, grants, cache, or logs.

Example commands, not a final interface:

```text
aws-guard init
aws-guard allow production-admin --ttl 5m
aws-guard policy list
aws-guard policy explain production-admin
aws-guard config validate
aws-guard recover policy
aws-guard recover grants
aws-guard recover cache
```

### On-disk files

Three directories, because the files have different lifetimes. User policy is something people keep and sync. Grants are machine-written and must not travel with that policy. Cache is disposable. They are not stored under `~/.aws`.

The same paths are used on Linux and macOS. The tool does not call `os.UserConfigDir()`, which on macOS points at `~/Library/Application Support`.

| Role | Default path | Respects |
| --- | --- | --- |
| User policy | `~/.config/aws-cli-guard/config.toml` | `XDG_CONFIG_HOME`, then `AWS_CLI_GUARD_CONFIG_DIR` |
| Grants | `~/.local/state/aws-cli-guard/grants.json` | `XDG_STATE_HOME`, then `AWS_CLI_GUARD_STATE_DIR` |
| Cache | `~/.cache/aws-cli-guard/` | `XDG_CACHE_HOME`, then `AWS_CLI_GUARD_CACHE_DIR` |

`AWS_CLI_GUARD_CONFIG_DIR`, `AWS_CLI_GUARD_STATE_DIR`, and `AWS_CLI_GUARD_CACHE_DIR` override the directory for tests and for dogfooding. A single home directory is not used, so a grant cannot sit beside a policy file that dotfiles sync would copy to another machine.

If `HOME` is unset and the relevant XDG variable is also unset, the wrapper stops. It does not fall back to the current directory or to `/tmp`.

The tool reads and writes these paths as regular files. A symlink is a diagnosis failure. Recovery renames the symlink aside and creates a regular file. It does not follow the symlink and write to the target.

Writes are atomic: a temporary file in the same directory, synced, then renamed into place. A crash leaves the previous file intact. Abandoned temporary files are kept and named as partials. They are not removed and they are not treated as the live file.

A partial's name is the live file name plus `.partial.` and a UTC timestamp `YYYYMMDDThhmmss.nnnnnnnnnZ`. If that name is taken, a numeric suffix is added. Partials are never deleted and never selected as the live file. The grants lock is a sibling file, the live name plus `.lock`, in the same directory. The lock file is not versioned, is not a live data file, and stays on disk after the lock is released. The lock is held across the whole grants read-modify-write. The atomic replace itself does not lock.

Grants are rewritten under a lock so two overlapping `aws-guard allow` commands cannot drop one update in a way that widens access. A lost update means the user runs allow again.

### Versions and schema changes

The policy file and the grants file each carry their own integer `version`. The version is required. Cache files that are structured carry a version too.

Decoding is strict. Unknown fields, duplicate keys, missing versions, and unknown action values are errors. The live `aws` path does not migrate or rewrite a file it does not understand.

A binary interprets the current version and any older version it still supports, in memory. An explicit migrate command, separate from recovery, can rewrite an older supported file. That command copies the current file aside first, then atomically replaces it. `aws` never calls it.

An unsupported version, including a newer one, uses the corrupted-file flow below. Writing a fresh file is not a migration. It drops the previous contents down to a default file, and the previous bytes stay on disk under the moved name.

Policy keys for behavior this build does not enforce are unknown fields. They fail validation.

### Missing policy

A missing policy file means the user has not run init. `aws` does not create the file, does not execute the real CLI, and does not treat the absence as an allow. It exits with a wrapper status and tells the user to run `aws-guard init`.

`aws-guard init` creates the config directory at mode `0700` and a versioned `config.toml` at mode `0600`. The new file uses the current schema, default action `prompt`, and credential-export action `prompt`. It has no profile allows. If the policy file already exists, init refuses to overwrite it.

A missing grants file means there are no grants. That is normal before the first allow, and it is not corruption. A missing cache means there is nothing cached. Cache misses do not affect authorization.

### Corrupted files

Corruption is a file that is present but unusable: empty, unreadable, not a regular file, invalid syntax, duplicate keys, a missing or unsupported version, unknown fields, or invalid values. Permission bits looser than `0600` are reported and are not themselves corruption. Expired grants are not corruption.

The tool diagnoses the file and prints the result to stderr. The note names the file class and path, the kind of failure, parser line and column when available, the version found, and the versions this binary supports. It names offending keys. It does not dump the file. A damaged file may contain a pasted access key or password.

The user has two choices:

1. Exit and fix the file. The file stays where it is.
2. Write a fresh file. The tool first moves the unusable file to a sibling name in the same directory, `config.toml.corrupt.<utc-timestamp>` or the equivalent for `grants.json` and cache files. If that name is taken, it picks another suffix. It never overwrites a preserved file and never deletes one. The fresh policy file is the init default. The fresh grants file is the current version with an empty grant list, which grants nothing. The fresh cache is empty.

Those two choices apply to the policy file, the grants file, and the cache. Choice 2 always moves the previous file aside for later diagnosis. The tool does not delete user data from these directories.

An `aws` invocation that hits a corrupt policy file or a corrupt grants file takes choice 1. It prints the diagnosis, explains both choices, and exits without executing AWS and without modifying the file. Choice 2 is an explicit command (`aws-guard recover policy`, `aws-guard recover grants`, or `aws-guard recover cache`). That command requires a terminal and a confirmation, so a headless process cannot reset policy or grants.

While the policy file is corrupt, nothing is executed. While the grants file is corrupt, nothing is executed either: a half-read grant file must not become an implicit allow. Cache corruption does not block `aws`. The wrapper prints a short warning and continues. The same recover command is available when the user wants a fresh cache.

## 5. Grants

A `prompt` profile runs only after the user has explicitly allowed it for this installation, for a limited time. That is the approval model. The `aws` process does not prompt. If the grant is missing or expired, the wrapper fails the call and tells the user to allow the profile first, then retry. The same failure happens in a terminal and in a headless script. Approving inside the failing process would hide the decision the tool exists to make explicit: this profile may be invoked anywhere on this machine for the TTL.

On a per-user install, "anywhere on this machine" means later invocations in any session of that user, until expiry. It does not mean every account on a shared machine. A reboot does not clear a grant early. Expiry is a wall-clock timestamp in the grants file. A clock step backwards can keep a grant alive longer. That is accepted for this threat model.

Grants are created only by the allow command. A successful `aws` invocation, an SSO login, and a config reload do not create one. Illustrative:

```text
aws-guard allow production-admin --ttl 5m
```

The command requires a controlling terminal. The user confirms by typing the profile name, so a headless caller cannot grant itself. The command then prints the profile, the expiry, and that the grant applies to later invocations on this installation, including headless ones. An agent that can already drive the user's terminal can still confirm. That limit is the same one documented for session-off in [Turn the wrapper off for one terminal session](#121-turn-the-wrapper-off-for-one-terminal-session).

`approval_ttl` in the profile's policy is the default duration when `--ttl` is omitted. It does not grant access on its own.

A grant record stores the profile name, the expiry, and a fingerprint of the credential source visible in the AWS config: account, role ARN, SSO account and role, and whether `credential_process` is set. The fingerprint does not include secrets or the credential-process command. If those source fields change, or if they cannot be read, the grant does not apply and the profile has to be allowed again.

Every invocation evaluates policy again. The grants file only answers a `prompt` action. Deny rules, including agent denials, beat a grant.

One-invocation prompts and per-shell prompts are not approval scopes. They would obscure a decision that is machine-wide for a period of time. Session identity is used for the session-off control, not for grants.

Illustrative failure:

```text
AWS CLI Guard

Profile: production-admin
Command: ec2 describe-instances
Policy: prompt

Prompt failed. This profile is not approved on this machine.
The command was not executed.

Allow it first, for a limited time, then retry. The grant covers later
invocations on this machine, including headless ones, until it expires:

  aws-guard allow production-admin --ttl 5m
```

The command words are illustrative. The message must say that the call failed because the profile is not allowed yet, and that the user allows it on the system before invoking it headlessly. It must not include a bypass, an environment variable that disables the wrapper, or the path of the real AWS CLI. It shows the service and operation, not the raw argument list.

## 6. Effective AWS Profile Resolution

Accurate profile resolution is one of the most important implementation requirements.

The wrapper should distinguish between:

- The profile explicitly requested on the command line.
- The profile selected by `AWS_PROFILE`.
- The default profile when no profile is explicitly selected.
- Credentials supplied through environment variables.
- Commands and authentication mechanisms that do not use a conventional profile.
- Credential-provider and role-chaining mechanisms that can change the eventual identity.

### 6.1 Command-line arguments

Support both forms:

```text
aws --profile production-admin ec2 describe-instances

aws ec2 describe-instances --profile production-admin

aws --profile=production-admin ec2 describe-instances
```

The wrapper should recognize global options in their supported positions and must not mistake an unrelated command argument for a profile option.

It should use a robust parser or a carefully tested parser for the AWS CLI's global argument structure.

Do not assume the profile argument is always the first argument.

If conflicting profile selections are supplied, behavior should be determined by the actual AWS CLI semantics rather than by an arbitrary wrapper rule. If the wrapper cannot safely determine which profile will be used, it should deny or require a safe resolution instead of guessing.

### 6.2 AWS_PROFILE

When no explicit `--profile` option is supplied, account for `AWS_PROFILE`.

Example:

```text
    AWS_PROFILE=staging-admin aws ec2 describe-instances
```

The wrapper should recognize that this invocation is associated with `staging-admin`.

Explicit command-line options and environment configuration must be evaluated according to the effective AWS CLI behavior.

### 6.3 Default profile

If no explicit profile is selected, the AWS CLI may use its default configuration or another applicable credential provider.

The wrapper must not assume that the absence of `--profile` means no credentials will be used.

The profile named `default` is a normal named profile. Policy can allow, prompt, or deny it. The wrapper does not forbid it. Whenever the effective subject is that profile, the wrapper writes one stderr warning that a named profile is the better practice. The warning is not a denial, and it is not a substitute for policy. Ambient credentials are not the `default` profile and do not produce this warning. See [Environment-based credentials](#64-environment-based-credentials).

### 6.4 Environment-based credentials

AWS CLI credential resolution can use environment variables such as:

```text
    AWS_ACCESS_KEY_ID
    AWS_SECRET_ACCESS_KEY
    AWS_SESSION_TOKEN
```

`AWS_ACCESS_KEY_ID` is not the named profile. When `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` are set, the AWS CLI uses those environment credentials instead of the access key stored for the profile selected by `--profile` or `AWS_PROFILE`. `AWS_SESSION_TOKEN` belongs to that same environment credential set. An allow rule or a grant for the named profile does not authorize that invocation, and the wrapper must not describe it as use of that profile.

The credential subject is one of:

- A named profile, including the profile whose name is `default`.
- Environment credentials, when the access-key variables above supply the keys.
- Another ambient source the wrapper recognizes, such as container credentials or instance metadata.
- Unresolved, when a credential source may be in use and the wrapper cannot classify it.

A grant or an allow for one subject does not cover another. The wrapper does not guess that an unresolved source is `default` or the profile named on the command line. An unresolved subject uses the default action, unless the user has set a policy for that subject. Environment credentials are that kind of subject: they can be named in policy, and they are not an entry under `profiles`.

The wrapper must never print secret access keys, session tokens, or passwords.

### 6.5 AWS configuration files

AWS CLI profiles may be defined across:

```text
    ~/.aws/config
    ~/.aws/credentials
```

The locations can also be changed through environment variables such as `AWS_CONFIG_FILE` and `AWS_SHARED_CREDENTIALS_FILE`.

The wrapper should respect the user's effective configuration locations and avoid assuming that every profile is defined in the default files.

It should account for named profile sections and the distinction between profile names in the config file and sections in the credentials file.

A profile may be defined in one file, both files, or through other credential-provider mechanisms.

### 6.6 Credential providers and role chaining

Profiles may use:

- `sso_session` and IAM Identity Center configuration.
- `sso_start_url`, `sso_region`, and related SSO settings.
- `role_arn` and `source_profile`.
- `credential_source`.
- `credential_process`.
- Web identity and container or instance metadata credentials.
- Other credential providers supported by the AWS CLI version in use.

A profile can also reference another profile or lead to a different role through a chain of credential providers.

The wrapper should distinguish between the requested profile and the eventual AWS identity.

At minimum, it should gate the requested profile before invoking the AWS CLI and make clear when the actual identity cannot be determined without further resolution.

A later version may offer optional identity verification by querying AWS STS, for example through `GetCallerIdentity`, when appropriate credentials are available.

This verification must be treated as an additional source of information, not as a prerequisite for every command. It introduces network access, latency, and additional credential-provider behavior.

### 6.7 `credential_process`

A profile may define `credential_process`, which runs an external command to obtain credentials.

This is especially important because credential providers can execute programs outside the direct AWS CLI command path.

The initial implementation should not attempt to intercept every nested process or rewrite credential-provider configuration.

However, it should:

- Recognize that a profile uses `credential_process`.
- Avoid assuming that the profile's name or static configuration fully describes the identity it returns.
- Avoid exposing credential-process output or secrets.
- Document that the wrapper cannot guarantee interception of every nested AWS invocation.
- Consider a warning or stricter policy for unclassified external credential providers.

The wrapper must not invoke a credential process merely to inspect it if doing so could have side effects or execute an untrusted command.

### 6.8 Profile resolution is not identity verification

A profile name identifies a configuration entry, not necessarily the final AWS account, role, or permission set.

For useful denial and guidance messages, the tool may optionally display:

- Requested profile.
- Configured account ID, if available.
- Configured role or role ARN, if available.
- Whether the profile appears to use SSO.
- Whether credential resolution or identity verification succeeded.

Any displayed identity information must be grounded in configuration or actual verification. The tool should not infer account IDs or permissions from profile names.

## 7. AWS SSO Behavior

AWS CLI Guard does not replace or manage AWS SSO authentication.

The initial goal is to control use of profiles whose credentials may already be available after authentication.

For example:

```text
    aws sso login --profile production-admin
```

may establish or refresh authentication state used by one or more configured profiles.

AWS CLI Guard should treat authentication and subsequent use as separate actions.

Possible behavior:

- Allow `aws sso login` by default, while documenting that login does not authorize subsequent profile usage.
- Optionally allow users to configure policies for login operations.
- Apply the normal profile policy to subsequent commands using that profile.
- Avoid assuming that a profile has no usable credentials just because the user has not explicitly approved it in AWS CLI Guard.

A successful SSO login must not automatically create a grant.

Likewise, revoking a grant does not revoke cached SSO credentials.

## 8. AI Agent Detection

The tool should support agent-aware policy as an additional defense-in-depth feature.

There is no universal, trustworthy environment variable or operating-system flag that identifies every command as originating from an AI agent.

Detection will therefore be heuristic and extensible.

### 8.1 Environment variables

Many coding agents expose environment variables to processes they launch.

Examples worth investigating include:

```text
    CLAUDECODE
    CODEX_THREAD_ID
    OPENCODE
    GOOSE_TERMINAL
    COPILOT_CLI
```

These are candidate signals, not a guaranteed or exhaustive list. Actual availability, meaning, and behavior must be verified against supported agent versions.

The implementation should:

- Maintain a configurable list of recognized markers.
- Treat a marker as a signal rather than cryptographic proof.
- Document which agents and versions have been tested.
- Avoid treating the absence of markers as proof of human execution.
- Allow additional markers to be configured by the user.
- Keep detection logic modular so new integrations can be added independently.

Environment variables can be inherited, removed, or forged. Detection should not be presented as tamper-proof.

### 8.2 Parent process inspection

On supported operating systems, the tool may inspect its parent process or process ancestry to look for known agent executables.

This may help identify an agent when environment markers are missing.

However:

- Process inspection differs across operating systems.
- Parent processes can exit or be reparented.
- Executable names can be changed or spoofed.
- Some agents launch intermediary shells or subprocesses.
- Process ancestry does not provide an authoritative identity in every environment.

This should be an optional enhancement, not a prerequisite for the core implementation.

### 8.3 TTY and interactive-session detection

The wrapper should detect whether standard input is attached to a terminal.

A non-interactive invocation may indicate automation, scripting, or an agent-driven workflow.

However, many legitimate human workflows are non-interactive, and agents may use interactive terminals.

TTY status should therefore influence policy only when explicitly configured. It must not be used as definitive proof of AI-agent execution.

### 8.4 Explicit agent integration

Future versions may support integrations where an agent is launched through a controlled wrapper or sandbox.

Such integrations can provide stronger provenance within their own boundaries than generic environment-variable detection.

They should remain optional.

### 8.5 Agent-specific policy

Illustrative TOML, for a build that enforces agent rules:

```toml
[agents]
default = "deny"

[agents.detection]
environment_markers = true
parent_process = true

[agents.profiles.dev-readonly]
action = "allow"

[agents.profiles.staging-admin]
action = "deny"

[agents.profiles.production-admin]
action = "deny"
```

The policy engine distinguishes:

- Recognized agent.
- Unknown caller.
- Interactive terminal.
- Non-interactive invocation.
- Caller context that cannot be determined.

If an agent is recognized, the tool can deny specific profiles even when those profiles are allowed for other callers.

An explicit deny takes precedence over a general allow rule or a current grant.

An unknown or uncertain caller uses the normal profile policy. A missing agent marker is not proof of a human, and it is not treated as a recognized agent. Applying the agent policy to an uncertain caller would make `allow` meaningless, because detection can never prove the caller is human. Only a positive recognition selects agent rules. Terminal and non-interactive state do not, by themselves, select those rules. See [TTY and interactive-session detection](#83-tty-and-interactive-session-detection).

### 8.6 Stronger isolation

If the objective later becomes preventing agents from accessing production credentials entirely, stronger measures may be needed.

Examples include running agents under separate operating-system accounts, in restricted containers, or within appropriately configured sandboxes that do not expose privileged credentials.

These approaches are outside the minimum viable product.

## 9. User Experience

Messages should state the security-relevant facts without overwhelming the user, and without asking the `aws` process to collect an approval.

A denial:

```text
AWS CLI Guard

Requested profile: production-admin
Account: 123456789012
Role: AdministratorAccess
Caller: recognized AI agent
Command: ec2 describe-instances

Policy: deny
Reason: production-admin is denied for recognized agents.

Command was not executed.
```

The account and role shown here are illustrative. They are included only when grounded in configuration or a real verification.

A `prompt` profile with no current grant fails the same way in a terminal and in a headless run. The message says the prompt failed and that the profile has to be allowed on this system before it is invoked, including headlessly. The full example is in [Grants](#5-grants). There is no menu, no default key, and no timeout that grants access. The process does not read stdin, so a piped payload is left untouched because the real CLI is not started.

A command that would print access keys fails until `[credential_export] action` is `allow` in the user policy. The message says so. It does not echo the key, the token, or the argument that would have selected it.

When the effective profile is `default`, stderr gets one warning line that a named profile is better practice. The command still follows policy for `default`.

Requirements:

- Denials and refusals name the policy rule that fired.
- Grants are created only by the explicit allow command, which confirms the profile name on a terminal.
- Errors and warnings go to stderr.
- When execution is allowed, the real AWS CLI's stdout and stderr stay as that program writes them. The wrapper has been replaced and does not filter them.
- The caller sees the real AWS CLI exit code. On Linux and macOS the wrapper replaces itself with that binary (`syscall.Exec`), so the code is the CLI's code, not a translation. Wrapper refusals, missing policy, corrupt policy, and corrupt grants exit with one documented wrapper status instead. That status must not be one of the AWS CLI's documented codes: 0, 1, 2, 130, 252, 253, 254, or 255. The stderr text identifies the failure as AWS CLI Guard's.
- Guidance must not include a wrapper-disabling environment variable or the path of the real AWS CLI.
- Displayed command text is the service and operation, not the raw argument vector. Profile names and other displayed fields are single-line and stripped of control characters.
- Logs and crash text follow [Secret handling](#secret-handling) and [Logging](#125-logging).

## 10. Bypass Limitations

AWS CLI Guard is a PATH-based local wrapper, not a complete security boundary.

Potential bypasses include:

- Invoking the real AWS CLI through an absolute path.
- Using a different AWS CLI installation or executable name.
- Changing PATH or environment variables.
- Calling AWS APIs through an SDK or another application.
- Using a credential provider directly.
- Running commands in a process that does not inherit the wrapper configuration.
- Reading or reusing credentials from other accessible sources.
- Removing or spoofing agent-detection markers.
- Running another process with sufficient permissions to bypass local controls.

The project should clearly document these limitations.

Intentional operator controls for turning the wrapper off are specified in [Recorded operational criteria](#12-recorded-operational-criteria). Those controls are separate from the bypasses above.

The objective is to make accidental misuse harder and more visible, not to guarantee that a determined user or process cannot reach AWS.

For actual authorization, AWS-side controls remain authoritative.

## 11. Security and Reliability Requirements

Even as a lightweight tool, the implementation should follow a few important rules.

### Fail-closed policy evaluation

A missing policy file, a corrupt policy file, a corrupt grants file, or any other failure to evaluate policy refuses the command. The wrapper does not execute the real CLI on a guess. Missing and corrupt files are different cases, specified in [Missing policy](#missing-policy) and [Corrupted files](#corrupted-files). Cache corruption does not refuse the command.

Session-off and persistent system-wide disable are recorded operator controls for convenience and recovery. They are specified in [Recorded operational criteria](#12-recorded-operational-criteria). While the wrapper is enabled, failure to evaluate policy still refuses the command.

### Safe process execution

- Do not build a shell command string from AWS arguments.
- Execute the real AWS CLI using argument arrays, forwarded unchanged.
- On Linux and macOS, replace the wrapper process so the AWS CLI's exit code and signal behavior are what the caller observes.
- Preserve the intended environment while applying any necessary wrapper-specific handling.
- Avoid recursive invocation.
- The `aws` entrypoint has no wrapper flags. See [Wrapper control plane](#wrapper-control-plane).

### Secret handling

Wrapper logs, guidance, diagnoses, and crash text must not contain passwords, secret access keys, session tokens, or credential-process output, as far as that is feasible.

- Do not put secrets, raw argument vectors, or the process environment into error strings. A Go panic prints argument values, so secret-bearing data stays out of function arguments and out of values that a stack trace would show.
- An unexpected failure prints a short wrapper error. A stack trace is emitted only in verbose troubleshooting, and only after redaction of environment values, argument values, and fields that look like keys, tokens, or passwords.
- This cannot cover a core dump, an uncaught runtime abort, or the real AWS CLI's own stdout and stderr after process replacement. The wrapper does not scrub the CLI's output, because it is no longer running. It also does not persist credentials.
- Displayed command text is the service and operation. Diagnosis of a corrupt file names keys and parser locations and does not print the file body.
- The `default` profile warning and other ordinary stderr lines follow the same redaction.

### Configuration integrity

- Detect malformed configuration and follow [Corrupted files](#corrupted-files).
- Create directories at mode `0700` and files at mode `0600`.
- Make policy changes and grants explicit.
- A failed reload must not leave a permissive policy in effect. Each invocation loads from disk, and a file it cannot trust contributes no allows and no grants.
- Recovery moves an unusable file aside and never deletes it. See [Reversibility](#reversibility).

### Policy transparency

Provide a way to explain why a particular invocation was allowed or denied.

Users should be able to inspect effective policy without executing the underlying AWS command.

### Testing

Tests should verify both the expected behavior and the absence of execution when a command is denied.

The real AWS CLI should be replaceable with a test executable during automated tests.

## 12. Recorded operational criteria

These behaviors are part of the intended product. They sit outside the phases in [Initial Roadmap](#15-initial-roadmap). This section records them so they are not forgotten. Session-off is specified in [Turn the wrapper off for one terminal session](#121-turn-the-wrapper-off-for-one-terminal-session). The other items are not designed in detail yet.

### 12.1 Turn the wrapper off for one terminal session

The user can turn the wrapper off for the current terminal session with a dedicated command.

- The off state ends with that session. A new terminal, login, or shell starts with the wrapper on.
- The command writes no policy. It writes a session record under the state directory from [On-disk files](#on-disk-files). A record for another session id has no effect, so a leftover file does not turn a new terminal off.
- For a caller that is not a recognized agent, the wrapper still starts, then replaces itself with the real AWS CLI: original arguments, original environment, no policy evaluation, and no grant check.
- Processes launched from that session share this behavior, except a recognized agent. A recognized agent ignores session-off and stays under normal policy. A different session stays on.
- The same session can turn the wrapper back on. `aws-guard on` writes a later record for that same session that turns enforcement back on. Older records are left on disk.
- The session shows that it is unprotected. `aws-guard off` says so, and each skipped invocation prints a short stderr line that this session is off and policy is not being applied. That line does not explain how to forge the off state.
- There is no environment variable whose presence disables policy. A boolean such as `AWS_CLI_GUARD_OFF=1` is something an agent can export on the same command that runs `aws`. Taking the wrapper off `PATH`, or pointing the shell at the real binary, would also skip the wrapper, which then cannot tell an agent from anything else.

Session-off has to stay inside the wrapper so the agent exception can run. The mechanism is:

1. `aws-guard off` requires a controlling terminal and an explicit confirmation. A headless caller cannot complete it. An agent that can already type on the user's terminal still can. That is an accepted limit.
2. The command records the kernel session id (`getsid` on Linux and macOS), the controlling terminal's device identity when there is one, and a random token created by that command. The token is not a well-known constant.
3. The wrapper honors the latest record for the current session only when the token matches and no recognized agent marker is present. An unknown caller with a matching record is off. That is the remaining hole: a missing agent marker is not proof of a human, so an unknown process in the confirmed session skips policy, including a process that stripped its markers. Detection remains heuristic. See [Agent-specific policy](#85-agent-specific-policy).
4. Policy refusals do not mention this mechanism, the record path, the token, or the real AWS CLI path.

Illustrative commands, not a final interface:

```text
aws-guard off
aws-guard on
```

### 12.2 Persistent system-wide disable and enable

The user can disable the wrapper for the whole installation and enable it again later. The disabled state persists across terminal sessions, logins, and reboots until enable. System-wide means every session that uses this installation. A per-user install covers that user's sessions. A machine-wide install covers the machine.

This is for troubleshooting, and for recovery when a published build is broken or corrupted.

- Disable and enable work when the installed wrapper binary does not run. The recovery path is documented and does not depend on executing that binary. The mechanism is still open. Candidates are a sentinel file that shell integration checks before invoking the wrapper, a small control script installed beside the binary, and invoking the real AWS CLI by absolute path.
- Earlier GitHub Release assets stay available, so a known-good build can be installed again by direct download. See [Publish pipeline](#123-publish-pipeline).
- The disabled state is explicit and visible.
- Persistent disable keeps every session off until enable, including a session that had been turned back on.
- Session off ([above](#121-turn-the-wrapper-off-for-one-terminal-session)) applies only in the session that requested it, and only while the installation is enabled.

This is an operator control for the person who administers the machine. [Bypass Limitations](#10-bypass-limitations) still describes what a determined local user can already do. While the wrapper is enabled, [fail-closed policy evaluation](#fail-closed-policy-evaluation) still refuses a command when policy cannot be evaluated.

Illustrative commands, not a final interface:

```text
aws-guard disable
aws-guard enable
```

### 12.3 Publish pipeline

Distribution follows [Aikido safe-chain](https://github.com/AikidoSec/safe-chain): GitHub Releases hold the artifacts, and each build can be downloaded directly.

Safe-chain publishes versioned release assets and treats those URLs as immutable. An installer script selects the platform, downloads the matching binary from `releases/download/<version>/...`, and verifies a checksum the pipeline produced. The binaries are also direct download assets. AWS CLI Guard uses that shape. GitHub Releases are the distribution channel.

- A release builds a binary for every platform the project supports at that time. Today that is Linux and macOS, amd64 and arm64, per [ADR 0001](decisions/0001-linux-and-macos-first.md). A later platform, including Windows once the project supports it, is added by extending that build matrix. The channel stays GitHub Releases.
- Artifacts are attached to a GitHub Release. The versioned asset URL is the canonical download. After publication, that version's bytes stay fixed.
- Producing the build can be automated. Publishing the release that new installs fetch is a deliberate step, so a broken build can be held back.
- Each release includes an installer that detects OS and architecture, downloads the matching binary, and verifies a checksum published with the release.
- Each release also includes the binaries and checksums as direct downloads, without the installer.
- Documentation install snippets name a specific version. The releases page lists the other versions. A latest-release URL may exist as a convenience. Installing a pinned version stays possible, including an older version during recovery.

Illustrative asset layout, not a final naming scheme:

```text
aws-cli-guard_<version>_linux_amd64.tar.gz
aws-cli-guard_<version>_linux_arm64.tar.gz
aws-cli-guard_<version>_darwin_amd64.tar.gz
aws-cli-guard_<version>_darwin_arm64.tar.gz
checksums.txt
install.sh
```

### 12.4 Auto-update and update notices

- A setting turns automatic updates on or off. The default is not decided yet.
- When a newer release exists and auto-update is off, the tool suggests the update.
- When auto-update is on, the tool downloads that release, verifies it with the same checksums as the installer, replaces the installed binary, and tells the user that it did so.
- The check and any suggestion leave the AWS command's stdout and exit code alone, and they do not stall the command. Suggestions are written to stderr.
- The check contacts the release channel only. It sends what it needs to learn the current release, such as its own version. It does not send AWS credentials, profile names, or command arguments.
- After an update, including a bad one, [persistent disable](#122-persistent-system-wide-disable-and-enable) still works, and an older release remains a direct download.

### 12.5 Logging

- Default logging is tight: concise, limited to security-relevant outcomes and errors, and quiet on the ordinary allow path. The exception is the `default` profile warning in [Default profile](#63-default-profile), which is always one stderr line when that profile is the subject.
- Verbose logging is for troubleshooting. It records the decision path: how the profile was resolved, which policy rule matched, which caller-context signals were seen, which real AWS CLI binary was selected, and why the wrapper handed off or refused.
- Both modes write to stderr, so AWS CLI stdout stays usable by scripts.
- Both modes omit passwords, secret access keys, session tokens, credential-process output, and other credential material. Verbose mode and any stack trace use the same redaction. Entire command lines stay out of the default log. See [Secret handling](#secret-handling).
- The trigger is `AWS_CLI_GUARD_LOG` or a policy-file setting. It is not a flag on `aws`. See [Wrapper control plane](#wrapper-control-plane).

Illustrative, not final:

```text
AWS_CLI_GUARD_LOG=verbose aws s3 ls
```

This is operational logging. The optional audit log in [Phase 6](#phase-6-optional-advanced-features) is a separate later idea, and it still has to redact sensitive data.

## 13. Go Implementation

Go is a good fit for a portable, self-contained command-line tool.

Potential implementation components:

- Standard Go process execution APIs for replacing the current process with the real CLI.
- A strict TOML decoder for user policy, and JSON for the grants file. Unknown fields and duplicate keys are errors.
- A robust parser for AWS CLI global options.
- OS-specific process and session inspection behind platform-specific files or interfaces.
- Unit tests for profile resolution, policy evaluation, and corrupt-file diagnosis.
- Integration tests using a fake AWS executable.
- Terminal confirmation for `aws-guard allow`, `aws-guard off`, and `aws-guard recover`. The `aws` path does not prompt.

The policy engine should be independent of process execution and terminal interaction.

Suggested conceptual components:

```text
    cmd/
      aws/
        main.go
      aws-guard/
        main.go

    internal/
      config/
      policy/
      profile/
      agent/
      grants/
      executor/
      terminal/
```

This is an illustrative layout and can be simplified for the first implementation.

The main execution path should be straightforward:

```text
    invocation = parseInvocation(args, environment)
    subject = resolveCredentialSubject(invocation, environment)
    config = loadConfig()
    if config is missing:
        explainRunInit()
        exitWithoutExecuting()
    if config or grants are corrupt:
        explainDiagnosisAndChoices()
        exitWithoutExecuting()

    context = detectCallerContext()
    if sessionOff(context) and context is not a recognized agent:
        executeRealAwsCli(originalArgs, originalEnvironment)

    decision = evaluatePolicy(config, subject, context, grants)
    if decision != allow:
        explainRefusal()
        exitWithoutExecuting()

    executeRealAwsCli(originalArgs, originalEnvironment)
```

`executeRealAwsCli` replaces the process. The caller sees the AWS CLI exit code. Wrapper exits use the documented wrapper status.

The actual implementation must handle unresolved subjects, credential-export commands, the `default` profile warning, corrupt-file recovery, and execution errors. This pseudocode is only the conceptual path. A `prompt` decision with a valid grant is an allow at the `decision != allow` check above. A `prompt` decision without one refuses.

## 14. Testing Strategy

The initial implementation should be tested without requiring access to a real AWS account.

### Profile resolution tests

- Explicit `--profile name`.
- Explicit `--profile=name`.
- Profile options before and after the service command.
- `AWS_PROFILE`.
- The default profile.
- Missing profile configuration.
- Conflicting or malformed arguments.
- `AWS_CONFIG_FILE` and `AWS_SHARED_CREDENTIALS_FILE`.
- Environment-based credentials.
- Profile references and role chains.
- Profiles with `credential_process`.
- Commands with no conventional profile.
- Authentication-related commands.

### Policy tests

- Allow, prompt, and deny.
- Unspecified profiles.
- Explicit deny precedence over a grant.
- System-wide grant expiry, and a grant whose credential fingerprint no longer matches.
- A named-profile grant does not authorize `AWS_ACCESS_KEY_ID` environment credentials.
- Credential-export commands fail unless `[credential_export] action` is `allow`, including when the profile action is `allow` or a grant is current.
- Agent-specific restrictions.
- Unknown caller with an ordinarily allowed profile follows the profile policy, not the agent policy.
- Missing policy file asks for init and does not execute.
- Corrupt policy and corrupt grants: diagnosis names the failure, the process exits, and the file bytes are unchanged.
- `aws-guard recover` moves the unusable file aside, writes a fresh file that grants nothing, and does not delete the moved file.
- A missing grants file means no grants, not corruption.

### Execution tests

- Approved command executes exactly once.
- Denied or unapproved command never executes.
- Original arguments are preserved, with no wrapper flags inserted.
- Environment variables are handled as intended.
- The process exit code after a handoff is the AWS CLI's exit code.
- A wrapper refusal uses the documented wrapper status, not an AWS CLI status.
- A refusal does not read stdin.
- Signals are handled correctly across process replacement.
- The wrapper does not recursively execute itself.
- The real executable cannot be confused with the wrapper.
- Error output is understandable and contains no secret material, raw argv, or a disabling environment variable.
- Use of the profile named `default` writes the named-profile warning and still follows policy.

### Agent detection tests

- Recognized environment markers.
- Missing markers.
- Conflicting signals.
- Marker values inherited from the environment.
- User-configured markers.
- Parent-process detection where supported.
- Non-interactive invocation without an agent marker.

The test suite should explicitly verify that no agent detector treats an absent marker as proof that the caller is human.

## 15. Initial Roadmap

### Phase 1: Minimum viable product

- Go executable.
- PATH-based wrapper with no wrapper flags on `aws`.
- Real AWS CLI discovery and process replacement, preserving the AWS CLI exit code.
- Profile argument parsing.
- `AWS_PROFILE`, the profile named `default` (with the stderr warning), and environment credentials as a separate subject from any named profile.
- On-disk layout, `aws-guard init`, corrupt-file diagnosis, and `aws-guard recover` (move the unusable file aside, never delete it).
- Strict TOML policy with `allow`, `prompt`, and `deny`.
- System-wide TTL grants, including the credential-source fingerprint. `prompt` without a current grant fails with guidance.
- Credential-export commands blocked unless the policy file allows that class.
- Safe execution of approved commands, arguments forwarded unchanged.
- Unit tests and a fake AWS CLI for integration tests.
- Clear documentation of bypass limitations.

### Phase 2: Better profile awareness

- Support custom AWS configuration locations.
- Detect common credential-provider mechanisms beyond the environment-key case covered in phase 1.
- Improve diagnostics for SSO profiles, role chaining, and `credential_process`.
- Optionally display verified AWS account and identity information.
- Add policy inspection and configuration-validation commands.

### Phase 3: Grant operations

- List current grants and show expiry and fingerprint mismatch.
- Explicit migrate command for an older policy version the binary still understands.
- Deny rules continue to beat grants. This is already required in phase 1; tests here cover reload and concurrent allow.

### Phase 4: AI-agent awareness

- Configurable environment-marker detection.
- Initial integrations for selected coding agents.
- Optional parent-process detection.
- Agent-specific profile rules.
- Unknown and uncertain callers follow the normal profile policy, as specified in [Agent-specific policy](#85-agent-specific-policy).
- Tests against actual supported agent environments.

### Phase 5: Installation and portability

- Install command or script. Uninstall stays out of scope.
- PATH-based setup that follows [Reversibility](#reversibility): never overwrite the real AWS CLI, write an install manifest, and mark any shell edits.
- Optional shell integration.
- Support for selected shells and operating systems.
- Safe detection of the actual AWS CLI executable.
- Upgrade behavior that keeps user policy in place and can run the explicit migrate command.

### Phase 6: Optional advanced features

- AWS identity verification.
- More detailed policy explanations.
- Additional agent integrations.
- Optional audit logging with sensitive-data redaction. Day-to-day tight and verbose logging is a separate recorded criterion in [Logging](#125-logging).
- Stronger integration with OS-level isolation mechanisms.

### Recorded, not yet scheduled

[Recorded operational criteria](#12-recorded-operational-criteria) are intentionally outside phases 1–6:

- A command that turns the wrapper off for the current terminal session only.
- A persistent system-wide disable and enable, usable when a published build is broken.
- A cross-platform GitHub Releases pipeline with an installer and direct downloads.
- Optional auto-update, with a suggestion when a newer version is available.
- Tight default logging and verbose logging.

Leave them unimplemented until a later pass pulls them into a phase.

The later phases should be driven by actual user needs rather than being prerequisites for a useful initial release.

## 16. Open Design Questions

Questions to resolve during implementation:

1. ~~Which operating systems are supported initially: Linux, macOS, Windows, or Windows through WSL?~~ **Resolved:** [ADR 0001](decisions/0001-linux-and-macos-first.md) — Linux and macOS first; cross-platform intent; Windows deferred.
2. Which shells should receive first-class integration?
3. ~~Should YAML or TOML be the canonical configuration format?~~ **Resolved:** TOML for user policy, JSON for grants. See [Example configuration](#example-configuration) and [On-disk files](#on-disk-files).
4. Should `aws sso login` be allowed by default, or should it have a separate policy?
5. Should identity verification be optional, and when should it run?
6. ~~How should unresolved profile and credential-source cases be handled?~~ **Resolved:** `AWS_ACCESS_KEY_ID` is not the named profile. Named profiles, environment credentials, other ambient sources, and unresolved subjects are different. A missing policy file asks for init. A corrupt policy or grants file uses the diagnosis flow and does not execute. See [Environment-based credentials](#64-environment-based-credentials), [Missing policy](#missing-policy), and [Corrupted files](#corrupted-files).
7. ~~How should session-scoped approval be tied to a terminal session?~~ **Resolved:** grants are not session-scoped. A grant is system-wide for the installation until its TTL. Session identity is used only for session-off. See [Grants](#5-grants) and [Turn the wrapper off for one terminal session](#121-turn-the-wrapper-off-for-one-terminal-session).
8. Which AI agents should be supported initially?
9. Should agent detection be enabled by default or require explicit configuration?
10. ~~What should happen when the caller is unknown and the requested profile is ordinarily allowed?~~ **Resolved:** the normal profile policy. Only a positively recognized agent uses agent rules. A missing marker is not proof of a human and is not treated as an agent. See [Agent-specific policy](#85-agent-specific-policy).
11. ~~Should the tool support an explicit, conspicuous bypass for debugging or recovery?~~ **Recorded:** yes. A dedicated command turns the wrapper off for the current terminal session only. The session-off mechanism is specified in [Turn the wrapper off for one terminal session](#121-turn-the-wrapper-off-for-one-terminal-session). A separate persistent control disables and re-enables the installation system-wide, including when a published build is broken. The persistent-disable mechanism is still open. See [Persistent system-wide disable and enable](#122-persistent-system-wide-disable-and-enable).
12. What minimum AWS CLI versions and configuration features must be supported?

## 17. Initial Product Principle

AWS CLI Guard should be small, understandable, local-first, and conservative when it cannot determine whether an invocation is permitted.

It should add friction to potentially dangerous profile usage without unnecessarily disrupting ordinary development workflows.

The primary value is the combination of profile-aware approval, persistent local policy, and best-effort AI-agent detection.

The wrapper should make accidental access less likely and help users retain control over which AWS identities their tools can use, while clearly acknowledging that stronger guarantees require controls outside the wrapper.
