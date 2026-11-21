# AWS CLI Guard

AWS CLI Guard is a lightweight, local-first wrapper around AWS CLI v2 that adds user-controlled authorization guardrails to AWS profile usage. It's small, understandable, local-first, and conservative when it cannot determine whether an invocation is permitted.

It adds friction to potentially dangerous profile usage without unnecessarily disrupting ordinary development workflows.

The primary value is the combination of profile-aware approval, persistent local policy, and best-effort AI-agent detection. It makes accidental access less likely and helps users retain control over which AWS identities their tools can use, while clearly acknowledging that stronger guarantees require controls outside the wrapper.

AWS CLI Guard is not a strict security boundary or a watertight security mechanism. It's a local safety mechanism designed to make accidental misuse less likely. See [Limitations](#limitations) for what the wrapper does not enforce.

Moreover, AWS CLI Guard does not check for what a profile can **actually** do, it has nothing to do with AWS IAM policies.

## Quick start

Prerequisites: Linux or macOS, `curl`, and AWS CLI v2 on `PATH`.

After the first release is published, pick the version from [GitHub Releases](https://github.com/joonarafael/aws-cli-guard/releases) and run one line (bash):

```bash
VERSION=0.1.0; curl -fsSL "https://github.com/joonarafael/aws-cli-guard/releases/download/v${VERSION}/install.sh" | sh -s -- --version "$VERSION" && export PATH="$HOME/.local/share/aws-cli-guard:$PATH" && aws-guard shell bash && aws-guard init
```

For zsh, use `aws-guard shell zsh` instead of `aws-guard shell bash`. Re-running the install line upgrades binaries in place. A second `aws-guard init` reports that the policy file already exists. Run `aws-guard help` to list every control-plane command.

## Limitations

AWS CLI Guard is a PATH-based local wrapper. It makes accidental misuse less likely; it does not guarantee that every process on the machine reaches AWS only through it.

- Invoking the real AWS CLI through an absolute path, a different installation, or a different executable name never hits the wrapper. Another `PATH` does the same.
- A process that does not inherit the wrapper's `PATH`, or a change to `PATH` or the environment, skips it.
- SDKs, other applications, and a credential provider used directly never go through the wrapper.
- Credentials already stored by the CLI or another tool can be read and reused outside it.
- Agent detection is best-effort. Omitting or spoofing its signals defeats it, and a missing signal does not prove the caller is a person.
- A process that can change the installation, `PATH`, or local files can step around the wrapper.
- An AWS CLI older than 2.9.0, or one whose version cannot be read, is handed through after a warning, with no policy.
- This build recognizes `aws sso login`, including with `--profile` or `--sso-session`, and hands it through with no policy and no grant. A login form this build does not recognize follows the normal guarded path. A later command that uses a profile is still subject to policy.
- `aws login` and the `login_session` profile key are not treated specially; they follow the normal guarded path.

For the current terminal session only, `aws-guard off` (with confirmation on the controlling terminal) and `aws-guard on` turn policy off and on. A new terminal starts with policy on. A recognized AI agent still follows policy when detection is on. Writing `sessions.json` under the state directory is a same-user bypass, like editing policy or grants locally. Persistent installation-wide disable is separate (see [Architecture](docs/ARCHITECTURE.md#install-upgrade-and-persistent-disable)). For actual authorization, AWS-side controls remain authoritative.

## When the wrapper stops a command

A cancellation means AWS CLI Guard stopped the command. The wrapper exits with status **3** and writes the reason to stderr. Every one of those lines starts with `aws-cli-guard: error:` and says the command was not executed. Warnings that leave the command running start with `aws-cli-guard: warning:`.

Status 3 is reserved for this. The real AWS CLI keeps its own codes, including 0, 1, 2, 130, 252, 253, 254, and 255. After an allowed command, the wrapper is gone and the exit code is the CLI's code. The CLI's own output is not prefixed.

The supported AWS CLI is 2.9.0 through the verified ceiling (2.40.3 in this release). An older CLI is run as-is after a warning, with no policy applied ([Limitations](#limitations)). Recognized `aws sso login` is handed through without policy ([Limitations](#limitations)). Behavior is documented in [Architecture](docs/ARCHITECTURE.md).

Logging is tight by default: an allowed command adds no extra stderr. Set `AWS_CLI_GUARD_LOG=verbose` for one invocation, or set `level = "verbose"` under `[logging]` in `settings.toml`, to print a troubleshooting decision block on stderr.

## Install

Local install uses `aws-guard install`, which writes the wrapper, the control plane binary, the persistent-disable script, and an install manifest under a new directory. Use `aws-guard disable` and `aws-guard enable` (or run `aws-cli-guard-disable` in the install directory when the Go binary does not run) to turn policy off or on for the whole installation. The default prefix is `~/.local/share/aws-cli-guard`. You can pass `--prefix` to choose another directory. The installer never replaces an existing file in that directory.

On success the command prints the prefix and a line you can paste to put that directory earlier on `PATH` for the current shell. To keep the prefix on `PATH` in new terminals, run `aws-guard shell bash` or `aws-guard shell zsh` after install. That adds a marked block to `.bashrc` or `.zshrc`. Remove it with `aws-guard shell bash --remove` or `aws-guard shell zsh --remove`. Run `aws-guard init` if you have not already.

After a maintainer publishes a release from GitHub, install from a pinned version with the release `install.sh` asset (not a floating latest URL). Pick the version from the project's GitHub Releases page. You need **curl**, **minisign**, and **sha256sum** or **shasum** on `PATH`.

```bash
sh install.sh --version X.Y.Z
```

`install.sh` downloads the platform archive, `checksums.txt`, and `checksums.txt.minisig`, verifies the minisign signature on `checksums.txt`, then checks the archive SHA-256 from that file. To verify checksums yourself before running the script:

```bash
minisign -Vm checksums.txt -P "$(sed -n '2p' path/to/minisign.pub)"
```

The pinned public key in `install.sh` matches [internal/update/minisign.pub](internal/update/minisign.pub). Compare `install.sh` to the git tag if you need assurance about the script itself.

`install.sh` accepts optional `--prefix` and `--base-url`. Tag pushes create a draft release only; maintainers attach `checksums.txt.minisig` and publish on GitHub ([CONTRIBUTING.md](CONTRIBUTING.md)).

## First run: everything waits for your approval

After `aws-guard init`, every profile in `config.toml` defaults to `prompt`, including `default`. Commands such as `aws --version`, `aws help`, `aws configure`, and `aws configure sso` use the `default` profile when you do not pass `--profile`, so they are blocked until you allow that profile or create a grant. `aws sso login` is handed through without a grant.

To proceed:

- Grant temporary access: `aws-guard allow <profile> --ttl 8h` (confirm on the controlling terminal).
- Allow a profile permanently in `~/.config/aws-cli-guard/config.toml`, for example `[profiles.dev]` with `action = "allow"`. You can set `approval_ttl` under `[defaults]` for grants created without `--ttl`. Quote profile names that contain dots.
- Inspect policy with `aws-guard policy explain <profile>` and grants with `aws-guard grants list`.
- Turn policy off for this terminal only with `aws-guard off` (with confirmation).

The `default` profile warning appears on stderr before each guarded command that resolves to `default`.

## Upgrade

When the install prefix already holds a valid installation, a second install upgrades in place:

```bash
sh install.sh --version X.Y.Z
# or, with the same prefix on PATH:
aws-guard install
```

You can also run `aws-guard upgrade` explicitly (`aws-guard upgrade [--prefix dir]`). Upgrade refreshes `aws`, `aws-guard`, the disable stub, and the manifest. It does not change the recorded real AWS CLI path. After a successful replace it runs `aws-guard migrate` logic on your policy and grants files in `~/.config/aws-cli-guard` and `~/.local/state/aws-cli-guard` (or your user-database paths).

If migrate fails after upgrade, stderr prints diagnosis plus recovery steps: use a `*.migrated.*` sibling if one was created, fix the file and run `aws-guard migrate`, use `aws-guard recover policy` or `aws-guard recover grants` on a terminal for corrupt files, or move the config, state, and cache directories aside and run `aws-guard init` to start over. Moving config also moves `settings.toml`; you can move that file back. A broken install prefix can be moved aside and reinstalled without deleting policy.

## Uninstall (manual)

1. If the installation is persistently disabled, run `aws-guard enable` first. `aws-guard shell --remove` refuses while the prefix is in the disabled layout.
2. Run `aws-guard shell bash --remove` and/or `aws-guard shell zsh --remove`. Pass `--prefix` when you used a custom install directory. If the marked block is missing, delete the lines from `# aws-cli-guard begin` through `# aws-cli-guard end` by hand in `.bashrc` or `.zshrc`.
3. Delete `~/.local/share/aws-cli-guard` (or your `--prefix` directory). For a shared prefix, remove only `aws`, `aws-guard`, `aws-cli-guard-disable`, `aws-cli-guard-manifest`, and any `aws.aside` or `aws-cli-guard-disabled` artifacts the disable swap created.
4. Optional: delete `~/.config/aws-cli-guard`, `~/.local/state/aws-cli-guard`, and `~/.cache/aws-cli-guard` if you do not want policy, grants, or cache data to remain.
5. Open a new terminal and run `command -v aws`. You should see the real AWS CLI; the wrapper never modified it.

## Documentation

- [Design](docs/DESIGN.md) — intent, goals, and principles
- [Architecture](docs/ARCHITECTURE.md) — how the software works today
- [Findings](docs/FINDINGS.md) — non-obvious lessons from implementation
- [Contributing](CONTRIBUTING.md) — build, test, maintain, and contribute
- [Coding style](docs/CODING_STYLE.md) — repository conventions
- [Security](SECURITY.md) — reporting vulnerabilities

Working agreements for AI-assisted changes are in [AGENTS.md](AGENTS.md).
