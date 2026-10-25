# AWS CLI Guard — Initial Design Document

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
- Support temporary approvals to reduce repetitive prompts.
- Apply stricter policies when a caller appears to be an AI agent.
- Provide a simple, user-controlled configuration format.
- Preserve normal AWS CLI behavior whenever a command is allowed.
- Make bypasses and limitations understandable rather than claiming comprehensive enforcement.

The implementation language will be Go.

## 2. Goals and Non-Goals

### Goals

- Intercept normal AWS CLI invocations.
- Identify the effective AWS profile as accurately as practical.
- Apply local policy before executing the real AWS CLI.
- Support allow, prompt, and deny decisions.
- Support user-defined profile policies.
- Support optional session-scoped and time-limited approvals.
- Detect known AI coding agents using available signals.
- Apply agent-specific policies, including unconditional profile denials.
- Handle common AWS CLI configuration and credential-provider mechanisms safely.
- Support both PATH-based installation and an optional shell-integration mode.
- Provide useful prompts explaining which profile and identity are about to be used.
- Fail closed when policy evaluation is unavailable or ambiguous in ways that could bypass a configured restriction.

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
5. Allows, prompts for approval, or denies the invocation.
6. If allowed, executes the real AWS CLI with the original arguments and appropriate environment.
7. Returns the subprocess exit status and propagates signals where practical.

The wrapper must locate the real AWS CLI without accidentally invoking itself recursively.

The original arguments should be forwarded without unnecessary modification. The implementation should use safe process execution rather than constructing a shell command from user-controlled arguments.

### Installation modes

Two modes should be considered:

**Mode A: PATH-based wrapper**

The installer places the wrapper in a directory earlier in PATH than the real AWS CLI.

Advantages:
- Works across shells and with many child processes.
- Requires little or no shell-specific code.
- Is the preferred default for the initial implementation.

Limitations:
- Can be bypassed by invoking the real AWS executable through an absolute path or another PATH.
- Does not intercept direct SDK usage or alternative credential access.
- Does not guarantee that every subprocess inherits the intended PATH.

**Mode B: Shell integration**

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

## 4. Policy Model

Policies are configured locally by the user.

Each profile has one of three basic actions:

- `allow`: execute without an approval prompt.
- `prompt`: require approval before execution.
- `deny`: refuse execution.

The default policy should be `prompt`.

A profile-specific `deny` must take precedence over a normal allow rule or temporary approval. Agent-specific restrictions must also be evaluated before an ordinary approval can authorize execution.

The initial policy evaluation order should be:

1. Resolve the effective invocation and profile, where possible.
2. Load and validate configuration.
3. Apply explicit unconditional deny rules.
4. Apply applicable caller-context restrictions.
5. Apply profile-specific policy.
6. Check any valid temporary approval.
7. Prompt when required.
8. Execute the real AWS CLI if permitted.

The exact precedence between global and profile-specific rules should be documented and deterministic. Explicit denials should never be silently overridden by broader allow rules.

### Example configuration

Illustrative YAML:

```yaml
    version: 1

    defaults:
      action: prompt

    profiles:
      dev-readonly:
        action: allow

      staging-admin:
        action: prompt

      production-admin:
        action: prompt
        approval_ttl: 300

      production-breakglass:
        action: deny

    agents:
      default: deny
      allow_readonly: true
```

This example is a starting point, not a finalized schema.

`allow_readonly` must not be interpreted as proof that a profile is read-only merely because its name contains a word such as `readonly`. If the tool supports a read-only classification, the user must explicitly configure it or the tool must establish it through a documented, sufficiently reliable method.

A more explicit future schema could distinguish a user's own profile labels from observed AWS identity information:

```yaml
    profiles:
      dev-readonly:
        action: allow
        classification: readonly
```

The `classification` field is a user assertion unless independently verified. It is not an AWS authorization guarantee.

### Configuration behavior

The tool should:

