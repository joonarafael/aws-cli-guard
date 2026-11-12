# ADR 0029: Tight and verbose logging

## Status

Accepted

## Context

[ADR 0023](0023-product-finish-before-agent-detection.md) schedules `OP-5` after persistent disable and before auto-update. [Logging](../DESIGN.md#125-logging) requires tight default logging, a verbose troubleshooting mode on stderr, and no flag on `aws`. It names `AWS_CLI_GUARD_LOG` or a policy-file setting as the trigger. [ADR 0006](0006-identity-verification-off-by-default.md) puts preferences that are not authorization in `settings.toml`, and the policy decoder rejects unknown fields. [ADR 0014](0014-guard-paths-from-user-database.md) already says verbose logging records whether guard-owned paths came from the user database or from `HOME`. [ADR 0027](0027-session-off.md) forbids the session-off warning and refusals from naming the session token, the sessions file, or how to forge the off state. Parent-process detection does not exist yet ([ADR 0023](0023-product-finish-before-agent-detection.md)).

## Decision

1. **Modes.** The only levels are `tight` and `verbose`. Tight is the stderr this binary already writes: cancellations, the `default` profile warning, loose-mode warnings, the CLI version warning, and the session-off warning. An ordinary allow adds no decision line. Verbose adds the decision block below. Both write to stderr through `output.Detail`, so each line starts with `aws-cli-guard:` and control characters are stripped. Stdout stays the AWS CLI's.

2. **Switch.** `AWS_CLI_GUARD_LOG` overrides the file for one process. An empty value, or a missing variable, means unset. `tight` and `verbose` are the only set values, matched exactly. Any other value cancels with status 3 and the reason `AWS_CLI_GUARD_LOG is invalid`. The message does not include the value. There is no flag on `aws`. The durable setting is optional `[logging]` in `settings.toml` schema version 1:

```toml
[logging]
level = "tight"
```

Absent file, absent table, and a successful load with no table all mean `tight`. A present table requires `level`, and the only values are `tight` and `verbose`. Anything else is `settings.KindInvalid` on key `logging.level`, the same fail-closed path as `identity.verify = true`. Init and `settings.Recover` keep the current default body and do not write `[logging]`. `[updates]` stays unknown until `OP-4`. The env var is read with `environ.Lookup`. The wrapper does not remove it before exec.

3. **When the file applies.** Invalid env cancels before discovery, including on `aws sso login` and when the real CLI cannot be found. Login, the version handoff, and discovery failure do not load settings, so a corrupt settings file still cannot block login ([ADR 0004](0004-login-out-of-scope.md)). On those paths verbose lines appear only when the env var is `verbose`. After `settings.Load` succeeds, an unset variable uses `logging.level`. A failed settings load still cancels; verbose lines for that cancellation come only from the env var.

4. **Block.** `verbose.Write` prints only the lines whose fields are set, in this order:

- `home: user database` or `home: HOME`
- `subject: named profile <name>`, `subject: environment`, or `subject: unresolved` plus the existing reason token when that token is non-empty
- `command: <service>` and `command: <service> <operation>`, the same text as a refusal
- `rule: <basis label> <action>`, using `policy.Basis.Label` and the action (`allow`, `prompt`, `deny`)
- `grant: absent`, `current`, `expired`, `source mismatch`, or `source unread`
- `stdin: interactive`, `non-interactive`, or `undetermined`
- `caller: recognized` or `caller: not recognized`
- `aws: <path>`
- `outcome: allow`, `refuse`, `session off`, `login handoff`, or `version handoff`

5. **Which lines each outcome includes.**

- `allow` and a policy `refuse` (profile, export, or agent rule): every line that applies. `grant:` is included only when the matched action is `prompt` and the subject is a named profile with a non-empty name. `credential_export` is the rule when export refuses, and the profile rule is the rule when execution reaches it.
- `session off`: home, subject, command, stdin, caller, aws, and `outcome: session off`. No rule line and no grant line. The existing warning is unchanged and still omits the token and the sessions path. The verbose block also omits both.
- `login handoff`: home, command, aws, and `outcome: login handoff`.
- `version handoff`: home, aws, and `outcome: version handoff`. This is the handoff where `cliversion.PolicyApplies` is false.
- A refusal before a policy rule (discovery failure, missing or unusable policy, unusable settings, unusable grants, unclassified caller, version-key failure): home when it resolves, subject and command when they are already known, aws when a binary was selected, and `outcome: refuse`.

6. **Grant word.** For a prompt named profile, no record for that profile is `absent`. Otherwise classify each of that profile's records with `grants.GrantStatus` and the fingerprint result `EvaluateGrant` already uses. `current` wins if any record is current. Otherwise use the first record's label. Do not call `Fingerprint` a second time.

7. **Home source.** The word comes from the same branch as `paths.homeDir`: `user database` when `os/user.Current()` returns an absolute home, otherwise `HOME`. Omit the line when home cannot be resolved. Omit it when a test-path override is set (`AWS_CLI_GUARD_CONFIG_DIR`, `AWS_CLI_GUARD_STATE_DIR`, or `AWS_CLI_GUARD_CACHE_DIR`), because those builds are not using that home. Do not print the directory.

8. **Secrets and fields that do not exist.** The block omits argument vectors, environment values, marker names and values, the session token, the sessions path, account and role lines, and any parent-process field. `outcome: allow` means the wrapper decided to execute the real CLI. If `ExecRealAWS` then fails, the existing cancellation follows that line and no second outcome is written. This build prints no stack trace. A later stack trace is verbose-only and redacted. Audit logging stays `AD-2`. `aws-guard` does not write this block.

9. **Order on the wire.** Existing tight lines stay where they are. The verbose block follows them on that same invocation.

## Consequences

- [ROADMAP.md](../ROADMAP.md) marks `OP-5` done. Next is `OP-4`.
- [DESIGN.md](../DESIGN.md#125-logging) points here for the switch and the line list.
- [DEVELOPMENT.md](../DEVELOPMENT.md) next ADR number is `0030`.
- A settings file that contains `[logging]` is unreadable to an older binary, which already rejects unknown fields and cancels.
