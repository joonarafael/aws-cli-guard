# Security policy

## Supported versions

Only the **latest published release** on GitHub receives security fixes.

## How to report

Use **GitHub private vulnerability reporting**: open the repository **Security** tab, choose **Report a vulnerability**.

Include steps to reproduce and affected version. **Never** attach real AWS credentials, session tokens, or full config file contents.

## In scope

- The wrapper runs a command that policy should have refused.
- A grant or confirmation can be created without the expected terminal interaction.
- Secrets appear in wrapper output, logs, or on-disk files.
- Guard files are written with loose permissions or through a symlink trap.
- The installer or updater replaces unexpected files or accepts unverified downloads.
- An environment variable disables the wrapper or redirects guard files in a **production** build.

## Out of scope

Documented limitations in [README.md](README.md#limitations), including:

- Calling the real AWS CLI by absolute path or another `PATH`.
- AWS SDKs and direct API access.
- The user editing their own policy or state files.
- Spoofing or omitting agent markers.
- The `HOME` fallback for guard path resolution.

## Response

Best effort from a single maintainer. No SLA.