- Support a default policy for unspecified profiles.
- Support profile-specific overrides.
- Validate configuration at startup or before each policy evaluation.
- Reject malformed policy rather than silently treating it as allow.
- Explain configuration errors clearly.
- Keep configuration local by default.
- Avoid silently creating persistent approvals.
- Provide a command to inspect the effective policy.
- Provide a command to validate the configuration.
- Consider restrictive file permissions for configuration and approval-state files.
- Never store AWS credentials in its own configuration.

Example diagnostic commands:

```text
    aws-guard policy list
    aws-guard policy explain production-admin
    aws-guard config validate
```

These commands are illustrative.

## 5. Approval Scopes

Approval should be explicit and should not automatically grant unrestricted future access.

Initial approval scopes:

**One invocation**

Approve only the current AWS CLI invocation. This should be the default for privileged profiles.

**Current shell session**

Allow subsequent invocations associated with the same recognized shell session, subject to expiry and policy changes.

**Timed approval**

Allow a profile for a specified period, such as five minutes.

The user interface should make the selected scope clear.

Example:

```text
    AWS CLI Guard

    Profile: production-admin
    Caller: interactive terminal
    Command: ec2 describe-instances

    This profile requires approval.

    [a] Allow this invocation
    [s] Allow for this shell session
    [t] Allow for five minutes
    [d] Deny
```

Approval persistence must be designed carefully. A PID alone is not a reliable session identifier because PIDs can be reused. Possible session mechanisms include an explicit shell integration, a wrapper-managed session token, or a suitable process/session identifier on supported operating systems.

Temporary approvals must not override explicit deny rules, including agent-specific restrictions.

For unattended invocations, prompts cannot be assumed to work. The policy should explicitly define behavior for non-interactive contexts. A sensible default is to deny a command that requires approval if no usable interactive approval mechanism exists.

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

    aws --profile production-admin ec2 describe-instances

    aws ec2 describe-instances --profile production-admin

    aws --profile=production-admin ec2 describe-instances

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

The default profile must be subject to policy like any other profile.

### 6.4 Environment-based credentials

AWS CLI credential resolution can use environment variables such as:

```text
    AWS_ACCESS_KEY_ID
    AWS_SECRET_ACCESS_KEY
    AWS_SESSION_TOKEN
```

These variables can provide credentials independently of the named profile mechanism.

The wrapper must detect when the invocation may use environment-supplied credentials instead of the expected profile and apply a documented policy.

It must not claim that it has gated `production-admin` if the command is actually using a different credential source.

Potential initial policy options:

- Deny invocations that use credential sources the tool cannot classify.
- Require approval for unclassified credential sources.
- Allow them only through an explicit configuration option.

The default should favor visibility and safety over silently assuming the selected profile determines the identity.

The wrapper must never print secret access keys or session tokens.

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

For useful approval prompts, the tool may optionally display:

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

A successful SSO login must not automatically create an AWS CLI Guard approval.

Likewise, revoking a wrapper approval does not revoke cached SSO credentials.

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

Example:

```yaml
    agents:
      default: deny
      detection:
        environment_markers: true
        parent_process: true
      profiles:
        dev-readonly:
          action: allow
        staging-admin:
          action: deny
        production-admin:
          action: deny
```

This is illustrative configuration.

The policy engine should distinguish between:

- Recognized agent.
- Unknown caller.
- Interactive terminal.
- Non-interactive invocation.
- Caller context that cannot be determined.

If an agent is recognized, the tool can deny specific profiles even if those profiles are ordinarily allowed for human use.

An explicit deny should take precedence over a general allow rule or an existing temporary approval.

If agent detection is uncertain, the tool must apply a documented fallback policy rather than assuming the caller is human.

### 8.6 Stronger isolation

If the objective later becomes preventing agents from accessing production credentials entirely, stronger measures may be needed.

Examples include running agents under separate operating-system accounts, in restricted containers, or within appropriately configured sandboxes that do not expose privileged credentials.

These approaches are outside the minimum viable product.

## 9. Prompt and User Experience

