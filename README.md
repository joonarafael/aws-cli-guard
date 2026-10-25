# AWS CLI Guard

AWS CLI Guard is a lightweight, local-first wrapper around AWS CLI v2 that adds user-controlled authorization guardrails to AWS profile usage. It's small, understandable, local-first, and conservative when it cannot determine whether an invocation is permitted.

It adds friction to potentially dangerous profile usage without unnecessarily disrupting ordinary development workflows.

The primary value is the combination of profile-aware approval, persistent local policy, and best-effort AI-agent detection. It makes accidental access less likely and helps users retain control over which AWS identities their tools can use, while clearly acknowledging that stronger guarantees require controls outside the wrapper.

AWS CLI Guard is not a strict security boundary or a watertight security mechanism. It's a local safety mechanism designed to make accidental misuse less likely. If needed, this wrapper can be bypassed by invoking the real AWS CLI executable through an absolute path or another PATH.

Moreover, AWS CLI Guard does not check for what a profile can **actually** do, it has nothing to do with AWS IAM policies.

## Developing

See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) for building, testing, and dogfooding the wrapper on Linux and macOS.
