# Security Policy

## Supported Versions

Only the latest released version of goforge receives security fixes.

| Version | Supported |
|---------|-----------|
| latest  | Yes       |
| older   | No        |

## Scope

goforge is a **code generator** — it runs locally and writes files to your filesystem. It does not start a server, open network ports, or handle user-supplied data at runtime.

Security issues that are in scope:

- **Malicious template injection** — a crafted config value that causes `text/template` to write unexpected content or escape the output directory
- **Path traversal** — template filenames or config values that write files outside the intended output directory
- **Dependency vulnerabilities** — known CVEs in goforge's direct dependencies (Cobra, huh, Lip Gloss)
- **Generated project security** — structural issues in generated templates that introduce vulnerabilities in the downstream project (e.g. missing CORS restrictions, insecure defaults in docker-compose)

Out of scope:

- Security issues in the user's own generated project code that are not caused by goforge templates
- Vulnerabilities in Go's standard library or toolchain itself

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, report them privately using one of the following:

1. **GitHub private vulnerability reporting** — go to the [Security tab](../../security/advisories/new) of this repository and click "Report a vulnerability".
2. **Email** — if private reporting is unavailable, email the maintainer directly. Check the repository owner's GitHub profile for a contact address.

### What to include

A useful report includes:

- A description of the vulnerability and its potential impact
- Steps to reproduce (minimal reproduction case preferred)
- The version of goforge affected
- Any suggested fix or mitigation (optional but appreciated)

### Response timeline

| Stage | Target |
|-------|--------|
| Acknowledgement | Within 3 business days |
| Initial assessment | Within 7 business days |
| Fix or remediation plan | Depends on severity |

We will coordinate a disclosure date with you once a fix is ready. We ask that you give us a reasonable window before any public disclosure.

## Security Best Practices for Generated Projects

goforge generates starter projects. Before shipping to production, review the generated code for:

- **Secrets management** — replace `.env.example` values with a proper secrets manager (Vault, AWS Secrets Manager, etc.); never commit real secrets
- **CORS configuration** — the generated CORS middleware uses permissive defaults; tighten `AllowOrigins` for your environment
- **Database credentials** — change default DSN values in `docker-compose.yml` and `application.yml`
- **Dependency updates** — run `go mod tidy` and audit dependencies with `govulncheck ./...` before deployment
- **TLS** — the generated HTTP server does not terminate TLS; put it behind a reverse proxy (nginx, Caddy, etc.) in production