Prompts should communicate the security-relevant information without overwhelming the user.

Example:

```text
    AWS CLI Guard

    Requested profile: production-admin
    Account: 123456789012
    Role: AdministratorAccess
    Caller: recognized AI agent
    Command: ec2 describe-instances

    Policy: DENY
    Reason: production-admin is denied for recognized agents.

    Command was not executed.
```

The account and role shown here are illustrative.

For an ordinary approval:

```text
    AWS CLI Guard

    Requested profile: staging-admin
    Command: ec2 describe-instances
    Caller: interactive terminal

    This profile requires approval.

    [a] Allow this invocation
    [s] Allow for this shell session
    [t] Allow for five minutes
    [d] Deny
```

Important UX requirements:

- Denials should explain the applicable policy.
- Approval should be explicit.
- The default choice should be deny or no action.
- Timeouts should not automatically grant permission.
- Non-interactive prompts must not hang indefinitely.
- Errors should be written to stderr.
- The AWS CLI's stdout and stderr should remain usable by scripts when execution is allowed.
- The wrapper should preserve the real AWS CLI exit code.
- Terminal prompts should not corrupt machine-readable command output.
- Logs must not contain credential values or unnecessary sensitive command arguments.

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

The objective is to make accidental misuse harder and more visible, not to guarantee that a determined user or process cannot reach AWS.

For actual authorization, AWS-side controls remain authoritative.

## 11. Security and Reliability Requirements

Even as a lightweight tool, the implementation should follow a few important rules.

### Fail-closed policy evaluation

If configuration cannot be parsed or policy evaluation fails, do not silently execute a command that requires approval.

A configurable emergency bypass, if ever implemented, should be explicit, visible, and carefully documented.

### Safe process execution

- Do not build a shell command string from AWS arguments.
- Execute the real AWS CLI using argument arrays.
- Preserve the intended environment while applying any necessary wrapper-specific handling.
- Avoid recursive invocation.
- Forward signals and return exit status correctly where supported.

### Secret handling

- Never log AWS secret access keys, session tokens, or credential-process output.
- Avoid recording entire command lines by default.
- Treat configuration and process environments as potentially sensitive.
- Do not persist credentials.

### Configuration integrity

- Detect malformed configuration.
- Use safe file permissions where practical.
- Make policy changes explicit.
- Ensure a failed configuration reload cannot accidentally leave a permissive policy active.

### Policy transparency

Provide a way to explain why a particular invocation was allowed or denied.

Users should be able to inspect effective policy without executing the underlying AWS command.

### Testing

Tests should verify both the expected behavior and the absence of execution when a command is denied.

The real AWS CLI should be replaceable with a test executable during automated tests.

## 12. Go Implementation

Go is a good fit for a portable, self-contained command-line tool.

Potential implementation components:

- Standard Go process execution APIs for launching the real CLI.
- A well-maintained YAML library if YAML is selected as the configuration format.
- A robust parser for AWS CLI global options.
- OS-specific process inspection behind platform-specific files or interfaces.
- Unit tests for profile resolution and policy evaluation.
- Integration tests using a fake AWS executable.
- A small terminal-prompting layer.

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
      approval/
      executor/
      terminal/
```

This is an illustrative layout and can be simplified for the first implementation.

The main execution path should be straightforward:

```text
    invocation = parseInvocation(args, environment)
    config = loadAndValidateConfig()
    context = detectCallerContext()
    profile = resolveEffectiveProfile(invocation, environment)
    decision = evaluatePolicy(config, profile, context)

    if decision == deny:
        explainDenial()
        exitWithoutExecuting()

    if decision == prompt:
        if not obtainExplicitApproval():
            exitWithoutExecuting()

    executeRealAwsCli(args, environment)
