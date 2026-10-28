# AWS CLI Guard

AWS CLI Guard is a lightweight, local-first wrapper around AWS CLI v2 that adds user-controlled authorization guardrails to AWS profile usage. It's small, understandable, local-first, and conservative when it cannot determine whether an invocation is permitted.

It adds friction to potentially dangerous profile usage without unnecessarily disrupting ordinary development workflows.

The primary value is the combination of profile-aware approval, persistent local policy, and best-effort AI-agent detection. It makes accidental access less likely and helps users retain control over which AWS identities their tools can use, while clearly acknowledging that stronger guarantees require controls outside the wrapper.

AWS CLI Guard is not a strict security boundary or a watertight security mechanism. It's a local safety mechanism designed to make accidental misuse less likely. See [Limitations](#limitations) for what the wrapper does not enforce.

Moreover, AWS CLI Guard does not check for what a profile can **actually** do, it has nothing to do with AWS IAM policies.

## Limitations

AWS CLI Guard is a PATH-based local wrapper. It makes accidental misuse less likely; it does not guarantee that every process on the machine reaches AWS only through it.

- Invoking the real AWS CLI through an absolute path, a different installation, or a different executable name never hits the wrapper. Another `PATH` does the same.
- A process that does not inherit the wrapper's `PATH`, or a change to `PATH` or the environment, skips it.
- SDKs, other applications, and a credential provider used directly never go through the wrapper.
- Credentials already stored by the CLI or another tool can be read and reused outside it.
- Agent detection is best-effort. Omitting or spoofing its signals defeats it, and a missing signal does not prove the caller is a person.
- A process that can change the installation, `PATH`, or local files can step around the wrapper.
- An AWS CLI older than 2.9.0, or one whose version cannot be read, is handed through after a warning, with no policy ([ADR 0002](docs/decisions/0002-aws-cli-version-floor.md)).
- Login commands, including `aws sso login`, are handed through with no policy and no grant. A later command that uses a profile is still subject to policy ([ADR 0004](docs/decisions/0004-login-out-of-scope.md)).

Purposeful off switches are separate operator controls. They are specified in the design and are not part of this list. For actual authorization, AWS-side controls remain authoritative.

## When the wrapper stops a command

A cancellation means AWS CLI Guard stopped the command. The wrapper exits with status **3** and writes the reason to stderr. Every one of those lines starts with `aws-cli-guard: error:` and says the command was not executed. Warnings that leave the command running start with `aws-cli-guard: warning:`.

Status 3 is reserved for this. The real AWS CLI keeps its own codes, including 0, 1, 2, 130, 252, 253, 254, and 255. After an allowed command, the wrapper is gone and the exit code is the CLI's code. The CLI's own output is not prefixed.

The supported AWS CLI is 2.9.0 and newer. An older CLI is run as-is after a warning, with no policy applied ([Limitations](#limitations)). Login commands are also handed through ([Limitations](#limitations)). The decision records under [docs/decisions/](docs/decisions/) are the detail.

## Developing

See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) for building, testing, and dogfooding the wrapper on Linux and macOS. Working agreements for changing the repo are in [AGENTS.md](AGENTS.md).
