# Security Policy

## Supported versions

Sumeru is experimental. Security fixes are applied to the latest **`main`** branch of [ProjectMeru/sumeru](https://github.com/ProjectMeru/sumeru). Older commits or forks are not regularly backported unless maintainers announce otherwise.

Related repositories (`sumeru_addons`, `sumeru_custom_addons`) should follow the same reporting process when the issue spans those trees; report against the repo that owns the vulnerable code when possible.

## Reporting a vulnerability

**Please do not open a public GitHub issue** for undisclosed security problems.

Prefer **GitHub Security Advisories** (private vulnerability reporting) on this repository:

https://github.com/ProjectMeru/sumeru/security/advisories/new

If private reporting is unavailable, contact the **ProjectMeru** organization maintainers via GitHub (https://github.com/ProjectMeru) and mark the communication as security-sensitive.

Include as much of the following as you can:

- Description of the issue and impact
- Steps to reproduce or a minimal proof of concept
- Affected version / commit / branch
- Any suggested fix (optional)

We will acknowledge valid reports as promptly as practical, assess impact, and coordinate disclosure after a fix is available when appropriate.

## Scope

**In scope (examples):**

- Authentication and session handling
- Access control / RBAC bypasses
- Injection or unsafe query construction in the ORM or JSON-RPC layer
- Cross-site scripting or CSRF in the web shell when caused by framework or core templates
- Known vulnerable dependencies that are reachable in default or documented configurations

**Out of scope (examples):**

- Misconfiguration (e.g. weak `db_password` or `db_sslmode = disable` in local/example configs)
- Issues that require already-compromised database credentials or host access
- Social engineering, physical attacks, or denial-of-service without a clear application bug
- Vulnerabilities only in third-party forks or unpublished custom addons outside ProjectMeru repositories

## Safe harbor

We appreciate good-faith research. Avoid privacy violations, data destruction, and disruption of production systems you do not own. Test against local or authorized environments only.
