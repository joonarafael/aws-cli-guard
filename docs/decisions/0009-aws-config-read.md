# ADR 0009: Reading AWS shared config and credentials files

## Status

Accepted

## Context

MVP grant fingerprints (`GR-2`) and later provider classification (`PR-5`) need to read the same AWS CLI config and shared credentials files the CLI uses. Those files are INI-shaped, not the strict TOML used for guard policy. [ADR 0008](0008-toml-decoder.md) keeps product dependencies minimal; botocore already defines the parse and merge rules in [`configloader.py`](https://github.com/boto/botocore/blob/develop/botocore/configloader.py) and `Session.full_config`.

The guard must not return secret access keys, session tokens, or `credential_process` command strings. Unreadable paths must not be treated as empty, or a grant could apply when the credential source cannot be verified.

## Decision

1. **Package.** `internal/awsconfig` reads both files from an environment slice. It does not call `os.Environ`. Paths follow `AWS_CONFIG_FILE` and `AWS_SHARED_CREDENTIALS_FILE`, then `$HOME/.aws/config` and `$HOME/.aws/credentials`. Missing files are empty. Existing non-regular or unreadable paths are errors.

2. **Symlinks.** Open with `os.Open`, which follows a final symlink to a regular file, matching `os.path.isfile` behavior for the CLI. This differs from `internal/files` `O_NOFOLLOW` reads used for guard-owned files.

3. **Parse rules.** Hand-rolled INI parsing aligned with botocore: profile sections in the config file (`[default]`, `[profile name]`), raw section names in the credentials file, `[DEFAULT]` inheritance, one-level nested blocks, strict duplicate keys and sections, last credentials keys merged over config per profile in `Session.full_config`.

4. **Secrets.** Strip `aws_access_key_id`, `aws_secret_access_key`, `aws_session_token`, `aws_security_token`, and `credential_process` values from returned data. Record presence flags only. Parse errors name paths, keys, and line numbers, never values or file bodies.

5. **Scope.** This ADR covers read and merge only. The wrapper entrypoint does not call `Load` until policy evaluation exists. Provider typing and fingerprint hashing stay in `PR-5` and `GR-2`.

## Consequences

- `GR-2` can depend on stable `Config` and `Section` types without re-parsing AWS files differently.
- Strict duplicate handling matches CLI refusal of broken files rather than last-wins semantics from permissive parsers.
- A separate `ErrHomeUnset` from `internal/paths` keeps AWS file errors distinct from XDG layout errors.