```

The actual implementation must handle unresolved profiles, credential-source ambiguity, non-interactive operation, signals, and execution errors. This pseudocode is only the conceptual happy path.

## 13. Testing Strategy

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
- Explicit deny precedence.
- Temporary approval expiration.
- Session-scoped approval behavior.
- Agent-specific restrictions.
- Invalid configuration.
- Non-interactive invocations.
- Unknown caller context.

### Execution tests

- Approved command executes exactly once.
- Denied command never executes.
- Original arguments are preserved.
- Environment variables are handled as intended.
- Exit status is propagated.
- Signals are handled correctly.
- The wrapper does not recursively execute itself.
- The real executable cannot be confused with the wrapper.
- Error output is understandable.

### Agent detection tests

- Recognized environment markers.
- Missing markers.
- Conflicting signals.
- Marker values inherited from the environment.
- User-configured markers.
- Parent-process detection where supported.
- Non-interactive invocation without an agent marker.

The test suite should explicitly verify that no agent detector treats an absent marker as proof that the caller is human.

## 14. Initial Roadmap

### Phase 1: Minimum viable product

- Go executable.
- PATH-based wrapper.
- Real AWS CLI discovery.
- Profile argument parsing.
- `AWS_PROFILE` and default-profile handling.
- Local configuration.
- `allow`, `prompt`, and `deny`.
- Safe execution of approved commands.
- Correct exit-status handling.
- Unit tests and a fake AWS CLI for integration tests.
- Clear documentation of bypass limitations.

### Phase 2: Better profile awareness

- Support custom AWS configuration locations.
- Recognize environment credential overrides.
- Detect common credential-provider mechanisms.
- Improve diagnostics for SSO profiles, role chaining, and `credential_process`.
- Optionally display verified AWS account and identity information.
- Add policy inspection and configuration-validation commands.

### Phase 3: Approval management

- One-invocation approvals.
- Session-scoped approvals.
- Timed approvals.
- Expiration and invalidation behavior.
- Clear display of existing approval state.
- Explicit precedence of deny rules over approval state.

### Phase 4: AI-agent awareness

- Configurable environment-marker detection.
- Initial integrations for selected coding agents.
- Optional parent-process detection.
- Agent-specific profile rules.
- Explicit behavior for unknown and ambiguous callers.
- Tests against actual supported agent environments.

### Phase 5: Installation and portability

- Install and uninstall commands or scripts.
- PATH-based setup.
- Optional shell integration.
- Support for selected shells and operating systems.
- Safe detection of the actual AWS CLI executable.
- Upgrade and configuration-migration behavior.

### Phase 6: Optional advanced features

- AWS identity verification.
- More detailed policy explanations.
- Additional agent integrations.
- Optional audit logging with sensitive-data redaction.
- Stronger integration with OS-level isolation mechanisms.

The later phases should be driven by actual user needs rather than being prerequisites for a useful initial release.

## 15. Open Design Questions

Questions to resolve during implementation:

1. Which operating systems are supported initially: Linux, macOS, Windows, or Windows through WSL?
2. Which shells should receive first-class integration?
3. Should YAML or TOML be the canonical configuration format?
4. Should `aws sso login` be allowed by default, or should it have a separate policy?
5. Should identity verification be optional, and when should it run?
6. How should unresolved profile and credential-source cases be handled?
7. How should session-scoped approval be tied to a terminal session?
8. Which AI agents should be supported initially?
9. Should agent detection be enabled by default or require explicit configuration?
10. What should happen when the caller is unknown and the requested profile is ordinarily allowed?
11. Should the tool support an explicit, conspicuous bypass for debugging or recovery?
12. What minimum AWS CLI versions and configuration features must be supported?

## 16. Initial Product Principle

AWS CLI Guard should be small, understandable, local-first, and conservative when it cannot determine whether an invocation is permitted.

It should add friction to potentially dangerous profile usage without unnecessarily disrupting ordinary development workflows.

The primary value is the combination of profile-aware approval, persistent local policy, and best-effort AI-agent detection.

The wrapper should make accidental access less likely and help users retain control over which AWS identities their tools can use, while clearly acknowledging that stronger guarantees require controls outside the wrapper.
