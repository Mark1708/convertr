# Security Policy

## Supported versions

Only the latest release receives security fixes. Older versions are not patched.

## Reporting a vulnerability

**Do not open a public GitHub/Gitea issue for security vulnerabilities.**

Send a private report to **mark1708.work@gmail.com** with the subject line `[convertr security]`.

Include:
- A description of the vulnerability and its potential impact
- Steps to reproduce
- Affected versions (if known)
- Any suggested fix (optional)

**Response timeline:**
- Acknowledgement within 72 hours
- Fix or mitigation within 90 days
- Credit in the release notes (unless you prefer to remain anonymous)

## Out of scope

- Vulnerabilities in upstream tools (pandoc, ffmpeg, LibreOffice, ImageMagick, etc.) — report those to their respective projects
- Issues requiring physical access to the machine
- Social engineering

## Hardening recommendations

- Keep the config file permissions at `0600`. `convertr config init` creates it with these permissions.
- Treat `extra_args` in `config.toml` as code — the values are passed directly to external binaries.
- Avoid placing sensitive data (API keys, passwords) in `extra_args`; use environment variables instead.
- Plugin executables (`convertr-*`) are run with your user's full permissions. Only install plugins you trust.

## Dependency scanning

The project uses `govulncheck` in CI. Run it locally:

```sh
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```
