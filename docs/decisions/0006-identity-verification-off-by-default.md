# ADR 0006: Identity verification is off unless the user turns it on

## Status

Accepted

## Context

Open question 5 in [DESIGN.md](../DESIGN.md) asked whether identity verification is optional and when it runs. The design already treats a profile name as the policy subject, and it already treats an STS query such as `GetCallerIdentity` as extra information. That query needs the network, adds latency, and can run credential-provider code. The wrapper must not run `credential_process` merely to inspect a profile.

The ordinary `aws` path should stay local. A user who wants a live account and role check can turn that on.

There is nowhere to store that choice today. User policy is authorization. Grants are machine-written and must not travel with policy. A preference that is not an allow or a deny does not belong in either file. The policy decoder rejects unknown fields, so adding the switch to `config.toml` would make that file unloadable on a build that does not enforce the key.

## Decision

1. **Optional, off by default.** Identity verification does not run unless the user has turned it on. The guarded `aws` path does not call AWS to decide or to decorate a message.

2. **Opt-in.** The user turns it on explicitly. While it is on, a verification result may be shown when it is grounded in that check. A failed or skipped check does not become an allow, and it does not become a deny by itself. Policy still uses the credential subject.

3. **Settings file.** Preferences that are not authorization live in `settings.toml`, in the same config directory as `config.toml`. The file is versioned TOML, mode `0600`, decoded strictly, and kept under the same atomic-write and symlink rules as the policy file. `aws-guard init` creates it when it is missing and does not overwrite an existing policy file to do so. A missing file means the defaults, including verification off.

4. **Corrupt settings.** A settings file that is present but unusable follows the corrupt-file path: the command is cancelled with the wrapper status, the file is left in place, and `aws-guard recover settings` can move it aside and write a fresh file of defaults.

Illustrative shape, not a frozen schema beyond the two switches this decision and [ADR 0007](0007-auto-update-asked-at-init.md) require. `[updates]` is omitted until init records the user's answer. A missing `auto` key means updates stay off.

```toml
version = 1

[identity]
verify = false
```

## Consequences

- Messages show an account or role from the config file when that file contains one, and from STS only after the user has opted in.
- `AD-1` stores the switch in this file. Auto-update stores its switch here too.
- A build that does not implement verification yet still accepts a settings file whose `verify` value is false. A file that turns verification on is a feature the build must enforce or reject, on the same rule as an unenforced policy section: the file must not claim a check the binary does not perform.
