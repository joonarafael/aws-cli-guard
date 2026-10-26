# AWS CLI Guard

AWS CLI Guard is a lightweight, local-first wrapper around AWS CLI v2 that adds user-controlled authorization guardrails to AWS profile usage. It's small, understandable, local-first, and conservative when it cannot determine whether an invocation is permitted.

It adds friction to potentially dangerous profile usage without unnecessarily disrupting ordinary development workflows.

The primary value is the combination of profile-aware approval, persistent local policy, and best-effort AI-agent detection. It makes accidental access less likely and helps users retain control over which AWS identities their tools can use, while clearly acknowledging that stronger guarantees require controls outside the wrapper.

AWS CLI Guard is not a strict security boundary or a watertight security mechanism. It's a local safety mechanism designed to make accidental misuse less likely. If needed, this wrapper can be bypassed by invoking the real AWS CLI executable through an absolute path or another PATH.

Moreover, AWS CLI Guard does not check for what a profile can **actually** do, it has nothing to do with AWS IAM policies.

## When the wrapper stops a command

A cancellation means AWS CLI Guard stopped the command. The wrapper exits with status **3** and writes the reason to stderr. Every one of those lines starts with `aws-cli-guard: error:` and says the command was not executed. Warnings that leave the command running start with `aws-cli-guard: warning:`.

Status 3 is reserved for this. The real AWS CLI keeps its own codes, including 0, 1, 2, 130, 252, 253, 254, and 255. After an allowed command, the wrapper is gone and the exit code is the CLI's code. The CLI's own output is not prefixed.

The supported AWS CLI is 2.9.0 and newer. An older CLI is run as-is after a warning, with no policy applied. Login commands are also handed through. The decision records under [docs/decisions/](docs/decisions/) are the detail.

## Developing

See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) for building, testing, and dogfooding the wrapper on Linux and macOS. Working agreements for changing the repo are in [AGENTS.md](AGENTS.md).
