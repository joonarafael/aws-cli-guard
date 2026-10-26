# ADR 0004: Login is outside the wrapper

## Status

Accepted

## Context

Open question 4 in [DESIGN.md](../DESIGN.md) asked whether `aws sso login` should be allowed by default or given its own policy. Login establishes or refreshes authentication. Later commands are what use a profile. The design already separates those two, and it already says a successful login does not create a grant and that revoking a grant does not clear cached SSO credentials.

The CLI has more than one way to log in. `aws sso login` can take `--profile` or `--sso-session`. Newer releases add other login commands. Governing each of those forms would be a second product.

## Decision

1. **Login is not guarded.** The wrapper does not allow, prompt, or deny a login. It does not read policy for that command. It hands the original arguments to the real AWS CLI.

2. **Which commands.** A login command is one whose purpose is to establish or refresh an authentication session. `aws sso login` is in that set. So is any other AWS CLI login form, including forms that do not take a profile. Choosing how to log in, and which account or role to log in to, belongs to the user.

3. **No grant.** A login does not create a grant, and it does not change policy. A later command that uses a profile, environment credentials, or another subject still follows the normal policy.

4. **Recognition.** The wrapper recognizes the login forms of the AWS CLI versions it knows, per [ADR 0002](0002-aws-cli-version-floor.md). A command it does not recognize as login follows the normal guarded path.

## Consequences

- An agent, or any other caller, can refresh cached credentials without a grant. Anything that then calls AWS outside this wrapper can use those credentials. That limit is the same class of bypass already documented for absolute paths and SDKs.
- `SS-1` is the work of recognizing login and handing it off. It does not add a login policy.
- A CLI newer than the known-supported release may add a login form this build does not recognize. That invocation is guarded like any other command, and the version warning from ADR 0002 still applies.
